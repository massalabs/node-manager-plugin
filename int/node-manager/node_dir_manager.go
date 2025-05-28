package nodeManager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	massaNodeFolderDir   = "node-massa"
	mainnetFolderPrefix  = "MAIN"
	buildnetFolderPrefix = "DEVN"
	nodeBinName          = "massa-node"
	nodeBinFolder        = "massa-node"
)

type nodeDirManager struct {
	nodeFolderPath string // the folder in which are stored massa node binaries for both mainnet and buildnet
	extension      string
	NodeBinPath    string
}

func (ndm *nodeDirManager) getNodeBinAndVersion(isMainnet bool) (string, string, error) {
	prefix := buildnetFolderPrefix
	expectedNetwork := "buildnet"
	binPath := ""
	version := ""
	if isMainnet {
		prefix = mainnetFolderPrefix
		expectedNetwork = "mainnet"
	}

	if ndm.nodeFolderPath == "" {
		if err := ndm.init(); err != nil {
			return "", "", err
		}
	}

	// retrieve all entries in the folder
	entries, err := os.ReadDir(ndm.nodeFolderPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read directory %s: %v", ndm.nodeFolderPath, err)
	}

	// Find the directory that starts with the prefix corresponding to the expected network
	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), prefix) {
			binPath = filepath.Join(ndm.nodeFolderPath, entry.Name(), nodeBinFolder, nodeBinName+ndm.extension)
			version = entry.Name()
			break
		}
	}

	if binPath == "" {
		return "", "", fmt.Errorf("failed to find %s bin in %s directory", expectedNetwork, ndm.nodeFolderPath)
	}

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("node binary not found at %s", binPath)
	}

	ndm.NodeBinPath = binPath

	return binPath, version, nil
}

func (ndm *nodeDirManager) init() error {
	// Determine the plugin's executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	massaNodeFolderPath := filepath.Join(filepath.Dir(execPath), massaNodeFolderDir)

	// Check if the directory exists
	if _, err := os.Stat(massaNodeFolderPath); os.IsNotExist(err) {
		return fmt.Errorf("massa node folder not found at %s", massaNodeFolderPath)
	}

	ndm.nodeFolderPath = massaNodeFolderPath

	// On Windows, add .exe extension
	if filepath.Ext(execPath) == ".exe" {
		ndm.extension = ".exe"
	}

	return nil
}
