package nodeDirManager

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/massalabs/node-manager-plugin/int/config"
)

const (
	massaNodeFolderDir = "node-massa"
	nodeBinName        = "massa-node"
	nodeBinFolder      = "massa-node"
	clientBinName      = "massa-client"
	clientBinFolder    = "massa-client"
	walletFolder       = "wallets"
)

type NodeDirManager interface {
	GetClientBin(isMainnet bool) (string, error)
	GetNodeBin(isMainnet bool) (string, error)
	HasClientAddresses(isMainnet bool) (bool, error)
}

type nodeDirManager struct {
	nodeFolderPath string // the folder in which are stored massa node binaries for both mainnet and buildnet
	extension      string
}

func NewNodeDirManager() (NodeDirManager, error) {
	ndm := &nodeDirManager{}
	if err := ndm.init(); err != nil {
		return nil, err
	}
	return ndm, nil
}

func (ndm *nodeDirManager) GetClientBin(isMainnet bool) (string, error) {
	version := config.GlobalPluginInfo.GetNetworkVersion(isMainnet)
	binPath := filepath.Join(ndm.nodeFolderPath, version, clientBinFolder, clientBinName+ndm.extension)

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		return "", fmt.Errorf("client binary not found at %s", binPath)
	}

	return binPath, nil
}

func (ndm *nodeDirManager) GetNodeBin(isMainnet bool) (string, error) {
	version := config.GlobalPluginInfo.GetNetworkVersion(isMainnet)

	binPath := filepath.Join(ndm.nodeFolderPath, version, nodeBinFolder, nodeBinName+ndm.extension)

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		return "", fmt.Errorf("node binary not found at %s", binPath)
	}

	return binPath, nil
}

func (ndm *nodeDirManager) HasClientAddresses(isMainnet bool) (bool, error) {
	version := config.GlobalPluginInfo.GetNetworkVersion(isMainnet)

	clientPath := filepath.Join(ndm.nodeFolderPath, version, clientBinFolder)

	if _, err := os.Stat(clientPath); os.IsNotExist(err) {
		return false, fmt.Errorf("client folder not found at %s", clientPath)
	}

	walletPath := filepath.Join(clientPath, walletFolder)

	if _, err := os.Stat(walletPath); os.IsNotExist(err) {
		return false, fmt.Errorf("wallet folder not found at %s", walletPath)
	}

	entries, err := os.ReadDir(walletPath)
	if err != nil {
		return false, fmt.Errorf("failed to read directory %s: %v", walletPath, err)
	}

	return len(entries) > 0, nil
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

	// clean old node versions if any
	if err := config.GlobalPluginInfo.RemoveOldNodeVersionsArtifacts(ndm.nodeFolderPath); err != nil {
		return fmt.Errorf("failed to remove old node versions bin folders: %v", err)
	}

	return nil
}
