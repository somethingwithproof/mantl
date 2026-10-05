// Package controlbundle packages and verifies independently versioned compliance content.
package controlbundle

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/thomasvincent/mantl/pkg/evidence"
)

const MaxBytes = 128 << 20
const (
	manifestName = "bundle.json"
	maxFiles     = 10000
	maxFileBytes = 32 << 20
)

var errBundleTooLarge = errors.New("bundle too large")

var semver = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?$`)

type File struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}
type Manifest struct {
	Schema  int    `json:"schemaVersion"`
	Version string `json:"version"`
	Files   []File `json:"files"`
}

func (m Manifest) Digest() string { b, _ := json.Marshal(m); return "sha256:" + evidence.Hash(b) }
func safe(name string) bool {
	return fs.ValidPath(name) && name != "." && name != manifestName &&
		!strings.ContainsAny(name, "\\\x00\r\n") && filepath.IsLocal(filepath.FromSlash(name))
}

type bundleContent struct {
	files    []File
	payloads map[string][]byte
	total    int
}

func (content *bundleContent) add(name string, data []byte) error {
	content.total += len(data)
	if content.total > MaxBytes || len(content.files) >= maxFiles {
		return errBundleTooLarge
	}
	content.payloads[name] = data
	content.files = append(content.files, File{name, evidence.Hash(data)})
	return nil
}

func (content *bundleContent) collectFile(root *os.Root, directory, prefix, filename string, entry fs.DirEntry) error {
	if entry.Type()&os.ModeSymlink != 0 {
		return fmt.Errorf("bundle cannot contain symlinks")
	}
	if entry.IsDir() {
		if entry.Name() == "charts" {
			return filepath.SkipDir
		}
		return nil
	}
	if !entry.Type().IsRegular() {
		return fmt.Errorf("bundle requires regular files")
	}
	if filepath.Ext(filename) != ".yaml" && filepath.Ext(filename) != ".yml" {
		return nil
	}
	name, err := filepath.Rel(directory, filename)
	if err != nil {
		return err
	}
	data, err := readBundleFile(root, filename, maxFileBytes)
	if err != nil {
		return err
	}
	return content.add(prefix+"/"+filepath.ToSlash(name), data)
}

func collectContent(source string) (bundleContent, error) {
	content := bundleContent{payloads: map[string][]byte{}}
	root, err := os.OpenRoot(source)
	if err != nil {
		return content, fmt.Errorf("open bundle source: %w", err)
	}
	defer func() { _ = root.Close() }()
	for _, directory := range []struct{ source, prefix string }{
		{"compliance/frameworks", "frameworks"},
		{"policies/kyverno", "policies"},
	} {
		err := fs.WalkDir(root.FS(), directory.source, func(filename string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			return content.collectFile(root, directory.source, directory.prefix, filename, entry)
		})
		if err != nil {
			return content, fmt.Errorf("collect bundle content: %w", err)
		}
	}
	return content, nil
}

// Build includes Mantl-owned framework/policy data; vendored chart trees are excluded.
func Build(source, version string, out io.Writer) (Manifest, error) {
	m := Manifest{Schema: 1, Version: version, Files: []File{}}
	if !semver.MatchString(version) {
		return m, fmt.Errorf("bundle version must be SemVer")
	}
	content, err := collectContent(source)
	if err != nil {
		return m, err
	}
	root := []byte("apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\nresources:\n  - policies/runtime\n  - frameworks/soc2/policies\n")
	if err = content.add("kustomization.yaml", root); err != nil {
		return m, err
	}
	m.Files = content.files
	sort.Slice(m.Files, func(i, j int) bool { return m.Files[i].Path < m.Files[j].Path })
	return m, writeArchive(m, content.payloads, out)
}

func writeArchive(m Manifest, payloads map[string][]byte, out io.Writer) error {
	gz := gzip.NewWriter(out)
	tw := tar.NewWriter(gz)
	write := func(name string, data []byte) error {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: int64(len(data)), Typeflag: tar.TypeReg}); err != nil {
			return err
		}
		_, err := tw.Write(data)
		return err
	}
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if err = write(manifestName, data); err != nil {
		return err
	}
	for _, file := range m.Files {
		if err = write(file.Path, payloads[file.Path]); err != nil {
			return err
		}
	}
	if err = tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

// readBundleFile confines each read to an open directory and bounds allocation.
func readBundleFile(root *os.Root, name string, limit int64) ([]byte, error) {
	info, err := root.Lstat(name)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, fmt.Errorf("invalid bundle file %q", name)
	}
	file, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	opened, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if !opened.Mode().IsRegular() || !os.SameFile(info, opened) {
		return nil, fmt.Errorf("bundle file changed while opening %q", name)
	}
	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > limit {
		return nil, fmt.Errorf("bundle file too large %q", name)
	}
	return data, nil
}

// Verify rejects missing, altered, duplicate and unexpected content.
func Verify(dir, expected string) (Manifest, error) {
	var m Manifest
	root, err := os.OpenRoot(dir)
	if err != nil {
		return m, err
	}
	defer func() { _ = root.Close() }()
	data, err := readBundleFile(root, manifestName, 4<<20)
	if err != nil {
		return m, err
	}
	if err = json.Unmarshal(data, &m); err != nil {
		return m, err
	}
	if m.Schema != 1 || !semver.MatchString(m.Version) || len(m.Files) == 0 || len(m.Files) > maxFiles {
		return m, fmt.Errorf("invalid bundle manifest")
	}
	if expected != "" && m.Digest() != expected {
		return m, fmt.Errorf("bundle digest mismatch")
	}
	seen, err := verifyFiles(root, m.Files)
	if err != nil {
		return m, err
	}
	return m, verifyDirectory(root, seen)
}

func verifyFiles(root *os.Root, files []File) (map[string]bool, error) {
	seen := map[string]bool{}
	total := 0
	for _, file := range files {
		if !safe(file.Path) || seen[file.Path] {
			return nil, fmt.Errorf("invalid or duplicate bundle path")
		}
		seen[file.Path] = true
		data, err := readBundleFile(root, filepath.FromSlash(file.Path), maxFileBytes)
		if err != nil {
			return nil, err
		}
		total += len(data)
		if total > MaxBytes {
			return nil, errBundleTooLarge
		}
		if evidence.Hash(data) != file.SHA256 {
			return nil, fmt.Errorf("bundle content mismatch: %q", file.Path)
		}
	}
	return seen, nil
}

func verifyDirectory(root *os.Root, seen map[string]bool) error {
	return fs.WalkDir(root.FS(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("bundle cannot contain symlinks")
		}
		if !entry.IsDir() && name != manifestName && !seen[name] {
			return fmt.Errorf("unexpected bundle content: %q", name)
		}
		return nil
	})
}

type archiveBudget struct {
	count int
	total int64
	seen  map[string]bool
}

func (budget *archiveBudget) accept(header *tar.Header) error {
	if budget.count >= maxFiles+1 || header.Typeflag != tar.TypeReg || header.Size < 0 || header.Size > maxFileBytes {
		return fmt.Errorf("unsafe bundle archive")
	}
	if (!safe(header.Name) && header.Name != manifestName) || budget.seen[header.Name] {
		return fmt.Errorf("invalid or duplicate archive path")
	}
	budget.seen[header.Name] = true
	budget.count++
	budget.total += header.Size
	if budget.total > MaxBytes {
		return errBundleTooLarge
	}
	return nil
}

func extractFile(root *os.Root, header *tar.Header, content io.Reader) error {
	name := filepath.FromSlash(header.Name)
	if err := root.MkdirAll(filepath.Dir(name), 0700); err != nil {
		return err
	}
	file, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		return err
	}
	_, copyErr := io.CopyN(file, content, header.Size)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func extractArchive(root *os.Root, content io.Reader) error {
	compressed, err := gzip.NewReader(content)
	if err != nil {
		return err
	}
	defer func() { _ = compressed.Close() }()
	archive := tar.NewReader(compressed)
	budget := archiveBudget{seen: map[string]bool{}}
	for {
		header, err := archive.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err = budget.accept(header); err != nil {
			return err
		}
		if err = extractFile(root, header, archive); err != nil {
			return err
		}
	}
}

// Unpack verifies in a private temporary directory before publishing a new directory.
func Unpack(in io.Reader, dest, expected string) (Manifest, error) {
	var m Manifest
	parent := filepath.Dir(dest)
	if err := os.MkdirAll(parent, 0750); err != nil {
		return m, err
	}
	dir, err := os.MkdirTemp(parent, ".mantl-bundle-")
	if err != nil {
		return m, err
	}
	defer func() { _ = os.RemoveAll(dir) }()
	root, err := os.OpenRoot(dir)
	if err != nil {
		return m, err
	}
	defer func() { _ = root.Close() }()
	if err = extractArchive(root, in); err != nil {
		return m, err
	}
	m, err = Verify(dir, expected)
	if err != nil {
		return m, err
	}
	if _, err = os.Lstat(dest); !os.IsNotExist(err) {
		return m, fmt.Errorf("bundle destination must not exist")
	}
	return m, os.Rename(dir, dest)
}
