package clientDriver

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	nodeDirManagerPkg "github.com/massalabs/node-manager-plugin/int/node-bin-dir-manager"
)

type Slot struct {
	Period uint64 `json:"period"`
	Thread uint8  `json:"thread"`
}

type DeferredCredits struct {
	Slot   Slot   `json:"slot"`
	Amount uint64 `json:"amount"`
}

// StakingAddress represents a staking address with its information
type StakingAddress struct {
	Address         string            `json:"address"`
	FinalRolls      uint64            `json:"final_roll_count"`
	FinalBalance    string            `json:"final_balance"`
	Thread          uint8             `json:"thread"`
	DeferredCredits []DeferredCredits `json:"deferred_credits"`
}

// ClientDriver handles interactions with the massa-client CLI tool
type ClientDriver struct {
	isMainnet      bool
	binPath        string
	nodeDirManager nodeDirManagerPkg.NodeDirManager
	timeout        time.Duration
}

// NewClientDriver creates a new ClientDriver instance
func NewClientDriver(
	isMainnet bool,
	nodeDirManager nodeDirManagerPkg.NodeDirManager,
	timeout time.Duration,
) (*ClientDriver, error) {
	binPath, err := nodeDirManager.GetClientBin(isMainnet)
	if err != nil {
		return nil, fmt.Errorf("failed to get client binary path: %v", err)
	}

	cd := &ClientDriver{
		isMainnet:      isMainnet,
		binPath:        binPath,
		nodeDirManager: nodeDirManager,
		timeout:        timeout,
	}

	return cd, nil
}

// executeCommand executes a massa-client command and returns the output
func (cd *ClientDriver) executeCommand(args ...string) ([]byte, error) {
	// Create command with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cd.timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, cd.binPath, args...)
	cmd.Dir = filepath.Dir(cd.binPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("command failed: %v, output: %s", err, string(output))
	}

	return output, nil
}

// GetStakingAddresses retrieves all staking addresses
func (cd *ClientDriver) GetStakingAddresses() ([]StakingAddress, error) {
	output, err := cd.executeCommand("node_get_staking_addresses")
	if err != nil {
		return nil, fmt.Errorf("failed to get staking addresses list: %v", err)
	}

	outputStr := strings.ReplaceAll(string(output), "\n", " ")

	addresses, err := cd.executeCommand("get_addresses", outputStr)
	if err != nil {
		return nil, fmt.Errorf("failed to get staking addresses info: %v", err)
	}

	resAddresses := []StakingAddress{}
	err = json.Unmarshal(addresses, &resAddresses)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal staking addresses info: %v", err)
	}

	return resAddresses, nil
}

// AddStakingAddress adds a new staking address
func (cd *ClientDriver) AddStakingAddress(pwd string, address string) error {
	_, err := cd.executeCommand("node_start_staking", "-p", pwd, address)
	if err != nil {
		return fmt.Errorf("failed to add staking address %s: %v", address, err)
	}

	return nil
}

// RemoveStakingAddress removes a staking address
func (cd *ClientDriver) RemoveStakingAddress(pwd string, address string) error {
	_, err := cd.executeCommand("node_stop_staking", "-p", pwd, address)
	if err != nil {
		return fmt.Errorf("failed to remove staking address %s: %v", address, err)
	}

	return nil
}

// BuyRolls buys rolls for a specific address
func (cd *ClientDriver) BuyRolls(pwd string, address string, amount int64, fee int64) (string, error) {
	output, err := cd.executeCommand("buy_rolls", "-p", pwd, "-j", address, fmt.Sprintf("%d", amount), fmt.Sprintf("%d", fee))
	if err != nil {
		return "", fmt.Errorf("failed to buy %d rolls for address %s with fee %d MAS, got error: %v", amount, address, fee, err)
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
func (cd *ClientDriver) SellRolls(pwd string, address string, amount int64, fee int64) (string, error) {
	output, err := cd.executeCommand("sell_rolls", "-p", pwd, "-j", address, fmt.Sprintf("%d", amount), fmt.Sprintf("%d", fee))
	if err != nil {
		return "", fmt.Errorf("failed to sell %d rolls for address %s with fee %d MAS, got error: %v", amount, address, fee, err)
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
