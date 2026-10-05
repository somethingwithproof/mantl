// Package controlbundle packages and verifies independently versioned compliance content.
package controlbundle

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/thomasvincent/mantl/pkg/evidence"
)

const MaxBytes = 128 << 20

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
	return name != "." && name != "bundle.json" && !strings.Contains(name, "\\") && !strings.HasPrefix(name, "/") && path.Clean(name) == name && !strings.HasPrefix(name, "../")
}

// Build includes Mantl-owned framework/policy data; vendored chart trees are excluded.
func Build(source, version string, out io.Writer) (Manifest, error) {
	m := Manifest{Schema: 1, Version: version, Files: []File{}}
	if !semver.MatchString(version) {
		return m, fmt.Errorf("bundle version must be SemVer")
	}
	payloads := map[string][]byte{}
	total := 0
	for _, prefix := range []string{"compliance/frameworks", "policies/kyverno"} {
		err := filepath.WalkDir(filepath.Join(source, prefix), func(p string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("bundle cannot contain symlinks")
			}
			if d.IsDir() {
				if d.Name() == "charts" {
					return filepath.SkipDir
				}
				return nil
			}
			if !d.Type().IsRegular() {
				return fmt.Errorf("bundle requires regular files")
			}
			if filepath.Ext(p) != ".yaml" && filepath.Ext(p) != ".yml" {
				return nil
			}
			name, err := filepath.Rel(filepath.Join(source, prefix), p)
			if err != nil {
				return err
			}
			name = filepath.ToSlash(name)
			if prefix == "compliance/frameworks" {
				name = "frameworks/" + name
			} else {
				name = "policies/" + name
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			if info.Size() > 32<<20 {
				return fmt.Errorf("bundle file too large")
			}
			b, err := os.ReadFile(p)
			if err != nil {
				return err
			}
			total += len(b)
			if total > MaxBytes {
				return fmt.Errorf("bundle too large")
			}
			payloads[name] = b
			m.Files = append(m.Files, File{name, evidence.Hash(b)})
			return nil
		})
		if err != nil {
			return m, fmt.Errorf("collect bundle content: %w", err)
		}
	}

	root := []byte("apiVersion: kustomize.config.k8s.io/v1beta1\nkind: Kustomization\nresources:\n  - policies/runtime\n  - frameworks/soc2/policies\n")
	payloads["kustomization.yaml"] = root
	m.Files = append(m.Files, File{"kustomization.yaml", evidence.Hash(root)})
	sort.Slice(m.Files, func(i, j int) bool { return m.Files[i].Path < m.Files[j].Path })
	gz := gzip.NewWriter(out)
	tw := tar.NewWriter(gz)
	write := func(name string, b []byte) error {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0600, Size: int64(len(b)), Typeflag: tar.TypeReg}); err != nil {
			return err
		}
		_, err := tw.Write(b)
		return err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return m, err
	}
	if err = write("bundle.json", b); err != nil {
		return m, err
	}
	for _, f := range m.Files {
		if err = write(f.Path, payloads[f.Path]); err != nil {
			return m, err
		}
	}
	if err = tw.Close(); err != nil {
		return m, err
	}
	return m, gz.Close()
}

// Verify rejects missing, altered, duplicate and unexpected content.
func Verify(dir, expected string) (Manifest, error) {
	var m Manifest
	info, err := os.Lstat(filepath.Join(dir, "bundle.json"))
	if err != nil {
		return m, err
	}
	if !info.Mode().IsRegular() || info.Size() > 4<<20 {
		return m, fmt.Errorf("invalid bundle manifest file")
	}
	b, err := os.ReadFile(filepath.Join(dir, "bundle.json"))
	if err != nil {
		return m, err
	}
	if len(b) > 4<<20 {
		return m, fmt.Errorf("bundle manifest too large")
	}
	if err = json.Unmarshal(b, &m); err != nil {
		return m, err
	}
	if m.Schema != 1 || !semver.MatchString(m.Version) || len(m.Files) == 0 || len(m.Files) > 10000 {
		return m, fmt.Errorf("invalid bundle manifest")
	}
	if expected != "" && m.Digest() != expected {
		return m, fmt.Errorf("bundle digest mismatch")
	}
	err = filepath.WalkDir(dir, func(_ string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("bundle cannot contain symlinks")
		}
		return nil
	})
	if err != nil {
		return m, err
	}
	seen := map[string]bool{}
	total := 0
	for _, f := range m.Files {
		if !safe(f.Path) || seen[f.Path] {
			return m, fmt.Errorf("invalid or duplicate bundle path")
		}
		seen[f.Path] = true
		p := filepath.Join(dir, filepath.FromSlash(f.Path))
		info, err := os.Lstat(p)
		if err != nil {
			return m, err
		}
		if !info.Mode().IsRegular() || info.Size() > 32<<20 {
			return m, fmt.Errorf("invalid bundle file")
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return m, err
		}
		total += len(b)
		if total > MaxBytes {
			return m, fmt.Errorf("bundle too large")
		}
		if evidence.Hash(b) != f.SHA256 {
			return m, fmt.Errorf("bundle content mismatch: %s", f.Path)
		}
	}
	err = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("bundle cannot contain symlinks")
		}
		if d.IsDir() {
			return nil
		}
		name, err := filepath.Rel(dir, p)
		if err != nil {
			return err
		}
		name = filepath.ToSlash(name)
		if name != "bundle.json" && !seen[name] {
			return fmt.Errorf("unexpected bundle content: %s", name)
		}
		return nil
	})
	return m, err
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
	gz, err := gzip.NewReader(in)
	if err != nil {
		return m, err
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	seen := map[string]bool{}
	total := int64(0)
	for count := 0; ; count++ {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return m, err
		}
		if count >= 10001 || h.Typeflag != tar.TypeReg || h.Size < 0 || h.Size > 32<<20 || (!safe(h.Name) && h.Name != "bundle.json") || seen[h.Name] {
			return m, fmt.Errorf("unsafe bundle archive")
		}
		seen[h.Name] = true
		total += h.Size
		if total > MaxBytes {
			return m, fmt.Errorf("bundle too large")
		}
		p := filepath.Join(dir, filepath.FromSlash(h.Name))
		if err = os.MkdirAll(filepath.Dir(p), 0700); err != nil {
			return m, err
		}
		f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
		if err != nil {
			return m, err
		}
		_, copyErr := io.CopyN(f, tr, h.Size)
		closeErr := f.Close()
		if copyErr != nil {
			return m, copyErr
		}
		if closeErr != nil {
			return m, closeErr
		}
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
