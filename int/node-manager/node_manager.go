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

// StartNode starts the nodeMana process
func (nodeMana *NodeManager) StartNode(isMainnet bool, pwd string) (string, error) {
	nodeMana.mu.Lock()
	defer nodeMana.mu.Unlock()

	if IsRunning(nodeMana.status) {
		logger.Infof("nodeMana is already running")
		return "", fmt.Errorf("nodeMana is already running")
	}

	// set nodeMana network
	nodeMana.nodeInfos.isMainnet = isMainnet
	nodeArgs := []string{"-p", pwd} // args for nodeMana
	networkName := "buildnet"
	if isMainnet {
		networkName = "mainnet"
		nodeArgs = append(nodeArgs, "-a")
	}
	logger.Infof("Starting nodeMana in %s mode", networkName)

	// Retrieve the massa nodeMana binary and version corresponding to selected network (defined by isMainnet param)
	nodeBinPath, version, err := nodeMana.massaNodeDirManager.getNodeBinAndVersion(isMainnet)
	if err != nil {
		return "", fmt.Errorf("failed to get nodeMana binary path: %v", err)
	}

	// store nodeMana version
	nodeMana.nodeInfos.version = version

	logger.Infof("Starting nodeMana process at %s", nodeBinPath)

	cmd := exec.Command(nodeBinPath, nodeArgs...)
	/* to run child process in a new process group.
	By default, nodeMana-manager-plugin process and it's massa nodeMana subprocess
	are in the same process group which means that all signals that are sent
	to one of them are also sent to the other one.
	This means that if the nodeMana-mananger-plugin is closed with ctrl-c from the terminal
	The massa nodeMana subprocess will also be closed independently from it's parent process.
	For clean shuttdown, we want the child process to be closed by it's parent, thus we
	launch it in it's own process group.
	*/
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // pgid: process group id.
	}
	cmd.Dir = filepath.Dir(nodeBinPath) // the command is executed in the folder of massa nodeMana binary

	nodeLogger := lumberjack.Logger{} // TODO

	cmd.Stdout = &nodeLogger
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("Failed to start massa nodeMana: %v", err)
	}

	logger.Infof("nodeMana process started with PID: %d", cmd.Process.Pid)

	go nodeMana.monitorBootstrapping()
	go nodeMana.handleNodeStoped(cmd)

	nodeMana.serverProcess = cmd.Process

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
	if !IsRunning(nodeMana.status) {
		logger.Infof("nodeMana is already stopped")
		return nil
	}

	logger.Infof("Stopping Massa nodeMana process...")

	if nodeMana.status == NodeStatusBootstrapping {
		nodeMana.closeBootstrapMonitorChan <- struct{}{}
	}
	nodeMana.status = NodeStatusStopping
	nodeMana.statusChan <- NodeStatusStopping

	// Send a SIGTERM signal to gracefully shut down
	if err := nodeMana.serverProcess.Signal(syscall.SIGTERM); err != nil {
		logger.Errorf("Failed to send SIGTERM: %v", err)

		// Force kill as a fallback
		if err = nodeMana.serverProcess.Kill(); err != nil {
			return err
		}
	}

	// Wait for the process to exit
	timeout := time.Now().Add(5 * time.Second)
	for time.Now().Before(timeout) && IsRunning(nodeMana.status) {
		time.Sleep(100 * time.Millisecond)
	}

	// If still running after timeout, force kill
	if IsRunning(nodeMana.status) {
		_ = nodeMana.serverProcess.Kill()
	}
	return nil
}

func (nodeMana *NodeManager) Logs() (string, error) {
	return "not implemented", nil
}

/*
Firsteval, it set the status to "bootstrapping"
Then it Read from the massa nodeMana's stdout and wait for the
"massa_bootstrap::client: Successful bootstrap" text to be printed.
Then it updates the nodeMana status from "bootstrapping" to "on" and return
*/
func (nodeMana *NodeManager) monitorBootstrapping() {
	nodeMana.setStatus(NodeStatusBootstrapping)

	logger.Info("Bootstrap started...")

	ticker := time.NewTicker(nodeStdoutReadInterval)
	defer ticker.Stop()

	client := node.NewClient(nodeURL)
	for {
		select {
		case <-nodeMana.closeBootstrapMonitorChan:
			logger.Debug("Stop bootstrap monitor goroutine because received closeBootstrapMonitor chan signal")
			return
		case <-ticker.C:
			if nodeMana.status == NodeStatusOn {
				return
			}

			/*Check if the nodeMana has finished bootstrapping by sending a request to it's api
			If the request fails, it means that the nodeMana is still bootstrapping*/
			logger.Debug("Send a get_status request to the nodeMana to check if it has bootstrapped")
			_, err := node.Status(client)
			if err != nil {
				if connRefused(err) {
					logger.Debug("Connection refused, the nodeMana is still bootstrapping")
					continue
				}
				nodeMana.setStatus(NodeManagerErrorStatus)
				logger.Errorf("attempted to retrieve the status of the massa nodeMana but got error: %w", err)
				continue
			}

			logger.Info("Bootstrap completed ! \n nodeMana is Up")

			nodeMana.setStatus(NodeStatusOn)
			/*Don't return here because a msg migth have been sent
			through the closeBootstrapMonitorChan while we were checking if node had bootstrapped.
			If we return here, closeBootstrapMonitorChan sender migth be blocked.
			This way we avoid locking with mutex all the "case <-ticker.C" logic.
			*/
		}
	}
}

/*
handleNodeStoped wait for the nodeMana process to exit.
If the process has exited with error, it handle this.
It update the status to off or error
*/
func (nodeMana *NodeManager) handleNodeStoped(cmd *exec.Cmd) {
	err := cmd.Wait() // Wait for the command to exit
	status := NodeStatusOff

	if err != nil && !isUserIntterupted(err) {
		logger.Errorf("massa nodeMana process exited with error: %v", err)
		status = NodeStatusError

		// if we are in bootstrapping phase, we need to close bootstrap monitor goroutine
		if nodeMana.status == NodeStatusBootstrapping {
			nodeMana.closeBootstrapMonitorChan <- struct{}{}
		}
	}

	nodeMana.mu.Lock()
	nodeMana.status = status
	nodeMana.statusChan <- status
	nodeMana.serverProcess = nil
	nodeMana.mu.Unlock()

	logger.Infof("massa nodeMana process exited")
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
