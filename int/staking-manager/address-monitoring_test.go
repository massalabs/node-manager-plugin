package stakingManager

import (
	"path/filepath"
	"testing"

	clientDriver "github.com/massalabs/node-manager-plugin/int/client-driver"
	"github.com/massalabs/station/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleRollsUpdates(t *testing.T) {
	// Initialize logger for testing
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")
	err := logger.InitializeGlobal(logPath)
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}
	defer func() {
		logger.Close()
	}()

	// Constants for testing
	const (
		minimalFees = 0.1
		rollPrice   = 100.0
	)

	tests := []struct {
		name          string
		newAddresses  []StakingAddress
		existingAddrs []StakingAddress
		expectedCalls func(*clientDriver.MockClientDriver, *testing.T)
	}{
		{
			name: "Should sell rolls when target is lower than current rolls",
			newAddresses: []StakingAddress{
				{
					Address:      "test_address_1",
					FinalRolls:   10,
					FinalBalance: 1000.0,
				},
			},
			existingAddrs: []StakingAddress{
				{
					Address:     "test_address_1",
					TargetRolls: 5,
				},
			},
			expectedCalls: func(mockClient *clientDriver.MockClientDriver, t *testing.T) {
				mockClient.On("SellRolls", mock.Anything, "test_address_1", uint64(5), float32(minimalFees)).Return("tx_hash", nil).Once()
			},
		},
		{
			name: "Should buy rolls when target is higher than current rolls",
			newAddresses: []StakingAddress{
				{
					Address:      "test_address_2",
					FinalRolls:   5,
					FinalBalance: 1000.0, // There is enough balance to buy 5 rolls
				},
			},
			existingAddrs: []StakingAddress{
				{
					Address:     "test_address_2",
					TargetRolls: 10,
				},
			},
			expectedCalls: func(mockClient *clientDriver.MockClientDriver, t *testing.T) {
				// Should buy 5 rolls (difference between target 10 and current 5)
				// But limited by available balance: 1000.0 / 100.0 = 10 rolls max
				// So should buy 5 rolls (the minimum of difference and available)
				mockClient.On("BuyRolls", mock.Anything, "test_address_2", uint64(5), float32(minimalFees)).Return("tx_hash", nil).Once()
			},
		},
		{
			name: "Should not sell rolls when insufficient balance for fees",
			newAddresses: []StakingAddress{
				{
					Address:      "test_address_3",
					FinalRolls:   5,
					FinalBalance: 0.05, // Less than minimal fees (0.1)
				},
			},
			existingAddrs: []StakingAddress{
				{
					Address:     "test_address_3",
					TargetRolls: 10,
				},
			},
			expectedCalls: func(mockClient *clientDriver.MockClientDriver, t *testing.T) {
				// Should not call SellRolls because insufficient balance
				mockClient.AssertNotCalled(t, "SellRolls")
				mockClient.AssertNotCalled(t, "BuyRolls")
			},
		},
		{
			name: "Should not buy rolls when insufficient balance for fees",
			newAddresses: []StakingAddress{
				{
					Address:      "test_address_4",
					FinalRolls:   10,
					FinalBalance: 0.05, // Less than minimal fees (0.1)
				},
			},
			existingAddrs: []StakingAddress{
				{
					Address:     "test_address_4",
					TargetRolls: 5,
				},
			},
			expectedCalls: func(mockClient *clientDriver.MockClientDriver, t *testing.T) {
				mockClient.AssertNotCalled(t, "BuyRolls")
				mockClient.AssertNotCalled(t, "SellRolls")
			},
		},
		{
			name: "Should limit buy rolls by available balance",
			newAddresses: []StakingAddress{
				{
					Address:      "test_address_5",
					FinalRolls:   5,
					FinalBalance: 301.13, // Can buy 3 rolls (300/100), but needs to buy 5
				},
			},
			existingAddrs: []StakingAddress{
				{
					Address:     "test_address_5",
					TargetRolls: 10,
				},
			},
			expectedCalls: func(mockClient *clientDriver.MockClientDriver, t *testing.T) {
				// Should buy 3 rolls (limited by balance: 300/100 = 3)
				mockClient.On("BuyRolls", mock.Anything, "test_address_5", uint64(3), float32(minimalFees)).Return("tx_hash", nil).Once()
			},
		},
		{
			name: "Should handle multiple addresses",
			newAddresses: []StakingAddress{
				{
					Address:      "test_address_6",
					FinalRolls:   10,
					FinalBalance: 1000.0,
				},
				{
					Address:      "test_address_7",
					FinalRolls:   5,
					FinalBalance: 1000.0,
				},
			},
			existingAddrs: []StakingAddress{
				{
					Address:     "test_address_6",
					TargetRolls: 5,
				},
				{
					Address:     "test_address_7",
					TargetRolls: 10,
				},
			},
			expectedCalls: func(mockClient *clientDriver.MockClientDriver, t *testing.T) {
				// Address 6: sell 5 rolls (10 target - 5 current)
				mockClient.On("SellRolls", mock.Anything, "test_address_6", uint64(5), float32(minimalFees)).Return("tx_hash", nil).Once()
				// Address 7: buy 5 rolls (10 current - 5 target)
				mockClient.On("BuyRolls", mock.Anything, "test_address_7", uint64(5), float32(minimalFees)).Return("tx_hash", nil).Once()
			},
		},
		{
			name: "Should handle client driver errors gracefully on sell rolls",
			newAddresses: []StakingAddress{
				{
					Address:      "test_address_8",
					FinalRolls:   7,
					FinalBalance: 1000.0,
				},
			},
			existingAddrs: []StakingAddress{
				{
					Address:     "test_address_8",
					TargetRolls: 3,
				},
			},
			expectedCalls: func(mockClient *clientDriver.MockClientDriver, t *testing.T) {
				// Should attempt to sell rolls but fail
				mockClient.On("SellRolls", mock.Anything, "test_address_8", uint64(4), float32(minimalFees)).Return("", assert.AnError).Once()
			},
		},
		{
			name: "Should handle client driver errors gracefully on buy rolls",
			newAddresses: []StakingAddress{
				{
					Address:      "test_address_9",
					FinalRolls:   3,
					FinalBalance: 1000.0,
				},
			},
			existingAddrs: []StakingAddress{
				{
					Address:     "test_address_9",
					TargetRolls: 7,
				},
			},
			expectedCalls: func(mockClient *clientDriver.MockClientDriver, t *testing.T) {
				// Should attempt to sell rolls but fail
				mockClient.On("BuyRolls", mock.Anything, "test_address_9", uint64(4), float32(minimalFees)).Return("", assert.AnError).Once()
			},
		},
		{
			name: "Should not perform any action when target equals current rolls",
			newAddresses: []StakingAddress{
				{
					Address:      "test_address_9",
					FinalRolls:   10,
					FinalBalance: 1000.0,
				},
			},
			existingAddrs: []StakingAddress{
				{
					Address:     "test_address_9",
					TargetRolls: 10,
				},
			},
			expectedCalls: func(mockClient *clientDriver.MockClientDriver, t *testing.T) {
				mockClient.AssertNotCalled(t, "BuyRolls")
				mockClient.AssertNotCalled(t, "SellRolls")
			},
		},
		{
			name:          "Should handle empty addresses list",
			newAddresses:  []StakingAddress{},
			existingAddrs: []StakingAddress{},
			expectedCalls: func(mockClient *clientDriver.MockClientDriver, t *testing.T) {
				mockClient.AssertNotCalled(t, "BuyRolls")
				mockClient.AssertNotCalled(t, "SellRolls")
			},
		},
		{
			name: "Should handle complex scenario with multiple conditions",
			newAddresses: []StakingAddress{
				{
					Address:      "addr1",
					FinalRolls:   15,
					FinalBalance: 1000.0,
				},
				{
					Address:      "addr2",
					FinalRolls:   3,
					FinalBalance: 1000.0,
				},
				{
					Address:      "addr3",
					FinalRolls:   8,
					FinalBalance: 1000.0,
				},
				{
					Address:      "addr4",
					FinalRolls:   10,
					FinalBalance: 99.99,
				},
			},
			existingAddrs: []StakingAddress{
				{
					Address:     "addr1",
					TargetRolls: 10,
				},
				{
					Address:     "addr2",
					TargetRolls: 5,
				},
				{
					Address:     "addr3",
					TargetRolls: 8,
				},
				{
					Address:     "addr4",
					TargetRolls: 15,
				},
			},
			expectedCalls: func(mockClient *clientDriver.MockClientDriver, t *testing.T) {
				// addr1: sell 5 rolls (15 current - 10 target)
				mockClient.On("SellRolls", mock.Anything, "addr1", uint64(5), float32(minimalFees)).Return("tx_hash1", nil).Once()
				// addr2: buy 2 rolls (5 target - 3 current)
				mockClient.On("BuyRolls", mock.Anything, "addr2", uint64(2), float32(minimalFees)).Return("tx_hash2", nil).Once()
				// addr3: no action needed (8 current = 8 target)
				// addr4: insufficient balance for buying rolls
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockClient := clientDriver.NewMockClientDriver(t)

			// Setup expected calls
			tt.expectedCalls(mockClient, t)

			// Create staking manager instance using the unexported struct
			sm := &stakingManager{
				clientDriver:     mockClient,
				stakingAddresses: tt.existingAddrs,
				miscellaneous: Miscellaneous{
					MinimalFees: minimalFees,
					RollPrice:   rollPrice,
				},
			}

			// Execute the function under test
			sm.handleRollsUpdates(tt.newAddresses)

			// Assert that all expected calls were made
			mockClient.AssertExpectations(t)
		})
	}
}
