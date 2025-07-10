package stakingManager

import (
	"slices"
	"testing"

	clientDriver "github.com/massalabs/node-manager-plugin/int/client-driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleRollsUpdates(t *testing.T) {
	cleanup := setupLog(t)
	defer cleanup()

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

func TestUpdateStakingAddresses(t *testing.T) {
	cleanup := setupLog(t)
	defer cleanup()

	tests := []struct {
		name           string
		existingAddrs  []StakingAddress
		newAddresses   []StakingAddress
		expectedResult bool
	}{
		{
			name: "Should return true when number of addresses changes",
			existingAddrs: []StakingAddress{
				{
					Address:     "addr1",
					TargetRolls: 10,
				},
			},
			newAddresses: []StakingAddress{
				{
					Address:     "addr1",
					TargetRolls: 10,
				},
				{
					Address:     "addr2",
					TargetRolls: 5,
				},
			},
			expectedResult: true,
		},
		{
			name: "Should return true when candidate rolls change",
			existingAddrs: []StakingAddress{
				{
					Address:        "addr1",
					CandidateRolls: 5,
					TargetRolls:    10,
				},
			},
			newAddresses: []StakingAddress{
				{
					Address:        "addr1",
					CandidateRolls: 8,
					TargetRolls:    10,
				},
			},
			expectedResult: true,
		},
		{
			name: "Should return true when candidate balance changes",
			existingAddrs: []StakingAddress{
				{
					Address:          "addr1",
					CandidateBalance: 100.0,
					TargetRolls:      10,
				},
			},
			newAddresses: []StakingAddress{
				{
					Address:          "addr1",
					CandidateBalance: 150.0,
					TargetRolls:      10,
				},
			},
			expectedResult: true,
		},
		{
			name: "Should return true when final rolls change",
			existingAddrs: []StakingAddress{
				{
					Address:     "addr1",
					FinalRolls:  5,
					TargetRolls: 10,
				},
			},
			newAddresses: []StakingAddress{
				{
					Address:     "addr1",
					FinalRolls:  7,
					TargetRolls: 10,
				},
			},
			expectedResult: true,
		},
		{
			name: "Should return true when active rolls change",
			existingAddrs: []StakingAddress{
				{
					Address:     "addr1",
					ActiveRolls: 3,
					TargetRolls: 10,
				},
			},
			newAddresses: []StakingAddress{
				{
					Address:     "addr1",
					ActiveRolls: 6,
					TargetRolls: 10,
				},
			},
			expectedResult: true,
		},
		{
			name: "Should return true when final balance changes",
			existingAddrs: []StakingAddress{
				{
					Address:      "addr1",
					FinalBalance: 200.0,
					TargetRolls:  10,
				},
			},
			newAddresses: []StakingAddress{
				{
					Address:      "addr1",
					FinalBalance: 250.0,
					TargetRolls:  10,
				},
			},
			expectedResult: true,
		},
		{
			name: "Should return true when deferred credits change",
			existingAddrs: []StakingAddress{
				{
					Address:     "addr1",
					TargetRolls: 10,
					DeferredCredits: []DeferredCredit{
						{
							Slot: Slot{
								Period: 100,
								Thread: 1,
							},
							Amount: 50.0,
						},
					},
				},
			},
			newAddresses: []StakingAddress{
				{
					Address:     "addr1",
					TargetRolls: 10,
					DeferredCredits: []DeferredCredit{
						{
							Slot: Slot{
								Period: 100,
								Thread: 1,
							},
							Amount: 75.0,
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "Should return true when deferred credits are added",
			existingAddrs: []StakingAddress{
				{
					Address:         "addr1",
					TargetRolls:     10,
					DeferredCredits: []DeferredCredit{},
				},
			},
			newAddresses: []StakingAddress{
				{
					Address:     "addr1",
					TargetRolls: 10,
					DeferredCredits: []DeferredCredit{
						{
							Slot: Slot{
								Period: 100,
								Thread: 1,
							},
							Amount: 50.0,
						},
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "Should return false when no changes occur",
			existingAddrs: []StakingAddress{
				{
					Address:          "addr1",
					CandidateRolls:   5,
					CandidateBalance: 100.0,
					FinalRolls:       7,
					ActiveRolls:      6,
					FinalBalance:     200.0,
					TargetRolls:      10,
					DeferredCredits: []DeferredCredit{
						{
							Slot: Slot{
								Period: 100,
								Thread: 1,
							},
							Amount: 50.0,
						},
					},
				},
			},
			newAddresses: []StakingAddress{
				{
					Address:          "addr1",
					CandidateRolls:   5,
					CandidateBalance: 100.0,
					FinalRolls:       7,
					ActiveRolls:      6,
					FinalBalance:     200.0,
					TargetRolls:      10,
					DeferredCredits: []DeferredCredit{
						{
							Slot: Slot{
								Period: 100,
								Thread: 1,
							},
							Amount: 50.0,
						},
					},
				},
			},
			expectedResult: false,
		},
		{
			name: "Should handle multiple addresses with mixed changes",
			existingAddrs: []StakingAddress{
				{
					Address:        "addr1",
					CandidateRolls: 5,
					TargetRolls:    10,
				},
				{
					Address:        "addr2",
					CandidateRolls: 8,
					TargetRolls:    10,
				},
			},
			newAddresses: []StakingAddress{
				{
					Address:        "addr1",
					CandidateRolls: 7, // Changed
					TargetRolls:    10,
				},
				{
					Address:        "addr2",
					CandidateRolls: 8, // Unchanged
					TargetRolls:    10,
				},
			},
			expectedResult: true,
		},
		{
			name:           "Should handle empty addresses",
			existingAddrs:  []StakingAddress{},
			newAddresses:   []StakingAddress{},
			expectedResult: false,
		},
		{
			name: "Should preserve target rolls when updating",
			existingAddrs: []StakingAddress{
				{
					Address:     "addr1",
					TargetRolls: 15, // This should be preserved
					FinalRolls:  5,
				},
			},
			newAddresses: []StakingAddress{
				{
					Address:     "addr1",
					TargetRolls: 20, // This should be ignored
					FinalRolls:  7,  // This should be updated
				},
			},
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create staking manager instance
			sm := &stakingManager{
				stakingAddresses: make([]StakingAddress, len(tt.existingAddrs)),
			}
			copy(sm.stakingAddresses, tt.existingAddrs)

			// Execute the function under test
			result := sm.updateStakingAddresses(tt.newAddresses)

			// Assert the result
			assert.Equal(t, tt.expectedResult, result, "Expected result mismatch")

			// Assert the updated addresses
			assert.Equal(t, len(tt.newAddresses), len(sm.stakingAddresses), "Number of addresses mismatch")

			for i, newAddr := range tt.newAddresses {
				if i < len(sm.stakingAddresses) {
					actualAddr := sm.stakingAddresses[i]
					assert.Equal(t, newAddr.Address, actualAddr.Address, "Address mismatch at index %d", i)
					assert.Equal(t, newAddr.CandidateRolls, actualAddr.CandidateRolls, "CandidateRolls mismatch at index %d", i)
					assert.Equal(t, newAddr.CandidateBalance, actualAddr.CandidateBalance, "CandidateBalance mismatch at index %d", i)
					assert.Equal(t, newAddr.FinalRolls, actualAddr.FinalRolls, "FinalRolls mismatch at index %d", i)
					assert.Equal(t, newAddr.ActiveRolls, actualAddr.ActiveRolls, "ActiveRolls mismatch at index %d", i)
					assert.Equal(t, newAddr.FinalBalance, actualAddr.FinalBalance, "FinalBalance mismatch at index %d", i)
					assert.Equal(t, len(newAddr.DeferredCredits), len(actualAddr.DeferredCredits), "DeferredCredits length mismatch at index %d", i)

					// Check that target rolls are preserved from existing addresses
					if len(tt.existingAddrs) > 0 {
						index := slices.IndexFunc(tt.existingAddrs, func(existingAddr StakingAddress) bool {
							return existingAddr.Address == newAddr.Address
						})
						if index >= 0 {
							assert.Equal(t, tt.existingAddrs[index].TargetRolls, actualAddr.TargetRolls, "TargetRolls should be preserved from existing address at index %d", i)
						}
					}

					for j, newCredit := range newAddr.DeferredCredits {
						if j < len(actualAddr.DeferredCredits) {
							actualCredit := actualAddr.DeferredCredits[j]
							assert.Equal(t, newCredit.Amount, actualCredit.Amount, "DeferredCredit amount mismatch at index %d, credit %d", i, j)
							assert.Equal(t, newCredit.Slot.Period, actualCredit.Slot.Period, "DeferredCredit slot period mismatch at index %d, credit %d", i, j)
							assert.Equal(t, newCredit.Slot.Thread, actualCredit.Slot.Thread, "DeferredCredit slot thread mismatch at index %d, credit %d", i, j)
						}
					}
				}
			}
		})
	}
}

func TestGetTotalValue(t *testing.T) {
	cleanup := setupLog(t)
	defer cleanup()

	const rollPrice = 100.0

	tests := []struct {
		name           string
		stakingAddrs   []StakingAddress
		expectedResult float64
	}{
		{
			name:           "Should return zero for empty addresses",
			stakingAddrs:   []StakingAddress{},
			expectedResult: 0.0,
		},
		{
			name: "Should calculate total value for single address with balance only",
			stakingAddrs: []StakingAddress{
				{
					Address:         "addr1",
					FinalBalance:    500.0,
					FinalRolls:      0,
					DeferredCredits: []DeferredCredit{},
				},
			},
			expectedResult: 500.0,
		},
		{
			name: "Should calculate total value for single address with rolls only",
			stakingAddrs: []StakingAddress{
				{
					Address:         "addr1",
					FinalBalance:    0.0,
					FinalRolls:      5,
					DeferredCredits: []DeferredCredit{},
				},
			},
			expectedResult: 500.0, // 5 rolls × 100.0 roll price
		},
		{
			name: "Should calculate total value for single address with deferred credits only",
			stakingAddrs: []StakingAddress{
				{
					Address:      "addr1",
					FinalBalance: 0.0,
					FinalRolls:   0,
					DeferredCredits: []DeferredCredit{
						{
							Slot: Slot{
								Period: 100,
								Thread: 1,
							},
							Amount: 250.0,
						},
						{
							Slot: Slot{
								Period: 101,
								Thread: 2,
							},
							Amount: 150.0,
						},
					},
				},
			},
			expectedResult: 400.0, // 250.0 + 150.0
		},
		{
			name: "Should calculate total value for single address with all components",
			stakingAddrs: []StakingAddress{
				{
					Address:      "addr1",
					FinalBalance: 1000.0,
					FinalRolls:   3,
					DeferredCredits: []DeferredCredit{
						{
							Slot: Slot{
								Period: 100,
								Thread: 1,
							},
							Amount: 200.0,
						},
					},
				},
			},
			expectedResult: 1500.0, // 1000.0 + (3 × 100.0) + 200.0
		},
		{
			name: "Should calculate total value for multiple addresses",
			stakingAddrs: []StakingAddress{
				{
					Address:      "addr1",
					FinalBalance: 500.0,
					FinalRolls:   2,
					DeferredCredits: []DeferredCredit{
						{
							Slot: Slot{
								Period: 100,
								Thread: 1,
							},
							Amount: 100.0,
						},
					},
				},
				{
					Address:      "addr2",
					FinalBalance: 750.0,
					FinalRolls:   1,
					DeferredCredits: []DeferredCredit{
						{
							Slot: Slot{
								Period: 101,
								Thread: 1,
							},
							Amount: 50.0,
						},
						{
							Slot: Slot{
								Period: 102,
								Thread: 2,
							},
							Amount: 75.0,
						},
					},
				},
			},
			expectedResult: 1775.0, // (500 + 200 + 100) + (750 + 100 + 50 + 75)
		},
		{
			name: "Should handle large numbers",
			stakingAddrs: []StakingAddress{
				{
					Address:      "addr1",
					FinalBalance: 1000000.0,
					FinalRolls:   1000,
					DeferredCredits: []DeferredCredit{
						{
							Slot: Slot{
								Period: 100,
								Thread: 1,
							},
							Amount: 500000.0,
						},
					},
				},
			},
			expectedResult: 1600000.0, // 1000000.0 + (1000 × 100.0) + 500000.0
		},
		{
			name: "Should handle addresses with no deferred credits",
			stakingAddrs: []StakingAddress{
				{
					Address:         "addr1",
					FinalBalance:    500.0,
					FinalRolls:      2,
					DeferredCredits: []DeferredCredit{},
				},
				{
					Address:         "addr2",
					FinalBalance:    300.0,
					FinalRolls:      1,
					DeferredCredits: []DeferredCredit{},
				},
			},
			expectedResult: 1100.0, // (500 + 200) + (300 + 100)
		},
		{
			name: "Should handle addresses with only deferred credits",
			stakingAddrs: []StakingAddress{
				{
					Address:      "addr1",
					FinalBalance: 0.0,
					FinalRolls:   0,
					DeferredCredits: []DeferredCredit{
						{
							Slot: Slot{
								Period: 100,
								Thread: 1,
							},
							Amount: 100.0,
						},
					},
				},
				{
					Address:      "addr2",
					FinalBalance: 0.0,
					FinalRolls:   0,
					DeferredCredits: []DeferredCredit{
						{
							Slot: Slot{
								Period: 101,
								Thread: 1,
							},
							Amount: 200.0,
						},
						{
							Slot: Slot{
								Period: 102,
								Thread: 2,
							},
							Amount: 300.0,
						},
					},
				},
			},
			expectedResult: 600.0, // 100.0 + (200.0 + 300.0)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create staking manager instance
			sm := &stakingManager{
				stakingAddresses: tt.stakingAddrs,
				miscellaneous: Miscellaneous{
					RollPrice: rollPrice,
				},
			}

			// Execute the function under test
			result := sm.getTotalValue()

			// Assert the result with tolerance for floating point precision
			assert.InDelta(t, tt.expectedResult, result, 0.01, "Total value calculation mismatch")
		})
	}
}
