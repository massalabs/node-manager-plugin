package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveOldNodeVersionsArtifacts(t *testing.T) {
	// Create temp root dir
	root := t.TempDir()

	// Prepare plugin info with current versions
	pi := &PluginInfo{
		MainnetVersion:  "MAIN.4.0",
		BuildnetVersion: "DEVN.29.0",
	}

	// Create directories representing versions
	mustMkdir(t, filepath.Join(root, "MAIN.4.0"))     // should stay
	mustMkdir(t, filepath.Join(root, "DEVN.29.0"))    // should stay
	mustMkdir(t, filepath.Join(root, "MAIN.3.0"))     // should be removed
	mustMkdir(t, filepath.Join(root, "DEVN.28.16"))   // should be removed
	mustMkdir(t, filepath.Join(root, "random-other")) // should be removed

	// Also add a file at root to ensure files are ignored
	mustWriteFile(t, filepath.Join(root, "README.txt"), []byte("noop"))

	// Run cleanup
	if err := pi.RemoveOldNodeVersionsArtifacts(root); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert expected directories
	assertExists(t, filepath.Join(root, "MAIN.4.0"), true)
	assertExists(t, filepath.Join(root, "DEVN.29.0"), true)
	assertExists(t, filepath.Join(root, "MAIN.3.0"), false)
	assertExists(t, filepath.Join(root, "DEVN.28.16"), false)
	assertExists(t, filepath.Join(root, "random-other"), false)
	// Files should be kept untouched
	assertExists(t, filepath.Join(root, "README.txt"), true)
}

func TestRemoveOldNodeVersionsArtifacts_ReadDirError(t *testing.T) {
	// Use a non-existent directory to trigger error
	pi := &PluginInfo{MainnetVersion: "M", BuildnetVersion: "B"}
	err := pi.RemoveOldNodeVersionsArtifacts(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatalf("expected error when reading non-existent dir")
	}
}

// helpers
func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir failed for %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write file failed for %s: %v", path, err)
	}
}

func assertExists(t *testing.T, path string, shouldExist bool) {
	t.Helper()
	_, err := os.Stat(path)
	if shouldExist && err != nil {
		t.Fatalf("expected %s to exist, but it does not: %v", path, err)
	}
	if !shouldExist && err == nil {
		t.Fatalf("expected %s to be removed, but it still exists", path)
	}
}
