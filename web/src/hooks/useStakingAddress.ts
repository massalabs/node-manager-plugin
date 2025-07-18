// EXTERNALS
import { toast } from '@massalabs/react-ui-kit';
import { useMutation } from '@tanstack/react-query';
import axios, { AxiosError } from 'axios';

// LOCALS
import { useError } from '@/contexts/ErrorContext';
import Intl from '@/i18n/i18n';
import {
  StakingAddress,
  AddStakingAddressBody,
  UpdateStakingAddressBody,
  RemoveStakingAddressBody,
} from '@/models/staking';
import { useStakingStore } from '@/store/stakingStore';
import { getErrorMessage } from '@/utils/error';
import { getApiUrl } from '@/utils/utils';

const STAKING_ADDRESS_ENDPOINT = getApiUrl() + '/stakingAddresses';

/**
 * Hook providing utilities for staking address operations
 * Uses react-query's useMutation for POST, PUT, and DELETE operations
 */
export function useStakingAddress() {
  const addStakingAddressToStore = useStakingStore(
    (state) => state.addStakingAddress,
  );
  const removeStakingAddressFromStore = useStakingStore(
    (state) => state.removeStakingAddress,
  );
  const updateStakingAddressInStore = useStakingStore(
    (state) => state.updateStakingAddress,
  );

  const { setError } = useError();
  // Add a new staking address
  const addStakingAddress = useMutation<
    StakingAddress,
    AxiosError,
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
    onError: (error: AxiosError) => {
      console.error('Failed adding staking address:', error);
      setError({
        title: Intl.t('errors.staking-address-add.title'),
        message: Intl.t('errors.staking-address-add.description', {
          error: getErrorMessage(error),
        }),
      });
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
    onSuccess: (_, payload) => {
      // Show success toast
      toast.success(
        Intl.t(
          'staking.stakingAddressDetails.updateRollTarget.confirmModal.rollTargetUpdated',
        ),
      );

      // Update the staking address in the store
      updateStakingAddressInStore(payload.address, {
        target_rolls: payload.target_rolls,
      });
    },
  });

  // Remove a staking address
  const removeStakingAddress = useMutation<
    void,
    AxiosError,
    RemoveStakingAddressBody,
    unknown
  >({
    mutationKey: ['staking-addresses', 'remove'],
    mutationFn: async (payload: RemoveStakingAddressBody) => {
      await axios.delete(STAKING_ADDRESS_ENDPOINT, { data: payload });
    },
    onSuccess: (_, payload) => {
      // Show success toast
      toast.success(Intl.t('staking.delete-address.address-deleted'));

      // Remove the staking address from the store
      removeStakingAddressFromStore(payload.address);
    },
    onError: (error: AxiosError) => {
      console.error('Failed deleting staking address:', error);
      setError({
        title: Intl.t('errors.staking-address-delete.title'),
        message: Intl.t('errors.staking-address-delete.description', {
          error: getErrorMessage(error),
        }),
      });
    },
  });

  return {
    addStakingAddress,
    updateStakingAddress,
    removeStakingAddress,
  };
}
