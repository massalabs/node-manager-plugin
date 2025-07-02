package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"gopkg.in/yaml.v2"
)

const (
	directoryName  = "station-node-manager-plugin"
	configFileName = "node_manager_config.yaml"
)

type PluginConfig struct {
	NodeLogPath                    string `yaml:"log_path"`
	NodeLogMaxSize                 int    `yaml:"log_max_size"`
	MaxLogBackups                  int    `yaml:"max_log_backups"`
	ClientTimeout                  int    `yaml:"client_timeout"`
	BootstrapCheckInterval         int    `yaml:"bootstrap_check_interval"`
	DesyncCheckInterval            int    `yaml:"desync_check_interval"`
	RestartCooldown                int    `yaml:"restart_cooldown"`
	StakingAddressDataPollInterval int    `yaml:"staking_address_data_poll_interval"`
	NodeStatusPollInterval         int    `yaml:"node_status_poll_interval"`
	DBPath                         string `yaml:"db_path"`
}

func defaultPluginConfig() (PluginConfig, error) {
	execPath, err := os.Executable()
	if err != nil {
		return PluginConfig{}, fmt.Errorf("failed to get executable path: %v", err)
	}
	return PluginConfig{
		NodeLogPath:                    filepath.Join(filepath.Dir(execPath), "./nodeLogs"),
		NodeLogMaxSize:                 1,
		MaxLogBackups:                  10,
		ClientTimeout:                  30,
		BootstrapCheckInterval:         30,   // Interval at which the node is checked if it has bootstrapped
		DesyncCheckInterval:            30,   // Interval at which the node is checked if it is desynced
		RestartCooldown:                5,    // Time to wait before restarting the node
		StakingAddressDataPollInterval: 30,   // Time to wait before polling the staking address data
		NodeStatusPollInterval:         1800, // Time to wait before polling the node status (30 minutes in seconds)
		DBPath:                         filepath.Join(filepath.Dir(execPath), "./db.sqlite"),
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

			if err = saveConfig(defaultConfig, configFilePath); err != nil {
				return PluginConfig{}, fmt.Errorf("saving default config: %w", err)
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

	config, hasChanged, err := FillDefaultValues(config)
	if err != nil {
		return PluginConfig{}, fmt.Errorf("filling empty plugin config with default values: %w", err)
	}

	if hasChanged {
		if err = saveConfig(config, configFilePath); err != nil {
			return PluginConfig{}, fmt.Errorf("saving config: %w", err)
		}
	}

	return config, nil
}

/*
FillDefaultValues takes a PluginConfig instance and replaces any zero values with the corresponding
default values from defaultPluginConfig(). Fields that already have non-zero values are kept unchanged.
*/
func FillDefaultValues(config PluginConfig) (PluginConfig, bool, error) {
	defaultConfig, err := defaultPluginConfig()
	if err != nil {
		return PluginConfig{}, false, fmt.Errorf("getting default config: %w", err)
	}

	hasChanged := false

	// Use reflection to iterate through all fields
	configValue := reflect.ValueOf(&config).Elem()
	defaultValue := reflect.ValueOf(defaultConfig)

	for i := 0; i < configValue.NumField(); i++ {
		configField := configValue.Field(i)
		defaultField := defaultValue.Field(i)

		// Check if the field has a zero value
		if configField.IsZero() {
			// Replace with default value
			configField.Set(defaultField)
			hasChanged = true
		}
	}

	return config, hasChanged, nil
}

func saveConfig(config PluginConfig, configFilePath string) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshaling config to YAML: %w", err)
	}

	err = os.WriteFile(configFilePath, data, 0o644)
	if err != nil {
		return fmt.Errorf("writing config to file: %w", err)
	}

	return nil
}
