import React from 'react';
import { Dropdown } from '@massalabs/react-ui-kit';
import { useNodeStore } from '@/store/nodeStore';
import { networks } from '@/utils/const';

/* ChooseNetwork allows to choose which on network the node will be launched: mainnet or buildnet
If the node is running, this component is disabled*/
export const SelectNetwork: React.FC = () => {
    const { network: currentNetwork, setNetwork, isRunning } = useNodeStore();

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
            readOnly={isRunning()}
            style={{ filter: 'brightness(1.5)' }}
        />
    );
};