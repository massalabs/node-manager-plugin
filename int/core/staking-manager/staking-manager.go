package stakingManager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	clientDriverPkg "github.com/massalabs/node-manager-plugin/int/client-driver"
	"github.com/massalabs/node-manager-plugin/int/config"
	nodeStatusPkg "github.com/massalabs/node-manager-plugin/int/core/NodeStatus"
	dbPkg "github.com/massalabs/node-manager-plugin/int/db"
	nodeManagerError "github.com/massalabs/node-manager-plugin/int/error"
	nodeAPI "github.com/massalabs/node-manager-plugin/int/node-api"
	nodeDirManagerPkg "github.com/massalabs/node-manager-plugin/int/node-bin-dir-manager"
	"github.com/massalabs/node-manager-plugin/int/utils"
	"github.com/massalabs/station/pkg/logger"
)

type Slot struct {
	Period uint64 `json:"period"`
	Thread uint8  `json:"thread"`
}

type DeferredCreditDtoNode struct {
	Slot   Slot   `json:"slot"`
	Amount string `json:"amount"`
}

type DeferredCredit struct {
	Slot   Slot    `json:"slot"`
	Amount float64 `json:"amount"`
}

// Miscellaneous contains various node-related data
type Miscellaneous struct {
	MinimalFees float32 `json:"minimal_fees"`
	RollPrice   float32 `json:"roll_price"`
}

type getAddressesResponse struct {
	Address          string                  `json:"address"`
	FinalRolls       uint64                  `json:"final_roll_count"`
	CandidateRolls   uint64                  `json:"candidate_roll_count"`
	FinalBalance     string                  `json:"final_balance"`
	CandidateBalance string                  `json:"candidate_balance"`
	Thread           uint8                   `json:"thread"`
	DeferredCredits  []DeferredCreditDtoNode `json:"deferred_credits"`
}

// StakingAddress represents a staking address with its information
type StakingAddress struct {
	Address            string           `json:"address"`
	FinalRolls         uint64           `json:"final_roll_count"`
	CandidateRolls     uint64           `json:"candidate_roll_count"`
	ActiveRolls        uint64           `json:"active_roll_count"`
	FinalBalance       float64          `json:"final_balance"`
	CandidateBalance   float64          `json:"candidate_balance"`
	Thread             uint8            `json:"thread"`
	DeferredCredits    []DeferredCredit `json:"deferred_credits"`
	TargetRolls        int64            `json:"target_rolls"` // if target rolls is negative, it means that the address is auto-compounding: buy as many rolls as possible
	pendingOperationId *string
}

type StakingManager interface {
	GetStakingAddresses(pwd string) ([]StakingAddress, AddressChangedDispatcher, error)
	AddStakingAddress(pwdNode, pwdAccount, nickname string) (StakingAddress, error)
	RemoveStakingAddress(pwd, address string) error
	SetTargetRolls(address string, targetRolls int64) error
	Close() error
}

type stakingManager struct {
	mu                             sync.Mutex
	muSellBuyRolls                 sync.Mutex
	nodeIsUp                       bool
	stakingAddresses               []StakingAddress
	clientDriver                   clientDriverPkg.ClientDriver
	nodeAPI                        nodeAPI.NodeAPI
	nodeStatusDispatcher           nodeStatusPkg.NodeStatusDispatcher
	addressChangedDispatcher       AddressChangedDispatcher
	stakingAddressDataPollInterval uint64
	miscellaneous                  Miscellaneous
	stopStakingMonitoringFunc      func()
	closeStakingManagerAsyncFunc   func()
	db                             dbPkg.DB
	nodeDirManager                 nodeDirManagerPkg.NodeDirManager
	clientTimeout                  uint64
	walletManager                  MassaWalletManager
	config                         *config.PluginConfig
}

func NewStakingManager(
	nodeAPI nodeAPI.NodeAPI,
	nodeStatusDispatcher nodeStatusPkg.NodeStatusDispatcher,
	database dbPkg.DB,
	nodeDirManager nodeDirManagerPkg.NodeDirManager,
	stakingAddressDataPollInterval uint64,
	clientTimeout uint64,
	walletManager MassaWalletManager,
	config *config.PluginConfig,
) StakingManager {
	sm := &stakingManager{
		nodeAPI:                        nodeAPI,
		nodeStatusDispatcher:           nodeStatusDispatcher,
		addressChangedDispatcher:       NewAddressChangedDispatcher(),
		stakingAddressDataPollInterval: stakingAddressDataPollInterval,
		miscellaneous:                  Miscellaneous{},
		db:                             database,
		nodeDirManager:                 nodeDirManager,
		clientTimeout:                  clientTimeout,
		walletManager:                  walletManager,
		config:                         config,
	}

	ctx, cancel := context.WithCancel(context.Background())
	sm.closeStakingManagerAsyncFunc = cancel
	go sm.asyncTask(ctx)

	return sm
}

