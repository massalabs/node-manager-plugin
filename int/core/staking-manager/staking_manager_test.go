package stakingManager

import (
	"fmt"
	"path/filepath"
	"testing"

	clientDriverPkg "github.com/massalabs/node-manager-plugin/int/client-driver"
	configPkg "github.com/massalabs/node-manager-plugin/int/config"
	dbPkg "github.com/massalabs/node-manager-plugin/int/db"
	nodeManagerError "github.com/massalabs/node-manager-plugin/int/error"
	nodeAPIPkg "github.com/massalabs/node-manager-plugin/int/node-api"
	"github.com/massalabs/node-manager-plugin/int/utils"
	"github.com/massalabs/station/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
		if err := logger.Close(); err != nil && !nodeManagerError.IsZapLoggerInvalidArgumentError(err) {
			t.Errorf("Failed to close logger: %v", err)
		}
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
				mockDB.On("GetRollsTarget", utils.NetworkMainnet).Return([]dbPkg.AddressInfo{
					{
						Address:    "test_address",
						RollTarget: 0,
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
				mockDB.On("DeleteRollOpHistoryByAddress", "test_address").Return(nil).Once()
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
				mockDB.On("DeleteRollOpHistoryByAddress", "test_address").Return(nil).Once()
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
			expectedError: "failed to sell the 5 candidate rolls of the address test_address. Can't remove it from staking: assert.AnError general error for testing",
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
			expectedError: "failed to remove address test_address from staking: assert.AnError general error for testing",
			// expectedError: "failed to remove address test_address from database, got following errors: failed to remove rolls target data for address test_address (mainnet) from database: ",
		},
		{
			name:     "Should return error when couldn't clean db",
			pwd:      "test_password",
			address:  "test_address",
			nodeIsUp: true,
			existingAddr: &StakingAddress{
				Address:        "test_address",
				CandidateRolls: 0,
			},
			setupMocks: func(mockClient *clientDriverPkg.MockClientDriver, mockDB *dbPkg.MockDB, t *testing.T) {
				mockClient.On("RemoveStakingAddress", "test_password", "test_address").Return(nil).Once()

				mockDB.On("DeleteRollsTarget", "test_address", utils.NetworkMainnet).Return(assert.AnError).Once()
				mockDB.On("DeleteRollOpHistoryByAddress", "test_address").Return(assert.AnError).Once()
			},
			expectedError: fmt.Sprintf(
				"%s%s%s",
				"failed to remove address test_address from database, got following errors: ",
				"failed to remove rolls target data for address test_address (mainnet) from database: assert.AnError general error for testing, ",
				"failed to remove rolls operation history for address test_address from database: assert.AnError general error for testing",
			),
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
		setupMocks    func(*dbPkg.MockDB, *MockAddressChangedDispatcher, *clientDriverPkg.MockClientDriver)
		expectedError string
	}{
		{
			name:        "Should successfully set target rolls",
			address:     "test_address",
			targetRolls: 10,
			existingAddr: &StakingAddress{
				Address:      "test_address",
				FinalBalance: 100000,
				TargetRolls:  5,
			},
			setupMocks: func(mockDB *dbPkg.MockDB, mockAddressChangedDispatcher *MockAddressChangedDispatcher, mockClient *clientDriverPkg.MockClientDriver) {
				mockDB.On("UpdateRollsTarget", "test_address", uint64(10), utils.NetworkMainnet).Return(nil).Once()
				mockAddressChangedDispatcher.On("Publish", []StakingAddress{
					{
						Address:      "test_address",
						FinalBalance: 100000,
						TargetRolls:  10,
					},
				}).Return().Once()
				mockClient.On("BuyRolls", "test_password", "test_address", uint64(10), float32(0.1)).Return("tx_hash", nil).Once()
				mockDB.On("AddRollOpHistory", "test_address", dbPkg.RollOpBuy, uint64(10), "tx_hash", utils.NetworkMainnet).Return(nil).Once()
			},
			expectedError: "",
		},
		{
			name:        "Should fail when address not found",
			address:     "non_existent_address",
			targetRolls: 10,
			setupMocks: func(mockDB *dbPkg.MockDB, mockAddressChangedDispatcher *MockAddressChangedDispatcher, mockClient *clientDriverPkg.MockClientDriver) {
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
			setupMocks: func(mockDB *dbPkg.MockDB, mockAddressChangedDispatcher *MockAddressChangedDispatcher, mockClient *clientDriverPkg.MockClientDriver) {
				// No mocks needed as the function should return early
			},
			expectedError: "",
		},
		{
			name:        "Should fallback to AddRollsTarget when UpdateRollsTarget fails",
			address:     "test_address",
			targetRolls: 10,
			existingAddr: &StakingAddress{
				Address:      "test_address",
				FinalBalance: 100000,
				TargetRolls:  5,
			},
			setupMocks: func(mockDB *dbPkg.MockDB, mockAddressChangedDispatcher *MockAddressChangedDispatcher, mockClient *clientDriverPkg.MockClientDriver) {
				mockDB.On("UpdateRollsTarget", "test_address", uint64(10), utils.NetworkMainnet).Return(
					nodeManagerError.New(nodeManagerError.ErrDBNotFoundItem, "target rolls for address test_address (mainnet) not found in database"),
				).Once()
				mockDB.On("AddRollsTarget", "test_address", uint64(10), utils.NetworkMainnet).Return(nil).Once()
				mockAddressChangedDispatcher.On("Publish", []StakingAddress{
					{
						Address:      "test_address",
						FinalBalance: 100000,
						TargetRolls:  10,
					},
				}).Return().Once()
				mockClient.On("BuyRolls", "test_password", "test_address", uint64(10), float32(0.1)).Return("tx_hash", nil).Once()
				mockDB.On("AddRollOpHistory", "test_address", dbPkg.RollOpBuy, uint64(10), "tx_hash", utils.NetworkMainnet).Return(nil).Once()
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
			setupMocks: func(mockDB *dbPkg.MockDB, mockAddressChangedDispatcher *MockAddressChangedDispatcher, mockClient *clientDriverPkg.MockClientDriver) {
				mockDB.On("UpdateRollsTarget", "test_address", uint64(10), utils.NetworkMainnet).Return(
					nodeManagerError.New(nodeManagerError.ErrDBNotFoundItem, "target rolls for address test_address (mainnet) not found in database"),
				).Once()
				mockDB.On("AddRollsTarget", "test_address", uint64(10), utils.NetworkMainnet).Return(assert.AnError).Once()
			},
			expectedError: "failed to add target rolls for address test_address (mainnet) to database: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			mockDB := dbPkg.NewMockDB(t)
			mockAddressChangedDispatcher := NewMockAddressChangedDispatcher(t)
			mockClient := clientDriverPkg.NewMockClientDriver(t)

			// Setup mocks
			tt.setupMocks(mockDB, mockAddressChangedDispatcher, mockClient)

			// Create staking manager instance
			sm := &stakingManager{
				db:                       mockDB,
				addressChangedDispatcher: mockAddressChangedDispatcher,
				clientDriver:             mockClient,
				miscellaneous: Miscellaneous{
					MinimalFees: 0.1,
					RollPrice:   100,
				},
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
			mockAddressChangedDispatcher.AssertExpectations(t)
		})
	}
}

func TestConvertToStakingAddress(t *testing.T) {
	cleanup := setupLog(t)
	defer cleanup()

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
					DeferredCredits: []DeferredCreditDtoNode{
						{
							Slot: Slot{
								Period: 100,
								Thread: 1,
							},
							Amount: "50.0",
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
					DeferredCredits:  []DeferredCreditDtoNode{},
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
					DeferredCredits:  []DeferredCreditDtoNode{},
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
					DeferredCredits:  []DeferredCreditDtoNode{},
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
					DeferredCredits:  []DeferredCreditDtoNode{},
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
			mockDB := dbPkg.NewMockDB(t)
			mockDB.On("GetRollsTarget", utils.NetworkMainnet).Return([]dbPkg.AddressInfo{}, nil).Maybe()

			// Create staking manager instance
			sm := &stakingManager{
				db: mockDB,
			}

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

func TestWithTargetRolls(t *testing.T) {
	cleanup := setupLog(t)
	defer cleanup()

	tests := []struct {
		name           string
		inputAddresses []StakingAddress
		dbAddresses    []dbPkg.AddressInfo
		isMainnet      bool
		setupMocks     func(*dbPkg.MockDB)
		expectedResult []StakingAddress
		expectedError  string
	}{
		{
			name: "Should successfully hydrate addresses with target rolls from database",
			inputAddresses: []StakingAddress{
				{
					Address:          "address1",
					FinalRolls:       10,
					CandidateRolls:   5,
					ActiveRolls:      8,
					FinalBalance:     1000.0,
					CandidateBalance: 500.0,
					Thread:           1,
					TargetRolls:      0, // Will be hydrated from DB
				},
				{
					Address:          "address2",
					FinalRolls:       20,
					CandidateRolls:   15,
					ActiveRolls:      18,
					FinalBalance:     2000.0,
					CandidateBalance: 1500.0,
					Thread:           2,
					TargetRolls:      0, // Will be hydrated from DB
				},
			},
			dbAddresses: []dbPkg.AddressInfo{
				{
					Address:    "address1",
					RollTarget: 15,
					Network:    "mainnet",
				},
				{
					Address:    "address2",
					RollTarget: 25,
					Network:    "mainnet",
				},
			},
			isMainnet: true,
			setupMocks: func(mockDB *dbPkg.MockDB) {
				mockDB.On("GetRollsTarget", utils.NetworkMainnet).Return([]dbPkg.AddressInfo{
					{
						Address:    "address1",
						RollTarget: 15,
						Network:    "mainnet",
					},
					{
						Address:    "address2",
						RollTarget: 25,
						Network:    "mainnet",
					},
				}, nil).Once()
			},
			expectedResult: []StakingAddress{
				{
					Address:          "address1",
					FinalRolls:       10,
					CandidateRolls:   5,
					ActiveRolls:      8,
					FinalBalance:     1000.0,
					CandidateBalance: 500.0,
					Thread:           1,
					TargetRolls:      15, // Hydrated from DB
				},
				{
					Address:          "address2",
					FinalRolls:       20,
					CandidateRolls:   15,
					ActiveRolls:      18,
					FinalBalance:     2000.0,
					CandidateBalance: 1500.0,
					Thread:           2,
					TargetRolls:      25, // Hydrated from DB
				},
			},
			expectedError: "",
		},
		{
			name: "Should handle buildnet network correctly",
			inputAddresses: []StakingAddress{
				{
					Address:          "address1",
					FinalRolls:       10,
					CandidateRolls:   5,
					ActiveRolls:      8,
					FinalBalance:     1000.0,
					CandidateBalance: 500.0,
					Thread:           1,
					TargetRolls:      0,
				},
			},
			dbAddresses: []dbPkg.AddressInfo{
				{
					Address:    "address1",
					RollTarget: 15,
					Network:    "buildnet",
				},
			},
			isMainnet: false,
			setupMocks: func(mockDB *dbPkg.MockDB) {
				mockDB.On("GetRollsTarget", utils.NetworkBuildnet).Return([]dbPkg.AddressInfo{
					{
						Address:    "address1",
						RollTarget: 15,
						Network:    "buildnet",
					},
				}, nil).Once()
			},
			expectedResult: []StakingAddress{
				{
					Address:          "address1",
					FinalRolls:       10,
					CandidateRolls:   5,
					ActiveRolls:      8,
					FinalBalance:     1000.0,
					CandidateBalance: 500.0,
					Thread:           1,
					TargetRolls:      15,
				},
			},
			expectedError: "",
		},
		{
			name: "Should fail when database GetRollsTarget fails",
			inputAddresses: []StakingAddress{
				{
					Address:          "address1",
					FinalRolls:       10,
					CandidateRolls:   5,
					ActiveRolls:      8,
					FinalBalance:     1000.0,
					CandidateBalance: 500.0,
					Thread:           1,
					TargetRolls:      0,
				},
			},
			isMainnet: true,
			setupMocks: func(mockDB *dbPkg.MockDB) {
				mockDB.On("GetRollsTarget", utils.NetworkMainnet).Return(nil, assert.AnError).Once()
			},
			expectedResult: nil,
			expectedError:  "failed to load rolul targets from database",
		},
		{
			name:           "Should handle empty input addresses",
			inputAddresses: []StakingAddress{},
			dbAddresses: []dbPkg.AddressInfo{
				{
					Address:    "orphaned_address",
					RollTarget: 20,
					Network:    "mainnet",
				},
			},
			isMainnet: true,
			setupMocks: func(mockDB *dbPkg.MockDB) {
				mockDB.On("GetRollsTarget", utils.NetworkMainnet).Return([]dbPkg.AddressInfo{
					{
						Address:    "orphaned_address",
						RollTarget: 20,
						Network:    "mainnet",
					},
				}, nil).Once()
			},
			expectedResult: []StakingAddress{},
			expectedError:  "",
		},
		{
			name: "Should handle empty database addresses",
			inputAddresses: []StakingAddress{
				{
					Address:          "address1",
					FinalRolls:       10,
					CandidateRolls:   5,
					ActiveRolls:      8,
					FinalBalance:     1000.0,
					CandidateBalance: 500.0,
					Thread:           1,
					TargetRolls:      0,
				},
			},
			dbAddresses: []dbPkg.AddressInfo{},
			isMainnet:   true,
			setupMocks: func(mockDB *dbPkg.MockDB) {
				mockDB.On("GetRollsTarget", utils.NetworkMainnet).Return([]dbPkg.AddressInfo{}, nil).Once()
			},
			expectedResult: []StakingAddress{
				{
					Address:          "address1",
					FinalRolls:       10,
					CandidateRolls:   5,
					ActiveRolls:      8,
					FinalBalance:     1000.0,
					CandidateBalance: 500.0,
					Thread:           1,
					TargetRolls:      0, // Should remain 0 since no DB entry
				},
			},
			expectedError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock
			mockDB := dbPkg.NewMockDB(t)

			// Setup mocks
			tt.setupMocks(mockDB)

			// Create staking manager instance
			sm := &stakingManager{
				db: mockDB,
			}

			// Set global config for network
			configPkg.GlobalPluginInfo.IsMainnet = tt.isMainnet

			// Execute the function under test
			result, err := sm.WithTargetRolls(tt.inputAddresses)

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
						assert.Equal(t, expected.TargetRolls, result[i].TargetRolls)
						assert.Equal(t, len(expected.DeferredCredits), len(result[i].DeferredCredits))
					}
				}
			}

			// Assert that all expected calls were made
			mockDB.AssertExpectations(t)
		})
	}
}
