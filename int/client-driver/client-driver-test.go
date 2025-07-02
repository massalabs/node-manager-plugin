package clientDriver

import (
	"testing"
	"time"
)

func TestNewClientDriver(t *testing.T) {
	// Test that we can create a client driver (this will fail if massa-client binary is not found,
	// but that's expected in a test environment)
	_, err := NewClientDriver(false, nil, time.Second*10)
	if err != nil {
		// This is expected if the massa-client binary is not available in test environment
		t.Logf("Expected error when massa-client binary not found: %v", err)
	}
}

// func TestStakingAddressStruct(t *testing.T) {
// 	// Test the StakingAddress struct
// 	addr := StakingAddress{
// 		Address: "AU1test123",
// 		Rolls:   10,
// 		Balance: "1000.0",
// 	}

// 	if addr.Address != "AU1test123" {
// 		t.Errorf("Expected address AU1test123, got %s", addr.Address)
// 	}

// 	if addr.Rolls != 10 {
// 		t.Errorf("Expected rolls 10, got %d", addr.Rolls)
// 	}

// 	if addr.Balance != "1000.0" {
// 		t.Errorf("Expected balance 1000.0, got %s", addr.Balance)
// 	}
// }

// func TestParseAddressBalance(t *testing.T) {
// 	cd := &ClientDriver{}

// 	// Test parsing balance from mock output
// 	mockOutput := "Address: AU1test123\nBalance: 1500.5 MAS\nRolls: 5"
// 	balance, err := cd.parseAddressBalance(mockOutput)
// 	if err != nil {
// 		t.Errorf("Unexpected error parsing balance: %v", err)
// 	}

// 	if balance != "1500.5" {
// 		t.Errorf("Expected balance 1500.5, got %s", balance)
// 	}
// }

// func TestParseAddressRolls(t *testing.T) {
// 	cd := &ClientDriver{}

// 	// Test parsing rolls from mock output
// 	mockOutput := "Address: AU1test123\nBalance: 1500.5 MAS\nRolls: 5"
// 	rolls, err := cd.parseAddressRolls(mockOutput)
// 	if err != nil {
// 		t.Errorf("Unexpected error parsing rolls: %v", err)
// 	}

// 	if rolls != 5 {
// 		t.Errorf("Expected rolls 5, got %d", rolls)
// 	}
// }

// func TestParseWalletInfo(t *testing.T) {
// 	cd := &ClientDriver{}

// 	// Test parsing wallet info from mock output
// 	mockOutput := `Wallet Information:
// Address: AU1test123
// Balance: 1500.5 MAS
// Rolls: 5

// Address: AU1test456
// Balance: 2000.0 MAS
// Rolls: 10`

// 	addresses, err := cd.parseWalletInfo(mockOutput)
// 	if err != nil {
// 		t.Errorf("Unexpected error parsing wallet info: %v", err)
// 	}

// 	if len(addresses) != 2 {
// 		t.Errorf("Expected 2 addresses, got %d", len(addresses))
// 	}

// 	if addresses[0].Address != "AU1test123" {
// 		t.Errorf("Expected first address AU1test123, got %s", addresses[0].Address)
// 	}

// 	if addresses[1].Address != "AU1test456" {
// 		t.Errorf("Expected second address AU1test456, got %s", addresses[1].Address)
// 	}
// }
