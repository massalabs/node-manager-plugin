import axios from 'axios';

import { nodeInfosResponse } from '@/models/nodeInfos';
import { networks } from '@/utils/const';

export async function getPluginInfos(): Promise<nodeInfosResponse> {
  const res = await axios.get<nodeInfosResponse>(
    `${import.meta.env.VITE_BASE_API}/pluginInfos`,
  );
  return res.data;
}

export function getNetworkFromVersion(version: string): networks {
  if (version.includes('MAIN')) {
    return networks.mainnet;
  }
  return networks.buildnet;
}
