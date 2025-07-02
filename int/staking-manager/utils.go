package stakingManager

import "slices"

func (sm *stakingManager) getAddressIndexFromRamList(address string) (int, bool) {
	for i, addr := range sm.stakingAddresses {
		if addr.Address == address {
			return i, true
		}
	}
	return -1, false
}

func (sm *stakingManager) ramAddressListContains(address string) bool {
	return slices.ContainsFunc(sm.stakingAddresses, func(addr StakingAddress) bool {
		return addr.Address == address
	})
}

func (sm *stakingManager) removeAddressFromRamList(address string) bool {
	sm.stakingAddresses = slices.DeleteFunc(sm.stakingAddresses, func(addr StakingAddress) bool {
		return addr.Address == address
	})

	return true
}

func (sm *stakingManager) getAddressesFromRamList() []string {
	addresses := make([]string, 0, len(sm.stakingAddresses))
	for _, addr := range sm.stakingAddresses {
		addresses = append(addresses, addr.Address)
	}
	return addresses
}
