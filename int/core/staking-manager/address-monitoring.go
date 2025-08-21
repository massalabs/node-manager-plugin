package stakingManager

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strconv"
	"time"

	configPkg "github.com/massalabs/node-manager-plugin/int/config"
	"github.com/massalabs/node-manager-plugin/int/db"
	errorPkg "github.com/massalabs/node-manager-plugin/int/error"
	"github.com/massalabs/node-manager-plugin/int/utils"
	"github.com/massalabs/station/pkg/logger"
)

// fetchMiscellaneousData fetches the node status and store some values in miscellaneous field
func (s *stakingManager) fetchMiscellaneousData() error {
	status, err := s.nodeAPI.GetStatus()
	if err != nil {
		return fmt.Errorf("failed to get node status: %v", err)
	}

	minimalFees, err := strconv.ParseFloat(*status.MinimalFees, 32)
	if err != nil {
		return fmt.Errorf("failed to parse minimal fees: %v", err)
	}

	s.miscellaneous.MinimalFees = float32(minimalFees)

	if s.miscellaneous.RollPrice == 0 {
		rollPrice, err := strconv.ParseFloat(*status.Config.RollPrice, 32)
		if err != nil {
			return fmt.Errorf("failed to parse roll price: %v", err)
		}

		s.miscellaneous.RollPrice = float32(rollPrice)
	}

	return nil
}

func (s *stakingManager) stakingAddressMonitoring(ctx context.Context) {
	s.mu.Lock()
	if len(s.stakingAddresses) == 0 {
		// Initialize staking addresses
		err := s.initStakingAddresses()
		if err != nil {
			logger.Error("failed to initialize staking addresses: %v", err)
		}
	}
	s.mu.Unlock()

	ticker := time.NewTicker(time.Duration(s.stakingAddressDataPollInterval) * time.Second)
	defer ticker.Stop()

	totValueTicker := time.NewTicker(time.Duration(s.config.TotValueRegisterInterval) * time.Second)
	defer totValueTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.Lock()
			currentAddresses := s.getAddressesFromRamList()

			if len(currentAddresses) == 0 {
				s.mu.Unlock()
				logger.Debug("no staking addresses found in ram list")
				continue
			}

			// Update staking addresses data
			newAddresses, err := s.getAddressesDataFromNode(currentAddresses)
			if err != nil {
				s.mu.Unlock()
				logger.Error("failed to retrieve staking addresses from node: %v", err)
				continue
			}

			if s.addressChangedDispatcher.HasSubscribers() {
				updated := s.updateStakingAddresses(newAddresses)
				if updated {
					s.addressChangedDispatcher.Publish(s.stakingAddresses)
				}
			}
			s.mu.Unlock()

			s.muSellBuyRolls.Lock()
			s.handleRollsUpdates(newAddresses)
			s.muSellBuyRolls.Unlock()

		case <-totValueTicker.C:
			totalValue := s.getTotalValue()
			currentNetwork := utils.NetworkMainnet
			if !configPkg.GlobalPluginInfo.GetIsMainnet() {
				currentNetwork = utils.NetworkBuildnet
			}
			if err := s.db.PostHistory(db.ValueHistory{
				Timestamp:  time.Now(),
				TotalValue: totalValue,
			}, currentNetwork); err != nil {
				logger.Errorf("failed to save total value to database: %v", err)
			}
		}
	}
}

// updateStakingAddresses updates the staking addresses list in ram with the new addresses data if required
// returns true if the staking addresses list has been updated
func (s *stakingManager) updateStakingAddresses(newAddresses []StakingAddress) bool {
	if len(newAddresses) != len(s.stakingAddresses) {
		logger.Debugf("number of staking addresses has changed from %d to %d", len(s.stakingAddresses), len(newAddresses))
		s.stakingAddresses = copyAddresses(newAddresses)
		return true
	}

	updated := false
	for _, newAddress := range newAddresses {
		index, _ := s.getAddressIndexFromRamList(newAddress.Address)
		if s.stakingAddresses[index].CandidateRolls != newAddress.CandidateRolls {
			updated = true
			s.stakingAddresses[index].CandidateRolls = newAddress.CandidateRolls
		}

		if s.stakingAddresses[index].CandidateBalance != newAddress.CandidateBalance {
			updated = true
			s.stakingAddresses[index].CandidateBalance = newAddress.CandidateBalance
		}

		if s.stakingAddresses[index].FinalRolls != newAddress.FinalRolls {
			updated = true
			s.stakingAddresses[index].FinalRolls = newAddress.FinalRolls
		}

		if s.stakingAddresses[index].ActiveRolls != newAddress.ActiveRolls {
			updated = true
			s.stakingAddresses[index].ActiveRolls = newAddress.ActiveRolls
		}

		if s.stakingAddresses[index].FinalBalance != newAddress.FinalBalance {
			updated = true
			s.stakingAddresses[index].FinalBalance = newAddress.FinalBalance
		}

		if !slices.EqualFunc(s.stakingAddresses[index].DeferredCredits, newAddress.DeferredCredits, func(a, b DeferredCredit) bool {
			return a.Amount == b.Amount
		}) {
			updated = true
			s.stakingAddresses[index].DeferredCredits = make([]DeferredCredit, len(newAddress.DeferredCredits))
			copy(s.stakingAddresses[index].DeferredCredits, newAddress.DeferredCredits)
		}
	}
	return updated
}

