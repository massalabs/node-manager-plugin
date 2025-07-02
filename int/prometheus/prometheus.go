package prometheus

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

// driver to interact with prometheus
type PrometheusDriver interface {
	HasDesync() (bool, error)
}

// Prometheus implements the PrometheusDriver interface
type Prometheus struct {
	client         *http.Client
	metricsIndexes map[string]int
}

// NewPrometheus creates a new Prometheus driver
func NewPrometheus() PrometheusDriver {
	return &Prometheus{
		client:         &http.Client{Timeout: 10 * time.Second},
		metricsIndexes: make(map[string]int),
	}
}

// HasDesync checks if the node is desynced or not
func (p *Prometheus) HasDesync() (bool, error) {
	prometheusData, err := p.getPrometheusMetrics()
	if err != nil {
		return false, fmt.Errorf("failed to get prometheus metrics: %w", err)
	}
	return p.checkDesync(prometheusData)
}

// getPrometheusMetrics fetches the prometheus metrics from the node
func (p *Prometheus) getPrometheusMetrics() ([]byte, error) {
	resp, err := p.client.Get("http://localhost:31248/metrics")
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
func (p *Prometheus) checkDesync(prometheusData []byte) (bool, error) {
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
findValueInPrometheusData finds the value of a metric in prometheus data.
Prometheus data should be provided as a []string corresponding to the list of lines.
It returns the value and true if the metric is found, otherwise returns an empty string and false
*/
func (p *Prometheus) findValueInPrometheusData(prometheusDataLines []string, metric string) (string, bool) {
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
