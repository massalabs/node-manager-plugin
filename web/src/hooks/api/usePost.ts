// STYLES

// EXTERNALS
import { UseMutationResult, useMutation } from '@tanstack/react-query';
import axios, { AxiosError } from 'axios';

// LOCALS
export function usePost<T>(
  resource: string,
): UseMutationResult<T, AxiosError, T, unknown> {
  var url = `${import.meta.env.VITE_BASE_API}/${resource}`;

  return useMutation<T, AxiosError, T, unknown>({
    mutationKey: [resource],
    mutationFn: async (payload) => {
      const { data } = await axios.post<T>(url, payload);

      return data;
    },
  });
}
