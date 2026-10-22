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
		MainnetVersion:  "MAIN.5.1",
		BuildnetVersion: "DEVN.30.2",
	}

	// Create directories representing versions
	mustMkdir(t, filepath.Join(root, "MAIN.5.1"))     // should stay
	mustMkdir(t, filepath.Join(root, "DEVN.30.2"))    // should stay
	mustMkdir(t, filepath.Join(root, "MAIN.3.0"))     // should be removed
	mustMkdir(t, filepath.Join(root, "DEVN.28.16"))   // should be removed
	mustMkdir(t, filepath.Join(root, "random-other")) // should be removed

	// Also add a file at root to ensure files are ignored
	mustWriteFile(t, filepath.Join(root, "MAIN.3.0", "README.txt"), []byte("noop"))
	mustMkdir(t, filepath.Join(root, "DEVN.28.16", "random-sub-folder"))
	mustWriteFile(t, filepath.Join(root, "DEVN.28.16", "README.txt"), []byte("noop"))

	// Run cleanup
	if err := pi.RemoveOldNodeVersionsArtifacts(root); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert expected directories
	assertExists(t, filepath.Join(root, "MAIN.5.1"), true)
	assertExists(t, filepath.Join(root, "DEVN.30.2"), true)
	assertExists(t, filepath.Join(root, "MAIN.3.0"), false)
	assertExists(t, filepath.Join(root, "DEVN.28.16"), false)
	assertExists(t, filepath.Join(root, "random-other"), false)
}

func TestRemoveOldNodeVersionsArtifacts_ReadDirError(t *testing.T) {
	// Use a non-existent directory to trigger error
	pi := &PluginInfo{MainnetVersion: "M", BuildnetVersion: "B"}
	err := pi.RemoveOldNodeVersionsArtifacts(filepath.Join(t.TempDir(), "does-not-exist"))
	if err == nil {
		t.Fatalf("expected error when reading non-existent dir")
	}
}

func TestMigrateNodeVersionsData(t *testing.T) {
	root := t.TempDir()

	pi := &PluginInfo{
		MainnetVersion:  "MAIN.5.1",
		BuildnetVersion: "DEVN.30.2",
	}

	stakingWallets := filepath.Join("massa-node", "config", "staking_wallets")
	clientWallets := filepath.Join("massa-client", "wallets")

	// previous mainnet versions: the latest one holds the keys to migrate
	mustMkdir(t, filepath.Join(root, "MAIN.4.10", stakingWallets))
	mustWriteFile(t, filepath.Join(root, "MAIN.4.10", stakingWallets, "wallet_AU1old.yaml"), []byte("old"))
	mustMkdir(t, filepath.Join(root, "MAIN.5.0", stakingWallets))
	mustWriteFile(t, filepath.Join(root, "MAIN.5.0", stakingWallets, "wallet_AU1staker.yaml"), []byte("staker"))
	mustMkdir(t, filepath.Join(root, "MAIN.5.0", clientWallets))
	mustWriteFile(t, filepath.Join(root, "MAIN.5.0", clientWallets, "wallet_AU1client.yaml"), []byte("client"))

	// the new version is shipped with empty config folders
	mustMkdir(t, filepath.Join(root, "MAIN.5.1", "massa-node", "config"))
	mustMkdir(t, filepath.Join(root, "MAIN.5.1", clientWallets))

	// previous buildnet version: the new version already has a client wallet, which must not be replaced
	mustMkdir(t, filepath.Join(root, "DEVN.30.0", stakingWallets))
	mustWriteFile(t, filepath.Join(root, "DEVN.30.0", stakingWallets, "wallet_AU1build.yaml"), []byte("build"))
	mustMkdir(t, filepath.Join(root, "DEVN.30.0", clientWallets))
	mustWriteFile(t, filepath.Join(root, "DEVN.30.0", clientWallets, "wallet_AU1oldclient.yaml"), []byte("old"))
	mustMkdir(t, filepath.Join(root, "DEVN.30.2", clientWallets))
	mustWriteFile(t, filepath.Join(root, "DEVN.30.2", clientWallets, "wallet_AU1newclient.yaml"), []byte("new"))

	if err := pi.MigrateNodeVersionsData(root); err != nil {
		t.Fatalf("MigrateNodeVersionsData returned error: %v", err)
	}

	assertContent(t, filepath.Join(root, "MAIN.5.1", stakingWallets, "wallet_AU1staker.yaml"), "staker")
	assertExists(t, filepath.Join(root, "MAIN.5.1", stakingWallets, "wallet_AU1old.yaml"), false)
	assertContent(t, filepath.Join(root, "MAIN.5.1", clientWallets, "wallet_AU1client.yaml"), "client")

	assertContent(t, filepath.Join(root, "DEVN.30.2", stakingWallets, "wallet_AU1build.yaml"), "build")
	assertContent(t, filepath.Join(root, "DEVN.30.2", clientWallets, "wallet_AU1newclient.yaml"), "new")
	assertExists(t, filepath.Join(root, "DEVN.30.2", clientWallets, "wallet_AU1oldclient.yaml"), false)

	// a second run (next plugin start) changes nothing
	if err := pi.MigrateNodeVersionsData(root); err != nil {
		t.Fatalf("second MigrateNodeVersionsData returned error: %v", err)
	}
	assertContent(t, filepath.Join(root, "MAIN.5.1", stakingWallets, "wallet_AU1staker.yaml"), "staker")

	// then the previous versions are removed, the keys stay in the current ones
	if err := pi.RemoveOldNodeVersionsArtifacts(root); err != nil {
		t.Fatalf("RemoveOldNodeVersionsArtifacts returned error: %v", err)
	}
	assertExists(t, filepath.Join(root, "MAIN.5.0"), false)
	assertContent(t, filepath.Join(root, "MAIN.5.1", stakingWallets, "wallet_AU1staker.yaml"), "staker")
}

func TestMigrateNodeVersionsData_NoPreviousVersion(t *testing.T) {
	root := t.TempDir()
	pi := &PluginInfo{MainnetVersion: "MAIN.5.1", BuildnetVersion: "DEVN.30.2"}

	mustMkdir(t, filepath.Join(root, "MAIN.5.1", "massa-client"))

	if err := pi.MigrateNodeVersionsData(root); err != nil {
		t.Fatalf("MigrateNodeVersionsData returned error: %v", err)
	}
	assertExists(t, filepath.Join(root, "MAIN.5.1", "massa-client", "wallets"), false)
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		sign int
	}{
		{"MAIN.5.0", "MAIN.4.2", 1},
		{"MAIN.4.10", "MAIN.4.2", 1},
		{"DEVN.29.2", "DEVN.30.0", -1},
		{"MAIN.5.0", "MAIN.5.0", 0},
	}

	for _, c := range cases {
		got := compareVersions(c.a, c.b)
		if (got > 0) != (c.sign > 0) || (got < 0) != (c.sign < 0) {
			t.Errorf("compareVersions(%s, %s) = %d, expected sign %d", c.a, c.b, got, c.sign)
		}
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

func assertContent(t *testing.T, path, expected string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file failed for %s: %v", path, err)
	}
	if string(data) != expected {
		t.Fatalf("expected %s to contain %q, got %q", path, expected, string(data))
	}
}
