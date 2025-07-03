// EXTERNALS
import { UseMutationResult, useMutation, useQueryClient } from '@tanstack/react-query';
import axios from 'axios';

// LOCALS
import { StakingAddress, AddStakingAddressBody, UpdateStakingAddressBody, RemoveStakingAddressBody } from '@/models/staking';

const STAKING_ADDRESS_ENDPOINT = import.meta.env.VITE_BASE_API + "/staking-addresses";

/**
 * Hook providing utilities for staking address operations
 * Uses react-query's useMutation for POST, PUT, and DELETE operations
 */
export function useStakingAddress() {
  const queryClient = useQueryClient();

  // Add a new staking address
  const addStakingAddress = useMutation<StakingAddress, unknown, AddStakingAddressBody, unknown>({
    mutationKey: ['staking-addresses', 'add'],
    mutationFn: async (payload) => {
      const { data } = await axios.post<StakingAddress>(STAKING_ADDRESS_ENDPOINT, payload);
      return data;
    },
    onSuccess: () => {
      // Invalidate and refetch staking addresses after successful addition
      queryClient.invalidateQueries({ queryKey: ['staking-addresses'] });
    },
  });

  // Update an existing staking address
  const updateStakingAddress = useMutation<void, unknown, UpdateStakingAddressBody, unknown>({
    mutationKey: ['staking-addresses', 'update'],
    mutationFn: async (payload) => {
      await axios.put(STAKING_ADDRESS_ENDPOINT, payload);
    },
    onSuccess: () => {
      // Invalidate and refetch staking addresses after successful update
      queryClient.invalidateQueries({ queryKey: ['staking-addresses'] });
    },
  });

  // Remove a staking address
  const removeStakingAddress = useMutation<void, unknown, RemoveStakingAddressBody, unknown>({
    mutationKey: ['staking-addresses', 'remove'],
    mutationFn: async (payload) => {
      await axios.delete(STAKING_ADDRESS_ENDPOINT, { data: payload });
    },
    onSuccess: () => {
      // Invalidate and refetch staking addresses after successful removal
      queryClient.invalidateQueries({ queryKey: ['staking-addresses'] });
    },
  });

  return {
    addStakingAddress,
    updateStakingAddress,
    removeStakingAddress,
  };
} 