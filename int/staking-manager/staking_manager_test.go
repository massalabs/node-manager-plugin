package stakingManager

import (
	"path/filepath"
	"testing"

	clientDriverPkg "github.com/massalabs/node-manager-plugin/int/client-driver"
	configPkg "github.com/massalabs/node-manager-plugin/int/config"
	dbPkg "github.com/massalabs/node-manager-plugin/int/db"
	nodeAPIPkg "github.com/massalabs/node-manager-plugin/int/node-api"
	"github.com/massalabs/node-manager-plugin/int/utils"
	"github.com/massalabs/station/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMassaWalletManager is a mock implementation of MassaWalletManager for testing
type MockMassaWalletManager struct {
	mock.Mock
}

func (m *MockMassaWalletManager) GetPrivateKeyFromNickname(pwd, nickname string) (string, string, error) {
	args := m.Called(pwd, nickname)
	return args.String(0), args.String(1), args.Error(2)
}

// setupLog initializes the logger for testing and returns a cleanup function
func setupLog(t *testing.T) func() {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")
	err := logger.InitializeGlobal(logPath)
	if err != nil {
		t.Fatalf("failed to initialize logger: %v", err)
	}

	// Initialize global config for testing
	configPkg.GlobalPluginInfo = &configPkg.PluginInfo{
		IsMainnet:  true,
		PwdMainnet: "test_password",
	}

	return func() {
		logger.Close()
	}
}

func TestAddStakingAddress(t *testing.T) {
	cleanup := setupLog(t)
	defer cleanup()

	tests := []struct {
		name          string
		pwdNode       string
		pwdAccount    string
		nickname      string
		nodeIsUp      bool
		existingAddr  *StakingAddress
		setupMocks    func(*clientDriverPkg.MockClientDriver, *dbPkg.MockDB, *nodeAPIPkg.MockNodeAPI, *MockMassaWalletManager, *testing.T)
		expectedError string
	}{
		{
			name:       "Should fail when node is not up",
			pwdNode:    "node_password",
			pwdAccount: "account_password",
			nickname:   "test_wallet",
			nodeIsUp:   false,
			setupMocks: func(mockClient *clientDriverPkg.MockClientDriver, mockDB *dbPkg.MockDB, mockNodeAPI *nodeAPIPkg.MockNodeAPI, mockWalletManager *MockMassaWalletManager, t *testing.T) {
				// No mocks needed as the function should fail early
			},
			expectedError: "massa node is not up",
		},
		{
			name:       "Should fail when address already exists in staking",
			pwdNode:    "node_password",
			pwdAccount: "account_password",
			nickname:   "test_wallet",
			nodeIsUp:   true,
			existingAddr: &StakingAddress{
				Address: "test_address",
			},
			setupMocks: func(mockClient *clientDriverPkg.MockClientDriver, mockDB *dbPkg.MockDB, mockNodeAPI *nodeAPIPkg.MockNodeAPI, mockWalletManager *MockMassaWalletManager, t *testing.T) {
				mockWalletManager.On("GetPrivateKeyFromNickname", "account_password", "test_wallet").Return("test_private_key", "test_address", nil).Once()
				mockClient.AssertNotCalled(t, "AddStakingAddress")
				mockDB.AssertNotCalled(t, "AddRollsTarget")
			},
			expectedError: "address test_address already in staking addresses",
		},
		{
			name:       "Should fail when GetPrivateKeyFromNickname fails",
			pwdNode:    "node_password",
			pwdAccount: "account_password",
			nickname:   "non_existent_wallet",
			nodeIsUp:   true,
			setupMocks: func(mockClient *clientDriverPkg.MockClientDriver, mockDB *dbPkg.MockDB, mockNodeAPI *nodeAPIPkg.MockNodeAPI, mockWalletManager *MockMassaWalletManager, t *testing.T) {
				mockWalletManager.On("GetPrivateKeyFromNickname", "account_password", "non_existent_wallet").Return("", "", assert.AnError).Once()
			},
			expectedError: "failed to get address and priv key from nickname non_existent_wallet",
		},
		{
			name:       "Should successfully add staking address",
			pwdNode:    "node_password",
			pwdAccount: "account_password",
			nickname:   "test_wallet",
			nodeIsUp:   true,
			setupMocks: func(mockClient *clientDriverPkg.MockClientDriver, mockDB *dbPkg.MockDB, mockNodeAPI *nodeAPIPkg.MockNodeAPI, mockWalletManager *MockMassaWalletManager, t *testing.T) {
				mockWalletManager.On("GetPrivateKeyFromNickname", "account_password", "test_wallet").Return("test_private_key", "test_address", nil).Once()
				mockClient.On("AddStakingAddress", "node_password", "test_private_key", "test_address").Return(nil).Once()
				mockDB.On("AddRollsTarget", "test_address", uint64(0), utils.NetworkMainnet).Return(nil).Once()
				mockNodeAPI.On("GetAddresses", []string{"test_address"}).Return([]byte(`[{"address":"test_address","final_roll_count":0,"candidate_roll_count":0,"final_balance":"100.0","candidate_balance":"100.0","thread":0,"deferred_credits":[]}]`), nil).Once()
				mockClient.On("WalletInfo", mock.Anything).Return(map[string]clientDriverPkg.WalletInfo{
					"test_address": {
						AddressInfo: clientDriverPkg.AddressInfo{
							ActiveRolls: 0,
						},
					},
				}, nil).Once()
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockClient := clientDriverPkg.NewMockClientDriver(t)
			mockDB := dbPkg.NewMockDB(t)
			mockNodeAPI := nodeAPIPkg.NewMockNodeAPI(t)
			mockWalletManager := &MockMassaWalletManager{}

			// Setup mocks
			tt.setupMocks(mockClient, mockDB, mockNodeAPI, mockWalletManager, t)

			// Create staking manager instance
			sm := &stakingManager{
				clientDriver:     mockClient,
				nodeAPI:          mockNodeAPI,
				db:               mockDB,
				nodeIsUp:         tt.nodeIsUp,
				stakingAddresses: []StakingAddress{},
				walletManager:    mockWalletManager,
			}

			// Add existing address if provided
			if tt.existingAddr != nil {
				sm.stakingAddresses = []StakingAddress{*tt.existingAddr}
			}

			// Execute the function under test
			result, err := sm.AddStakingAddress(tt.pwdNode, tt.pwdAccount, tt.nickname)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}

			// Assert that all expected calls were made
			mockClient.AssertExpectations(t)
			mockDB.AssertExpectations(t)
			mockNodeAPI.AssertExpectations(t)
			mockWalletManager.AssertExpectations(t)
		})
	}
}

