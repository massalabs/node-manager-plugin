package db

import (
	"database/sql"
	"fmt"
	"time"

	nodeManagerError "github.com/massalabs/node-manager-plugin/int/error"
	"github.com/massalabs/node-manager-plugin/int/utils"
	_ "github.com/mattn/go-sqlite3"
)

type DB interface {
	Close() error
	GetRollsTarget(network utils.Network) ([]AddressInfo, error)
	UpdateRollsTarget(address string, rollTarget uint64, network utils.Network) error
	AddRollsTarget(address string, rollTarget uint64, network utils.Network) error
	DeleteRollsTarget(address string, network utils.Network) error
	PostHistory(histories []ValueHistory, network utils.Network) error
	GetHistory(since time.Time, network utils.Network) ([]ValueHistory, error)
}

type dB struct {
	db *sql.DB
}

type ValueHistory struct {
	Timestamp  time.Time `json:"timestamp"`
	TotalValue float64   `json:"total_value"`
}

type AddressInfo struct {
	Address    string `json:"address"`
	RollTarget uint64 `json:"roll_target"`
	Network    string `json:"network"`
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
	// Create value_history_mainnet table
	valueHistoryMainnetTable := `
	CREATE TABLE IF NOT EXISTS value_history_mainnet (
		timestamp DATETIME PRIMARY KEY,
		total_value INTEGER NOT NULL
	);`

	// Create value_history_buildnet table
	valueHistoryBuildnetTable := `
	CREATE TABLE IF NOT EXISTS value_history_buildnet (
		timestamp DATETIME PRIMARY KEY,
		total_value INTEGER NOT NULL
	);`

	// Create rolls_target table
	rollsTargetTable := `
	CREATE TABLE IF NOT EXISTS rolls_target (
		address TEXT,
		roll_target INTEGER NOT NULL,
		network TEXT NOT NULL,
		PRIMARY KEY (address, network)
	);`

	if _, err := d.db.Exec(valueHistoryMainnetTable); err != nil {
		return fmt.Errorf("failed to create value_history_mainnet table: %w", err)
	}

	if _, err := d.db.Exec(valueHistoryBuildnetTable); err != nil {
		return fmt.Errorf("failed to create value_history_buildnet table: %w", err)
	}

	if _, err := d.db.Exec(rollsTargetTable); err != nil {
		return fmt.Errorf("failed to create rolls_target table: %w", err)
	}

	return nil
}

// GetRollsTarget returns a list of address and roll_target pairs for a specific network
func (d *dB) GetRollsTarget(network utils.Network) ([]AddressInfo, error) {
	query := `SELECT address, roll_target, network FROM rolls_target WHERE network = ? ORDER BY address`

	rows, err := d.db.Query(query, string(network))
	if err != nil {
		return nil, fmt.Errorf("failed to query rolls_target: %w", err)
	}
	defer rows.Close()

	var addresses []AddressInfo
	for rows.Next() {
		var addr AddressInfo
		if err := rows.Scan(&addr.Address, &addr.RollTarget, &addr.Network); err != nil {
			return nil, fmt.Errorf("failed to scan rolls_target row: %w", err)
		}
		addresses = append(addresses, addr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rolls_target rows: %w", err)
	}

	return addresses, nil
}

// ExistsRollsTarget checks if an address exists in the rolls_target table for a specific network

func (d *dB) existsRollsTarget(address string, network utils.Network) (bool, error) {
	query := `SELECT COUNT(*) FROM rolls_target WHERE address = ? AND network = ?`

	var count int
	err := d.db.QueryRow(query, address, string(network)).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if address %s exists for network %s: %w", address, string(network), err)
	}

	return count > 0, nil
}

// UpdateRollsTarget updates the roll_target for a specific address and network
func (d *dB) UpdateRollsTarget(address string, rollTarget uint64, network utils.Network) error {
	query := `UPDATE rolls_target SET roll_target = ? WHERE address = ? AND network = ?`

	result, err := d.db.Exec(query, rollTarget, address, string(network))
	if err != nil {
		return fmt.Errorf("failed to update roll_target: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("address %s not found for network %s", address, string(network))
	}

	return nil
}

// AddRollsTarget adds a new address with roll_target for a specific network
func (d *dB) AddRollsTarget(address string, rollTarget uint64, network utils.Network) error {
	query := `INSERT INTO rolls_target (address, roll_target, network) VALUES (?, ?, ?)`

	_, err := d.db.Exec(query, address, rollTarget, string(network))
	if err != nil {
		return fmt.Errorf("failed to add rolls_target: %w", err)
	}

	return nil
}

// DeleteRollsTarget deletes an address from the rolls_target table for a specific network
func (d *dB) DeleteRollsTarget(address string, network utils.Network) error {

	exists, err := d.existsRollsTarget(address, network)
	if err != nil {
		return fmt.Errorf("failed to check if address %s exists for network %s: %w", address, string(network), err)
	}

	if !exists {
		return nodeManagerError.New(nodeManagerError.ErrDBNotFoundItem, fmt.Sprintf("address %s not found for network %s", address, string(network)))
	}

	query := `DELETE FROM rolls_target WHERE address = ? AND network = ?`

	result, err := d.db.Exec(query, address, string(network))
	if err != nil {
		return fmt.Errorf("failed to delete rolls_target: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("address %s not found for network %s", address, string(network))
	}

	return nil
}

// PostHistory adds a list of value history records to the appropriate network table
func (d *dB) PostHistory(histories []ValueHistory, network utils.Network) error {
	if len(histories) == 0 {
		return nil
	}

	var tableName string
	switch network {
	case utils.NetworkMainnet:
		tableName = "value_history_mainnet"
	case utils.NetworkBuildnet:
		tableName = "value_history_buildnet"
	default:
		return fmt.Errorf("unsupported network: %s", string(network))
	}

	// Build a single INSERT statement with multiple VALUES
	query := fmt.Sprintf(`INSERT INTO %s (timestamp, total_value) VALUES `, tableName)
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
		return fmt.Errorf("failed to insert %s: %w", tableName, err)
	}

	return nil
}

// GetHistory retrieves all value history records after a given timestamp for a specific network, ordered chronologically
func (d *dB) GetHistory(since time.Time, network utils.Network) ([]ValueHistory, error) {
	var tableName string
	switch network {
	case utils.NetworkMainnet:
		tableName = "value_history_mainnet"
	case utils.NetworkBuildnet:
		tableName = "value_history_buildnet"
	default:
		return nil, fmt.Errorf("unsupported network: %s", string(network))
	}

	query := fmt.Sprintf(`SELECT timestamp, total_value FROM %s WHERE timestamp > ? ORDER BY timestamp ASC`, tableName)

	rows, err := d.db.Query(query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query %s: %w", tableName, err)
	}
	defer rows.Close()

	var histories []ValueHistory
	for rows.Next() {
		var history ValueHistory
		if err := rows.Scan(&history.Timestamp, &history.TotalValue); err != nil {
			return nil, fmt.Errorf("failed to scan %s row: %w", tableName, err)
		}
		histories = append(histories, history)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over %s rows: %w", tableName, err)
	}

	return histories, nil
}