func (s *stakingManager) GetStakingAddresses(pwd string) ([]StakingAddress, AddressChangedDispatcher, error) {
	if !s.nodeIsUp {
		return nil, nil, fmt.Errorf("massa node is not up")
	}

	if len(s.stakingAddresses) == 0 {
		if err := s.initStakingAddresses(); err != nil {
			return nil, nil, err
		}
	}

	return copyAddresses(s.stakingAddresses), s.addressChangedDispatcher, nil
}

// AddStakingAddress add an address to the massa node for staking
func (s *stakingManager) AddStakingAddress(pwdNode, pwdAccount, nickname string) (StakingAddress, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.nodeIsUp {
		return StakingAddress{}, fmt.Errorf("massa node is not up")
	}

	privateKey, address, err := s.walletManager.GetPrivateKeyFromNickname(pwdAccount, nickname)
	if err != nil {
		return StakingAddress{}, fmt.Errorf("failed to get address and priv key from nickname %s: %w", nickname, err)
	}

	if s.ramAddressListContains(address) {
		return StakingAddress{}, fmt.Errorf("address %s already in staking addresses", address)
	}

	// add address to node staking addresses
	err = s.clientDriver.AddStakingAddress(pwdNode, privateKey, address)
	if err != nil {
		return StakingAddress{}, fmt.Errorf("failed to add address %s to node staking addresses: %w", address, err)
	}

	// Add to database
	currentNetwork := utils.NetworkMainnet
	if !config.GlobalPluginInfo.GetIsMainnet() {
		currentNetwork = utils.NetworkBuildnet
	}

	// Default to -1 rolls target (auto-compounding) -> buy as many rolls as possible
	if err := s.db.AddRollsTarget(address, -1, currentNetwork); err != nil {
		return StakingAddress{}, fmt.Errorf("address added to node staking addresses but failed to add address to rolls_target table in local database: %w", err)
	}

	// get address data from node
	addressData, err := s.getAddressesDataFromNode([]string{address})
	if err != nil {
		return StakingAddress{}, fmt.Errorf("address added to node staking addresses but failed to get address data from node to update ram list: %w", err)
	}

	// add address to staking addresses list in ram
	s.stakingAddresses = append(s.stakingAddresses, addressData[0])

	return addressData[0], nil
}

// RemoveStakingAddress remove an address from the massa node. The address will be removed from staking.
func (s *stakingManager) RemoveStakingAddress(pwd, address string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.nodeIsUp {
		return fmt.Errorf("massa node is not up")
	}

	// check if the address to remove exist in staking addresses
	index, ok := s.getAddressIndexFromRamList(address)
	if !ok {
		return fmt.Errorf("address %s not found in staking addresses", address)
	}

	// sell rolls if there are any
	if s.stakingAddresses[index].CandidateRolls > 0 {
		logger.Info("address %s has %d candidate rolls. Must sell them before removing from staking", address, s.stakingAddresses[index].CandidateRolls)
		_, err := s.clientDriver.SellRolls(pwd, address, s.stakingAddresses[index].CandidateRolls, s.miscellaneous.MinimalFees)
		if err != nil {
			return fmt.Errorf("failed to sell the %d candidate rolls of the address %s. Can't remove it from staking: %w", s.stakingAddresses[index].CandidateRolls, address, err)
		}
	}

	// remove address from massa-node and massa-client
	err := s.clientDriver.RemoveStakingAddress(pwd, address)
	if err != nil {
		return fmt.Errorf("failed to remove address %s from staking: %w", address, err)
	}

	// remove address from staking addresses list
	s.removeAddressFromRamList(address)

	// Remove from database if available
	currentNetwork := utils.NetworkMainnet
	if !config.GlobalPluginInfo.GetIsMainnet() {
		currentNetwork = utils.NetworkBuildnet
	}

	errs := []error{}

	if err := s.db.DeleteRollsTarget(address, currentNetwork); err != nil {
		if nodeManagerError.Is(err, nodeManagerError.ErrDBNotFoundItem) {
			logger.Info("[RemoveStakingAddress] target rolls for address %s and network %s is not in database. Nothing to remove from DB", address, string(currentNetwork))
		} else {
			errs = append(errs, fmt.Errorf("failed to remove rolls target data for address %s (%s) from database: %w", address, string(currentNetwork), err))
		}
	}

	// Remove roll operation history from database
	if err := s.db.DeleteRollOpHistoryByAddress(address); err != nil {
		if nodeManagerError.Is(err, nodeManagerError.ErrDBNotFoundItem) {
			logger.Info("[RemoveStakingAddress] roll operation history for address %s is not in database. Nothing to remove from DB", address)
		} else {
			errs = append(errs, fmt.Errorf("failed to remove rolls operation history for address %s from database: %w", address, err))
		}
	}

	if len(errs) > 0 {
		errMsg := fmt.Sprintf("failed to remove address %s from database, got following errors: ", address)
		for i, err := range errs {
			errMsg += err.Error()
			if i < len(errs)-1 {
				errMsg += ", "
			}
		}
		return errors.New(errMsg)
	}

	return nil
}

