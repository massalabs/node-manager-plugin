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
      final_balance: 1000000.773,
      target_rolls: 2,
      final_roll_count: 1,
      candidate_roll_count: 1,
      active_roll_count: 1,
      candidate_balance: 1000000.33,
      thread: 1,
      deferred_credits: [
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
