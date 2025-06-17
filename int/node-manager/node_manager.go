package nodeManager

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/massalabs/node-manager-plugin/api/models"
	"github.com/massalabs/node-manager-plugin/int/config"
	"github.com/massalabs/station/pkg/logger"
	"github.com/massalabs/station/pkg/node"
)

type NodeInfos struct {
	isMainnet bool
	pwd       string
}

type NodeManager struct {
	mu                     sync.Mutex
	statusChan             chan NodeStatus
	nodeInfos              NodeInfos
	status                 NodeStatus
	massaNodeDirManager    nodeDirManager
	nodeMonitor            *NodeMonitor
	autoRestart            bool
	nodeLogger             *NodeLogger
	cancelNodeAndAsyncTask context.CancelFunc // cancel function to stop node subprocess and all concurrent tasks
	nodeProcess            *os.Process        // store the node process
}

const (
	nodeURL                = "http://localhost:33035"
	bootstrapCheckInterval = 5 * time.Second
	statusChanCapacity     = 100              // To make sure the channel will not be blocking
	desyncCheckInterval    = 30 * time.Second // Interval for desync check
	restartCooldown        = 5 * time.Second  // Time to wait before restarting the node
)

// NewNodeManager creates a new NodeManager instance
func NewNodeManager(config config.PluginConfig) (*NodeManager, error) {
	nodeDirManag := nodeDirManager{}
	if err := nodeDirManag.init(); err != nil {
		return nil, err
	}

	nodeLogger, err := NewNodeLogger(config)
	if err != nil {
		return nil, err
	}

	return &NodeManager{
		status:              NodeStatusOff,
		statusChan:          make(chan NodeStatus, statusChanCapacity),
		massaNodeDirManager: nodeDirManag,
		nodeLogger:          nodeLogger,
		nodeMonitor:         NewNodeMonitor(),
	}, nil
}

// StartNode starts the massa node process
func (nodeMana *NodeManager) StartNode(isMainnet bool, pwd string) (string, error) {
	nodeMana.mu.Lock()
	defer nodeMana.mu.Unlock()

	if IsRunning(nodeMana.status) {
		logger.Infof("massa node is already running")
		return "", fmt.Errorf("massa node is already running")
	}

	nodeMana.status = NodeStatusStarting
	nodeMana.statusChan <- NodeStatusStarting

	// Set node parameters
	nodeMana.nodeInfos.isMainnet = isMainnet
	nodeMana.nodeInfos.pwd = pwd
	nodeArgs := []string{"-p", pwd} // args for massa node process
	networkName := "buildnet"
	if isMainnet {
		networkName = "mainnet"
		nodeArgs = append(nodeArgs, "-a")
	}
	logger.Infof("Starting massa node in %s mode", networkName)

	// Retrieve the massa node binary and version corresponding to selected network (defined by isMainnet param)
	nodeBinPath, version, err := nodeMana.massaNodeDirManager.getNodeBinAndVersion(isMainnet)
	if err != nil {
		return "", fmt.Errorf("failed to get massa node binary path: %v", err)
	}

	// Prepare the node subprocess
	logger.Infof("Starting massa node process at %s", nodeBinPath)

	// Create a new context for this node instance
	ctx, cancel := context.WithCancel(context.Background())
	nodeMana.cancelNodeAndAsyncTask = cancel

	cmd := exec.CommandContext(ctx, nodeBinPath, nodeArgs...)

	/* to run child process in a new process group.
	By default, node-manager-plugin process and it's massa node subprocess
	are in the same process group which means that all signals that are sent
	to one of them are also sent to the other one.
	This means that if the node-manager-plugin is closed with ctrl-c from the terminal
	The massa node subprocess will also be closed independently from it's parent process.
	For clean shuttdown, we want the child process to be closed by it's parent, thus we
	launch it in it's own process group.
	*/
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // pgid: process group id.
	}
	cmd.Dir = filepath.Dir(nodeBinPath) // the command is executed in the folder of massa node binary

	// Set the node logger as the stdout and stderr of the node process
	nodeLogger := nodeMana.nodeLogger.newLogger(version)

	nodeLogger.Write([]byte(fmt.Sprintf("\n \n>>> new node session (%s): \n", time.Now().Format("2006-01-02 15:04:05"))))
	cmd.Stdout = nodeLogger
	cmd.Stderr = nodeLogger

	// Launch the node subprocess
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("Failed to start massa node: %v", err)
	}

	logger.Infof("massa node process started with PID: %d", cmd.Process.Pid)

	// launch bootstrap monitor goroutine
	go nodeMana.monitorBootstrapping(ctx)

	// launch node stopped monitor goroutine
	go nodeMana.handleNodeStopped(cmd)

	nodeMana.nodeProcess = cmd.Process

	return version, nil
}

/*
Return the current status and an unidirectional buffered channel that return
the new status when it has been updated/
*/
func (nodeMana *NodeManager) GetStatus() (NodeStatus, <-chan NodeStatus) {
	return nodeMana.status, nodeMana.statusChan
}

func (nodeMana *NodeManager) StopNode() error {
	nodeMana.mu.Lock()
	defer nodeMana.mu.Unlock()

	if !IsRunning(nodeMana.status) {
		logger.Infof("massa node process is already stopped")
		return fmt.Errorf("massa node process is already stopped")
	}

	if nodeMana.status == NodeStatusStopping {
		logger.Infof("massa node process is already stopping")
		return fmt.Errorf("massa node process is already stopping")
	}

	nodeMana.status = NodeStatusStopping
	nodeMana.statusChan <- NodeStatusStopping

	logger.Infof("Stopping Massa node process...")
	nodeMana.cancelNodeAndAsyncTask()
	return nil
}