// SetTargetRolls sets the target rolls for a staking address
func (s *stakingManager) SetTargetRolls(address string, targetRolls int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	index, ok := s.getAddressIndexFromRamList(address)
	if !ok {
		return fmt.Errorf("address not found for address %s", address)
	}

	if s.stakingAddresses[index].TargetRolls == targetRolls {
		return nil
	}

	// Update database
	currentNetwork := config.GlobalPluginInfo.GetNetwork()

	err := s.db.UpdateRollsTarget(address, targetRolls, currentNetwork)
	if err != nil {
		if nodeManagerError.Is(err, nodeManagerError.ErrDBNotFoundItem) {
			logger.Info("[SetTargetRolls] target rolls for address %s (%s) not found in database. Adding it", address, string(currentNetwork))

			// If database update fails, try to add the record
			if err = s.db.AddRollsTarget(address, targetRolls, currentNetwork); err != nil {
				return fmt.Errorf("failed to add target rolls for address %s (%s) to database: %w", address, string(currentNetwork), err)
			}
		} else {
			return fmt.Errorf("failed to update target rolls for address %s (%s) in database: %w", address, string(currentNetwork), err)
		}
	}

	// update target rolls in ram list
	s.stakingAddresses[index].TargetRolls = targetRolls

	// publish the new staking addresses list to the front
	s.addressChangedDispatcher.Publish(s.stakingAddresses)

	// sell or buy rolls for the address according to new target rolls
	s.muSellBuyRolls.Lock()
	defer s.muSellBuyRolls.Unlock()

	err = s.sellBuyRollsAddress(s.stakingAddresses[index])
	if err != nil {
		return fmt.Errorf("failed to sell or buy rolls for address %s: %w", address, err)
	}

	return nil
}

// Close closes the staking manager and its database connection
func (s *stakingManager) Close() error {
	if s.closeStakingManagerAsyncFunc != nil {
		s.closeStakingManagerAsyncFunc()
	}

	if s.db != nil {
		return s.db.Close()
	}

	return nil
}

func (s *stakingManager) asyncTask(ctx context.Context) {
	// when node is up
	statusOnChan, _ := s.nodeStatusDispatcher.Subscribe([]nodeStatusPkg.NodeStatus{nodeStatusPkg.NodeStatusOn}, "staking-manager-status-on")

	// when node is down and can no more be used for staking
	nodeDownChan, _ := s.nodeStatusDispatcher.Subscribe(
		[]nodeStatusPkg.NodeStatus{ // don't listen to off status because it can only be active after NodeStatusStopping or NodeStatusDesynced
			nodeStatusPkg.NodeStatusCrashed,
			nodeStatusPkg.NodeStatusDesynced,
			nodeStatusPkg.NodeStatusStopping,
		},
		"staking-manager-node-down",
	)

	for {
		select {
		case <-ctx.Done():
			logger.Debugf("staking manager async task stopped: %v", ctx.Err())
			return

		case <-statusOnChan:
			s.mu.Lock()
			s.nodeIsUp = true
			s.mu.Unlock()

			clientDriver, err := clientDriverPkg.NewClientDriver(
				config.GlobalPluginInfo.GetIsMainnet(),
				s.nodeDirManager,
				time.Duration(s.clientTimeout)*time.Second,
			)
			if err != nil {
				logger.Error("failed to create client driver: %v", err)
				continue
			}

			s.clientDriver = clientDriver

			// Check if miscellaneous data has been initialized
			if s.miscellaneous.RollPrice == 0 {
				// if not, fetch it
				if err := s.fetchMiscellaneousData(); err != nil {
					logger.Error("failed to fetch miscellaneous data: %v", err)
				}
			}

			ctxMonitoring, cancel := context.WithCancel(ctx)
			s.stopStakingMonitoringFunc = cancel
			go s.stakingAddressMonitoring(ctxMonitoring)

		case <-nodeDownChan:
			s.mu.Lock()
			s.nodeIsUp = false
			s.mu.Unlock()

			if s.stopStakingMonitoringFunc != nil {
				s.stopStakingMonitoringFunc()
			}
		}
	}
}

/* initStakingAddresses Retrieve data from the node and the db to initialize the staking addresses list in ram
 */
