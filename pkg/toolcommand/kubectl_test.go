package toolcommand

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestResolutionIgnoresPATHAndFailsClosedOnInvalidOverride(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "kubectl")
	if err := os.WriteFile(binary, []byte("fixture"), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", directory)
	if _, err := resolveKubectl("", nil); err == nil {
		t.Fatal("PATH influenced executable resolution")
	}
	if _, err := resolveKubectl("./kubectl", []string{directory}); err == nil {
		t.Fatal("relative override fell back to another executable")
	}
	resolved, err := resolveKubectl("", []string{directory})
	expected, resolveErr := filepath.EvalSymlinks(binary)
	if err != nil || resolveErr != nil || resolved != expected {
		t.Fatalf("fixed directory resolution: %q, %v", resolved, err)
	}
}

func TestResolutionRejectsUnsafeExecutableModes(t *testing.T) {
	directory := t.TempDir()
	binary := filepath.Join(directory, "kubectl")
	for _, mode := range []os.FileMode{0600, 0720, 0702} {
		if err := os.WriteFile(binary, []byte("fixture"), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(binary, mode); err != nil {
			t.Fatal(err)
		}
		if _, err := resolveKubectl(binary, nil); err == nil {
			t.Fatalf("accepted unsafe mode %o", mode)
		}
	}
	if _, err := resolveKubectl(directory, nil); err == nil {
		t.Fatal("accepted directory as executable")
	}
	if _, err := resolveKubectl(filepath.Join(directory, "missing"), nil); err == nil {
		t.Fatal("accepted missing executable")
	}
}

func TestCommandUsesExplicitBinaryAndPreservesArguments(t *testing.T) {
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := filepath.EvalSymlinks(binary)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("MANTL_KUBECTL_PATH", binary)
	command := Kubectl(context.Background(), "get", "--", "pods")
	if command.Err != nil || command.Path != resolved || command.Args[2] != "--" {
		t.Fatalf("unexpected command: %#v", command)
	}
	t.Setenv("MANTL_KUBECTL_PATH", "relative")
	if err := Kubectl(context.Background(), "get").Run(); err == nil {
		t.Fatal("invalid binary configuration started a process")
	}
}

func TestResolutionSkipsFilesNotExecutableByOwner(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root bypasses owner execute permission")
	}
	first, second := t.TempDir(), t.TempDir()
	for _, dir := range []string{first, second} {
		if err := os.WriteFile(filepath.Join(dir, "kubectl"), []byte("fixture"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Chmod(filepath.Join(first, "kubectl"), 0641); err != nil {
		t.Fatal(err)
	}
	path, err := resolveKubectl("", []string{first, second})
	expected, _ := filepath.EvalSymlinks(filepath.Join(second, "kubectl"))
	if err != nil || path != expected {
		t.Fatalf("unexecutable candidate selected: %q %v", path, err)
	}
	if _, err := resolveKubectl(filepath.Join(first, "kubectl"), []string{second}); err == nil {
		t.Fatal("invalid override fell back")
	}
}
