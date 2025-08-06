package nodeManager

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/massalabs/node-manager-plugin/int/config"
	nodeStatusPkg "github.com/massalabs/node-manager-plugin/int/core/NodeStatus"
	NodeDirManager "github.com/massalabs/node-manager-plugin/int/node-bin-dir-manager"
	nodeDriver "github.com/massalabs/node-manager-plugin/int/node-driver"
	"github.com/massalabs/station/pkg/logger"
)

type INodeManager interface {
	StartNode(isMainnet bool, pwd string) error
	StopNode() error
	Logs(isMainnet bool) (string, error)

	GetStatus() nodeStatusPkg.NodeStatus
	Close() error
}

type NodeManager struct {
	mu                sync.Mutex
	config            *config.PluginConfig
	status            nodeStatusPkg.NodeStatus
	buildnetLogger    io.WriteCloser
	mainnetLogger     io.WriteCloser
	processExitedChan <-chan nodeDriver.ProcessExitedResult
	nodeMonitor       NodeMonitoring
	NodeLogManager    *NodeLogManager
	nodeDirManager    NodeDirManager.NodeDirManager
	cancelAsyncTask   context.CancelFunc // cancel function to stop node subprocess and all concurrent tasks
	nodeDriver        nodeDriver.NodeDriver
	statusDispatcher  nodeStatusPkg.NodeStatusDispatcher
}

// NewNodeManager creates a new NodeManager instance
func NewNodeManager(
	config *config.PluginConfig,
	nodeDirManager NodeDirManager.NodeDirManager,
	nodeMonitor NodeMonitoring,
	nodeDriver nodeDriver.NodeDriver,
	statusDispatcher nodeStatusPkg.NodeStatusDispatcher,
) (*NodeManager, error) {
	nodeLogManager, err := NewNodeLogManager(config)
	if err != nil {
		return nil, err
	}

	return &NodeManager{
		status:           nodeStatusPkg.NodeStatusOff,
		config:           config,
		NodeLogManager:   nodeLogManager,
		nodeMonitor:      nodeMonitor,
		nodeDirManager:   nodeDirManager,
		nodeDriver:       nodeDriver,
		statusDispatcher: statusDispatcher,
	}, nil
}

// StartNode starts the massa node process
func (nodeMana *NodeManager) StartNode(isMainnet bool, pwd string) error {
	nodeMana.mu.Lock()
	defer nodeMana.mu.Unlock()

	if IsRunning(nodeMana.status) {
		logger.Infof("massa node is already running")
		return fmt.Errorf("massa node is already running")
	}

	nodeMana.setStatus(nodeStatusPkg.NodeStatusStarting)

	nodeLogger, err := nodeMana.getLogger(isMainnet)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(nodeLogger, "\n\n>>> new node session (%s): \n", time.Now().Format("2006-01-02 15:04:05"))
	if err != nil {
		return fmt.Errorf("failed to write to node logger: %v", err)
	}

	nodeMana.processExitedChan, err = nodeMana.nodeDriver.StartNode(isMainnet, pwd, nodeLogger)
	if err != nil {
		return fmt.Errorf("failed to start node: %v", err)
	}

	nodeMana.setStatus(nodeStatusPkg.NodeStatusBootstrapping)

	// Update global plugin info
	config.GlobalPluginInfo.SetIsMainnet(isMainnet)
	config.GlobalPluginInfo.SetPwd(pwd)

	ctx, cancel := context.WithCancel(context.Background())
	nodeMana.cancelAsyncTask = cancel

	go nodeMana.HandleBootstrapping(ctx)

	// launch node stopped handler goroutine
	go nodeMana.handleNodeStopped()

	return nil
}

func (nodeMana *NodeManager) StopNode() error {
	nodeMana.mu.Lock()
	defer nodeMana.mu.Unlock()

	if !IsRunning(nodeMana.status) {
		logger.Infof("massa node process is already stopped")
		return fmt.Errorf("massa node process is already stopped")
	}

	if nodeMana.status == nodeStatusPkg.NodeStatusStopping {
		logger.Infof("massa node process is already stopping")
		return fmt.Errorf("massa node process is already stopping")
	}

	nodeMana.setStatus(nodeStatusPkg.NodeStatusStopping)

	logger.Infof("Stopping Massa node process...")
	nodeMana.cancelAsyncTask()

	if err := nodeMana.nodeDriver.StopNode(); err != nil {
		return fmt.Errorf("failed to stop node: %v", err)
	}

	return nil
}

func (nodeMana *NodeManager) Logs(isMainnet bool) (string, error) {
	version, err := nodeMana.nodeDirManager.GetVersion(isMainnet)
	if err != nil {
		return "", fmt.Errorf("failed to get massa node binary path: %v", err)
	}
	return nodeMana.NodeLogManager.getLogs(version)
}

func (nodeMana *NodeManager) GetStatus() nodeStatusPkg.NodeStatus {
	return nodeMana.status
}

