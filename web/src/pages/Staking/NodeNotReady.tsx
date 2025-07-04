import React from 'react';

import { FetchingLine } from '@massalabs/react-ui-kit';
import { FiPower } from 'react-icons/fi';

import Intl from '@/i18n/i18n';

const NodeNotReady: React.FC = () => {
  return (
    <div className="flex items-center justify-center h-full">
      <div className="text-center">
        <div className="flex justify-center mb-4">
          <FiPower className="text-6xl text-gray-400" />
        </div>
        <div className="flex justify-center w-full">
          <FetchingLine width={40} height={4} />
        </div>
        <p className="mas-body text-f-primary text-lg">
          {Intl.t('staking.node-not-running')}
        </p>
      </div>
    </div>
  );
};

export default NodeNotReady;