func (s *stakingManager) initStakingAddresses() error {
	// get staking addresses from node. These are the ultimate source of truth.
	stakingAddresses, err := s.clientDriver.GetStakingAddresses()
	if err != nil {
		return fmt.Errorf("failed to get staking addresses list from node: %w", err)
	}

	// clean addresses not in node but in db
	if err := s.cleanAddressesNotInNodeButInDB(stakingAddresses); err != nil {
		return fmt.Errorf("failed to clean addresses not in node but in db: %w", err)
	}

	if len(stakingAddresses) == 0 {
		logger.Info("no staking addresses found in node")
		return nil
	}

	// get addresses data from node
	addresses, err := s.getAddressesDataFromNode(stakingAddresses)
	if err != nil {
		return fmt.Errorf("failed to get addresses data from node: %w", err)
	}

	// init staking addresses list in ram
	s.stakingAddresses = addresses
	return nil
}

/*
cleanAddressesNotInNodeButInDB cleans all addresses that are registered in db but not in the node if any
This usualy happens when the plugins has been updated and some addresses are still registered in db but not in the new node
*/
func (s *stakingManager) cleanAddressesNotInNodeButInDB(addressesInNode []string) error {
	currentNetwork := config.GlobalPluginInfo.GetNetwork()
	dbAddresses, err := s.db.GetRollsTarget(currentNetwork)
	if err != nil {
		return fmt.Errorf("failed to get staking addresses registered in db: %w", err)
	}

	for _, dbAddr := range dbAddresses {
		found := false
		for _, stakingAddr := range addressesInNode {
			if dbAddr.Address == stakingAddr {
				found = true
				break
			}
		}
		if !found {
			logger.Warnf("address %s is registered in db but not in node. Deleting it from db...", dbAddr.Address)
			if err := s.db.DeleteRollsTarget(dbAddr.Address, currentNetwork); err != nil {
				return fmt.Errorf("failed to delete address %s from db: %w", dbAddr.Address, err)
			}
			logger.Infof("address %s deleted from db", dbAddr.Address)
		}
	}

	return nil
}

// getAddressesDataFromNode gets the addresses data from the node and convert it to the StakingAddress struct
func (s *stakingManager) getAddressesDataFromNode(addresses []string) ([]StakingAddress, error) {
	js, err := s.nodeAPI.GetAddresses(addresses)
	if err != nil {
		return nil, err
	}

	var res []getAddressesResponse

	err = json.Unmarshal(js, &res)
	if err != nil {
		return nil, err
	}

	walletInfos, err := s.clientDriver.WalletInfo(config.GlobalPluginInfo.GetPwd())
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet info: %v", err)
	}

	return s.convertToStakingAddress(res, walletInfos)
}

func (s *stakingManager) convertToStakingAddress(addresses []getAddressesResponse, walletInfos map[string]clientDriverPkg.WalletInfo) ([]StakingAddress, error) {
	stakingAddresses := make([]StakingAddress, len(addresses))
	for i, addr := range addresses {
		finalBalance, err := strconv.ParseFloat(addr.FinalBalance, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse final balance: %v", err)
		}
		candidateBalance, err := strconv.ParseFloat(addr.CandidateBalance, 64)
		if err != nil {
			return nil, fmt.Errorf("failed to parse candidate balance: %v", err)
		}

		stakingAddresses[i] = StakingAddress{
			Address:          addr.Address,
			FinalRolls:       addr.FinalRolls,
			CandidateRolls:   addr.CandidateRolls,
			ActiveRolls:      walletInfos[addr.Address].AddressInfo.ActiveRolls,
			FinalBalance:     finalBalance,
			CandidateBalance: candidateBalance,
			Thread:           addr.Thread,
		}

		stakingAddresses[i].DeferredCredits = make([]DeferredCredit, len(addr.DeferredCredits))
		for j, credit := range addr.DeferredCredits {
			amount, err := strconv.ParseFloat(credit.Amount, 64)
			if err != nil {
				return nil, fmt.Errorf("failed to parse deferred credit amount: %v", err)
			}
			stakingAddresses[i].DeferredCredits[j] = DeferredCredit{
				Slot:   credit.Slot,
				Amount: amount,
			}
		}

	}

	return s.WithTargetRolls(stakingAddresses)
}

// WithTargetRolls hydrates the addresses with the target rolls from the database
func (s *stakingManager) WithTargetRolls(addresses []StakingAddress) ([]StakingAddress, error) {
	currentNetwork := config.GlobalPluginInfo.GetNetwork()
	dbAddresses, err := s.db.GetRollsTarget(currentNetwork)
	if err != nil {
		return nil, fmt.Errorf("failed to load rolul targets from database: %w", err)
	}

	// Update in-memory addresses with database roll targets
	for _, dbAddr := range dbAddresses {
		for i := range addresses {
			if addresses[i].Address == dbAddr.Address {
				addresses[i].TargetRolls = dbAddr.RollTarget
				break
			}
		}
	}
	return addresses, nil
}
