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

// // GetAddressBalance gets the balance of a specific address
// func (cd *ClientDriver) GetAddressBalance(address string) (string, error) {
// 	output, err := cd.executeCommand("wallet", "info", "--address", address)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to get balance for address %s: %v", address, err)
// 	}

// 	// Parse the output to extract balance
// 	balance, err := cd.parseAddressBalance(output)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to parse balance for address %s: %v", address, err)
// 	}

// 	return balance, nil
// }

// // GetAddressRolls gets the number of rolls for a specific address
// func (cd *ClientDriver) GetAddressRolls(address string) (int64, error) {
// 	output, err := cd.executeCommand("wallet", "info", "--address", address)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to get rolls for address %s: %v", address, err)
// 	}

// 	// Parse the output to extract rolls
// 	rolls, err := cd.parseAddressRolls(output)
// 	if err != nil {
// 		return 0, fmt.Errorf("failed to parse rolls for address %s: %v", address, err)
// 	}

// 	return rolls, nil
// }

// // parseWalletInfo parses the wallet info output to extract staking addresses
// func (cd *ClientDriver) parseWalletInfo(output string) ([]StakingAddress, error) {
// 	var addresses []StakingAddress

// 	// This is a simplified parser - the actual output format may need adjustment
// 	lines := strings.Split(output, "\n")

// 	for _, line := range lines {
// 		line = strings.TrimSpace(line)
// 		if strings.Contains(line, "Address:") {
// 			// Extract address and related information
// 			parts := strings.Fields(line)
// 			if len(parts) >= 2 {
// 				address := parts[1]
// 				// Try to get rolls and balance for this address
// 				rolls, _ := cd.GetAddressRolls(address)
// 				balance, _ := cd.GetAddressBalance(address)

// 				addresses = append(addresses, StakingAddress{
// 					Address: address,
// 					Rolls:   rolls,
// 					Balance: balance,
// 				})
// 			}
// 		}
// 	}

// 	return addresses, nil
// }

// // parseAddressBalance parses the address info output to extract balance
// func (cd *ClientDriver) parseAddressBalance(output string) (string, error) {
// 	// This is a simplified parser - adjust based on actual output format
// 	lines := strings.Split(output, "\n")
// 	for _, line := range lines {
// 		if strings.Contains(line, "Balance:") {
// 			parts := strings.Fields(line)
// 			if len(parts) >= 2 {
// 				return parts[1], nil
// 			}
// 		}
// 	}
// 	return "0", nil
// }

// // parseAddressRolls parses the address info output to extract rolls
// func (cd *ClientDriver) parseAddressRolls(output string) (int64, error) {
// 	// This is a simplified parser - adjust based on actual output format
// 	lines := strings.Split(output, "\n")
// 	for _, line := range lines {
// 		if strings.Contains(line, "Rolls:") {
// 			parts := strings.Fields(line)
// 			if len(parts) >= 2 {
// 				var rolls int64
// 				_, err := fmt.Sscanf(parts[1], "%d", &rolls)
// 				if err != nil {
// 					return 0, fmt.Errorf("failed to parse rolls number: %v", err)
// 				}
// 				return rolls, nil
// 			}
// 		}
// 	}
// 	return 0, nil
// }

// // TestConnection tests if the client can connect to the node
// func (cd *ClientDriver) TestConnection() error {
// 	_, err := cd.executeCommand("wallet", "info")
// 	if err != nil {
// 		return fmt.Errorf("failed to test connection: %v", err)
// 	}
// 	return nil
// }
