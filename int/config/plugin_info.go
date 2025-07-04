package config

import "sync"

// PluginInfo contains runtime plugin information
type PluginInfo struct {
	Pwd         string `json:"pwd"`
	AutoRestart bool   `json:"auto_restart"`
	IsMainnet   bool   `json:"is_mainnet"`
	mu          sync.RWMutex
}

// Global instance of PluginInfo
var GlobalPluginInfo *PluginInfo

// init function to initialize the global PluginInfo instance
func init() {
	GlobalPluginInfo = &PluginInfo{
		Pwd:         "",
		AutoRestart: false,
		IsMainnet:   false,
	}
}

// GetPwd returns the current password
func (pi *PluginInfo) GetPwd() string {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	return pi.Pwd
}

// SetPwd sets the password
func (pi *PluginInfo) SetPwd(pwd string) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	pi.Pwd = pwd
}

// GetAutoRestart returns the current auto restart setting
func (pi *PluginInfo) GetAutoRestart() bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	return pi.AutoRestart
}

// SetAutoRestart sets the auto restart setting
func (pi *PluginInfo) SetAutoRestart(autoRestart bool) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	pi.AutoRestart = autoRestart
}

// GetIsMainnet returns the current mainnet setting
func (pi *PluginInfo) GetIsMainnet() bool {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	return pi.IsMainnet
}

// SetIsMainnet sets the mainnet setting
func (pi *PluginInfo) SetIsMainnet(isMainnet bool) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	pi.IsMainnet = isMainnet
}
