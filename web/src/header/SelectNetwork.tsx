import React from 'react';

import { Dropdown } from '@massalabs/react-ui-kit';
import shallow from 'zustand/shallow';

import { useNodeStore } from '@/store/nodeStore';
import { getNetworkFromVersion, isRunning } from '@/utils';

function getNetworkNameFromVersion(version: string) {
  return getNetworkFromVersion(version) + ' v' + version.slice(5);
}

/* SelectNetwork allows to choose on which network the node will be launched: mainnet or buildnet
If the node is running, this component is disabled*/
export const SelectNetwork: React.FC = () => {
  const { currentNetwork, versions, setNetwork, status } = useNodeStore(
    (state) => ({
      currentNetwork: state.currentNetwork,
      versions: state.networksData.map((network) => network.version),
      setNetwork: state.setNetwork,
      status: state.status,
    }),
    shallow,
  );

  const nodeIsRunning = isRunning(status);

  const availableNetworksItems = versions.map((version) => ({
    item: getNetworkNameFromVersion(version),
    onClick: () => {
      setNetwork(getNetworkFromVersion(version));
    },
  }));

  const selectedNetworkKey: number = parseInt(
    Object.keys(versions).find(
      (_, idx) => getNetworkFromVersion(versions[idx]) === currentNetwork,
    ) || '0',
  );

  return (
    <Dropdown
      options={availableNetworksItems}
      select={selectedNetworkKey}
      size="xs"
      readOnly={nodeIsRunning}
      style={{
        filter: nodeIsRunning
          ? 'grayscale(0.7) brightness(1.2)'
          : 'brightness(1.5)',
        opacity: nodeIsRunning ? 0.6 : 1,
      }}
    />
  );
};
