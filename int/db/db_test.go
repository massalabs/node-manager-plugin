package db

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/massalabs/node-manager-plugin/int/utils"
)

func TestNewDB(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "testdb.db")

	// Test creating a new database
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()
}

func TestAddressInfoOperations(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "testdb.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test adding address info for mainnet
	err = db.AddRollsTarget("address1", 100, utils.NetworkMainnet)
	if err != nil {
		t.Fatalf("Failed to add address info: %v", err)
	}

	err = db.AddRollsTarget("address2", 200, utils.NetworkMainnet)
	if err != nil {
		t.Fatalf("Failed to add address info: %v", err)
	}

	// Test adding address info for buildnet
	err = db.AddRollsTarget("address1", 150, utils.NetworkBuildnet)
	if err != nil {
		t.Fatalf("Failed to add address info: %v", err)
	}

	// Test getting rolls target for mainnet
	addresses, err := db.GetRollsTarget(utils.NetworkMainnet)
	if err != nil {
		t.Fatalf("Failed to get rolls target: %v", err)
	}

	if len(addresses) != 2 {
		t.Fatalf("Expected 2 addresses for mainnet, got %d", len(addresses))
	}

	// Test getting rolls target for buildnet
	addresses, err = db.GetRollsTarget(utils.NetworkBuildnet)
	if err != nil {
		t.Fatalf("Failed to get rolls target: %v", err)
	}

	if len(addresses) != 1 {
		t.Fatalf("Expected 1 address for buildnet, got %d", len(addresses))
	}

	// Test updating rolls target for mainnet
	err = db.UpdateRollsTarget("address1", 150, utils.NetworkMainnet)
	if err != nil {
		t.Fatalf("Failed to update rolls target: %v", err)
	}

	// Test deleting rolls target for mainnet
	err = db.DeleteRollsTarget("address2", utils.NetworkMainnet)
	if err != nil {
		t.Fatalf("Failed to delete rolls target: %v", err)
	}

	// Verify deletion for mainnet
	addresses, err = db.GetRollsTarget(utils.NetworkMainnet)
	if err != nil {
		t.Fatalf("Failed to get rolls target: %v", err)
	}

	if len(addresses) != 1 {
		t.Fatalf("Expected 1 address after deletion for mainnet, got %d", len(addresses))
	}

	// Verify buildnet still has its address
	addresses, err = db.GetRollsTarget(utils.NetworkBuildnet)
	if err != nil {
		t.Fatalf("Failed to get rolls target: %v", err)
	}

	if len(addresses) != 1 {
		t.Fatalf("Expected 1 address for buildnet after mainnet deletion, got %d", len(addresses))
	}
}

func TestValueHistoryOperations(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "testdb.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test posting history for mainnet
	now := time.Now()
	histories := []ValueHistory{
		{Timestamp: now.Add(-2 * time.Hour), TotalValue: 1000},
		{Timestamp: now.Add(-1 * time.Hour), TotalValue: 1100},
		{Timestamp: now, TotalValue: 1200},
	}

	err = db.PostHistory(histories, utils.NetworkMainnet)
	if err != nil {
		t.Fatalf("Failed to post history: %v", err)
	}

	// Test posting history for buildnet
	buildnetHistories := []ValueHistory{
		{Timestamp: now.Add(-1 * time.Hour), TotalValue: 500},
		{Timestamp: now, TotalValue: 600},
	}

	err = db.PostHistory(buildnetHistories, utils.NetworkBuildnet)
	if err != nil {
		t.Fatalf("Failed to post buildnet history: %v", err)
	}

	// Test getting history for mainnet
	since := now.Add(-3 * time.Hour)
	retrievedHistories, err := db.GetHistory(since, utils.NetworkMainnet)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(retrievedHistories) != 3 {
		t.Fatalf("Expected 3 history records for mainnet, got %d", len(retrievedHistories))
	}

	// Test getting history for buildnet
	retrievedHistories, err = db.GetHistory(since, utils.NetworkBuildnet)
	if err != nil {
		t.Fatalf("Failed to get buildnet history: %v", err)
	}

	if len(retrievedHistories) != 2 {
		t.Fatalf("Expected 2 history records for buildnet, got %d", len(retrievedHistories))
	}

	// Test getting history with a later since time for mainnet
	since = now.Add(-30 * time.Minute)
	retrievedHistories, err = db.GetHistory(since, utils.NetworkMainnet)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(retrievedHistories) != 1 {
		t.Fatalf("Expected 1 history record for mainnet, got %d", len(retrievedHistories))
	}
}
