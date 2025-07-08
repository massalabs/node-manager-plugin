import { create } from 'zustand';

import { getPluginInfos, NodeStatus, getNetworkFromVersion } from '@/utils';
import { networks } from '@/utils/const';
export interface NodeStoreState {
  status: NodeStatus;
  network: networks;
  version: string;
  autoRestart: boolean;
  hasPwdMainnet: boolean;
  hasPwdBuildnet: boolean;
  pluginVersion: string;
  setStatus: (status: NodeStatus) => void;
  setNetwork: (network: networks) => void;
  setVersion: (version: string) => void;
  setAutoRestart: (autoRestart: boolean) => void;
  getHasPwd: () => boolean;
  setHasPwd: (hasPwd: boolean, network?: networks) => void;
  setPluginVersion: (pluginVersion: string) => void;
}

export const useNodeStore = create<NodeStoreState>((set, get) => ({
  status: NodeStatus.UNSET,
  network: networks.mainnet,
  version: '',
  autoRestart: false,
  hasPwdMainnet: false,
  hasPwdBuildnet: false,
  pluginVersion: '',
  setStatus: (status: NodeStatus) => {
    /*
    if the first status update is not off, it means the node have been launched and that we have reloaded the page
    This means that various store values are not default and we need to retrieve them.
    */
    if (get().status === NodeStatus.UNSET) {
      getPluginInfos().then((data) => {
        set({
          autoRestart: data.autoRestart ?? false,
          version: data.version,
          network: getNetworkFromVersion(data.version),
          hasPwdMainnet: data.hasPwdMainnet,
          hasPwdBuildnet: data.hasPwdBuildnet,
          pluginVersion: data.pluginVersion,
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
  setHasPwd: (hasPwd: boolean, network?: networks) => {
    const net = network ?? get().network;
    if (net === networks.mainnet) {
      set({ hasPwdMainnet: hasPwd });
    } else if (net === networks.buildnet) {
      set({ hasPwdBuildnet: hasPwd });
    }
  },
  getHasPwd: () => {
    return get().network === networks.mainnet
      ? get().hasPwdMainnet
      : get().hasPwdBuildnet;
  },
  setPluginVersion: (pluginVersion: string) => {
    set({ pluginVersion });
  },
}));
