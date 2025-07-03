import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { useNodeStore } from '@/store/nodeStore';
import { NodeStatus, Path, routeFor } from '@/utils';
import { isMassaWalletInstalled } from './utils/station';

import DownloadMassaWallet from './DownloadMassaWallet';
import NodeNotReady from '@/pages/Staking/NodeNotReady';
import StakingDashboard from './StakingDashboard';
import Loading from './Loading';
import Intl from '@/i18n/i18n';

const StakingPage: React.FC = () => {
  const status = useNodeStore((state) => state.status);
  const [massaWalletInstalled, setMassaWalletInstalled] = useState<boolean | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    isMassaWalletInstalled().then((isInstalled) => {
      setMassaWalletInstalled(isInstalled);
    }).catch((error) => {
      console.error('Error checking if Massa Wallet is installed:', error);
      navigate(routeFor(Path.error), {
        replace: true,
        state: {
          error: {
            title: Intl.t('errors.massa-wallet-check.title'),
            message: Intl.t('errors.massa-wallet-check.description', {
              error: error instanceof Error ? error.message : String(error),
            }),
          },
        },
      }); 
    });
  }, [navigate]);

  // If massaWalletInstalled is null, show fetching round
  if (massaWalletInstalled === null) {
    return <Loading message={Intl.t('staking.loading')} />;
  }

  // If massaWalletInstalled is false, show download wallet
  if (massaWalletInstalled === false) {
    return <DownloadMassaWallet />;
  }

  // If status is ON and massa wallet is installed, show staking dashboard
  return <StakingDashboard />;

  // If status is not NodeStatus.ON, show node not ready
  if (status !== NodeStatus.ON) {
    return <NodeNotReady />;
  }  

  
};

export default StakingPage; 