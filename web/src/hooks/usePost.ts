// STYLES

// EXTERNALS
import { UseMutationResult, useMutation } from '@tanstack/react-query';
import axios, { AxiosError } from 'axios';

// LOCALS
import { getApiUrl } from '@/utils/utils';

export function usePost<T>(
  resource: string,
): UseMutationResult<T, AxiosError, T, unknown> {
  var url = `${getApiUrl()}/${resource}`;

  return useMutation<T, AxiosError, T, unknown>({
    mutationKey: [resource],
    mutationFn: async (payload) => {
      const { data } = await axios.post<T>(url, payload);

      return data;
    },
  });
}
