package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

const (
	directoryName  = "station-node-manager-plugin"
	configFileName = "node_manager_config.yaml"
)

type PluginConfig struct {
	NodeLogPath            string `yaml:"log_path"`
	NodeLogMaxSize         int    `yaml:"log_max_size"`
	MaxLogBackups          int    `yaml:"max_log_backups"`
	ClientTimeout          int    `yaml:"client_timeout"`
	BootstrapCheckInterval int    `yaml:"bootstrap_check_interval"`
	DesyncCheckInterval    int    `yaml:"desync_check_interval"`
	RestartCooldown        int    `yaml:"restart_cooldown"`
}

func defaultPluginConfig() (PluginConfig, error) {
	execPath, err := os.Executable()
	if err != nil {
		return PluginConfig{}, fmt.Errorf("failed to get executable path: %v", err)
	}
	return PluginConfig{
		NodeLogPath:            filepath.Join(filepath.Dir(execPath), "./nodeLogs"),
		NodeLogMaxSize:         1,
		MaxLogBackups:          10,
		ClientTimeout:          30,
		BootstrapCheckInterval: 30, // Interval at which the node is checked if it has bootstrapped
		DesyncCheckInterval:    30, // Interval at which the node is checked if it is desynced
		RestartCooldown:        5,  // Time to wait before restarting the node
	}, nil
}

/*
Retrieve the directory in which the node manager plugin config is stored
If the folder doesn't exist, it is created.
*/
func PluginDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("getting user config directory: %w", err)
	}

	path := filepath.Join(configDir, directoryName)

	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(path, os.ModePerm)
			if err != nil {
				return "", fmt.Errorf("creating node manager plugin configuration directory '%s': %w", path, err)
			}
		} else {
			return "", fmt.Errorf("checking directory '%s': %w", path, err)
		}
	}

	return path, nil
}

/*
RetrieveConfig retrieves the plugin configuration. If the configuration file does not exist,
it creates the file with a default configuration and returns it.
*/
func RetrieveConfig() (PluginConfig, error) {
	pluginDir, err := PluginDir()
	if err != nil {
		return PluginConfig{}, fmt.Errorf("getting plugin directory: %w", err)
	}

	configFilePath := filepath.Join(pluginDir, configFileName)

	// Check if the file exists
	_, err = os.Stat(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, create it with default config
			defaultConfig, err := defaultPluginConfig()
			if err != nil {
				return PluginConfig{}, fmt.Errorf("getting default config: %w", err)
			}

			data, err := yaml.Marshal(defaultConfig)
			if err != nil {
				return PluginConfig{}, fmt.Errorf("marshaling default config to YAML: %w", err)
			}

			err = os.WriteFile(configFilePath, data, 0o644)
			if err != nil {
				return PluginConfig{}, fmt.Errorf("writing default config to file: %w", err)
			}

			return defaultConfig, nil
		}
		return PluginConfig{}, fmt.Errorf("checking config file: %w", err)
	}

	// File exists, read and parse it
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return PluginConfig{}, fmt.Errorf("reading config file: %w", err)
	}

	var config PluginConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return PluginConfig{}, fmt.Errorf("unmarshaling config file: %w", err)
	}

	return config, nil
}
