package metrics

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	activeCursorMetric = "active_cursor_period"
	finalCursorMetric  = "final_cursor_period"
)

// driver to interact with node's returned metrics
type MetricsDriver interface {
	HasDesync() (bool, error)
}

// Metrics implements the MetricsDriver interface
type Metrics struct {
	client         *http.Client
	metricsIndexes map[string]int
}

// NewMetrics creates a new Metrics driver
func NewMetrics() MetricsDriver {
	return &Metrics{
		client:         &http.Client{Timeout: 10 * time.Second},
		metricsIndexes: make(map[string]int),
	}
}

// HasDesync checks if the node is desynced or not
func (p *Metrics) HasDesync() (bool, error) {
	prometheusData, err := p.getPrometheusMetrics()
	if err != nil {
		return false, fmt.Errorf("failed to get prometheus metrics: %w", err)
	}
	return p.checkDesync(prometheusData)
}

// getPrometheusMetrics fetches the prometheus metrics from the node's metrics endpoint
func (p *Metrics) getPrometheusMetrics() ([]byte, error) {
	resp, err := p.client.Get("http://localhost:31248/metrics")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metrics: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			fmt.Printf("warning: failed to close metrics response body: %v\n", cerr)
		}
	}()

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
*/
func (p *Metrics) checkDesync(prometheusData []byte) (bool, error) {
	lines := strings.Split(string(prometheusData), "\n")

	activeCursorStr, found := p.findValueInPrometheusData(lines, activeCursorMetric)
	if !found {
		return false, fmt.Errorf("failed to find %s metric in prometheus data", activeCursorMetric)
	}

	activeCursor, err := strconv.Atoi(activeCursorStr)
	if err != nil {
		return false, fmt.Errorf("failed to convert %s string value %s to int: %w", activeCursorMetric, activeCursorStr, err)
	}

	finalCursorStr, found := p.findValueInPrometheusData(lines, finalCursorMetric)
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
findValueInPrometheusData finds the value of a metric in prometheus formated data.
Prometheus data should be provided as a []string corresponding to the list of lines.
It returns the value and true if the metric is found, otherwise returns an empty string and false
*/
func (p *Metrics) findValueInPrometheusData(prometheusDataLines []string, metric string) (string, bool) {
	// If the metric line index is found in the map, we can use it
	if index, ok := p.metricsIndexes[metric]; ok {
		if strings.HasPrefix(prometheusDataLines[index], metric) {
			return strings.Fields(prometheusDataLines[index])[1], true
		}
	}

	// If the metric line index is not found in the map or if it has changed, we need to find it
	for index, line := range prometheusDataLines {
		if strings.HasPrefix(line, metric) {
			p.metricsIndexes[metric] = index // store the index of the metric in the map
			return strings.Fields(prometheusDataLines[index])[1], true
		}
	}

	return "", false
}