/*
HandleBootstrapping set the status to NodeStatusBootstrapping then it subscribe to the channel returned by MonitorBootstrapping.
When the node has bootstrapped, it updates the status to NodeStatusOn and starts the desync monitor goroutine.
*/
func (nodeMana *NodeManager) HandleBootstrapping(ctx context.Context) {
	nodeMana.setStatus(nodeStatusPkg.NodeStatusBootstrapping)

	logger.Info("Bootstrap started...")
	for {
		select {
		case <-ctx.Done():
			return
		case <-nodeMana.nodeMonitor.MonitorBootstrapping(ctx, time.Duration(nodeMana.config.BootstrapCheckInterval)*time.Second):
			logger.Info("Bootstrap completed !")
			nodeMana.mu.Lock()
			defer nodeMana.mu.Unlock()

			if IsClosedOrClosing(nodeMana.status) {
				return
			}

			nodeMana.setStatus(nodeStatusPkg.NodeStatusOn)

			logger.Info("Massa Node is Up")

			go nodeMana.handleNodeDesync(ctx)
			return
		}
	}
}

/*
handleNodeDesync subscribe to the channel returned by MonitorDesync.
When the node is desynced, it updates the status to NodeStatusDesynced and restart the node if autoRestart is enabled.
*/
func (nodeMana *NodeManager) handleNodeDesync(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-nodeMana.nodeMonitor.MonitorDesync(ctx, time.Duration(nodeMana.config.DesyncCheckInterval)*time.Second):
			logger.Warn("Node is desynced")

			nodeMana.mu.Lock()
			if IsClosedOrClosing(nodeMana.status) {
				nodeMana.mu.Unlock()
				return
			}
			nodeMana.setStatus(nodeStatusPkg.NodeStatusDesynced)
			nodeMana.mu.Unlock()

			if config.GlobalPluginInfo.GetAutoRestart() {
				// Wait for the restart cooldown
				time.Sleep(time.Duration(nodeMana.config.RestartCooldown) * time.Second)

				// Stop the node
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

				// Start the node
				err := nodeMana.StartNode(config.GlobalPluginInfo.GetIsMainnet(), "")
				if err != nil {
					logger.Errorf("Failed to restart node: %v", err)
				}
			}
			return
		}
	}
}

/*
handleNodeStoped wait for the massa node process to exit.
If the process has exited with error, it handle it.
It update the status to NodeStatusOff or NodeStatusCrashed
*/
func (nodeMana *NodeManager) handleNodeStopped() {
	result := <-nodeMana.processExitedChan // Wait for the command to exit
	status := nodeStatusPkg.NodeStatusOff

	if result.Err != nil && !isUserInterrupted(result.Err) {
		logger.Errorf("massa node process exited with error: %v", result.Err)
		status = nodeStatusPkg.NodeStatusCrashed

		// if auto-restart option is enabled, restart the node
		if config.GlobalPluginInfo.GetAutoRestart() {
			logger.Info("Auto-restarting node due to error")
			nodeMana.setStatus(status)
			time.Sleep(time.Duration(nodeMana.config.RestartCooldown) * time.Second)
			err := nodeMana.StartNode(
				config.GlobalPluginInfo.GetIsMainnet(),
				config.GlobalPluginInfo.GetPwd(),
			)
			if err != nil {
				logger.Errorf("Failed to restart node: %v", err)
			}
			return
		}
	}

	nodeMana.setStatus(status)

	logger.Infof("massa node process exited")
}

// cleanup
func (nodeMana *NodeManager) Close() error {
	logger.Debug("Node manager cleanup")

	nodeMana.mu.Lock()
	defer nodeMana.mu.Unlock()

	if !IsClosedOrClosing(nodeMana.status) {
		logger.Debug("Stopping node")
		nodeMana.cancelAsyncTask()
		if err := nodeMana.nodeDriver.StopNode(); err != nil {
			return fmt.Errorf("failed to stop node: %v", err)
		}

		// if is still running, wait a little to let the time to subprocess to close.
		if IsRunning(nodeMana.status) {
			time.Sleep(3 * time.Second)
		}
	}

	// close the node logger
	if err := nodeMana.closeLoggers(); err != nil {
		return fmt.Errorf("failed to close node loggers: %v", err)
	}

	return nil
}

// update the status of the node and dispatch it to other services that are subscribed to the status change
func (nodeMana *NodeManager) setStatus(status nodeStatusPkg.NodeStatus) {
	nodeMana.status = status
	nodeMana.statusDispatcher.Publish(status)
}

// isUserInterrupted checks if the error is due to user interruption SIGTERM or SIGKILL
// func isUserInterrupted(err error) bool {
// 	if exitErr, ok := err.(*exec.ExitError); ok {
// 		if runtime.GOOS == "windows" {
// 			// Windows specific check
// 			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
// 				// Check for CTRL_C_EVENT (0xC000013A) or other Windows exit codes
// 				return status.ExitStatus() == windows.CTRL_C_EVENT ||
// 					status.ExitStatus() == windows.ERROR_OPERATION_ABORTED
// 			}
// 		} else {
// 			// Unix-like systems (Linux, macOS)
// 			if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
// 				return ws.Signaled() &&
// 					(ws.Signal() == syscall.SIGTERM ||
// 						ws.Signal() == syscall.SIGINT ||
// 						ws.Signal() == syscall.SIGQUIT)
// 			}
// 		}
// 	}
// 	return false
// }
