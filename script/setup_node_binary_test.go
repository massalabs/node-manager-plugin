package main

import (
	"os"
	"path/filepath"
	"testing"
)

const massaVersion = "MAIN.2.5"

func getDownloadURL(os, arch string) string {
	switch os {
	case "linux":
		if arch == "arm64" {
			return "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_linux_arm64.tar.gz"
		}
		return "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_linux.tar.gz"
	case "darwin":
		if arch == "arm64" {
			return "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_macos_aarch64.tar.gz"
		}
		return "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_macos.tar.gz"
	case "windows":
		return "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_windows.zip"
	default:
		return ""
	}
}

func TestDownloadURLGeneration(t *testing.T) {
	tests := []struct {
		os     string
		arch   string
		want   string
	}{
		{
			os:   "linux",
			arch: "amd64",
			want: "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_linux.tar.gz",
		},
		{
			os:   "linux",
			arch: "arm64",
			want: "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_linux_arm64.tar.gz",
		},
		{
			os:   "darwin",
			arch: "amd64",
			want: "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_macos.tar.gz",
		},
		{
			os:   "darwin",
			arch: "arm64",
			want: "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_macos_aarch64.tar.gz",
		},
		{
			os:   "windows",
			arch: "amd64",
			want: "https://github.com/massalabs/massa/releases/download/" + massaVersion + "/massa_" + massaVersion + "_release_windows.zip",
		},
	}

	for _, tt := range tests {
		t.Run(tt.os+"_"+tt.arch, func(t *testing.T) {
			got := getDownloadURL(tt.os, tt.arch)
			if got != tt.want {
				t.Errorf("getDownloadURL(%s, %s) = %v, want %v", tt.os, tt.arch, got, tt.want)
			}
		})
	}
}

func TestSetupNodeDirectories(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "massa-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test directory structure
	nodeMassaDir := filepath.Join(tempDir, "node-massa")
	if err := os.MkdirAll(nodeMassaDir, 0755); err != nil {
		t.Fatalf("Failed to create node-massa dir: %v", err)
	}

	// Create massa-node directory with some test files
	nodeDir := filepath.Join(nodeMassaDir, "massa-node")
	if err := os.MkdirAll(nodeDir, 0755); err != nil {
		t.Fatalf("Failed to create massa-node dir: %v", err)
	}

	// Create a test file in massa-node
	testFile := filepath.Join(nodeDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create massa-client directory (should be removed)
	clientDir := filepath.Join(nodeMassaDir, "massa-client")
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		t.Fatalf("Failed to create massa-client dir: %v", err)
	}

	// Run setupNodeDirectories
	if err := setupNodeDirectories(nodeMassaDir); err != nil {
		t.Fatalf("setupNodeDirectories failed: %v", err)
	}

	// Verify the structure
	tests := []struct {
		path string
		want bool
	}{
		{filepath.Join(nodeMassaDir, "mainnet"), true},
		{filepath.Join(nodeMassaDir, "buildnet"), true},
		{filepath.Join(nodeMassaDir, "massa-client"), false},
		{filepath.Join(nodeMassaDir, "mainnet", "test.txt"), true},
		{filepath.Join(nodeMassaDir, "buildnet", "test.txt"), true},
	}

	for _, tt := range tests {
		_, err := os.Stat(tt.path)
		exists := err == nil
		if exists != tt.want {
			t.Errorf("Path %s exists = %v, want %v", tt.path, exists, tt.want)
		}
	}
}

func TestDownloadBuildnetConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "massa-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test directory structure
	nodeMassaDir := filepath.Join(tempDir, "node-massa")
	buildnetDir := filepath.Join(nodeMassaDir, "buildnet")
	if err := os.MkdirAll(buildnetDir, 0755); err != nil {
		t.Fatalf("Failed to create buildnet dir: %v", err)
	}

	// Run downloadBuildnetConfig
	if err := downloadBuildnetConfig(nodeMassaDir); err != nil {
		t.Fatalf("downloadBuildnetConfig failed: %v", err)
	}

	// Verify the config file exists
	configPath := filepath.Join(buildnetDir, "base_config", "config.toml")
	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("Config file not found at %s: %v", configPath, err)
	}
} 