import React from 'react';

import { Dropdown } from '@massalabs/react-ui-kit';
import shallow from 'zustand/shallow';

import { useNodeStore } from '@/store/nodeStore';
import { getNetworkFromVersion, isRunning } from '@/utils';

function getNetworkNameFromVersion(version: string) {
  return getNetworkFromVersion(version) + ' (' + version + ')';
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

  const selectedNetworkKey: number =
    versions.length > 0
      ? parseInt(
          Object.keys(versions).find(
            (_, idx) => getNetworkFromVersion(versions[idx]) === currentNetwork,
          ) || '0',
        )
      : 0;

  // If no networks are loaded yet, don't render the dropdown
  if (versions.length === 0) {
    return (
      <div
        className="h-8 w-32 bg-gray-200 animate-pulse rounded"
        role="status"
        aria-label="Loading networks"
      />
    );
  }

  return (
    <Dropdown
      options={availableNetworksItems}
      select={selectedNetworkKey}
      size="xs"
      readOnly={nodeIsRunning}
      style={{
        filter: nodeIsRunning
          ? 'grayscale(0.7) brightness(1.2) opacity(0.5)'
          : 'brightness(1.5) opacity(1)',
      }}
    />
  );
};
