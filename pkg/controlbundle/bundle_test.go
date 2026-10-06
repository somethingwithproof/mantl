// SPDX-License-Identifier: Apache-2.0

package controlbundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestRoundTripAndTampering(t *testing.T) {
	source := t.TempDir()
	for _, p := range []string{"compliance/frameworks/soc2/framework.yaml", "policies/kyverno/runtime/policy.yaml", "policies/kyverno/charts/vendor/policy.yaml"} {
		file := filepath.Join(source, p)
		if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte("content"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var one, two bytes.Buffer
	m, err := Build(source, "1.0.0", &one)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = Build(source, "1.0.0", &two); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(one.Bytes(), two.Bytes()) || len(m.Files) != 3 {
		t.Fatal("bundle not deterministic or includes vendor content")
	}
	dest := filepath.Join(t.TempDir(), "bundle")
	if _, err = Unpack(&one, dest, m.Digest()); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(dest, "sha256:incorrect"); err == nil {
		t.Fatal("wrong digest accepted")
	}
	if err = os.WriteFile(filepath.Join(dest, m.Files[0].Path), []byte("changed"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = Verify(dest, m.Digest()); err == nil {
		t.Fatal("changed content accepted")
	}
}
func TestRejectUnsafeArchives(t *testing.T) {
	for _, h := range []*tar.Header{{Name: "../escape", Mode: 0600, Typeflag: tar.TypeReg}, {Name: "frameworks/link", Linkname: "/etc/passwd", Typeflag: tar.TypeSymlink}} {
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		tw := tar.NewWriter(gz)
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
		if err := tw.Close(); err != nil {
			t.Fatal(err)
		}
		if err := gz.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := Unpack(&b, filepath.Join(t.TempDir(), "bundle"), ""); err == nil {
			t.Fatal("unsafe archive accepted")
		}
	}
}

func TestBundleFileReadsStayBoundedAndRegular(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "content"), []byte("bounded"), 0600); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	data, err := readBundleFile(root, "content", 7)
	if err != nil || string(data) != "bounded" {
		t.Fatalf("regular content rejected: %q, %v", data, err)
	}
	for _, name := range []string{"content", ".", "missing"} {
		if _, err := readBundleFile(root, name, 6); err == nil {
			t.Fatalf("invalid or oversized content accepted: %q", name)
		}
	}
	if err := os.Symlink("content", filepath.Join(dir, "link")); err != nil {
		t.Fatal(err)
	}
	if _, err := readBundleFile(root, "link", 7); err == nil {
		t.Fatal("symlink accepted")
	}
}

func TestArchiveBudgetRejectsDuplicatesAndLimits(t *testing.T) {
	budget := archiveBudget{seen: map[string]bool{}}
	header := &tar.Header{Name: "frameworks/control.yaml", Typeflag: tar.TypeReg, Size: 1}
	if err := budget.accept(header); err != nil {
		t.Fatal(err)
	}
	if err := budget.accept(header); err == nil {
		t.Fatal("duplicate accepted")
	}
	for _, limits := range []archiveBudget{
		{count: 10001, seen: map[string]bool{}},
		{total: MaxBytes, seen: map[string]bool{}},
	} {
		if err := limits.accept(header); err == nil {
			t.Fatal("archive budget exceeded")
		}
	}
}

func TestSafeBundlePaths(t *testing.T) {
	for _, name := range []string{"", ".", "..", "bundle.json", "/absolute", "nested/../file", "nested\\file", "nested/file\n"} {
		if safe(name) {
			t.Fatalf("noncanonical path accepted: %q", name)
		}
	}
	if !safe("frameworks/soc2/control.yaml") {
		t.Fatal("canonical content path rejected")
	}
}

func bundleFixture(t *testing.T) (string, Manifest) {
	t.Helper()
	source := t.TempDir()
	for _, name := range []string{"compliance/frameworks/control.yaml", "policies/kyverno/policy.yaml"} {
		filename := filepath.Join(source, name)
		if err := os.MkdirAll(filepath.Dir(filename), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filename, []byte("fixture"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	var archive bytes.Buffer
	manifest, err := Build(source, "1.0.0", &archive)
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "verified")
	if _, err := Unpack(&archive, dir, manifest.Digest()); err != nil {
		t.Fatal(err)
	}
	return dir, manifest
}

func TestVerifyRejectsInvalidManifests(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*Manifest)
	}{
		{"schema", func(m *Manifest) { m.Schema = 0 }},
		{"version", func(m *Manifest) { m.Version = "invalid" }},
		{"empty", func(m *Manifest) { m.Files = nil }},
		{"file count", func(m *Manifest) { m.Files = make([]File, maxFiles+1) }},
		{"duplicate", func(m *Manifest) { m.Files = append(m.Files, m.Files[0]) }},
		{"noncanonical", func(m *Manifest) { m.Files[0].Path = "." }},
		{"missing", func(m *Manifest) { m.Files[0].Path = "missing.yaml" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir, manifest := bundleFixture(t)
			tc.mutate(&manifest)
			data, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, manifestName), data, 0600); err != nil {
				t.Fatal(err)
			}
			if _, err := Verify(dir, ""); err == nil {
				t.Fatal("invalid manifest accepted")
			}
		})
	}
}

func TestVerifyRejectsUnexpectedAndMalformedContent(t *testing.T) {
	dir, _ := bundleFixture(t)
	if err := os.WriteFile(filepath.Join(dir, "unexpected.yaml"), []byte("extra"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(dir, ""); err == nil {
		t.Fatal("unexpected content accepted")
	}
	if err := os.Remove(filepath.Join(dir, "unexpected.yaml")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("kustomization.yaml", filepath.Join(dir, "untracked-link")); err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(dir, ""); err == nil {
		t.Fatal("untracked symlink accepted")
	}
	if err := os.WriteFile(filepath.Join(dir, manifestName), []byte("not JSON"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Verify(dir, ""); err == nil {
		t.Fatal("malformed manifest accepted")
	}
	if _, err := Verify(filepath.Join(dir, "missing"), ""); err == nil {
		t.Fatal("missing root accepted")
	}
}

func TestBuildAndExtractionPropagateIOFailures(t *testing.T) {
	if _, err := Build(t.TempDir(), "invalid", io.Discard); err == nil {
		t.Fatal("invalid version accepted")
	}
	if _, err := Build(t.TempDir(), "1.0.0", io.Discard); err == nil {
		t.Fatal("missing source accepted")
	}
	root, err := os.OpenRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	if err := extractArchive(root, bytes.NewBufferString("not gzip")); err == nil {
		t.Fatal("malformed compression accepted")
	}
	if err := extractFile(root, &tar.Header{Name: "content", Size: 3}, bytes.NewBufferString("x")); err == nil {
		t.Fatal("short content accepted")
	}
	if err := extractFile(root, &tar.Header{Name: "content", Size: 0}, bytes.NewReader(nil)); err == nil {
		t.Fatal("existing content overwritten")
	}
}
