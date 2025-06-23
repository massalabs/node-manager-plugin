import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

import { nodeInfosResponse } from '@/models/nodeInfos';

export const usePluginInfos = () => {
  return useQuery({
    queryKey: ['pluginInfos'],
    queryFn: async () => {
      const { data } = await axios.get<nodeInfosResponse>(
        `${import.meta.env.VITE_BASE_API}/pluginInfos`,
      );
      return data;
    },
  });
};
