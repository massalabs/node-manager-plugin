package handlers

import (
	"net/http"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	nodeManager "github.com/massalabs/node-manager-plugin/int/node-manager"
)

func HandleNodeStatusFeeder(nodeManager *nodeManager.INodeManager) func(operations.GetMassaNodeStatusParams) middleware.Responder {
	return func(params operations.GetMassaNodeStatusParams) middleware.Responder {
		return middleware.ResponderFunc(
			func(w http.ResponseWriter, _ runtime.Producer) {
				// Set SSE headers
				w.Header().Set("Content-Type", "text/event-stream")
				w.Header().Set("Cache-Control", "no-cache")
				w.Header().Set("Connection", "keep-alive")

				// retrieve the current status and the channel to listen for updates
				currentStatus, statusChan := (*nodeManager).GetStatus()
				flush(w, currentStatus)

				for status := range statusChan { // receive values until the statusChan channel is closed
					// If severals statues are in the buffered channel, we send only the last one.
					if len(statusChan) > 0 {
						continue
					}
					// Send the new status to the client
					flush(w, status)
				}
			},
		)
	}
}

func flush(w http.ResponseWriter, status nodeManager.NodeStatus) {
	_, err := w.Write([]byte(status))
	if err != nil {
		panic(err)
	}
	w.(http.Flusher).Flush()
}
