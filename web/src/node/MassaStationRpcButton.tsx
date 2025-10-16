import React from 'react';

import { Button, Tooltip } from '@massalabs/react-ui-kit';
import { FiExternalLink } from 'react-icons/fi';

import { DEFAULT_JSON_RPC_PORT, DEFAULT_RPC_NAME } from '../utils';
import { getMassaStationUrl } from '../utils/station';
import Intl from '@/i18n/i18n';

export const MassaStationRpcButton: React.FC = () => {
  const handleClick = () => {
    const jsonRpcUrl = `http://localhost:${DEFAULT_JSON_RPC_PORT}`;
    const nodeUrl = encodeURIComponent(jsonRpcUrl);
    const name = encodeURIComponent(DEFAULT_RPC_NAME);
    const configUrl = `${getMassaStationUrl()}web/config?name=${name}&url=${nodeUrl}`;
    window.open(configUrl, '_blank');
  };

  return (
    <Tooltip body={Intl.t('node.massaStationRpc.description')}>
      <Button
        variant="secondary"
        onClick={handleClick}
        className="flex items-center gap-2"
      >
        <FiExternalLink className="w-4 h-4" />
        {Intl.t('node.massaStationRpc.button')}
      </Button>
    </Tooltip>
  );
};