// For each new address, it check whether it needs to buy or sell rolls to reach it's roll target
// and perform the action if the address has enough MAS to pay the minimal fees
func (s *stakingManager) handleRollsUpdates(newAddresses []StakingAddress) {
	for _, newAddress := range newAddresses {
		err := s.sellBuyRollsAddress(newAddress)
		if err != nil {
			if errorPkg.Is(err, errorPkg.ErrStakingManagerPendingOperationNotCompleted) {
				logger.Debugf(err.Error())
				continue
			}
			logger.Errorf("failed to handle rolls update operation (if required) for address %s: %v", newAddress.Address, err)
		}
	}
}

func (s *stakingManager) sellBuyRollsAddress(address StakingAddress) error {
	index, _ := s.getAddressIndexFromRamList(address.Address)
	currentRollTarget := s.stakingAddresses[index].TargetRolls

	// Before sending a new roll operation, we need to check if there is a pending operation and if it is completed
	pendingOpCompleted, err := s.checkIfPendingOperationIsCompleted(index)
	if err != nil {
		logger.Errorf("failed to check if there is a pending operation and if it is completed for address %s: %v", address.Address, err)
		return errorPkg.New(
			errorPkg.ErrStakingManagerPendingOperationNotCompleted,
			fmt.Sprintf("failed to check if there is a pending operation and if it is completed for address %s", address.Address),
		)
	}

	// If there is a pending operation and it is not completed, we skip the rolls update
	if !pendingOpCompleted {
		logger.Debugf("pending operation for address %s is not completed, skipping rolls update", address.Address)
		return errorPkg.New(
			errorPkg.ErrStakingManagerPendingOperationNotCompleted,
			fmt.Sprintf("pending operation for address %s is not completed, skipping rolls update", address.Address),
		)
	}

	// When the target rolls is negative -> auto compound: buy as many rolls as possible
	if currentRollTarget < 0 {
		// Calculate maximum rolls that can be bought with available balance
		maxRollsToBuy := uint64(address.FinalBalance / float64(s.miscellaneous.RollPrice))

		if maxRollsToBuy > 0 {
			// Check if the address has enough balance to pay minimal fees
			if float64(s.miscellaneous.MinimalFees) > address.FinalBalance {
				return fmt.Errorf("address %s need to buy rolls but has %f mas which is less than minimal fees (%.2f mas)", address.Address, address.FinalBalance, s.miscellaneous.MinimalFees)
			}

			logger.Infof("Address %s (balance: %f) has %d rolls and target is maximum: Need to buy %d rolls (max possible)", address.Address, address.FinalBalance, address.FinalRolls, maxRollsToBuy)
			opId, err := s.clientDriver.BuyRolls(configPkg.GlobalPluginInfo.GetPwd(), address.Address, maxRollsToBuy, float32(s.miscellaneous.MinimalFees))
			if err != nil {
				return fmt.Errorf("failed to buy rolls for address %s: %v", address.Address, err)
			}

			if err := s.handleRollOpMonitoring(index, opId, db.RollOpBuy, maxRollsToBuy); err != nil {
				return fmt.Errorf("failed to handle roll op monitoring for address %s: %v", address.Address, err)
			}

			logger.Infof("Bought %d rolls for address %s", maxRollsToBuy, address.Address)
		}
		return nil
	}

	// Sell rolls
	if currentRollTarget < int64(address.CandidateRolls) {
		// Check if the address has enough balance to pay minimal fees
		if float64(s.miscellaneous.MinimalFees) > address.FinalBalance {
			return fmt.Errorf("address %s need to sell rolls but has %f mas which is less than minimal fees (%.2f mas)", address.Address, address.FinalBalance, s.miscellaneous.MinimalFees)
		}
		rollsToSell := address.CandidateRolls - uint64(currentRollTarget)
		logger.Infof("Address %s had %d rolls and %d target rolls: Need to sell %d rolls", address.Address, address.FinalRolls, currentRollTarget, rollsToSell)

		opId, err := s.clientDriver.SellRolls(configPkg.GlobalPluginInfo.GetPwd(), address.Address, rollsToSell, float32(s.miscellaneous.MinimalFees))
		if err != nil {
			return fmt.Errorf("failed to sell rolls for address %s: %v", address.Address, err)
		}

		if err := s.handleRollOpMonitoring(index, opId, db.RollOpSell, rollsToSell); err != nil {
			return fmt.Errorf("failed to handle roll op monitoring for address %s: %v", address.Address, err)
		}

		logger.Infof("Sold %d rolls for address %s", rollsToSell, address.Address)

		// Buy rolls
	} else if currentRollTarget > int64(address.CandidateRolls) {
		// Check if the address has enough balance to pay minimal fees
		if float64(s.miscellaneous.MinimalFees) > address.FinalBalance {
			return fmt.Errorf("address %s need to buy rolls but has %f mas which is less than minimal fees (%.2f mas)", address.Address, address.FinalBalance, s.miscellaneous.MinimalFees)
		}

		rollsToBuy := uint64(min(
			float64(currentRollTarget-int64(address.CandidateRolls)),
			address.FinalBalance/float64(s.miscellaneous.RollPrice),
		))

		if rollsToBuy > 0 {
			logger.Infof("Address %s (balance: %f) has %d rolls and %d target rolls: Need to buy %d rolls", address.Address, address.FinalBalance, address.FinalBalance, address.FinalRolls, currentRollTarget, rollsToBuy)
			opId, err := s.clientDriver.BuyRolls(configPkg.GlobalPluginInfo.GetPwd(), address.Address, rollsToBuy, float32(s.miscellaneous.MinimalFees))
			if err != nil {
				return fmt.Errorf("failed to buy rolls for address %s: %v", address.Address, err)
			}

			if err := s.handleRollOpMonitoring(index, opId, db.RollOpBuy, rollsToBuy); err != nil {
				return fmt.Errorf("failed to handle roll op monitoring for address %s: %v", address.Address, err)
			}

			logger.Infof("Bought %d rolls for address %s", rollsToBuy, address.Address)
		}
	}
	return nil
}

