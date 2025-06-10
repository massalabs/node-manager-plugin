import { create } from 'zustand';
import { networks } from '@/utils/const';
import { NodeStatus } from '@/utils';
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

export const useNodeStore = create<NodeStoreState>((set) => ({
    status: NodeStatus.OFF,
    network: networks.mainnet,
    version: '',
    autoRestart: false,
    setStatus: (status: NodeStatus) => {
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