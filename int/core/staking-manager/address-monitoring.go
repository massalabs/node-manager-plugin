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
	totalValue := s.getTotalValue()
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

			s.handleRollsUpdates(newAddresses)

			totalValue = s.getTotalValue()
		case <-totValueTicker.C:
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
		index, _ := s.getAddressIndexFromRamList(newAddress.Address)
		currentRollTarget := s.stakingAddresses[index].TargetRolls

		pendingOpCompleted, err := s.checkIfPendingOperationIsCompleted(index, newAddress.CandidateRolls)
		if err != nil {
			logger.Errorf("failed to check if there is a pending operation and if it is completed for address %s: %v", newAddress.Address, err)
			continue
		}

		if !pendingOpCompleted {
			logger.Debugf("pending operation for address %s is not completed, skipping rolls update", newAddress.Address)
			continue
		}

		// Sell rolls
		if currentRollTarget < newAddress.CandidateRolls {
			// Check if the address has enough balance to pay minimal fees
			if float64(s.miscellaneous.MinimalFees) > newAddress.FinalBalance {
				logger.Errorf("address %s need to sell rolls but has %f mas which is less than minimal fees (%.2f mas)", newAddress.Address, newAddress.FinalBalance, s.miscellaneous.MinimalFees)
				continue
			}
			rollsToSell := newAddress.CandidateRolls - currentRollTarget
			logger.Infof("Address %s had %d rolls and %d target rolls: Need to sell %d rolls", newAddress.Address, newAddress.FinalRolls, currentRollTarget, rollsToSell)

			opId, err := s.clientDriver.SellRolls(configPkg.GlobalPluginInfo.GetPwd(), newAddress.Address, uint64(rollsToSell), float32(s.miscellaneous.MinimalFees))
			if err != nil {
				logger.Errorf("failed to sell rolls for address %s: %v", newAddress.Address, err)
				continue
			}

			/* if the sellRolls op has been sent, we need to wait for it to be completed.
			so we save it's op id to be able tocheck later if it has been completed */
			s.stakingAddresses[index].pendingOperation = &pendingOperation{
				id:            opId,
				expectedRolls: currentRollTarget,
			}

			logger.Infof("Sold %d rolls for address %s", rollsToSell, newAddress.Address)
			// Buy rolls
		} else if currentRollTarget > newAddress.CandidateRolls {
			// Check if the address has enough balance to pay minimal fees
			if float64(s.miscellaneous.MinimalFees) > newAddress.FinalBalance {
				logger.Errorf("Address %s need to buy rolls but has %f mas which is less than minimal fees (%.2f mas)", newAddress.Address, newAddress.FinalBalance, s.miscellaneous.MinimalFees)
				continue
			}

			rollsToBuy := uint64(min(
				float64(currentRollTarget-newAddress.CandidateRolls),
				newAddress.FinalBalance/float64(s.miscellaneous.RollPrice),
			))

			if rollsToBuy > 0 {
				logger.Infof("Address %s (balance: %f) had %d rolls and %d target rolls: Need to buy %d rolls", newAddress.Address, newAddress.FinalBalance, newAddress.FinalBalance, newAddress.FinalRolls, currentRollTarget, rollsToBuy)
				opId, err := s.clientDriver.BuyRolls(configPkg.GlobalPluginInfo.GetPwd(), newAddress.Address, rollsToBuy, float32(s.miscellaneous.MinimalFees))
				if err != nil {
					logger.Errorf("failed to buy rolls for address %s: %v", newAddress.Address, err)
					continue
				}

				/* if the buyRolls op has been sent, we need to wait for it to be completed.
				so we save it's op id to be able tocheck later if it has been completed */
				s.stakingAddresses[index].pendingOperation = &pendingOperation{
					id:            opId,
					expectedRolls: newAddress.CandidateRolls + rollsToBuy,
				}

				logger.Infof("Bought %d rolls for address %s", rollsToBuy, newAddress.Address)
			}
		}
	}
}

/*
	checkIfPendingOperationIsCompleted checks if the staking address at "index" has a

pending operation and if it has been completed.
It return whether the rolls update process can be pursued or not
*/
func (s *stakingManager) checkIfPendingOperationIsCompleted(index int, candidateRolls uint64) (bool, error) {
	pendingOp := s.stakingAddresses[index].pendingOperation

	if pendingOp == nil {
		return true, nil
	}

	// if there is a pending operation, check if it has been completed
	if pendingOp.expectedRolls == candidateRolls {
		// the pending operation is completed, so we can remove it from the staking addresses list
		s.stakingAddresses[index].pendingOperation = nil
		logger.Infof("Pending operation for address %s has been completed", s.stakingAddresses[index].Address)
		return true, nil
	} else {
		// the pending operation is not completed, so we need to check if it has been expired
		operation, err := s.nodeAPI.GetOperation(pendingOp.id)
		if err != nil {
			return false, fmt.Errorf("failed to get operation %s: %v", pendingOp.id, err)
		}

		if operation == nil || operation.Detail == nil {
			return false, fmt.Errorf("operation or operation detail is nil")
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
			s.stakingAddresses[index].pendingOperation = nil
			logger.Debugf("Pending operation '%s' for address %s has been expired", pendingOp.id, s.stakingAddresses[index].Address)
			return true, nil
		} else {
			s.stakingAddresses[index].pendingOperation = nil
			logger.Debugf("Pending operation '%s' for address %s is still pending", pendingOp.id, s.stakingAddresses[index].Address)
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