func (nodeMana *NodeManager) Logs(isMainnet bool) (string, error) {
	_, version, err := nodeMana.massaNodeDirManager.getNodeBinAndVersion(isMainnet)
	if err != nil {
		return "", fmt.Errorf("failed to get massa node binary path: %v", err)
	}
	return nodeMana.nodeLogger.getLogs(version)
}

func (nodeMana *NodeManager) SetAutoRestart(autoRestart bool) {
	nodeMana.autoRestart = autoRestart
}

func (nodeMana *NodeManager) GetNodeManagerConfig() models.Config {
	return models.Config{
		AutoRestart: nodeMana.autoRestart,
	}
}

/*
First, it set the status to "bootstrapping"
Then it call the massa node api on get_status endpoint to check if the node has bootstrapped.
If the node has bootstrapped, it starts the desync monitor goroutine.
*/
func (nodeMana *NodeManager) monitorBootstrapping(ctx context.Context) {
	nodeMana.setStatus(NodeStatusBootstrapping)

	logger.Info("Bootstrap started...")

	ticker := time.NewTicker(bootstrapCheckInterval)
	defer ticker.Stop()

	client := node.NewClient(nodeURL)
	for {
		select {
		case <-ctx.Done():
			logger.Debug("Stop bootstrap monitor goroutine because received close node sub process signal")
			return
		case <-ticker.C:
			/*Check if the massa node process has finished bootstrapping by sending a request to it's api
			If the request fails, it means that the node is still bootstrapping*/
			logger.Debug("Send a get_status request to the massa node to check if it has bootstrapped")
			_, err := node.Status(client)
			if err != nil {
				if connRefused(err) {
					logger.Debug("Connection refused, the massa node is still bootstrapping")
					continue
				}
				logger.Errorf("attempted to retrieve the status of the massa node but got error: %w", err)
				continue
			}

			nodeMana.mu.Lock()
			defer nodeMana.mu.Unlock()
			if IsClosedOrClosing(nodeMana.status) {
				return
			}

			logger.Info("Bootstrap completed ! \n Massa Node is Up")

			// Start desync monitor goroutine after bootstrapping
			go nodeMana.handleNodeDesync(ctx)

			nodeMana.status = NodeStatusOn
			nodeMana.statusChan <- NodeStatusOn
			return
		}
	}
}

// handleNodeDesync monitors the node for desync and restarts if autoRestart is enabled
func (nodeMana *NodeManager) handleNodeDesync(ctx context.Context) {
	if nodeMana.status != NodeStatusOn {
		logger.Errorf("handleNodeDesync: node is not running, cannot monitor desync")
		return
	}
	ticker := time.NewTicker(desyncCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Debug("Stop desync monitor goroutine because context was cancelled")
			return
		case <-nodeMana.nodeMonitor.MonitorDesync(ctx, desyncCheckInterval):
			logger.Warn("Node is desynced")

			nodeMana.mu.Lock()
			if IsClosedOrClosing(nodeMana.status) {
				nodeMana.mu.Unlock()
				return
			}
			nodeMana.status = NodeStatusDesynced
			nodeMana.statusChan <- NodeStatusDesynced
			nodeMana.mu.Unlock()

			if nodeMana.autoRestart {
				logger.Info("Auto-restarting node due to desync")
				if err := nodeMana.StopNode(); err != nil {
					logger.Errorf("Failed to stop node for auto-restart: %v", err)
					continue
				}

				// Wait for the node to stop
				for IsRunning(nodeMana.status) {
					logger.Debug("Waiting for node to stop")
					time.Sleep(5 * time.Second)
				}

				_, err := nodeMana.StartNode(nodeMana.nodeInfos.isMainnet, "")
				if err != nil {
					logger.Errorf("Failed to restart node: %v", err)
				}
			}
		}
	}
}

/*
handleNodeStoped wait for the massa node process to exit.
If the process has exited with error, it handle this.
It update the status to off or error
*/
func (nodeMana *NodeManager) handleNodeStopped(cmd *exec.Cmd) {
	err := cmd.Wait() // Wait for the command to exit
	status := NodeStatusOff

	if err != nil && !isUserIntterupted(err) {
		logger.Errorf("massa node process exited with error: %v", err)
		status = NodeStatusCrashed

		// // close all concurrent tasks
		// nodeMana.cancelAsyncTask()

		// if auto-restart option is enabled, restart the node
		if nodeMana.autoRestart {
			logger.Info("Auto-restarting node due to error")
			nodeMana.setStatus(status)
			time.Sleep(restartCooldown)
			_, err := nodeMana.StartNode(nodeMana.nodeInfos.isMainnet, nodeMana.nodeInfos.pwd)
			if err != nil {
				logger.Errorf("Failed to restart node: %v", err)
			}
			return
		}
	}

	nodeMana.mu.Lock()
	nodeMana.status = status
	nodeMana.statusChan <- status
	nodeMana.nodeProcess = nil
	nodeMana.mu.Unlock()

	logger.Infof("massa node process exited")
}

/*
set a new status and send it through the status channel.
Should not be called inside the nodeManager's mutext
*/
func (nodeMana *NodeManager) setStatus(status NodeStatus) {
	nodeMana.mu.Lock()
	nodeMana.status = status
	nodeMana.statusChan <- status
	nodeMana.mu.Unlock()
}

// isUserIntterupted checks if the error is due to user interruption SIGTERM or SIGKILL
func isUserIntterupted(err error) bool {
	if exitErr, ok := err.(*exec.ExitError); ok {
		if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
			if ws.Signal() == syscall.SIGTERM || ws.Signal() == syscall.SIGKILL {
				return true
			}
		}
	}
	return false
}