func (s *stakingManager) handleRollOpMonitoring(index int, opId string, operationType db.RollOp, amount uint64) error {
	/* if the buyRolls or sellRolls op has been sent, we need to wait for it to be completed.
	so we save it's op id to be able tocheck later if it has been completed */
	s.stakingAddresses[index].pendingOperationId = &opId

	// Record the roll operation in the database
	currentNetwork := utils.NetworkMainnet
	if !configPkg.GlobalPluginInfo.GetIsMainnet() {
		currentNetwork = utils.NetworkBuildnet
	}

	if err := s.db.AddRollOpHistory(s.stakingAddresses[index].Address, operationType, amount, opId, currentNetwork); err != nil {
		if operationType == db.RollOpBuy {
			return fmt.Errorf("failed to record buy roll operation for address %s (amount: %d): %v", s.stakingAddresses[index].Address, amount, err)
		} else {
			return fmt.Errorf("failed to record sell roll operation for address %s (amount: %d): %v", s.stakingAddresses[index].Address, amount, err)
		}
	}

	return nil
}

/*
	checkIfPendingOperationIsCompleted checks if the staking address at "index" has a

pending operation and if it has been completed.
It return whether the rolls update process can be pursued or not
*/
func (s *stakingManager) checkIfPendingOperationIsCompleted(index int) (bool, error) {
	pendingOpId := s.stakingAddresses[index].pendingOperationId

	if pendingOpId == nil {
		return true, nil
	}

	operation, err := s.nodeAPI.GetOperation(*pendingOpId)
	if err != nil {
		return false, fmt.Errorf("failed to get operation %s: %v", *pendingOpId, err)
	}

	if operation == nil {
		return false, fmt.Errorf("retrieved operation %s is nil", *pendingOpId)
	}

	if operation.IsFinal {
		s.stakingAddresses[index].pendingOperationId = nil
		return true, nil
	} else {
		// if the op is not final, check if it has expired
		if operation.Detail == nil {
			return false, fmt.Errorf("detail field of retrieved operation %s is nil", *pendingOpId)
		}

		status, err := s.nodeAPI.GetStatus()
		if err != nil {
			return false, fmt.Errorf("failed to get node status: %v", err)
		}

		if status.LastSlot == nil {
			return false, fmt.Errorf("node status last slot is nil")
		}

		// if the operation has been expired, we can remove it from the staking addresses list
		if operation.Detail.Content.ExpirePeriod < uint(status.LastSlot.Period) {
			s.stakingAddresses[index].pendingOperationId = nil
			logger.Debugf("Pending operation '%s' for address %s has been expired", *pendingOpId, s.stakingAddresses[index].Address)
			return true, nil
		} else {
			s.stakingAddresses[index].pendingOperationId = nil
			logger.Debugf("Pending operation '%s' for address %s is still pending", *pendingOpId, s.stakingAddresses[index].Address)
			return false, nil
		}
	}
}

// getTotalValue returns the total value of all staking addresses
// it takes into account the final balance, the final rolls and the deferred credits
func (s *stakingManager) getTotalValue() float64 {
	totalValue := float64(0)
	for _, address := range s.stakingAddresses {
		deferredCredits := float64(0)
		for _, defCredit := range address.DeferredCredits {
			deferredCredits += defCredit.Amount
		}
		totalValue += address.FinalBalance + float64(address.FinalRolls)*float64(s.miscellaneous.RollPrice) + deferredCredits
	}
	return math.Floor(totalValue*1000) / 1000 // keep only 3 digit after the comma
}
