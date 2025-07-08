package config

import "sync"

// PluginInfo contains runtime plugin information
type PluginInfo struct {
	PwdMainnet  string `json:"pwd_mainnet"`
	PwdBuildnet string `json:"pwd_buildnet"`
	AutoRestart bool   `json:"auto_restart"`
	IsMainnet   bool   `json:"is_mainnet"`
	mu          sync.RWMutex
}

// Global instance of PluginInfo
var GlobalPluginInfo *PluginInfo

// init function to initialize the global PluginInfo instance
func init() {
	GlobalPluginInfo = &PluginInfo{
		PwdMainnet:  "",
		PwdBuildnet: "",
		AutoRestart: false,
		IsMainnet:   true,
	}
}

func (pi *PluginInfo) GetPwdByNetwork(isMainnet bool) string {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	if isMainnet {
		return pi.PwdMainnet
	}
	return pi.PwdBuildnet
}

// GetPwd returns the current password
func (pi *PluginInfo) GetPwd() string {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	if pi.IsMainnet {
		return pi.PwdMainnet
	}
	return pi.PwdBuildnet
}

// SetPwd sets the password
func (pi *PluginInfo) SetPwd(pwd string) {
	pi.mu.Lock()
	defer pi.mu.Unlock()
	if pi.IsMainnet {
		pi.PwdMainnet = pwd
	} else {
		pi.PwdBuildnet = pwd
	}
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
