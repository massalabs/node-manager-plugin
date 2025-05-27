import { create } from 'zustand';
import { networks } from '@/utils/const';

export enum NodeStatus {
    ON = 'on',
    OFF = 'off',
    BOOTSTRAPPING = 'bootstrapping',
    STOPPING = 'stopping',
    CRASHED = 'crashed',
    DESYNCED = 'desynced',
    PLUGINERROR = 'pluginError',
}

export interface NodeStoreState {
    status: NodeStatus;
    network: networks;
    autoRestart: boolean;
    setStatus: (status: NodeStatus) => void;
    setNetwork: (network: networks) => void;
    setAutoRestart: (autoRestart: boolean) => void;
    isRunning: () => boolean;
}

export const useNodeStore = create<NodeStoreState>((set, get) => ({
    status: NodeStatus.OFF,
    network: networks.mainnet,
    autoRestart: false,
    setStatus: (status: NodeStatus) => {
        set({ status });
    },
    setNetwork: (network: networks) => {
        set({ network });
    },
    setAutoRestart: (autoRestart: boolean) => {
        set({ autoRestart });
    },
    isRunning: () => {
        const { status } = get();
        return status !== NodeStatus.OFF &&
        status !== NodeStatus.CRASHED
    }
}));