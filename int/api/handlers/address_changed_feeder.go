package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	configPkg "github.com/massalabs/node-manager-plugin/int/config"
	stakingManagerPkg "github.com/massalabs/node-manager-plugin/int/staking-manager"
	"github.com/massalabs/station/pkg/logger"
)

// TODO: Replace with actual operation when GetStakingAddressesParams is created
func HandleAddressChangedFeeder(stakingManager stakingManagerPkg.StakingManager) func(operations.GetStakingAddressesParams) middleware.Responder {
	return func(params operations.GetStakingAddressesParams) middleware.Responder {
		return middleware.ResponderFunc(
			func(w http.ResponseWriter, _ runtime.Producer) {
				logger.Info("Call GET api/stakingAddresses")
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

				// Get current staking addresses
				currentAddresses, addressDispatcher, err := stakingManager.GetStakingAddresses(configPkg.GlobalPluginInfo.GetPwd())
				if err != nil {
					logger.Errorf("Failed to get current staking addresses: %v", err)
					http.Error(w, "Failed to get staking addresses", http.StatusInternalServerError)
					return
				}

				// Subscribe to address changes
				addressChan, unsubscribe := addressDispatcher.Subscribe("address-Server-Side-Event-feeder")
				defer unsubscribe() // Ensure cleanup

				flushAddresses(w, flusher, currentAddresses)

				for {
					select {
					case <-params.HTTPRequest.Context().Done():
						logger.Debug("SSE connection closed")
						return
					case addresses, ok := <-addressChan:
						if !ok {
							logger.Debug("Address channel closed")
							return
						}

						// If several address updates are in the buffered channel, we send only the last one.
						if len(addressChan) > 0 {
							logger.Debugf("Skipping address update, buffer size: %d", len(addressChan))
							continue
						}

						// Send the new addresses to the client
						logger.Infof("Sending address update: %d addresses", len(addresses))
						flushAddresses(w, flusher, addresses)
					}
				}
			},
		)
	}
}

func flushAddresses(w http.ResponseWriter, flusher http.Flusher, addresses []stakingManagerPkg.StakingAddress) {
	// Convert addresses to JSON
	addressesJSON, err := json.Marshal(addresses)
	if err != nil {
		logger.Errorf("Failed to marshal addresses to JSON: %v", err)
		return
	}

	_, err = fmt.Fprintf(w, "data: %s\n\n", string(addressesJSON))
	if err != nil {
		logger.Errorf("Failed to flush addresses, got error: %v", err)
		return
	}
	flusher.Flush()
}
