package db

import (
	"path/filepath"
	"testing"
	"time"
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

	// Test adding address info
	err = db.AddRollsTarget("address1", 100)
	if err != nil {
		t.Fatalf("Failed to add address info: %v", err)
	}

	err = db.AddRollsTarget("address2", 200)
	if err != nil {
		t.Fatalf("Failed to add address info: %v", err)
	}

	// Test getting rolls target
	addresses, err := db.GetRollsTarget()
	if err != nil {
		t.Fatalf("Failed to get rolls target: %v", err)
	}

	if len(addresses) != 2 {
		t.Fatalf("Expected 2 addresses, got %d", len(addresses))
	}

	// Test updating rolls target
	err = db.UpdateRollsTarget("address1", 150)
	if err != nil {
		t.Fatalf("Failed to update rolls target: %v", err)
	}

	// Test deleting rolls target
	err = db.DeleteRollsTarget("address2")
	if err != nil {
		t.Fatalf("Failed to delete rolls target: %v", err)
	}

	// Verify deletion
	addresses, err = db.GetRollsTarget()
	if err != nil {
		t.Fatalf("Failed to get rolls target: %v", err)
	}

	if len(addresses) != 1 {
		t.Fatalf("Expected 1 address after deletion, got %d", len(addresses))
	}
}

func TestBalanceHistoryOperations(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "testdb.db")

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test posting history
	now := time.Now()
	histories := []BalanceHistory{
		{Timestamp: now.Add(-2 * time.Hour), TotalValue: 1000},
		{Timestamp: now.Add(-1 * time.Hour), TotalValue: 1100},
		{Timestamp: now, TotalValue: 1200},
	}

	err = db.PostHistory(histories)
	if err != nil {
		t.Fatalf("Failed to post history: %v", err)
	}

	// Test getting history
	since := now.Add(-3 * time.Hour)
	retrievedHistories, err := db.GetHistory(since)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(retrievedHistories) != 3 {
		t.Fatalf("Expected 3 history records, got %d", len(retrievedHistories))
	}

	// Test getting history with a later since time
	since = now.Add(-30 * time.Minute)
	retrievedHistories, err = db.GetHistory(since)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(retrievedHistories) != 1 {
		t.Fatalf("Expected 1 history record, got %d", len(retrievedHistories))
	}
}
