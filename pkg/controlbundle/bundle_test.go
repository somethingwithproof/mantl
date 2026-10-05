package controlbundle

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
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
