import { useQuery } from '@tanstack/react-query';
import axios from 'axios';

import { nodeInfosResponse } from '@/models/nodeInfos';
import { getApiUrl } from '@/utils/utils';

export const usePluginInfos = () => {
  return useQuery({
    queryKey: ['pluginInfos'],
    queryFn: async () => {
      const { data } = await axios.get<nodeInfosResponse>(
        `${getApiUrl()}/pluginInfos`,
      );
      return data;
    },
  });
};
