import React from 'react';

import { Spinner } from '@massalabs/react-ui-kit';

import Intl from '@/i18n/i18n';
import { useNodeStore } from '@/store/nodeStore';
import { NodeStatus } from '@/utils';

export const Status: React.FC = () => {
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
        return 'bg-gray-400';
    }
  };

  const getStatusTextColor = (status: NodeStatus) => {
    switch (status) {
      case NodeStatus.ON:
        return 'text-green-500';
      case NodeStatus.OFF:
        return 'text-gray-500';
      case NodeStatus.CRASHED:
      case NodeStatus.DESYNCED:
        return 'text-red-500';
      case NodeStatus.STOPPING:
        return 'text-yellow-500';
      case NodeStatus.STARTING:
        return 'text-blue-300';
      case NodeStatus.BOOTSTRAPPING:
        return 'text-blue-500';
      default:
        return 'text-gray-400';
    }
  };

  const getStatusText = (status: NodeStatus) => {
    switch (status) {
      case NodeStatus.ON:
        return Intl.t('node.status.up');
      case NodeStatus.OFF:
        return Intl.t('node.status.down');
      case NodeStatus.CRASHED:
        return Intl.t('node.status.crashed');
      case NodeStatus.DESYNCED:
        return Intl.t('node.status.desynced');
      case NodeStatus.STOPPING:
        return Intl.t('node.status.stopping');
      case NodeStatus.STARTING:
        return Intl.t('node.status.starting');
      case NodeStatus.BOOTSTRAPPING:
        return Intl.t('node.status.bootstrapping');
      default:
        return Intl.t('node.status.unknown');
    }
  };

  const isLoading =
    status === NodeStatus.STARTING ||
    status === NodeStatus.BOOTSTRAPPING ||
    status === NodeStatus.STOPPING;

  return (
    <div
      className={
        `w-[20%] inline-flex items-center justify-center gap-2 px-6` +
        `py-1 rounded-full text-xs font-medium bg-opacity-20 ${getStatusColor(
          status,
        )} ${getStatusTextColor(status)}`
      }
    >
      {isLoading && <Spinner size={16} />}
      {getStatusText(status)}
    </div>
  );
};
