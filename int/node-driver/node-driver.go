package nodeDriver

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	NodeDirManagerPkg "github.com/massalabs/node-manager-plugin/int/node-bin-dir-manager"
	"github.com/massalabs/station/pkg/logger"
)

// NodeDriver defines the interface for managing node processes
type NodeDriver interface {
	/*
		StartNode starts a node process with the given password and node logger.
		It returns a channel that will send a ProcessExitedResult when the node process exits.
	*/
	StartNode(isMainnet bool, pwd string, nodeLogger io.Writer) (<-chan ProcessExitedResult, error)

	/*
		StopNode stops the currently running node process
		It returns any error that occurred during the stop operation
	*/
	StopNode() error
}

type ProcessExitedResult struct {
	Err error
}

// NodeDriverImpl implements the NodeDriver interface
type NodeDriverImpl struct {
	cancelNodeProcess context.CancelFunc
	nodeDirManager    NodeDirManagerPkg.NodeDirManager
}

// NewNodeDriver creates a new NodeDriver instance
func NewNodeDriver(nodeDirManager NodeDirManagerPkg.NodeDirManager) NodeDriver {
	return &NodeDriverImpl{
		nodeDirManager: nodeDirManager,
	}
}

/*
StartNode starts the node process
It returns a channel that will send a ProcessExitedResult when the node process exits.
*/
func (nd *NodeDriverImpl) StartNode(isMainnet bool, pwd string, nodeLogger io.Writer) (<-chan ProcessExitedResult, error) {
	// Set node parameters
	nodeArgs := []string{"-p", pwd, "-a"} // args for node process
	networkName := "buildnet"
	if isMainnet {
		networkName = "mainnet"
	}
	logger.Infof("Starting node in %s mode", networkName)

	// Retrieve the massa node binary corresponding to selected network (defined by isMainnet param)
	nodeBinPath, err := nd.nodeDirManager.GetNodeBin(isMainnet)
	if err != nil {
		return nil, fmt.Errorf("failed to get massa node binary path: %v", err)
	}

	// Prepare the node subprocess
	logger.Infof("Starting node process at %s", nodeBinPath)

	// Create a new context for this node instance
	ctx, cancel := context.WithCancel(context.Background())
	nd.cancelNodeProcess = cancel

	cmd := exec.CommandContext(ctx, nodeBinPath, nodeArgs...)
	cmd.Dir = filepath.Dir(nodeBinPath) // the command is executed in the folder of node binary

	cmd.Stdout = nodeLogger
	cmd.Stderr = nodeLogger

	// Launch the node subprocess
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start node: %v", err)
	}

	logger.Infof("node process started with PID: %d", cmd.Process.Pid)

	processExitedChan := make(chan ProcessExitedResult)
	go func() {
		err := cmd.Wait()
		processExitedChan <- ProcessExitedResult{
			Err: err,
		}
		close(processExitedChan)
	}()

	return processExitedChan, nil
}

/*
StopNode stops the node process
It returns any error that occurred during the stop operation
*/
func (nd *NodeDriverImpl) StopNode() error {
	nd.cancelNodeProcess()

	return nil
}