func TestRemoveStakingAddress(t *testing.T) {
	cleanup := setupLog(t)
	defer cleanup()

	tests := []struct {
		name          string
		pwd           string
		address       string
		nodeIsUp      bool
		existingAddr  *StakingAddress
		setupMocks    func(*clientDriverPkg.MockClientDriver, *dbPkg.MockDB, *testing.T)
		expectedError string
	}{
		{
			name:     "Should successfully remove staking address",
			pwd:      "test_password",
			address:  "test_address",
			nodeIsUp: true,
			existingAddr: &StakingAddress{
				Address:        "test_address",
				CandidateRolls: 0,
			},
			setupMocks: func(mockClient *clientDriverPkg.MockClientDriver, mockDB *dbPkg.MockDB, t *testing.T) {
				mockClient.On("RemoveStakingAddress", "test_password", "test_address").Return(nil).Once()
				mockDB.On("DeleteRollsTarget", "test_address", utils.NetworkMainnet).Return(nil).Once()
			},
			expectedError: "",
		},
		{
			name:     "Should fail when node is not up",
			pwd:      "test_password",
			address:  "test_address",
			nodeIsUp: false,
			setupMocks: func(mockClient *clientDriverPkg.MockClientDriver, mockDB *dbPkg.MockDB, t *testing.T) {
				// No mocks needed as the function should fail early
			},
			expectedError: "massa node is not up",
		},
		{
			name:     "Should fail when address not found in staking addresses",
			pwd:      "test_password",
			address:  "non_existent_address",
			nodeIsUp: true,
			setupMocks: func(mockClient *clientDriverPkg.MockClientDriver, mockDB *dbPkg.MockDB, t *testing.T) {
				// No mocks needed as the function should fail early
			},
			expectedError: "address non_existent_address not found in staking addresses",
		},
		{
			name:     "Should sell candidate rolls before removing address",
			pwd:      "test_password",
			address:  "test_address",
			nodeIsUp: true,
			existingAddr: &StakingAddress{
				Address:        "test_address",
				CandidateRolls: 5,
			},
			setupMocks: func(mockClient *clientDriverPkg.MockClientDriver, mockDB *dbPkg.MockDB, t *testing.T) {
				mockClient.On("SellRolls", "test_password", "test_address", uint64(5), float32(0.1)).Return("tx_hash", nil).Once()
				mockClient.On("RemoveStakingAddress", "test_password", "test_address").Return(nil).Once()
				mockDB.On("DeleteRollsTarget", "test_address", utils.NetworkMainnet).Return(nil).Once()
			},
			expectedError: "",
		},
		{
			name:     "Should fail when selling candidate rolls fails",
			pwd:      "test_password",
			address:  "test_address",
			nodeIsUp: true,
			existingAddr: &StakingAddress{
				Address:        "test_address",
				CandidateRolls: 5,
			},
			setupMocks: func(mockClient *clientDriverPkg.MockClientDriver, mockDB *dbPkg.MockDB, t *testing.T) {
				mockClient.On("SellRolls", "test_password", "test_address", uint64(5), float32(0.1)).Return("", assert.AnError).Once()
			},
			expectedError: "failed to sell the 5 candidate rolls of the address test_address",
		},
		{
			name:     "Should fail when RemoveStakingAddress returns error",
			pwd:      "test_password",
			address:  "test_address",
			nodeIsUp: true,
			existingAddr: &StakingAddress{
				Address:        "test_address",
				CandidateRolls: 0,
			},
			setupMocks: func(mockClient *clientDriverPkg.MockClientDriver, mockDB *dbPkg.MockDB, t *testing.T) {
				mockClient.On("RemoveStakingAddress", "test_password", "test_address").Return(assert.AnError).Once()
			},
			expectedError: "failed to remove address test_address from staking",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockClient := clientDriverPkg.NewMockClientDriver(t)
			mockDB := dbPkg.NewMockDB(t)

			// Setup mocks
			tt.setupMocks(mockClient, mockDB, t)

			// Create staking manager instance
			sm := &stakingManager{
				clientDriver: mockClient,
				db:           mockDB,
				nodeIsUp:     tt.nodeIsUp,
				miscellaneous: Miscellaneous{
					MinimalFees: 0.1,
				},
			}

			// Add existing address if provided
			if tt.existingAddr != nil {
				sm.stakingAddresses = []StakingAddress{*tt.existingAddr}
			}

			// Execute the function under test
			err := sm.RemoveStakingAddress(tt.pwd, tt.address)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Assert that all expected calls were made
			mockClient.AssertExpectations(t)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestSetTargetRolls(t *testing.T) {
	cleanup := setupLog(t)
	defer cleanup()

	tests := []struct {
		name          string
		address       string
		targetRolls   uint64
		existingAddr  *StakingAddress
		setupMocks    func(*dbPkg.MockDB, *testing.T)
		expectedError string
	}{
		{
			name:        "Should successfully set target rolls",
			address:     "test_address",
			targetRolls: 10,
			existingAddr: &StakingAddress{
				Address:     "test_address",
				TargetRolls: 5,
			},
			setupMocks: func(mockDB *dbPkg.MockDB, t *testing.T) {
				mockDB.On("UpdateRollsTarget", "test_address", uint64(10), utils.NetworkMainnet).Return(nil).Once()
			},
			expectedError: "",
		},
		{
			name:        "Should fail when address not found",
			address:     "non_existent_address",
			targetRolls: 10,
			setupMocks: func(mockDB *dbPkg.MockDB, t *testing.T) {
				// No mocks needed as the function should fail early
			},
			expectedError: "address not found for address non_existent_address",
		},
		{
			name:        "Should not update when target rolls is the same",
			address:     "test_address",
			targetRolls: 5,
			existingAddr: &StakingAddress{
				Address:     "test_address",
				TargetRolls: 5,
			},
			setupMocks: func(mockDB *dbPkg.MockDB, t *testing.T) {
				// No mocks needed as the function should return early
			},
			expectedError: "",
		},
		{
			name:        "Should fallback to AddRollsTarget when UpdateRollsTarget fails",
			address:     "test_address",
			targetRolls: 10,
			existingAddr: &StakingAddress{
				Address:     "test_address",
				TargetRolls: 5,
			},
			setupMocks: func(mockDB *dbPkg.MockDB, t *testing.T) {
				mockDB.On("UpdateRollsTarget", "test_address", uint64(10), utils.NetworkMainnet).Return(assert.AnError).Once()
				mockDB.On("AddRollsTarget", "test_address", uint64(10), utils.NetworkMainnet).Return(nil).Once()
			},
			expectedError: "",
		},
		{
			name:        "Should fail when both UpdateRollsTarget and AddRollsTarget fail",
			address:     "test_address",
			targetRolls: 10,
			existingAddr: &StakingAddress{
				Address:     "test_address",
				TargetRolls: 5,
			},
			setupMocks: func(mockDB *dbPkg.MockDB, t *testing.T) {
				mockDB.On("UpdateRollsTarget", "test_address", uint64(10), utils.NetworkMainnet).Return(assert.AnError).Once()
				mockDB.On("AddRollsTarget", "test_address", uint64(10), utils.NetworkMainnet).Return(assert.AnError).Once()
			},
			expectedError: "failed to update database",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockDB := dbPkg.NewMockDB(t)

			// Setup mocks
			tt.setupMocks(mockDB, t)

			// Create staking manager instance
			sm := &stakingManager{
				db: mockDB,
			}

			// Add existing address if provided
			if tt.existingAddr != nil {
				sm.stakingAddresses = []StakingAddress{*tt.existingAddr}
			}

			// Execute the function under test
			err := sm.SetTargetRolls(tt.address, tt.targetRolls)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			// Assert that all expected calls were made
			mockDB.AssertExpectations(t)
		})
	}
}

func TestConvertToStakingAddress(t *testing.T) {
	tests := []struct {
		name           string
		addresses      []getAddressesResponse
		walletInfos    map[string]clientDriverPkg.WalletInfo
		expectedResult []StakingAddress
		expectedError  string
	}{
		{
			name: "Should successfully convert addresses",
			addresses: []getAddressesResponse{
				{
					Address:          "test_address_1",
					FinalRolls:       10,
					CandidateRolls:   5,
					FinalBalance:     "1000.5",
					CandidateBalance: "500.25",
					Thread:           1,
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
				{
					Address:          "test_address_2",
					FinalRolls:       20,
					CandidateRolls:   15,
					FinalBalance:     "2000.75",
					CandidateBalance: "1500.50",
					Thread:           2,
					DeferredCredits:  []DeferredCredit{},
				},
			},
			walletInfos: map[string]clientDriverPkg.WalletInfo{
				"test_address_1": {
					AddressInfo: clientDriverPkg.AddressInfo{
						ActiveRolls: 8,
					},
				},
				"test_address_2": {
					AddressInfo: clientDriverPkg.AddressInfo{
						ActiveRolls: 18,
					},
				},
			},
			expectedResult: []StakingAddress{
				{
					Address:          "test_address_1",
					FinalRolls:       10,
					CandidateRolls:   5,
					ActiveRolls:      8,
					FinalBalance:     1000.5,
					CandidateBalance: 500.25,
					Thread:           1,
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
				{
					Address:          "test_address_2",
					FinalRolls:       20,
					CandidateRolls:   15,
					ActiveRolls:      18,
					FinalBalance:     2000.75,
					CandidateBalance: 1500.50,
					Thread:           2,
					DeferredCredits:  []DeferredCredit{},
				},
			},
			expectedError: "",
		},
		{
			name: "Should fail when final balance is invalid",
			addresses: []getAddressesResponse{
				{
					Address:          "test_address_1",
					FinalRolls:       10,
					CandidateRolls:   5,
					FinalBalance:     "invalid_balance",
					CandidateBalance: "500.25",
					Thread:           1,
					DeferredCredits:  []DeferredCredit{},
				},
			},
			walletInfos: map[string]clientDriverPkg.WalletInfo{
				"test_address_1": {
					AddressInfo: clientDriverPkg.AddressInfo{
						ActiveRolls: 8,
					},
				},
			},
			expectedError: "failed to parse final balance",
		},
		{
			name: "Should fail when candidate balance is invalid",
			addresses: []getAddressesResponse{
				{
					Address:          "test_address_1",
					FinalRolls:       10,
					CandidateRolls:   5,
					FinalBalance:     "1000.5",
					CandidateBalance: "invalid_balance",
					Thread:           1,
					DeferredCredits:  []DeferredCredit{},
				},
			},
			walletInfos: map[string]clientDriverPkg.WalletInfo{
				"test_address_1": {
					AddressInfo: clientDriverPkg.AddressInfo{
						ActiveRolls: 8,
					},
				},
			},
			expectedError: "failed to parse candidate balance",
		},
		{
			name:           "Should handle empty addresses list",
			addresses:      []getAddressesResponse{},
			walletInfos:    map[string]clientDriverPkg.WalletInfo{},
			expectedResult: []StakingAddress{},
			expectedError:  "",
		},
		{
			name: "Should handle address not found in wallet infos",
			addresses: []getAddressesResponse{
				{
					Address:          "test_address_1",
					FinalRolls:       10,
					CandidateRolls:   5,
					FinalBalance:     "1000.5",
					CandidateBalance: "500.25",
					Thread:           1,
					DeferredCredits:  []DeferredCredit{},
				},
			},
			walletInfos: map[string]clientDriverPkg.WalletInfo{
				// Empty wallet infos - address not found
			},
			expectedResult: []StakingAddress{
				{
					Address:          "test_address_1",
					FinalRolls:       10,
					CandidateRolls:   5,
					ActiveRolls:      0, // Default value when not found
					FinalBalance:     1000.5,
					CandidateBalance: 500.25,
					Thread:           1,
					DeferredCredits:  []DeferredCredit{},
				},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create staking manager instance
			sm := &stakingManager{}

			// Execute the function under test
			result, err := sm.convertToStakingAddress(tt.addresses, tt.walletInfos)

			// Assert results
			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.expectedResult), len(result))

				for i, expected := range tt.expectedResult {
					if i < len(result) {
						assert.Equal(t, expected.Address, result[i].Address)
						assert.Equal(t, expected.FinalRolls, result[i].FinalRolls)
						assert.Equal(t, expected.CandidateRolls, result[i].CandidateRolls)
						assert.Equal(t, expected.ActiveRolls, result[i].ActiveRolls)
						assert.Equal(t, expected.FinalBalance, result[i].FinalBalance)
						assert.Equal(t, expected.CandidateBalance, result[i].CandidateBalance)
						assert.Equal(t, expected.Thread, result[i].Thread)
						assert.Equal(t, len(expected.DeferredCredits), len(result[i].DeferredCredits))

						// Check deferred credits if they exist
						if len(expected.DeferredCredits) > 0 {
							assert.Equal(t, expected.DeferredCredits[0].Amount, result[i].DeferredCredits[0].Amount)
							assert.Equal(t, expected.DeferredCredits[0].Slot.Period, result[i].DeferredCredits[0].Slot.Period)
							assert.Equal(t, expected.DeferredCredits[0].Slot.Thread, result[i].DeferredCredits[0].Slot.Thread)
						}
					}
				}
			}
		})
	}
}
