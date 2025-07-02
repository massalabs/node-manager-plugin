package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	nodeStatus "github.com/massalabs/node-manager-plugin/int/NodeStatus"
	"github.com/massalabs/station/pkg/logger"
)

const (
	flushCooldown = 5 * time.Second
)

func HandleNodeStatusFeeder(statusDispatcher nodeStatus.NodeStatusDispatcher) func(operations.GetMassaNodeStatusParams) middleware.Responder {
	return func(params operations.GetMassaNodeStatusParams) middleware.Responder {
		return middleware.ResponderFunc(
			func(w http.ResponseWriter, _ runtime.Producer) {
				logger.Info("Call GET api/status")
				flusher, ok := w.(http.Flusher)
				if !ok {
					logger.Error("ResponseWriter does not implement http.Flusher, cannot handle SSE")
					http.Error(w, "ResponseWriter does not implement http.Flusher, cannot handle SSE", http.StatusInternalServerError)
					return
				}

				// Set SSE headers
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")
				w.Header().Set("Access-Control-Allow-Origin", "*")

				// subscribe to all status changes
				currentStatus := statusDispatcher.GetCurrentStatus()
				statusChan, unsubscribe := statusDispatcher.SubscribeAll("status-Server-Side-Event-feeder")
				defer unsubscribe() // Ensure cleanup

				flush(w, flusher, currentStatus)

				lastFlushTime := time.Now()
				for {
					select {
					case <-params.HTTPRequest.Context().Done():
						logger.Debug("SSE connection closed")
						return
					case status, ok := <-statusChan:
						if !ok {
							logger.Debug("Status channel closed")
							return
						}

						// If severals statuses are in the buffered channel, we send only the last one.
						if len(statusChan) > 0 {
							logger.Debugf("Skipping status update, buffer size: %d", len(statusChan))
							continue
						}

						// Ensure at least flushCooldown has passed since the last flush
						timeSinceLastFlush := time.Since(lastFlushTime)
						if timeSinceLastFlush < flushCooldown {
							logger.Debugf("Waiting %v before next flush", flushCooldown-timeSinceLastFlush)
							time.Sleep(flushCooldown - timeSinceLastFlush)
						}

						// Send the new status to the client
						logger.Infof("Sending status update: %s", status)
						flush(w, flusher, status)
						lastFlushTime = time.Now()
					}
				}
			},
		)
	}
}

func flush(w http.ResponseWriter, flusher http.Flusher, status nodeStatus.NodeStatus) {
	_, err := fmt.Fprintf(w, "data: %s\n\n", status)
	if err != nil {
		logger.Errorf("Failed to flush status %s, got error: %v", status, err)
		return
	}
	flusher.Flush()
}
