package clientDriver

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	nodeDirManagerPkg "github.com/massalabs/node-manager-plugin/int/node-bin-dir-manager"
)

type Slot struct {
	Period uint64 `json:"period"`
	Thread uint8  `json:"thread"`
}

type AddressInfo struct {
	ActiveRolls uint64 `json:"active_rolls"`
}

// WalletInfo represents wallet information for an address
type WalletInfo struct {
	AddressInfo AddressInfo `json:"address_info"`
}

// ClientDriver handles interactions with the massa-client CLI tool

type ClientDriver interface {
	GetStakingAddresses() ([]string, error)
	AddStakingAddress(pwd string, secKey, address string) error
	RemoveStakingAddress(pwd string, address string) error
	BuyRolls(pwd string, address string, amount uint64, fee float32) (string, error)
	SellRolls(pwd string, address string, amount uint64, fee float32) (string, error)
	WalletInfo(pwd string) (map[string]WalletInfo, error)
	WalletInfoWithoutNode(pwd string) (string, error)
}

type clientDriver struct {
	binPath        string
	nodeDirManager nodeDirManagerPkg.NodeDirManager
	timeout        time.Duration
}

// NewClientDriver creates a new ClientDriver instance
func NewClientDriver(
	isMainnet bool,
	nodeDirManager nodeDirManagerPkg.NodeDirManager,
	timeout time.Duration,
) (ClientDriver, error) {
	binPath, err := nodeDirManager.GetClientBin(isMainnet)
	if err != nil {
		return nil, fmt.Errorf("failed to get client binary path: %v", err)
	}

	cd := &clientDriver{
		binPath:        binPath,
		nodeDirManager: nodeDirManager,
		timeout:        timeout,
	}

	return cd, nil
}

// executeCommand executes a massa-client command and returns the output
func (cd *clientDriver) executeCommand(args ...string) ([]byte, error) {
	// Create command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cd.timeout)
	defer cancel()

	args = append(args, "-a")

	cmd := exec.CommandContext(ctx, cd.binPath, args...)
	cmd.Dir = filepath.Dir(cd.binPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}

	return output, nil
}

// GetStakingAddresses retrieves all staking addresses
func (cd *clientDriver) GetStakingAddresses() ([]string, error) {
	output, err := cd.executeCommand("node_get_staking_addresses", "-j")
	if err != nil {
		return nil, fmt.Errorf("failed to get staking addresses list: %v", err)
	}

	var addresses []string
	err = json.Unmarshal(output, &addresses)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal staking addresses: %v", err)
	}

	return addresses, nil
}

// AddStakingAddress adds a new staking address
func (cd *clientDriver) AddStakingAddress(pwd string, secKey, address string) error {
	_, err := cd.executeCommand("wallet_add_secret_keys", "-p", pwd, secKey)
	if err != nil {
		return fmt.Errorf("failed to add address %s to massa client: %v", address, err)
	}

	_, err = cd.executeCommand("node_start_staking", "-p", pwd, address)
	if err != nil {
		return fmt.Errorf("failed to add staking address %s to massa node: %v", address, err)
	}

	return nil
}

// RemoveStakingAddress removes a staking address
func (cd *clientDriver) RemoveStakingAddress(pwd string, address string) error {
	_, err := cd.executeCommand("node_stop_staking", "-p", pwd, address)
	if err != nil {
		return fmt.Errorf("failed to remove staking address %s from massa node: %v", address, err)
	}

	_, err = cd.executeCommand("wallet_remove_addresses", "-p", pwd, address)
	if err != nil {
		return fmt.Errorf("failed to remove address %s from massa client: %v", address, err)
	}

	return nil
}

// BuyRolls buys rolls for a specific address
func (cd *clientDriver) BuyRolls(pwd string, address string, amount uint64, fee float32) (string, error) {
	output, err := cd.executeCommand("buy_rolls", "-p", pwd, "-j", address, fmt.Sprintf("%d", amount), fmt.Sprintf("%f", fee))
	if err != nil {
		return "", fmt.Errorf("failed to buy %d rolls for address %s with fee %f MAS, got error: %v", amount, address, fee, err)
	}

	// retrieve operation id from response
	var res []string
	err = json.Unmarshal(output, &res)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal buy rolls response: %v", err)
	}
	if len(res) == 0 {
		return "", fmt.Errorf("no operation id response from buy rolls")
	}

	return res[0], nil
}

// SellRolls sells rolls for a specific address
func (cd *clientDriver) SellRolls(pwd string, address string, amount uint64, fee float32) (string, error) {
	output, err := cd.executeCommand("sell_rolls", "-p", pwd, "-j", address, fmt.Sprintf("%d", amount), fmt.Sprintf("%f", fee))
	if err != nil {
		return "", fmt.Errorf("failed to sell %d rolls for address %s with fee %f MAS, got error: %v", amount, address, fee, err)
	}

	// retrieve operation id from response
	var res []string
	err = json.Unmarshal(output, &res)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal sell rolls response: %v", err)
	}
	if len(res) == 0 {
		return "", fmt.Errorf("no operation id response from sell rolls")
	}

	return res[0], nil
}

// WalletInfo retrieves wallet information for all addresses
func (cd *clientDriver) WalletInfo(pwd string) (map[string]WalletInfo, error) {
	output, err := cd.executeCommand("wallet_info", "-j", "-p", pwd)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet info: %v", err)
	}

	// Parse the JSON response which is an object with address keys
	var walletData map[string]WalletInfo

	err = json.Unmarshal(output, &walletData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal wallet info: %v", err)
	}

	return walletData, nil
}

func (cd *clientDriver) WalletInfoWithoutNode(pwd string) (string, error) {
	output, err := cd.executeCommand("wallet_info", "-j", "-p", pwd)
	if err != nil {
		return "", fmt.Errorf("failed to get wallet info: %v", err)
	}
	return string(output), nil
}
