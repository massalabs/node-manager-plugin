package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/massalabs/node-manager-plugin/int/utils"
)

// PluginInfo contains runtime plugin information
type PluginInfo struct {
	PwdMainnet      string `json:"pwd_mainnet"`
	PwdBuildnet     string `json:"pwd_buildnet"`
	AutoRestart     bool   `json:"auto_restart"`
	IsMainnet       bool   `json:"is_mainnet"`
	MainnetVersion  string `json:"mainnet_version"`
	BuildnetVersion string `json:"buildnet_version"`
	mu              sync.RWMutex
}

// Global instance of PluginInfo
var GlobalPluginInfo *PluginInfo

// init function to initialize the global PluginInfo instance
func init() {
	GlobalPluginInfo = &PluginInfo{
		PwdMainnet:      "",
		PwdBuildnet:     "",
		AutoRestart:     false,
		IsMainnet:       true,
		MainnetVersion:  "MAIN.5.0",
		BuildnetVersion: "DEVN.30.0",
	}
}

func (pi *PluginInfo) GetNetworkVersion(isMainnet bool) string {
	if isMainnet {
		return pi.MainnetVersion
	}
	return pi.BuildnetVersion
}

/*
RemoveOldNodeVersionsArtifacts RemoveOldNodeVersionsArtifacts removes node version artifact folders not matching the current
mainnet or buildnet versions (e.g. after a plugin update).
*/
func (pi *PluginInfo) RemoveOldNodeVersionsArtifacts(folderPath string) error {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", folderPath, err)
	}

	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() && name != pi.MainnetVersion && name != pi.BuildnetVersion {
			if err := os.RemoveAll(filepath.Join(folderPath, name)); err != nil {
				return fmt.Errorf("failed to remove directory %s: %v", filepath.Join(folderPath, name), err)
			}
		}
	}

	return nil
}

// nodeWalletFolders are the folders, relative to a node version folder, that hold the keys of the node user:
// the staking keys of the node and the wallet of the client. Both are encrypted with the node password.
var nodeWalletFolders = []string{
	filepath.Join("massa-node", "config", "staking_wallets"),
	filepath.Join("massa-client", "wallets"),
}

/*
MigrateNodeVersionsData copies the staking keys and the client wallet of the previous node version of each network
into the current version folder (e.g. after a plugin update), so that the node keeps staking the same addresses
with the same password. A folder is only copied when the current version does not have it yet or has it empty.
It must run before RemoveOldNodeVersionsArtifacts, which deletes the previous version folders.
*/
func (pi *PluginInfo) MigrateNodeVersionsData(folderPath string) error {
	entries, err := os.ReadDir(folderPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", folderPath, err)
	}

	for _, currentVersion := range []string{pi.MainnetVersion, pi.BuildnetVersion} {
		previousVersion := latestPreviousVersion(entries, currentVersion)
		if previousVersion == "" {
			continue
		}

		for _, walletFolder := range nodeWalletFolders {
			src := filepath.Join(folderPath, previousVersion, walletFolder)
			dst := filepath.Join(folderPath, currentVersion, walletFolder)

			if isEmptyOrMissingDir(src) || !isEmptyOrMissingDir(dst) {
				continue
			}

			if err := copyDirAtomically(src, dst); err != nil {
				return fmt.Errorf("failed to copy %s to %s: %w", src, dst, err)
			}
		}
	}

	return nil
}

// latestPreviousVersion returns the highest version folder of the same network as currentVersion
// (same prefix, e.g. MAIN or DEVN), other than currentVersion. It returns an empty string if there is none.
func latestPreviousVersion(entries []os.DirEntry, currentVersion string) string {
	prefix, _, found := strings.Cut(currentVersion, ".")
	if !found {
		return ""
	}

	latest := ""
	for _, entry := range entries {
		name := entry.Name()
		if !entry.IsDir() || name == currentVersion || !strings.HasPrefix(name, prefix+".") {
			continue
		}

		if latest == "" || compareVersions(name, latest) > 0 {
			latest = name
		}
	}

	return latest
}

// compareVersions compares the numeric parts of two node versions of the same network (e.g. MAIN.4.10 > MAIN.4.2).
func compareVersions(a, b string) int {
	partsA := strings.Split(a, ".")[1:]
	partsB := strings.Split(b, ".")[1:]

	for i := 0; i < len(partsA) && i < len(partsB); i++ {
		numA, errA := strconv.Atoi(partsA[i])
		numB, errB := strconv.Atoi(partsB[i])
		if errA != nil || errB != nil {
			return strings.Compare(partsA[i], partsB[i])
		}

		if numA != numB {
			return numA - numB
		}
	}

	return len(partsA) - len(partsB)
}

func isEmptyOrMissingDir(path string) bool {
	entries, err := os.ReadDir(path)
	return err != nil || len(entries) == 0
}

// copyDirAtomically copies src to a temporary folder next to dst, then renames it to dst,
// so that an interrupted copy does not leave a partial dst that would prevent the next attempt.
func copyDirAtomically(src, dst string) error {
	tmp := dst + ".migrating"
	if err := os.RemoveAll(tmp); err != nil {
		return err
	}

	if err := copyDir(src, tmp); err != nil {
		return err
	}

	// dst is missing or empty
	if err := os.RemoveAll(dst); err != nil {
		return err
	}

	return os.Rename(tmp, dst)
}

// copyDir copies the src folder content into dst, keeping the files permissions.
func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		target := filepath.Join(dst, relPath)

		info, err := entry.Info()
		if err != nil {
			return err
		}

		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm())
		}

		return copyFile(path, target, info.Mode().Perm())
	})
}

func copyFile(src, dst string, perm os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}

	return out.Close()
}

func (pi *PluginInfo) GetPwdByNetwork(isMainnet bool) string {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	if isMainnet {
		return pi.PwdMainnet
	}
	return pi.PwdBuildnet
}

// GetPwd returns the current password
func (pi *PluginInfo) GetPwd() string {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	if pi.IsMainnet {
		return pi.PwdMainnet
	}
	return pi.PwdBuildnet
}

// SetPwd sets the password
func (pi *PluginInfo) SetPwd(pwd string) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	if pi.IsMainnet {
		pi.PwdMainnet = pwd
	} else {
		pi.PwdBuildnet = pwd
	}
}

// GetAutoRestart returns the current auto restart setting
func (pi *PluginInfo) GetAutoRestart() bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	return pi.AutoRestart
}

// SetAutoRestart sets the auto restart setting
func (pi *PluginInfo) SetAutoRestart(autoRestart bool) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	pi.AutoRestart = autoRestart
}

// GetIsMainnet returns the current mainnet setting
func (pi *PluginInfo) GetIsMainnet() bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	return pi.IsMainnet
}

// SetIsMainnet sets the mainnet setting
func (pi *PluginInfo) SetIsMainnet(isMainnet bool) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	pi.IsMainnet = isMainnet
}

func (pi *PluginInfo) GetNetwork() utils.Network {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	if pi.IsMainnet {
		return utils.NetworkMainnet
	}
	return utils.NetworkBuildnet
}
