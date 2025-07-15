package api

import (
	"os"

	"github.com/go-openapi/loads"
	"github.com/massalabs/node-manager-plugin/api/restapi"
	"github.com/massalabs/node-manager-plugin/api/restapi/operations"
	nodeStatusPkg "github.com/massalabs/node-manager-plugin/int/NodeStatus"
	"github.com/massalabs/node-manager-plugin/int/api/handlers"
	"github.com/massalabs/node-manager-plugin/int/api/html"
	"github.com/massalabs/node-manager-plugin/int/config"
	"github.com/massalabs/node-manager-plugin/int/db"
	historymanager "github.com/massalabs/node-manager-plugin/int/history-manager"
	nodeAPI "github.com/massalabs/node-manager-plugin/int/node-api"
	nodeDirManager "github.com/massalabs/node-manager-plugin/int/node-bin-dir-manager"
	nodeDriverPkg "github.com/massalabs/node-manager-plugin/int/node-driver"
	nodeManagerPkg "github.com/massalabs/node-manager-plugin/int/node-manager"
	prometheusPkg "github.com/massalabs/node-manager-plugin/int/prometheus"
	stakingManagerPkg "github.com/massalabs/node-manager-plugin/int/staking-manager"
	"github.com/massalabs/station-massa-hello-world/pkg/plugin"
	"github.com/massalabs/station/pkg/logger"
)

const (
	NodeURL = "http://localhost:33035"
)

type API struct {
	apiServer        *restapi.Server
	api              *operations.NodeManagerPluginAPI
	nodeManager      nodeManagerPkg.INodeManager
	nodeDirManager   nodeDirManager.NodeDirManager
	statusDispatcher nodeStatusPkg.NodeStatusDispatcher
	config           *config.PluginConfig
	stakingManager   stakingManagerPkg.StakingManager
	db               db.DB
	historyMgr       *historymanager.HistoryManager
}

// NewAPI creates a new API with the provided plugin directory
func NewAPI(config *config.PluginConfig) *API {
	nodeManagerAPI, err := createAPI()
	if err != nil {
		logger.Fatalf("could not create the api, got : %s", err)
	}

	apiServer := restapi.NewServer(nodeManagerAPI)

	nodeDirManager, err := nodeDirManager.NewNodeDirManager()
	if err != nil {
		logger.Fatalf("could not create a node dir manager instance, got : %s", err)
	}

	prometheusDriver := prometheusPkg.NewPrometheus()
	statusDispatcher := nodeStatusPkg.NewNodeStatusDispatcher()
	nodeAPI := nodeAPI.NewNodeAPI()
	nodeMonitor := nodeManagerPkg.NewNodeMonitor(prometheusDriver, statusDispatcher, nodeAPI)

	nodeDriver := nodeDriverPkg.NewNodeDriver(nodeDirManager)
	db, err := db.NewDB(config.DBPath)
	if err != nil {
		logger.Fatalf("could not create a database instance, got : %s", err)
	}

	historyMgr := historymanager.NewHistoryManager(db, int64(config.TotValueDelAfter), int64(config.TotValueRegisterInterval))

	stakingManager := stakingManagerPkg.NewStakingManager(
		nodeAPI,
		statusDispatcher,
		db,
		nodeDirManager,
		uint64(config.StakingAddressDataPollInterval),
		uint64(config.ClientTimeout),
		stakingManagerPkg.NewMassaWalletManager(),
		config,
	)

	// create the node manager instance
	nodeManager, err := nodeManagerPkg.NewNodeManager(
		config,
		nodeDirManager,
		nodeMonitor,
		nodeDriver,
		statusDispatcher,
	)
	if err != nil {
		logger.Fatalf("could not create a node manager instance, got : %s", err)
	}

	return &API{
		apiServer:        apiServer,
		api:              nodeManagerAPI,
		nodeManager:      nodeManager,
		nodeDirManager:   nodeDirManager,
		statusDispatcher: statusDispatcher,
		config:           config,
		stakingManager:   stakingManager,
		db:               db,
		historyMgr:       historyMgr,
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
	a.api.StartNodeHandler = operations.StartNodeHandlerFunc(handlers.HandleStartNode(a.nodeManager, a.statusDispatcher, &a.nodeDirManager, a.config))
	a.api.StopNodeHandler = operations.StopNodeHandlerFunc(handlers.HandleStopNode(a.nodeManager, a.statusDispatcher))
	a.api.GetMassaNodeStatusHandler = operations.GetMassaNodeStatusHandlerFunc(handlers.HandleNodeStatusFeeder(a.statusDispatcher))
	a.api.GetNodeLogsHandler = operations.GetNodeLogsHandlerFunc(handlers.HandleGetNodeLogs(&a.nodeManager))
	a.api.SetAutoRestartHandler = operations.SetAutoRestartHandlerFunc(handlers.HandleSetAutoRestart())
	a.api.GetPluginInfosHandler = operations.GetPluginInfosHandlerFunc(handlers.HandleGetPluginInfos(&a.nodeDirManager))

	a.api.GetStakingAddressesHandler = operations.GetStakingAddressesHandlerFunc(handlers.HandleAddressChangedFeeder(a.stakingManager))
	a.api.AddStakingAddressHandler = operations.AddStakingAddressHandlerFunc(handlers.HandlePostStakingAddresses(a.stakingManager))
	a.api.UpdateStakingAddressHandler = operations.UpdateStakingAddressHandlerFunc(handlers.HandlePutStakingAddresses(a.stakingManager))
	a.api.RemoveStakingAddressHandler = operations.RemoveStakingAddressHandlerFunc(handlers.HandleDeleteStakingAddresses(a.stakingManager))
	a.api.GetValueHistoryHandler = operations.GetValueHistoryHandlerFunc(handlers.HandleGetValueHistory(a.db, a.historyMgr, a.config))
}

func (a *API) Cleanup() {
	if err := a.nodeManager.Close(); err != nil {
		logger.Errorf("Failed to cleanup node manager: %v", err)
	}

	if err := a.stakingManager.Close(); err != nil {
		logger.Errorf("Failed to cleanup staking manager: %v", err)
	}

	logger.Debug("Closing plugin logger")
	logger.Close()
}
