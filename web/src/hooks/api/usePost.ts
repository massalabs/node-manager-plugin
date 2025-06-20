// STYLES

// EXTERNALS
import { UseMutationResult, useMutation } from '@tanstack/react-query';
import axios from 'axios';

// LOCALS
export function usePost<T>(
  resource: string,
): UseMutationResult<T, unknown, T, unknown> {
  var url = `${import.meta.env.VITE_BASE_API}/${resource}`;

  return useMutation<T, unknown, T, unknown>({
    mutationKey: [resource],
    mutationFn: async (payload) => {
      const { data } = await axios.post<T>(url, payload);

      return data;
    },
  });
}
