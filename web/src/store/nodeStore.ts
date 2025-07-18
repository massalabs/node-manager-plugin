import { create } from 'zustand';

import { networkData } from '@/models/nodeInfos';
import { NodeStatus, getNetworkFromVersion } from '@/utils';
import { networks } from '@/utils/const';
export interface NodeStoreState {
  status: NodeStatus;
  networksData: networkData[];
  currentNetwork: networks;
  autoRestart: boolean;
  pluginVersion: string;
  initInfos: (
    autoRestart: boolean,
    networksData: networkData[],
    currentNetwork: networks,
    pluginVersion: string,
  ) => void;
  setStatus: (status: NodeStatus) => void;
  setNetwork: (network: networks) => void;
  setAutoRestart: (autoRestart: boolean) => void;
  getHasPwd: () => boolean;
  setHasPwd: (hasPwd: boolean, network?: networks) => void;
  setPluginVersion: (pluginVersion: string) => void;
}

export const useNodeStore = create<NodeStoreState>((set, get) => ({
  status: NodeStatus.UNSET,
  networksData: [],
  currentNetwork: networks.mainnet,
  autoRestart: false,
  pluginVersion: '',
  initInfos: (
    autoRestart: boolean,
    networksData: networkData[],
    currentNetwork: networks,
    pluginVersion: string,
  ) => {
    set({
      autoRestart: autoRestart,
      networksData: networksData,
      currentNetwork: currentNetwork,
      pluginVersion: pluginVersion,
    });
  },
  setStatus: (status: NodeStatus) => {
    set({ status });
  },
  setNetwork: (network: networks) => {
    set({ currentNetwork: network });
  },
  setAutoRestart: (autoRestart: boolean) => {
    set({ autoRestart });
  },
  setHasPwd: (hasPwd: boolean, network?: networks) => {
    const net = network ?? get().currentNetwork;
    const networkData = get().networksData.map((networkData) => {
      if (getNetworkFromVersion(networkData.version) === net) {
        return { ...networkData, hasPwd: hasPwd };
      }
      return networkData;
    });
    set({ networksData: networkData });
  },
  getHasPwd: () => {
    return (
      get().networksData.find(
        (networkData) =>
          getNetworkFromVersion(networkData.version) === get().currentNetwork,
      )?.hasPwd ?? false
    );
  },
  setPluginVersion: (pluginVersion: string) => {
    set({ pluginVersion });
  },
}));
