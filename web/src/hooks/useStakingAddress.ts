// EXTERNALS
import { toast } from '@massalabs/react-ui-kit';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';

// LOCALS
import Intl from '@/i18n/i18n';
import {
  StakingAddress,
  AddStakingAddressBody,
  UpdateStakingAddressBody,
  RemoveStakingAddressBody,
} from '@/models/staking';
import { useStakingStore } from '@/store/stakingStore';
import { goToErrorPage } from '@/utils/routes';

const STAKING_ADDRESS_ENDPOINT =
  import.meta.env.VITE_BASE_API + '/stakingAddresses';

/**
 * Hook providing utilities for staking address operations
 * Uses react-query's useMutation for POST, PUT, and DELETE operations
 */
export function useStakingAddress() {
  const queryClient = useQueryClient();
  const addStakingAddressToStore = useStakingStore(
    (state) => state.addStakingAddress,
  );
  const navigate = useNavigate();
  // Add a new staking address
  const addStakingAddress = useMutation<
    StakingAddress,
    Error,
    AddStakingAddressBody,
    unknown
  >({
    mutationKey: ['staking-addresses', 'add'],
    mutationFn: async (payload: AddStakingAddressBody) => {
      const { data } = await axios.post<StakingAddress>(
        STAKING_ADDRESS_ENDPOINT,
        payload,
      );
      return data;
    },
    onSuccess: (data: StakingAddress) => {
      // Add the received StakingAddress to the stakingStore
      addStakingAddressToStore(data);

      // Show success toast
      toast.success(Intl.t('staking.add-address.address-added'));
    },
    onError: (error: Error) => {
      console.error('Failed adding staking address:', error);
      goToErrorPage(
        navigate,
        Intl.t('errors.staking-address-add.title'),
        Intl.t('errors.staking-address-add.description', {
          error: error.message,
        }),
      );
    },
  });

  // Update an existing staking address
  const updateStakingAddress = useMutation<
    void,
    Error,
    UpdateStakingAddressBody,
    unknown
  >({
    mutationKey: ['staking-addresses', 'update'],
    mutationFn: async (payload: UpdateStakingAddressBody) => {
      await axios.put(STAKING_ADDRESS_ENDPOINT, payload);
    },
    onSuccess: () => {
      // Invalidate and refetch staking addresses after successful update
      queryClient.invalidateQueries({ queryKey: ['staking-addresses'] });
    },
  });

  // Remove a staking address
  const removeStakingAddress = useMutation<
    void,
    Error,
    RemoveStakingAddressBody,
    unknown
  >({
    mutationKey: ['staking-addresses', 'remove'],
    mutationFn: async (payload: RemoveStakingAddressBody) => {
      await axios.delete(STAKING_ADDRESS_ENDPOINT, { data: payload });
    },
    onSuccess: () => {
      // Invalidate and refetch staking addresses after successful removal
      queryClient.invalidateQueries({ queryKey: ['staking-addresses'] });
      // Show success toast
      toast.success(Intl.t('staking.delete-address.address-deleted'));
    },
    onError: (error: Error) => {
      console.error('Failed deleting staking address:', error);
      goToErrorPage(
        navigate,
        Intl.t('errors.staking-address-delete.title'),
        Intl.t('errors.staking-address-delete.description', {
          error: error.message,
        }),
      );
    },
  });

  return {
    addStakingAddress,
    updateStakingAddress,
    removeStakingAddress,
  };
}
