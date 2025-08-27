package main

import (
	_ "embed"
	"log"
	"path/filepath"

	"github.com/massalabs/node-manager-plugin/int/api"
	"github.com/massalabs/node-manager-plugin/int/config"
	"github.com/massalabs/node-manager-plugin/int/utils"
	"github.com/massalabs/station/pkg/logger"
)

const pluginLogPath = "pluginLogs"

func main() {
	execDir, err := utils.GetExecDirPath()
	if err != nil {
		log.Fatalf("failed to get executable directory path: %v", err)
	}
	logPath := filepath.Join(execDir, pluginLogPath, "node-manager-plugin.log")

	err = logger.InitializeGlobal(logPath)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	pluginConfig, err := config.RetrieveConfig()
	if err != nil {
		logger.Fatalf("failed to load node manager configuration : %v", err)
	}

	// Create and start the API with the plugin directory
	nodePlugin := api.NewAPI(pluginConfig)
	nodePlugin.Start()

	logger.Warnf("node manager plugin stopped")
}
