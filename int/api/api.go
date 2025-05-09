package api

import (
	"log"

	"github.com/go-openapi/loads"
	"github.com/massalabs/node-manager-plugin/api/restapi"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	"github.com/massalabs/node-manager-plugin/int/config"
)

type API struct {
	apiServer      *restapi.Server
	nodeManagerAPI *operations.NodeManagerPluginAPI
	config         config.PluginConfig
}

// NewAPI creates a new API with the provided plugin directory
func NewAPI(config config.PluginConfig) *API {
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	nodeManagerAPI := operations.NewNodeManagerPluginAPI(swaggerSpec)
	apiServer := restapi.NewServer(nodeManagerAPI)

	// manager, err := server.NewServerManager(configDir)
	// if err != nil {
	// 	logger.Errorf("Failed to create server manager: %v", err)
	// }

	return &API{
		apiServer:      apiServer,
		nodeManagerAPI: nodeManagerAPI,
		config:         config,
	}
}
