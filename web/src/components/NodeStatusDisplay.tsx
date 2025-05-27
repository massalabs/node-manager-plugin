import React from 'react';
import { useNodeStore, NodeStatus } from '@/store/nodeStore';

export const NodeStatusDisplay: React.FC = () => {
  const status = useNodeStore((state) => state.status);

  const getStatusColor = (status: NodeStatus) => {
    switch (status) {
      case NodeStatus.ON:
      return 'bg-green-500'; // Green for active/on
      case NodeStatus.OFF:
      return 'bg-gray-500'; // Gray for inactive/off
      case NodeStatus.CRASHED, NodeStatus.DESYNCED, NodeStatus.PLUGINERROR:
      return 'bg-red-500'; // Red for crashed/error
      case NodeStatus.STOPPING:
      return 'bg-yellow-500'; // Yellow for starting
      case NodeStatus.BOOTSTRAPPING:
      return 'bg-blue-500'; // Blue for updating
      default:
      return 'bg-gray-400'; // Light gray for unknown status
    }
  };

  return (
    <div
      className={`p-4 rounded shadow ${getStatusColor(status)} w-36 text-center opacity-70`}
    >
      <span className="font-bold text-white">{status}</span>
    </div>
  );
};