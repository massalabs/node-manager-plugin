import React from 'react';

import { Spinner } from '@massalabs/react-ui-kit';

import { useNodeStore } from '@/store/nodeStore';
import { NodeStatus } from '@/utils';

export const NodeStatusDisplay: React.FC = () => {
  const status = useNodeStore((state) => state.status);

  const getStatusColor = (status: NodeStatus) => {
    switch (status) {
      case NodeStatus.ON:
        return 'bg-green-500';
      case NodeStatus.OFF:
        return 'bg-gray-500';
      case NodeStatus.CRASHED:
      case NodeStatus.DESYNCED:
        return 'bg-red-500';
      case NodeStatus.STOPPING:
        return 'bg-yellow-500';
      case NodeStatus.STARTING:
        return 'bg-blue-300';
      case NodeStatus.BOOTSTRAPPING:
        return 'bg-blue-500';
      default:
        return 'bg-gray-400'; // Light gray for unknown status
    }
  };

  return (
    <>
      <div
        className={`relative p-4 rounded shadow ${getStatusColor(
          status,
        )} w-36 text-center opacity-70`}
        style={{ minHeight: '64px' }} // Ensure enough height for centering
      >
        {(status === NodeStatus.STOPPING ||
          status === NodeStatus.BOOTSTRAPPING ||
          status === NodeStatus.STARTING) && (
          <div className="absolute inset-0 flex items-center justify-center z-20">
            <Spinner />
          </div>
        )}
        <span
          className="
          font-bold text-white z-10 opacity-50 absolute inset-0 flex items-center justify-center pointer-events-none
          text-sm
          "
        >
          {status}
        </span>
      </div>
    </>
  );
};
