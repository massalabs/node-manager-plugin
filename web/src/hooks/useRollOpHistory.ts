import { useEffect } from 'react';

import { useQuery } from '@tanstack/react-query';
import axios from 'axios';
import { AxiosError } from 'axios';

import { useError } from '@/contexts/ErrorContext';
import { RollOpHistoryResponse } from '@/models/staking';
import { getErrorMessage } from '@/utils/error';
import { getApiUrl } from '@/utils/utils';

/**
 * Hook for fetching roll operation history for a specific address and network
 */
export function useRollOpHistory(address: string, isMainnet: boolean) {
  const { setError } = useError();

  const useQueryResult = useQuery<RollOpHistoryResponse, AxiosError>({
    queryKey: ['roll-op-history', address, isMainnet],
    queryFn: async () => {
      const { data } = await axios.get<RollOpHistoryResponse>(
        `${getApiUrl()}/rollOpHistory`,
        {
          params: {
            address,
            isMainnet,
          },
        },
      );
      return data;
    },
    enabled: !!address,
    refetchInterval: 30000, // Refetch every 30 seconds
  });

  useEffect(() => {
    if (useQueryResult.error) {
      setError({
        title: 'Error fetching roll operation history',
        message: getErrorMessage(useQueryResult.error),
      });
    }
  }, [useQueryResult.error, setError]);

  return useQueryResult;
}
