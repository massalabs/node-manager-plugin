package nodeManager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/massalabs/station/pkg/logger"
	"github.com/massalabs/station/pkg/node"
	"gopkg.in/natefinch/lumberjack.v2"
)

type NodeInfos struct {
	isMainnet bool
	version   string
}

type NodeManager struct {
	mu                        sync.Mutex
	serverProcess             *os.Process
	statusChan                chan NodeStatus
	closeBootstrapMonitorChan chan struct{}
	nodeInfos                 NodeInfos
	status                    NodeStatus
	massaNodeDirManager       nodeDirManager
}

const (
	nodeURL                = "http://localhost:33035"
	bootstrapCompleteStr   = "massa_bootstrap::client: Successful bootstrap"
	nodeStdoutReadInterval = 5 * time.Second
	statusChanCapacity     = 100 // To make sure the channel will not be blocking
	stdoutReadBufferSize   = 512 // Same size used in io.ReadAll
)

// NewNodeManager creates a new NodeManager instance
func NewNodeManager() (*NodeManager, error) {
	nodeDirManag := nodeDirManager{}
	if err := nodeDirManag.init(); err != nil {
		return nil, err
	}

	return &NodeManager{
		status:                    NodeStatusOff,
		statusChan:                make(chan NodeStatus, statusChanCapacity),
		closeBootstrapMonitorChan: make(chan struct{}),
		massaNodeDirManager:       nodeDirManag,
	}, nil
}

// StartNode starts the node process
func (nm *NodeManager) StartNode(isMainnet bool, pwd string) (string, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if IsRunning(nm.status) {
		logger.Infof("Node is already running")
		return "", fmt.Errorf("node is already running")
	}

	// set node network
	nm.nodeInfos.isMainnet = isMainnet
	nodeArgs := []string{"-p", pwd} // args for node
	networkName := "buildnet"
	if isMainnet {
		networkName = "mainnet"
		nodeArgs = append(nodeArgs, "-a")
	}
	logger.Infof("Starting node in %s mode", networkName)

	// Retrieve the massa node binary and version corresponding to selected network (defined by isMainnet param)
	nodeBinPath, version, err := nm.massaNodeDirManager.getNodeBinAndVersion(isMainnet)
	if err != nil {
		return "", fmt.Errorf("failed to get node binary path: %v", err)
	}

	// store node version
	nm.nodeInfos.version = version

	logger.Infof("Starting node process at %s", nodeBinPath)

	cmd := exec.Command(nodeBinPath, nodeArgs...)
	/* to run child process in a new process group.
	By default, node-manager-plugin process and it's massa node subprocess
	are in the same process group which means that all signals that are sent
	to one of them are also sent to the other one.
	This means that if the node-mananger-plugin is closed with ctrl-c from the terminal
	The massa node subprocess will also be closed independently from it's parent process.
	For clean shuttdown, we want the child process to be closed by it's parent, thus we
	launch it in it's own process group.
	*/
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // pgid: process group id.
	}
	cmd.Dir = filepath.Dir(nodeBinPath) // the command is executed in the folder of massa node binary

	nodeLogger := lumberjack.Logger{} // TODO

	cmd.Stdout = &nodeLogger
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("Failed to start massa node: %v", err)
	}

	logger.Infof("Node process started with PID: %d", cmd.Process.Pid)

	go nm.monitorBootstrapping()
	go nm.handleNodeStoped(cmd)

	nm.serverProcess = cmd.Process

	return version, nil
}

/*
Return the current status and an unidirectional buffered channel that return
the new status when it has been updated/
*/
func (nm *NodeManager) GetStatus() (NodeStatus, <-chan NodeStatus) {
	return nm.status, nm.statusChan
}

func (nm *NodeManager) StopNode() error {
	if !IsRunning(nm.status) {
		logger.Infof("Node is already stopped")
		return nil
	}

	logger.Infof("Stopping Massa node process...")

	nm.mu.Lock()
	if nm.status == NodeStatusBootstrapping {
		nm.closeBootstrapMonitorChan <- struct{}{}
	}
	nm.status = NodeStatusStopping
	nm.statusChan <- NodeStatusStopping
	nm.mu.Unlock()

	// Send a SIGTERM signal to gracefully shut down
	if err := nm.serverProcess.Signal(syscall.SIGTERM); err != nil {
		logger.Errorf("Failed to send SIGTERM: %v", err)

		// Force kill as a fallback
		if err = nm.serverProcess.Kill(); err != nil {
			return err
		}
	}

	// Wait for the process to exit
	timeout := time.Now().Add(5 * time.Second)
	for time.Now().Before(timeout) && IsRunning(nm.status) {
		time.Sleep(100 * time.Millisecond)
	}

	// If still running after timeout, force kill
	if IsRunning(nm.status) {
		_ = nm.serverProcess.Kill()
	}
	return nil
}

func (nm *NodeManager) Logs() (string, error) {
	return "not implemented", nil
}

/*
Firsteval, it set the status to "bootstrapping"
Then it Read from the massa node's stdout and wait for the
"massa_bootstrap::client: Successful bootstrap" text to be printed.
Then it updates the node status from "bootstrapping" to "on" and return
*/
func (nm *NodeManager) monitorBootstrapping() {
	nm.setStatus(NodeStatusBootstrapping)

	logger.Info("Bootstrap started...")

	ticker := time.NewTicker(nodeStdoutReadInterval)
	defer ticker.Stop()

	client := node.NewClient(nodeURL)
	for {
		select {
		case <-nm.closeBootstrapMonitorChan:
			logger.Debug("Stop bootstrap monitor goroutine because received stop chan signal")
			return
		case <-ticker.C:
			if nm.status == NodeStatusOn {
				return
			}

			/*Check if the node has finished bootstrapping by sending a request to it's api
			If the request fails, it means that the node is still bootstrapping*/
			logger.Debug("Send a get_status request to the node to check if it has bootstrapped")
			_, err := node.Status(client)
			if err != nil {
				if connRefused(err) {
					logger.Debug("Connection refused, the node is still bootstrapping")
					continue
				}
				nm.setStatus(NodeManagerErrorStatus)
				logger.Errorf("attempted to retrieve the status of the massa node but got error: %w", err)
				continue
			}

			logger.Info("Bootstrap completed ! \n Node is Up")

			nm.setStatus(NodeStatusOn)
			/*Don't return here because a msg migth have been sent
			through the closeBootstrapMonitorChan while we were reading the stdout.
			If we return here, closeBootstrapMonitorChan sender migth be blocked.
			This way we avoid locking with mutex all the "case <-ticker.C" logic.
			*/
		}
	}
}

/*
handleNodeStoped wait for the node process to exit.
If the process has exited with error, it handle this.
It update the status to off or error
*/
func (nm *NodeManager) handleNodeStoped(cmd *exec.Cmd) {
	err := cmd.Wait() // Wait for the command to exit
	status := NodeStatusOff

	if err != nil && !isUserIntterupted(err) {
		logger.Errorf("massa node process exited with error: %v", err)
		status = NodeStatusError
	}

	nm.mu.Lock()
	nm.status = status
	nm.statusChan <- status
	nm.serverProcess = nil
	nm.mu.Unlock()

	logger.Infof("massa node process exited")
}

/*
set a new status and send it through the status channel.
Should not be called inside the nodeManager's mutext
*/
func (nm *NodeManager) setStatus(status NodeStatus) {
	nm.mu.Lock()
	nm.status = status
	nm.statusChan <- status
	nm.mu.Unlock()
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
