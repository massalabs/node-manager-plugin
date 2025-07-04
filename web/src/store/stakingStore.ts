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
  stakingAddresses: [
    {
      address: 'AU12h9c7mHSf8jVZLCpVGFXzqzMZvqXG8t8kfn8dqSwFu5PVxgqnc',
      finalBalance: 1000000,
      targetRolls: 2,
      finalRolls: 1,
      candidateRolls: 1,
      activeRolls: 1,
      candidateBalance: 1000000,
      thread: 1,
      deferredCredits: [
        {
          amount: 500000,
          slot: {
            period: 100,
            thread: 1,
          },
        },
      ],
    },
  ],

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
