package stakingManager

import (
	"context"
	"fmt"
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

			/* If the front is listening to address data changes (if the addressChangedDispatcher has subscribers),
			update the staking addresses list in ram and publish the new staking address data to front
			*/
			if s.addressChangedDispatcher.HasSubscribers() {
				updated := s.updateStakingAddresses(newAddresses)
				// if the staking addresses list has been updated, publish the new staking address data to front
				if updated {
					s.addressChangedDispatcher.Publish(s.stakingAddresses)
				}
			}
			s.mu.Unlock()

			// Handle whether we need to buy or sell rolls to reach the roll target
			s.handleRollsUpdates(newAddresses)

			// get total value of all staking addresses and save to database if it has changed
			newTotalValue := s.getTotalValue()
			if newTotalValue != totalValue {
				logger.Debugf("total value has changed from %f to %f, saving to database", totalValue, newTotalValue)
				currentNetwork := utils.NetworkMainnet
				if !configPkg.GlobalPluginInfo.GetIsMainnet() {
					currentNetwork = utils.NetworkBuildnet
				}
				if err := s.db.PostHistory([]db.ValueHistory{
					{
						Timestamp:  time.Now(),
						TotalValue: newTotalValue,
					},
				}, currentNetwork); err != nil {
					logger.Errorf("failed to save total value to database: %v", err)
				}
				totalValue = newTotalValue
			}

		}
	}
}

// updateStakingAddresses updates the staking addresses list in ram with the new addresses data if required
// returns true if the staking addresses list has been updated
func (s *stakingManager) updateStakingAddresses(newAddresses []StakingAddress) bool {
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

		// Sell rolls
		if currentRollTarget > newAddress.FinalRolls {
			// Check if the address has enough balance to pay minimal fees
			if float64(s.miscellaneous.MinimalFees) > newAddress.FinalBalance {
				logger.Errorf("address %s need to sell rolls but has %f mas which is less than minimal fees (%.2f mas)", newAddress.Address, newAddress.FinalBalance, s.miscellaneous.MinimalFees)
				continue
			}
			rollsToSell := currentRollTarget - newAddress.FinalRolls
			logger.Infof("Address %s had %d rolls and %d target rolls: Need to sell %d rolls", newAddress.Address, newAddress.FinalRolls, currentRollTarget, rollsToSell)

			_, err := s.clientDriver.SellRolls(configPkg.GlobalPluginInfo.GetPwd(), newAddress.Address, uint64(rollsToSell), float32(s.miscellaneous.MinimalFees))
			if err != nil {
				logger.Errorf("failed to sell rolls for address %s: %v", newAddress.Address, err)
				continue
			}

			logger.Infof("Sold %d rolls for address %s", rollsToSell, newAddress.Address)
			// Buy rolls
		} else if currentRollTarget < newAddress.FinalRolls {
			// Check if the address has enough balance to pay minimal fees
			if float64(s.miscellaneous.MinimalFees) > newAddress.FinalBalance {
				logger.Errorf("Address %s need to buy rolls but has %f mas which is less than minimal fees (%.2f mas)", newAddress.Address, newAddress.FinalBalance, s.miscellaneous.MinimalFees)
				continue
			}

			rollsToBuy := min(
				float64(newAddress.FinalRolls-currentRollTarget),
				newAddress.FinalBalance/float64(s.miscellaneous.RollPrice),
			)

			if rollsToBuy > 0 {
				logger.Infof("Address %s (balance: %f) had %d rolls and %d target rolls: Need to buy %d rolls", newAddress.Address, newAddress.FinalBalance, newAddress.FinalBalance, newAddress.FinalRolls, currentRollTarget, rollsToBuy)
				_, err := s.clientDriver.BuyRolls(configPkg.GlobalPluginInfo.GetPwd(), newAddress.Address, uint64(rollsToBuy), float32(s.miscellaneous.MinimalFees))
				if err != nil {
					logger.Errorf("failed to buy rolls for address %s: %v", newAddress.Address, err)
					continue
				}

				logger.Infof("Bought %d rolls for address %s", rollsToBuy, newAddress.Address)
			}
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
	return totalValue
}
