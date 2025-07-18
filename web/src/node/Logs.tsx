import React, { useState } from 'react';

import { Spinner } from '@massalabs/react-ui-kit';

import { useLogs } from '@/hooks/node-manager/useLogs';
import Intl from '@/i18n/i18n';
import { networks } from '@/utils/const';

export const Logs: React.FC = () => {
  const [isOpen, setIsOpen] = useState(false);
  const { isLoading, downloadLogs } = useLogs();

  const handleDownload = (network: networks) => {
    downloadLogs(network);
    setIsOpen(false);
  };

  return (
    <div className="relative">
      <button
        onClick={() => setIsOpen(!isOpen)}
        disabled={isLoading}
        className={
          'flex items-center gap-2 text-white text-sm hover:text-gray-300' +
          'disabled:opacity-50 disabled:cursor-not-allowed'
        }
      >
        {isLoading && <Spinner />}
        {Intl.t('node.logs.export')}
        <svg
          className={`w-4 h-4 transition-transform ${
            isOpen ? 'rotate-180' : ''
          }`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M19 9l-7 7-7-7"
          />
        </svg>
      </button>

      {isOpen && (
        <div
          className={
            'absolute bottom-full left-0 mb-2 bg-gray-800 border' +
            'border-gray-700 rounded-lg shadow-lg z-10 min-w-[120px]'
          }
        >
          <div className="py-1">
            <button
              onClick={() => handleDownload(networks.mainnet)}
              disabled={isLoading}
              className={
                'w-full px-3 py-2 text-left text-sm text-white hover:bg-gray-700' +
                'disabled:opacity-50 disabled:cursor-not-allowed'
              }
            >
              {Intl.t('node.logs.mainnet')}
            </button>
            <button
              onClick={() => handleDownload(networks.buildnet)}
              disabled={isLoading}
              className={
                'w-full px-3 py-2 text-left text-sm text-white hover:bg-gray-700' +
                'disabled:opacity-50 disabled:cursor-not-allowed'
              }
            >
              {Intl.t('node.logs.buildnet')}
            </button>
          </div>
        </div>
      )}
    </div>
  );
};
