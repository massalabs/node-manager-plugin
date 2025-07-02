import { create } from 'zustand';

import { getPluginInfos, NodeStatus, getNetworkFromVersion } from '@/utils';
import { networks } from '@/utils/const';
export interface NodeStoreState {
  status: NodeStatus;
  network: networks;
  version: string;
  autoRestart: boolean;
  setStatus: (status: NodeStatus) => void;
  setNetwork: (network: networks) => void;
  setVersion: (version: string) => void;
  setAutoRestart: (autoRestart: boolean) => void;
}

export const useNodeStore = create<NodeStoreState>((set, get) => ({
  status: NodeStatus.UNSET,
  network: networks.mainnet,
  version: '',
  autoRestart: false,
  setStatus: (status: NodeStatus) => {
    /*
    if the first status update is not off, it means the node have been launched and that we have reloaded the page
    This means that various store values are not default and we need to retrieve them.
    */
    if (status !== NodeStatus.OFF && get().status === NodeStatus.UNSET) {
      getPluginInfos().then((data) => {
        set({
          autoRestart: data.autoRestart ?? false,
          version: data.version,
          network: getNetworkFromVersion(data.version),
        });
      });
    }

    // if the node is closed, don't display version
    if (status === NodeStatus.OFF) {
      set({
        version: '',
      });
    }
    set({ status });
  },
  setNetwork: (network: networks) => {
    set({ network });
  },
  setVersion: (version: string) => {
    set({ version });
  },
  setAutoRestart: (autoRestart: boolean) => {
    set({ autoRestart });
  },
}));
