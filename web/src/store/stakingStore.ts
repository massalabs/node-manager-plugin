import { create } from 'zustand';

import { StakingAddress } from '@/models/staking';

export interface StakingStoreState {
  stakingAddresses: StakingAddress[];
  setStakingAddresses: (addresses: StakingAddress[]) => void;
  updateStakingAddress: (
    address: string,
    updates: Partial<StakingAddress>,
  ) => void;
  addStakingAddress: (address: StakingAddress) => void;
  removeStakingAddress: (address: string) => void;
  clearStakingAddresses: () => void;
}

export const useStakingStore = create<StakingStoreState>((set) => ({
  stakingAddresses: [],

  setStakingAddresses: (addresses: StakingAddress[]) => {
    set({ stakingAddresses: addresses });
  },

  updateStakingAddress: (address: string, updates: Partial<StakingAddress>) => {
    set((state) => ({
      stakingAddresses: state.stakingAddresses.map((addr) =>
        addr.address === address ? { ...addr, ...updates } : addr,
      ),
    }));
  },

  addStakingAddress: (address: StakingAddress) => {
    set((state) => ({
      stakingAddresses: [...state.stakingAddresses, address],
    }));
  },

  removeStakingAddress: (address: string) => {
    set((state) => ({
      stakingAddresses: state.stakingAddresses.filter(
        (addr) => addr.address !== address,
      ),
    }));
  },

  clearStakingAddresses: () => {
    set({ stakingAddresses: [] });
  },
}));
