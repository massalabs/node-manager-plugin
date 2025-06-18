package nodeManager

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/massalabs/station/pkg/logger"
	"github.com/massalabs/station/pkg/node"
)

const (
	activeCursorMetric = "active_cursor_period"
	finalCursorMetric  = "final_cursor_period"
)

// NodeMonitoring defines the interface for monitoring node status
type NodeMonitoring interface {
	/*
		MonitorDesync fetches prometheus metrics from the node and check if the node is desynced
		It returns a notification channel that will send a struct{} if the node is desynced
	*/
	MonitorDesync(ctx context.Context, interval time.Duration) <-chan struct{}
}

// NodeMonitor implements the NodeMonitoring interface
type NodeMonitor struct {
	client         *http.Client
	metricsIndexes map[string]int
}

// NewNodeMonitor creates a new NodeMonitor instance
func NewNodeMonitor() *NodeMonitor {
	return &NodeMonitor{
		client:         &http.Client{Timeout: 10 * time.Second},
		metricsIndexes: make(map[string]int),
	}
}

func (nm *NodeMonitor) MonitorBootstrapping(ctx context.Context, interval time.Duration) <-chan struct{} {
	bootstrappingChan := make(chan struct{})

	go func() {
		defer close(bootstrappingChan)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		client := node.NewClient(nodeURL)
		for {
			select {
			case <-ctx.Done():
				logger.Debug("Stop bootstrap monitor goroutine because received cancelAsyncTask signal")
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
				select {
				case bootstrappingChan <- struct{}{}:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return bootstrappingChan
}

// MonitorDesync fetch prometheus metrics from the node and check if the node is desynced
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
				return
			case <-ticker.C:
				data, err := nm.getPrometheusMetrics()
				if err != nil {
					logger.Error("failed to fetch prometheus metrics for desync check, got error: %v", err)
					continue
				}

				isTemporaryDesynced, err := nm.checkDesync(data)
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

// getPrometheusMetrics fetches the prometheus metrics from the node
func (nm *NodeMonitor) getPrometheusMetrics() ([]byte, error) {
	resp, err := nm.client.Get("http://localhost:31248/metrics")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("metrics endpoint returned status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return body, nil
}

/*
checkDesync checks if the node is desynced by comparing current_cursor_period and final_cursor_period
Prometheus data is parsing is custom because the desync formula is simple.
But if more complex task are required, using a prometheus library may be a better option.
*/
func (nm *NodeMonitor) checkDesync(prometheusData []byte) (bool, error) {
	lines := strings.Split(string(prometheusData), "\n")

	activeCursorStr, found := nm.findValueInPrometheusData(lines, activeCursorMetric)
	if !found {
		return false, fmt.Errorf("failed to find %s metric in prometheus data", activeCursorMetric)
	}

	activeCursor, err := strconv.Atoi(activeCursorStr)
	if err != nil {
		return false, fmt.Errorf("failed to convert %s string value %s to int: %w", activeCursorMetric, activeCursorStr, err)
	}

	finalCursorStr, found := nm.findValueInPrometheusData(lines, finalCursorMetric)
	if !found {
		return false, fmt.Errorf("failed to find %s metric in prometheus data", finalCursorMetric)
	}

	finalCursor, err := strconv.Atoi(finalCursorStr)
	if err != nil {
		return false, fmt.Errorf("failed to convert %s string value %s to int: %w", finalCursorMetric, finalCursorStr, err)
	}

	return activeCursor-finalCursor > 10, nil
}

/*
and parses
Finds the value of a metric in prometheus data.
Prometheus data should be provided a []string corresponding to the list of lines.
Returns the value and true if the metric is found, otherwise returns an empty string and false
*/
func (nm *NodeMonitor) findValueInPrometheusData(prometheusDataLines []string, metric string) (string, bool) {
	// If the metric line index is found in the map, we can use it
	if index, ok := nm.metricsIndexes[metric]; ok {
		if strings.HasPrefix(prometheusDataLines[index], metric) {
			return strings.Fields(prometheusDataLines[index])[1], true
		}
	}

	// If the metric line index is not found in the map or if it has changed, we need to find it
	for index, line := range prometheusDataLines {
		if strings.HasPrefix(line, metric) {
			nm.metricsIndexes[metric] = index // store the index of the metric in the map
			return strings.Fields(prometheusDataLines[index])[1], true
		}
	}

	return "", false
}
