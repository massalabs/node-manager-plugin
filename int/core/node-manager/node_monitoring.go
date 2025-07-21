package nodeManager

import (
	"context"
	"time"

	nodeStatusPkg "github.com/massalabs/node-manager-plugin/int/core/NodeStatus"
	nodeAPI "github.com/massalabs/node-manager-plugin/int/node-api"
	"github.com/massalabs/node-manager-plugin/int/node-api/metrics"
	"github.com/massalabs/station/pkg/logger"
)

// NodeMonitoring defines the interface for monitoring node status
type NodeMonitoring interface {
	/*
		MonitorDesync fetches prometheus formated metrics from the node and check if the node is desynced
		It returns a notification channel that will send a struct{} if the node is desynced
	*/
	MonitorDesync(ctx context.Context, interval time.Duration) <-chan struct{}

	/*
		MonitorBootstrapping call the massa node api on get_status endpoint to check if the node has bootstrapped.
		It returns a notification channel that will send a struct{} if the node has bootstrapped.
	*/
	MonitorBootstrapping(ctx context.Context, interval time.Duration) <-chan struct{}
}

// NodeMonitor implements the NodeMonitoring interface
type NodeMonitor struct {
	metricsDriver    metrics.MetricsDriver
	statusDispatcher nodeStatusPkg.NodeStatusDispatcher
	nodeAPI          nodeAPI.NodeAPI
}

// NewNodeMonitor creates a new NodeMonitor instance
func NewNodeMonitor(
	metricsDriver metrics.MetricsDriver,
	statusDispatcher nodeStatusPkg.NodeStatusDispatcher,
	nodeAPI nodeAPI.NodeAPI,
) NodeMonitoring {
	return &NodeMonitor{
		metricsDriver:    metricsDriver,
		statusDispatcher: statusDispatcher,
		nodeAPI:          nodeAPI,
	}
}

/*
MonitorBootstrapping return a channel that will send a struct{} when the node has bootstrapped.
It launch a goroutine that continuously calls the massa node api on get_status endpoint to check if the node has bootstrapped.
*/
func (nm *NodeMonitor) MonitorBootstrapping(ctx context.Context, interval time.Duration) <-chan struct{} {
	bootstrappingChan := make(chan struct{})

	go func() {
		defer close(bootstrappingChan)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				logger.Debug("Stop bootstrap monitor goroutine because received cancelAsyncTask signal")
				return
			case <-ticker.C:
				/*Check if the massa node process has finished bootstrapping by sending a request to it's api
				If the request fails, it means that the node is still bootstrapping*/
				logger.Debug("Send a get_status request to the massa node to check if it has bootstrapped")
				_, err := nm.nodeAPI.GetStatus()
				if err != nil {
					if connRefused(err) {
						logger.Debug("Connection refused, the massa node is still bootstrapping")
						continue
					}
					logger.Errorf("attempted to retrieve the status of the massa node but got error: %w", err)
					continue
				}
				select {
				case bootstrappingChan <- struct{}{}:
				case <-ctx.Done():
				}
				return
			}
		}
	}()

	return bootstrappingChan
}

// MonitorDesync return a channel that will send a struct{} when the node is desynced
// It launch a goroutine that continuously fetch metrics from the node and check if the node is desynced
func (nm *NodeMonitor) MonitorDesync(ctx context.Context, interval time.Duration) <-chan struct{} {
	desyncChan := make(chan struct{})
	wasTemporaryDesynced := false
	desyncStartTime := time.Time{}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		defer close(desyncChan)

		for {
			select {
			case <-ctx.Done():
				logger.Debug("Stop desync monitor goroutine because received cancelAsyncTask signal")
				return
			case <-ticker.C:
				isTemporaryDesynced, err := nm.metricsDriver.HasDesync()
				if err != nil {
					logger.Error("failed to check desync, got error: %v", err)
					continue
				}

				if isTemporaryDesynced {
					if !wasTemporaryDesynced {
						/* If the desync formulas is true for the first time, we start the timer */
						desyncStartTime = time.Now()
						wasTemporaryDesynced = true
						/* If the desync formulas is true for more than 1 minute, we consider the node is desynced */
					} else if time.Since(desyncStartTime) > time.Minute {
						select {
						case desyncChan <- struct{}{}:
						case <-ctx.Done():
						}
						return // When a node is desynced it is definitive, we can stop the monitoring
					}
				} else if wasTemporaryDesynced {
					wasTemporaryDesynced = false // If the desync formulas is false, we reset the timer
				}
			}
		}
	}()

	return desyncChan
}
