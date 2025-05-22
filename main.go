package main

import (
	_ "embed"
	"log"
	"path/filepath"

	"github.com/massalabs/node-manager-plugin/int/api"
	"github.com/massalabs/node-manager-plugin/int/config"
	"github.com/massalabs/station/pkg/logger"
)

func main() {
	pluginDir, err := config.PluginDir()
	if err != nil {
		log.Fatalf("failed to get plugin directory: %v", err)
	}
	logPath := filepath.Join(pluginDir, "./node-manager-plugin.log")

	err = logger.InitializeGlobal(logPath)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	logger.Infof("retrieving node manager plugin configuration")
	pluginConfig, err := config.RetrieveConfig()
	if err != nil {
		log.Fatalf("failed to load node manager configuration : %v", err)
	}

	// Create and start the API with the plugin directory
	nodePlugin := api.NewAPI(pluginConfig)
	nodePlugin.Start()

	logger.Warnf("node manager plugin stopped")
}
