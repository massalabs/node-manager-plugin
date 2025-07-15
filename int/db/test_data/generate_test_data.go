package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	// Configuration
	totalEntries    = 5000
	intervalSeconds = 180 // 3 minutes interval
	groupSizeMean   = 20  // Mean entries per group
	groupSizeStdDev = 5   // Standard deviation for group size
	oldestEntryDays = 366 // 1 year and 1 day
)

func main() {
	// Get the directory where this script file is located
	_, scriptFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatalf("Failed to get script file path")
	}
	scriptDir := filepath.Dir(scriptFile)

	// Database path in the same directory as this script file
	dbPath := filepath.Join(scriptDir, "test_data_total_value_history.db")

	// Remove existing test.db if it exists
	if _, err := os.Stat(dbPath); err == nil {
		if err := os.Remove(dbPath); err != nil {
			log.Fatalf("Failed to remove existing test.db: %v", err)
		}
		log.Println("Removed existing test.db")
	}

	// Open database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create tables
	if err := createTables(db); err != nil {
		log.Fatalf("Failed to create tables: %v", err)
	}

	// Generate and insert test data
	if err := generateTestData(db); err != nil {
		log.Fatalf("Failed to generate test data: %v", err)
	}

	log.Printf("Successfully created test database at %s with %d entries", dbPath, totalEntries)
}

func createTables(db *sql.DB) error {
	// Create value_history_buildnet table
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS value_history_buildnet (
		timestamp DATETIME PRIMARY KEY,
		total_value REAL NOT NULL
	);`

	if _, err := db.Exec(createTableSQL); err != nil {
		return fmt.Errorf("failed to create value_history_buildnet table: %w", err)
	}

	log.Println("Created value_history_buildnet table")
	return nil
}

func generateTestData(db *sql.DB) error {
	now := time.Now()
	oldestTime := now.AddDate(0, 0, -oldestEntryDays)

	log.Printf("Generating data from %s to %s", oldestTime.Format("2006-01-02 15:04:05"), now.Format("2006-01-02 15:04:05"))

	// Start transaction for better performance
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare insert statement
	insertStmt, err := tx.Prepare("INSERT INTO value_history_buildnet (timestamp, total_value) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer insertStmt.Close()

	entriesGenerated := 0
	baseValue := 10000.0 // Starting value in MAS

	// Calculate how many groups we need to evenly distribute 5000 entries
	// Each group has approximately groupSizeMean entries
	numGroups := totalEntries / groupSizeMean
	if totalEntries%groupSizeMean != 0 {
		numGroups++ // Add one more group for remaining entries
	}

	// Calculate the time span for each group to be evenly distributed
	totalTimeSpan := now.Sub(oldestTime)
	groupTimeSpan := totalTimeSpan / time.Duration(numGroups)

	log.Printf("Distributing %d groups evenly across %v timespan", numGroups, totalTimeSpan)
	log.Printf("Each group will span approximately %v", groupTimeSpan)

	// Generate groups evenly distributed across the timespan
	for groupIndex := 0; groupIndex < numGroups && entriesGenerated < totalEntries; groupIndex++ {
		// Calculate the start time for this group
		groupStartTime := oldestTime.Add(time.Duration(groupIndex) * groupTimeSpan)

		// Determine group size (normal distribution around mean)
		groupSize := int(rand.NormFloat64()*groupSizeStdDev + float64(groupSizeMean))
		if groupSize < 1 {
			groupSize = 1
		}
		if groupSize > 40 {
			groupSize = 40
		}

		// Ensure we don't exceed total entries
		if entriesGenerated+groupSize > totalEntries {
			groupSize = totalEntries - entriesGenerated
		}

		// Calculate the time span for this specific group
		groupEndTime := groupStartTime.Add(groupTimeSpan)
		if groupIndex == numGroups-1 {
			groupEndTime = now // Last group extends to now
		}

		// Calculate how many entries can fit in this group with fixed interval
		groupDuration := groupEndTime.Sub(groupStartTime)
		maxEntriesInGroup := int(groupDuration.Seconds())/intervalSeconds + 1

		// Use the smaller of calculated group size or max possible entries
		entriesInGroup := groupSize
		if entriesInGroup > maxEntriesInGroup {
			entriesInGroup = maxEntriesInGroup
		}

		// Generate entries for this group using fixed interval
		for i := 0; i < entriesInGroup; i++ {
			entryTime := groupStartTime.Add(time.Duration(i*intervalSeconds) * time.Second)

			// Skip if entry time is beyond the group end time
			if entryTime.After(groupEndTime) {
				break
			}

			// May add some randomness to the value (small variations)
			valueVariation := 0.0
			if rand.Float32() < 0.5 {
				valueVariation = (rand.Float64() - 0.5) * 100 // ±50 MAS variation
			}

			totalValue := baseValue + valueVariation

			// Insert the entry
			_, err := insertStmt.Exec(entryTime, totalValue)
			if err != nil {
				return fmt.Errorf("failed to insert entry at %s: %w", entryTime.Format("2006-01-02 15:04:05"), err)
			}

			entriesGenerated++
		}

		// Progress update
		if groupIndex%10 == 0 || groupIndex == numGroups-1 {
			log.Printf("Generated group %d/%d with %d entries (total: %d/%d)",
				groupIndex+1, numGroups, entriesInGroup, entriesGenerated, totalEntries)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Printf("Successfully generated %d entries in %d evenly distributed groups", entriesGenerated, numGroups)
	return nil
}
