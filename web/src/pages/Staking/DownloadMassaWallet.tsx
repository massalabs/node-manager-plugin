import React from 'react';

import { MassaWallet } from '@massalabs/react-ui-kit';
import { FiDownload } from 'react-icons/fi';

import Intl from '@/i18n/i18n';

const DownloadMassaWallet: React.FC = () => {
  const handleDownloadClick = () => {
    // Open Massa Station Wallet download page
    window.open('https://station.massa/web/store', '_blank');
  };

  return (
    <div className="flex items-center justify-center h-full">
      <div className="text-center max-w-md">
        <div className="mb-6">
          <div className="flex justify-center">
            <MassaWallet className="w-400 h-300"/>
          </div>
          
         
        </div>
        
        <h2 className="mas-title mb-4">
          {Intl.t('staking.download-wallet.title')}
        </h2>
        
        <p className="mas-body text-f-primary mb-6">
          {Intl.t('staking.download-wallet.description')}
        </p>
        
        <button
          onClick={handleDownloadClick}
          className="bg-green-500 text-white font-bold px-6 py-3 rounded-lg hover:bg-green-600 transition-colors flex items-center gap-2 mx-auto"
        >
          <FiDownload className="w-5 h-5" />
          {Intl.t('staking.download-wallet.button')}
        </button>
      </div>
    </div>
  );
};

export default DownloadMassaWallet; 