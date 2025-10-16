import React from 'react';

import { Toggle, Tooltip } from '@massalabs/react-ui-kit';
import { Clipboard } from '@massalabs/react-ui-kit';

import { Logs } from './Logs';
import { MassaStationRpcButton } from './MassaStationRpcButton';
import { OnOffBtn } from './OnOffBtn';
import { SelectNetwork } from './SelectNetwork';
import { Status } from './Status';
import { useAutoRestart } from '../hooks/node-manager/useAutoRestart';
import Intl from '@/i18n/i18n';
import { useNodeStore } from '@/store/nodeStore';
import {
  DEFAULT_GRPC_PORT,
  DEFAULT_JSON_RPC_PORT,
  showRpcAddButton,
} from '@/utils';

export const NodeManager: React.FC = () => {
  const { autoRestart, handleToggleChange } = useAutoRestart();
  const status = useNodeStore((state) => state.status);

  const jsonRpcUrl = `http://localhost:${DEFAULT_JSON_RPC_PORT}`;
  const grpcUrl = `http://localhost:${DEFAULT_GRPC_PORT}`;

  return (
    <div className="bg-secondary rounded-2xl p-8 w-full h-full relative border border-gray-700 flex flex-col">
      {/* Title */}
      <h2 className="text-2xl font-bold text-white mb-8">
        {Intl.t('node.title')}
      </h2>

      {/* Network Row */}
      <div className="flex justify-between items-center mb-8">
        <span className="text-white">{Intl.t('node.select-network')}</span>
        <SelectNetwork />
      </div>

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
      {showRpcAddButton(status) && (
        <div className="mb-8">
          <div className="text-white text-sm flex justify-between items-center mb-3">
            <span className="font-medium">JsonRPC API:</span>
            <div className="w-[50%]">
              <Clipboard
                rawContent={jsonRpcUrl}
                displayedContent={jsonRpcUrl}
                customClass="h-7"
              />
            </div>
          </div>
          <div className="text-white text-sm flex justify-between items-center mb-3">
            <span className="font-medium">gRPC API:</span>
            <div className="w-[50%]">
              <Clipboard
                rawContent={grpcUrl}
                displayedContent={grpcUrl}
                customClass="h-7"
              />
            </div>
          </div>
          <div className="text-white text-sm flex justify-between items-center">
            <span className="font-medium">
              {Intl.t('node.massaStationRpc.label')}
            </span>
            <MassaStationRpcButton />
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
