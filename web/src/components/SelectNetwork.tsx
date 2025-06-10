import React from 'react';
import { Dropdown } from '@massalabs/react-ui-kit';
import { useNodeStore } from '@/store/nodeStore';
import { networks } from '@/utils/const';
import shallow from 'zustand/shallow';
import { isRunning } from '@/utils';


/* ChooseNetwork allows to choose which on network the node will be launched: mainnet or buildnet
If the node is running, this component is disabled*/
export const SelectNetwork: React.FC = () => {
    const { currentNetwork, setNetwork, status } = useNodeStore(
        state => ({ currentNetwork: state.network, setNetwork: state.setNetwork, status: state.status}),
        shallow
    );

    const nodeIsRunning = isRunning(status);

    const availableNetworks = Object.values(networks)

    const availableNetworksItems = availableNetworks.map((network) => ({
        item: network,
        onClick: () => {
          setNetwork(network);
        },
    }));

    const selectedNetworkKey: number = parseInt(
        Object.keys(availableNetworks).find(
          (_, idx) => availableNetworks[idx] === currentNetwork,
        ) || '0',
    );

    return (
        <Dropdown
            options={availableNetworksItems}
            select={selectedNetworkKey}
            size="xs"
            readOnly={nodeIsRunning}
            style={{
                filter: nodeIsRunning ? 'grayscale(0.7) brightness(1.2)' : 'brightness(1.5)',
                opacity: nodeIsRunning ? 0.6 : 1,
            }}
        />
    );
};