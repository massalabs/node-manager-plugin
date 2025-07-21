import React from 'react';

import { Toggle, Tooltip } from '@massalabs/react-ui-kit';

import { Logs } from './Logs';
import { OnOffBtn } from './OnOffBtn';
import { Status } from './Status';
import { useAutoRestart } from '../hooks/node-manager/useAutoRestart';
import Intl from '@/i18n/i18n';
import { useNodeStore } from '@/store/nodeStore';
import { NodeStatus } from '@/utils';

export const NodeManager: React.FC = () => {
  const { autoRestart, handleToggleChange } = useAutoRestart();
  const status = useNodeStore((state) => state.status);
  const nodeRunning = status === NodeStatus.ON;

  return (
    <div className="bg-secondary rounded-2xl p-8 w-full h-full relative border border-gray-700 flex flex-col">
      {/* Title */}
      <h2 className="text-2xl font-bold text-white mb-8">
        {Intl.t('node.title')}
      </h2>

      {/* Status Rows */}
      <div className="space-y-4 mb-8">
        {/* Status Row */}
        <div className="flex justify-between items-center">
          <span className="text-white">{Intl.t('node.status.label')}</span>
          <Status />
        </div>

        {/* Auto Restart Row */}
        <div className="flex justify-between items-center">
          <div className="flex items-center gap-2">
            <span className="text-white">
              {Intl.t('node.autoRestart.label')}
            </span>
            <Tooltip body={Intl.t('node.autoRestart.description')} />
          </div>
          <Toggle checked={autoRestart} onChange={handleToggleChange} />
        </div>
      </div>

      {/* Public API - only show when node is running */}
      {nodeRunning && (
        <div className="mb-8">
          <div className="text-white text-sm flex justify-between">
            <span className="font-medium">Public API:</span>
            <span className="text-gray-300">localhost:33035</span>
          </div>
        </div>
      )}

      {/* Spacer to push button to bottom */}
      <div className="flex-1"></div>

      {/* Action Button - positioned at bottom with constant distance */}
      <div className="mb-20">
        <OnOffBtn />
      </div>

      {/* Logs Dropdown - positioned at bottom left */}
      <div className="absolute bottom-6 left-6">
        <Logs />
      </div>
    </div>
  );
};
