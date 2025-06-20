package api

import (
	"os"
	"os/signal"
	"syscall"

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
	nodeManager, err := nodeManagerPkg.NewNodeManager(config)
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

	// Gracefuly shutdown the node manager plugin on SIGTERM and SIGINT
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		// launch the plugin API
		if err := a.apiServer.Serve(); err != nil {
			logger.Fatalf("Failed to start node manager plugin: %v", err)
		}
	}()

	sig := <-sigChan
	logger.Debugf("Node manager plugin received closing signal: %v", sig)
	a.Cleanup()
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
	a.api.StartNodeHandler = operations.StartNodeHandlerFunc(handlers.HandleStartNode(&a.nodeManager))
	a.api.StopNodeHandler = operations.StopNodeHandlerFunc(handlers.HandleStopNode(&a.nodeManager))
	a.api.GetMassaNodeStatusHandler = operations.GetMassaNodeStatusHandlerFunc(handlers.HandleNodeStatusFeeder(&a.nodeManager))
	a.api.GetNodeLogsHandler = operations.GetNodeLogsHandlerFunc(handlers.HandleGetNodeLogs(&a.nodeManager))
	a.api.SetAutoRestartHandler = operations.SetAutoRestartHandlerFunc(handlers.HandleSetAutoRestart(&a.nodeManager))
	a.api.GetNodeInfosHandler = operations.GetNodeInfosHandlerFunc(handlers.HandleGetNodeInfos(&a.nodeManager))
}

func (a *API) Cleanup() {
	if err := a.nodeManager.Close(); err != nil {
		logger.Errorf("Failed to cleanup node manager: %v", err)
	}

	logger.Debug("Closing plugin logger")
	logger.Close()
}
