package api

import (
	"log"
	"os"

	"github.com/go-openapi/loads"
	"github.com/massalabs/node-manager-plugin/api/restapi"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	"github.com/massalabs/node-manager-plugin/int/api/handlers"
	"github.com/massalabs/node-manager-plugin/int/api/html"
	"github.com/massalabs/node-manager-plugin/int/config"
	nodeManagerPkg "github.com/massalabs/node-manager-plugin/int/node-manager"
	"github.com/massalabs/station-massa-hello-world/pkg/plugin"
	"github.com/massalabs/station/pkg/logger"
)

type API struct {
	apiServer   *restapi.Server
	api         *operations.NodeManagerPluginAPI
	nodeManager nodeManagerPkg.INodeManager
	config      config.PluginConfig
}

// NewAPI creates a new API with the provided plugin directory
func NewAPI(config config.PluginConfig) *API {
	nodeManagerAPI, err := createAPI()
	if err != nil {
		logger.Fatalf("could not create the api, got : %s", err)
	}

	apiServer := restapi.NewServer(nodeManagerAPI)

	// create the node manager instance
	nodeManager, err := nodeManagerPkg.NewNodeManager()
	if err != nil {
		logger.Fatalf("could not create a node manager instance, got : %s", err)
	}

	return &API{
		apiServer:   apiServer,
		api:         nodeManagerAPI,
		nodeManager: nodeManager,
		config:      config,
	}
}

func (a *API) Start() {
	defer a.Cleanup()

	if os.Getenv("STANDALONE") == "1" {
		// If the plugin is run without being linked to Massa Station
		a.apiServer.Port = 8080
	} else {
		// We don't care about the port of the plugin API as MassaStation will handle the port mapping
		a.apiServer.Port = 0
	}
	a.registerHandlers()
	a.apiServer.ConfigureAPI()

	a.apiServer.SetHandler(a.api.Serve(nil))

	// Register the plugin to massa station
	listener, err := a.apiServer.HTTPListener()
	if err != nil {
		logger.Fatalf("Failed to get HTTP listener: %v", err)
	}

	logger.Info("Registering node manager plugin to Massa Station")
	if err := plugin.RegisterPlugin(listener); err != nil {
		logger.Fatalf("Failed to register plugin: %v", err)
	}

	logger.Infof("Starting node manager plugin API on port %d", a.apiServer.Port)

	// launch the plugin API
	if err := a.apiServer.Serve(); err != nil {
		logger.Fatalf("Failed to start node manager plugin: %v", err)
	}
}

// createAPI creates a new NodeManagerPluginAPI instance and loads the Swagger specification
func createAPI() (*operations.NodeManagerPluginAPI, error) {
	// Load the Swagger specification
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		return nil, err
	}

	// Create a new NodeManagerPluginAPI instance
	return operations.NewNodeManagerPluginAPI(swaggerSpec), nil
}

func (a API) registerHandlers() {
	// Set web endpoints
	html.AppendEndpoints(a.api)

	// Set API handlers
	a.api.StartNodeHandler = operations.StartNodeHandlerFunc(handlers.HandleStartNode(&a.nodeManager, a.config.Password))
	a.api.StopNodeHandler = operations.StopNodeHandlerFunc(handlers.HandleStopNode(&a.nodeManager))
	a.api.GetMassaNodeStatusHandler = operations.GetMassaNodeStatusHandlerFunc(handlers.HandleNodeStatusFeeder(&a.nodeManager))
	a.api.GetNodeLogsHandler = operations.GetNodeLogsHandlerFunc(handlers.HandleGetNodeLogs(&a.nodeManager))
}

func (a *API) Cleanup() {
	// Shutdown the server manager if it is running
	if a.nodeManager != nil && isRunning(a.nodeManager) {
		if err := a.nodeManager.StopNode(); err != nil {
			log.Fatalln(err)
		}
	}

	logger.Close()
}

// isRunning checks if the node manager is running
func isRunning(nodeManager nodeManagerPkg.INodeManager) bool {
	status, _ := nodeManager.GetStatus()
	return nodeManagerPkg.IsRunning(status)
}
