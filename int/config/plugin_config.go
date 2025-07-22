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
	pluginLogPath  = "pluginLogs"
	nodeLogPath    = "nodeLogs"
	dbName         = "db.sqlite"
)

type PluginConfig struct {
	PluginLogPath                  string `yaml:"plugin_log_path"`
	NodeLogPath                    string `yaml:"node_log_path"`
	NodeLogMaxSize                 int    `yaml:"log_max_size"`
	MaxLogBackups                  int    `yaml:"max_log_backups"`
	ClientTimeout                  int    `yaml:"client_timeout"`
	BootstrapCheckInterval         int    `yaml:"bootstrap_check_interval"`
	DesyncCheckInterval            int    `yaml:"desync_check_interval"`
	RestartCooldown                int    `yaml:"restart_cooldown"`
	StakingAddressDataPollInterval int    `yaml:"staking_address_data_poll_interval"`
	DBPath                         string `yaml:"db_path"`
	TotValueRegisterInterval       int    `yaml:"tot_value_register_interval"`
	TotValueDelAfter               int    `yaml:"tot_value_del_after"`
}

func defaultPluginConfig() (PluginConfig, error) {
	execPath, err := os.Executable()
	if err != nil {
		return PluginConfig{}, fmt.Errorf("failed to get executable path: %v", err)
	}
	execDir := filepath.Dir(execPath)
	return PluginConfig{
		PluginLogPath:                  filepath.Join(execDir, pluginLogPath),
		NodeLogPath:                    filepath.Join(execDir, nodeLogPath),
		NodeLogMaxSize:                 1,
		MaxLogBackups:                  10,
		ClientTimeout:                  30,
		BootstrapCheckInterval:         30, // Interval at which the node is checked if it has bootstrapped
		DesyncCheckInterval:            30, // Interval at which the node is checked if it is desynced
		RestartCooldown:                5,  // Time to wait before restarting the node
		StakingAddressDataPollInterval: 30, // Time to wait before polling the staking address data
		DBPath:                         filepath.Join(execDir, dbName),
		TotValueRegisterInterval:       180,      // 3 minutes
		TotValueDelAfter:               31536000, // 1 year
	}, nil
}

/*
RetrieveConfig retrieves the plugin configuration. If the configuration file does not exist,
it creates the file with a default configuration and returns it.
*/
func RetrieveConfig() (*PluginConfig, error) {
	configFilePath, err := getConfPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get config file path: %v", err)
	}

	// Check if the file exists
	_, err = os.Stat(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, create it with default config
			defaultConfig, err := defaultPluginConfig()
			if err != nil {
				return &PluginConfig{}, fmt.Errorf("getting default config: %w", err)
			}

			if err = saveConfig(defaultConfig, configFilePath); err != nil {
				return &PluginConfig{}, fmt.Errorf("saving default config: %w", err)
			}

			return &defaultConfig, nil
		}
		return &PluginConfig{}, fmt.Errorf("checking config file: %w", err)
	}

	// File exists, read and parse it
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return &PluginConfig{}, fmt.Errorf("reading config file: %w", err)
	}

	var config PluginConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return &PluginConfig{}, fmt.Errorf("unmarshaling config file: %w", err)
	}

	config, hasChanged, err := FillDefaultValues(config)
	if err != nil {
		return &PluginConfig{}, fmt.Errorf("filling empty plugin config with default values: %w", err)
	}

	if hasChanged {
		if err = saveConfig(config, configFilePath); err != nil {
			return &PluginConfig{}, fmt.Errorf("saving config: %w", err)
		}
	}

	return &config, nil
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

// UpdateConfigField updates a specific field in the config file
func UpdateConfigField(fieldName string, value interface{}) error {
	configFilePath, err := getConfPath()
	if err != nil {
		return fmt.Errorf("getting config file path: %w", err)
	}

	// Read current config
	config, err := readConfigFile(configFilePath)
	if err != nil {
		return fmt.Errorf("reading config file: %w", err)
	}

	// Use reflection to update the field
	configValue := reflect.ValueOf(&config).Elem()
	field := configValue.FieldByName(fieldName)
	if !field.IsValid() {
		return fmt.Errorf("field %s not found in config", fieldName)
	}

	if !field.CanSet() {
		return fmt.Errorf("field %s cannot be set", fieldName)
	}

	// Convert value to the correct type
	valueReflect := reflect.ValueOf(value)
	if field.Type() != valueReflect.Type() {
		return fmt.Errorf("type mismatch: expected %v, got %v", field.Type(), valueReflect.Type())
	}

	field.Set(valueReflect)

	// Save updated config
	return saveConfig(config, configFilePath)
}

func getConfPath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %v", err)
	}
	return filepath.Join(filepath.Dir(execPath), configFileName), nil
}

func readConfigFile(configFilePath string) (PluginConfig, error) {
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
