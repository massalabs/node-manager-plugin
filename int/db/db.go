package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB interface {
	Close() error
	GetRollsTarget() ([]AddressInfo, error)
	UpdateRollsTarget(address string, rollTarget uint64) error
	AddRollsTarget(address string, rollTarget uint64) error
	DeleteRollsTarget(address string) error
	PostHistory(histories []BalanceHistory) error
	GetHistory(since time.Time) ([]BalanceHistory, error)
}

type dB struct {
	db *sql.DB
}

type BalanceHistory struct {
	Timestamp  time.Time `json:"timestamp"`
	TotalValue float64   `json:"total_value"`
}

type AddressInfo struct {
	Address    string `json:"address"`
	RollTarget uint64 `json:"roll_target"`
}

// NewDB creates a new database connection and initializes tables
func NewDB(dbPath string) (DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	database := &dB{db: db}

	if err := database.initTables(); err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %w", err)
	}

	return database, nil
}

// Close closes the database connection
func (d *dB) Close() error {
	return d.db.Close()
}

// initTables creates the required tables if they don't exist
func (d *dB) initTables() error {
	// Create balance_history table
	balanceHistoryTable := `
	CREATE TABLE IF NOT EXISTS balance_history (
		timestamp DATETIME PRIMARY KEY,
		total_value INTEGER NOT NULL
	);`

	// Create rolls_target table
	rollsTargetTable := `
	CREATE TABLE IF NOT EXISTS rolls_target (
		address TEXT PRIMARY KEY,
		roll_target INTEGER NOT NULL
	);`

	if _, err := d.db.Exec(balanceHistoryTable); err != nil {
		return fmt.Errorf("failed to create balance_history table: %w", err)
	}

	if _, err := d.db.Exec(rollsTargetTable); err != nil {
		return fmt.Errorf("failed to create rolls_target table: %w", err)
	}

	return nil
}

// GetRollsTarget returns a list of address and roll_target pairs
func (d *dB) GetRollsTarget() ([]AddressInfo, error) {
	query := `SELECT address, roll_target FROM rolls_target ORDER BY address`

	rows, err := d.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query rolls_target: %w", err)
	}
	defer rows.Close()

	var addresses []AddressInfo
	for rows.Next() {
		var addr AddressInfo
		if err := rows.Scan(&addr.Address, &addr.RollTarget); err != nil {
			return nil, fmt.Errorf("failed to scan rolls_target row: %w", err)
		}
		addresses = append(addresses, addr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rolls_target rows: %w", err)
	}

	return addresses, nil
}

// UpdateRollsTarget updates the roll_target for a specific address
func (d *dB) UpdateRollsTarget(address string, rollTarget uint64) error {
	query := `UPDATE rolls_target SET roll_target = ? WHERE address = ?`

	result, err := d.db.Exec(query, rollTarget, address)
	if err != nil {
		return fmt.Errorf("failed to update roll_target: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("address %s not found", address)
	}

	return nil
}

// AddRollsTarget adds a new address with roll_target
func (d *dB) AddRollsTarget(address string, rollTarget uint64) error {
	query := `INSERT INTO rolls_target (address, roll_target) VALUES (?, ?)`

	_, err := d.db.Exec(query, address, rollTarget)
	if err != nil {
		return fmt.Errorf("failed to add rolls_target: %w", err)
	}

	return nil
}

// DeleteRollsTarget deletes an address from the rolls_target table
func (d *dB) DeleteRollsTarget(address string) error {
	query := `DELETE FROM rolls_target WHERE address = ?`

	result, err := d.db.Exec(query, address)
	if err != nil {
		return fmt.Errorf("failed to delete rolls_target: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("address %s not found", address)
	}

	return nil
}

// PostHistory adds a list of balance history records
func (d *dB) PostHistory(histories []BalanceHistory) error {
	if len(histories) == 0 {
		return nil
	}

	// Build a single INSERT statement with multiple VALUES
	query := `INSERT INTO balance_history (timestamp, total_value) VALUES `
	args := make([]interface{}, 0, len(histories)*2)

	for i, history := range histories {
		if i > 0 {
			query += ", "
		}
		query += "(?, ?)"
		args = append(args, history.Timestamp, history.TotalValue)
	}

	_, err := d.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to insert balance_history: %w", err)
	}

	return nil
}

// GetHistory retrieves all balance history records after a given timestamp, ordered chronologically
func (d *dB) GetHistory(since time.Time) ([]BalanceHistory, error) {
	query := `SELECT timestamp, total_value FROM balance_history WHERE timestamp > ? ORDER BY timestamp ASC`

	rows, err := d.db.Query(query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query balance_history: %w", err)
	}
	defer rows.Close()

	var histories []BalanceHistory
	for rows.Next() {
		var history BalanceHistory
		if err := rows.Scan(&history.Timestamp, &history.TotalValue); err != nil {
			return nil, fmt.Errorf("failed to scan balance_history row: %w", err)
		}
		histories = append(histories, history)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over balance_history rows: %w", err)
	}

	return histories, nil
}
