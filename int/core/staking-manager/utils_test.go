package stakingManager

import (
	"testing"

	"slices"
)

// createTestStakingManager creates a stakingManager instance for testing
func createTestStakingManager() *stakingManager {
	return &stakingManager{
		stakingAddresses: []StakingAddress{
			{
				Address:          "AU1test123",
				FinalRolls:       10,
				CandidateRolls:   8,
				FinalBalance:     1000.0,
				CandidateBalance: 950.0,
				Thread:           0,
				TargetRolls:      15,
			},
			{
				Address:          "AU1test456",
				FinalRolls:       5,
				CandidateRolls:   5,
				FinalBalance:     500.0,
				CandidateBalance: 500.0,
				Thread:           1,
				TargetRolls:      10,
			},
			{
				Address:          "AU1test789",
				FinalRolls:       20,
				CandidateRolls:   18,
				FinalBalance:     2000.0,
				CandidateBalance: 1900.0,
				Thread:           2,
				TargetRolls:      25,
			},
		},
	}
}

func TestGetAddressIndexFromRamList(t *testing.T) {
	sm := createTestStakingManager()

	tests := []struct {
		name     string
		address  string
		expected int
		found    bool
	}{
		{
			name:     "existing address first position",
			address:  "AU1test123",
			expected: 0,
			found:    true,
		},
		{
			name:     "existing address middle position",
			address:  "AU1test456",
			expected: 1,
			found:    true,
		},
		{
			name:     "existing address last position",
			address:  "AU1test789",
			expected: 2,
			found:    true,
		},
		{
			name:     "non-existing address",
			address:  "AU1nonexistent",
			expected: -1,
			found:    false,
		},
		{
			name:     "empty address",
			address:  "",
			expected: -1,
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, found := sm.getAddressIndexFromRamList(tt.address)
			if index != tt.expected {
				t.Errorf("getAddressIndexFromRamList() index = %v, expected %v", index, tt.expected)
			}
			if found != tt.found {
				t.Errorf("getAddressIndexFromRamList() found = %v, expected %v", found, tt.found)
			}
		})
	}
}

func TestRamAddressListContains(t *testing.T) {
	sm := createTestStakingManager()

	tests := []struct {
		name     string
		address  string
		expected bool
	}{
		{
			name:     "existing address",
			address:  "AU1test123",
			expected: true,
		},
		{
			name:     "another existing address",
			address:  "AU1test456",
			expected: true,
		},
		{
			name:     "non-existing address",
			address:  "AU1nonexistent",
			expected: false,
		},
		{
			name:     "empty address",
			address:  "",
			expected: false,
		},
		{
			name:     "case sensitive test",
			address:  "au1test123",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sm.ramAddressListContains(tt.address)
			if result != tt.expected {
				t.Errorf("ramAddressListContains() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestRemoveAddressFromRamList(t *testing.T) {
	tests := []struct {
		name            string
		addressToRemove string
		expectedLength  int
		stakingManager  *stakingManager
	}{
		{
			name:            "remove existing address",
			addressToRemove: "AU1test456",
			expectedLength:  2,
			stakingManager:  createTestStakingManager(),
		},
		{
			name:            "remove non-existing address",
			addressToRemove: "AU1nonexistent",
			expectedLength:  3, // Should remain unchanged
			stakingManager:  createTestStakingManager(),
		},
		{
			name:            "nil stakingAddress slice",
			addressToRemove: "AU1test123",
			expectedLength:  0,
			stakingManager:  &stakingManager{stakingAddresses: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Remove the address
			result := tt.stakingManager.removeAddressFromRamList(tt.addressToRemove)

			// Check return value
			if !result {
				t.Errorf("removeAddressFromRamList() should always return true")
			}

			// Check final length
			if len(tt.stakingManager.stakingAddresses) != tt.expectedLength {
				t.Errorf("Expected length after removal: %d, got: %d", tt.expectedLength, len(tt.stakingManager.stakingAddresses))
			}

			// Check if address still exists
			existsAfter := tt.stakingManager.ramAddressListContains(tt.addressToRemove)
			if existsAfter {
				t.Errorf("Address %s should not exist after removal", tt.addressToRemove)
			}
		})
	}
}

func TestRemoveAddressFromRamList_EmptyList(t *testing.T) {
	sm := &stakingManager{
		stakingAddresses: []StakingAddress{},
	}

	result := sm.removeAddressFromRamList("AU1test123")
	if !result {
		t.Errorf("removeAddressFromRamList() should return true even for empty list")
	}
}

func TestGetAddressesFromRamList(t *testing.T) {
	tests := []struct {
		name              string
		stakingManager    *stakingManager
		expectedLength    int
		expectedAddresses []string
	}{
		{
			name:              "get addresses from normal list",
			stakingManager:    createTestStakingManager(),
			expectedLength:    3,
			expectedAddresses: []string{"AU1test123", "AU1test456", "AU1test789"},
		},
		{
			name: "get address from single element list",
			stakingManager: &stakingManager{
				stakingAddresses: []StakingAddress{
					{
						Address: "AU1single",
					},
				},
			},
			expectedLength:    1,
			expectedAddresses: []string{"AU1single"},
		},
		{
			name: "get addresses from empty list",
			stakingManager: &stakingManager{
				stakingAddresses: []StakingAddress{},
			},
			expectedLength:    0,
			expectedAddresses: []string{},
		},
		{
			name: "get addresses from nil list",
			stakingManager: &stakingManager{
				stakingAddresses: nil,
			},
			expectedLength:    0,
			expectedAddresses: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			addresses := tt.stakingManager.getAddressesFromRamList()

			// Check length
			if len(addresses) != tt.expectedLength {
				t.Errorf("Expected %d addresses, got %d", tt.expectedLength, len(addresses))
			}

			// Check that addresses match expected
			if !slices.Equal(addresses, tt.expectedAddresses) {
				t.Errorf("Addresses mismatch.\nExpected: %v\nGot: %v", tt.expectedAddresses, addresses)
			}
		})
	}
}
