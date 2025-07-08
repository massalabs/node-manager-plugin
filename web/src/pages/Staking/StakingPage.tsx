import React, { useEffect, useState } from 'react';

import { useNavigate } from 'react-router-dom';

import DownloadMassaWallet from './DownloadMassaWallet';
import Loading from './Loading';
import StakingDashboard from './StakingDashboard';
import { isMassaWalletInstalled } from './utils/station';
import { useStakingListener } from '@/hooks/useStakingListener';
import Intl from '@/i18n/i18n';
import NodeNotReady from '@/pages/Staking/NodeNotReady';
import { useNodeStore } from '@/store/nodeStore';
import { getErrorMessage, goToErrorPage, NodeStatus } from '@/utils';

const StakingPage: React.FC = () => {
  const status = useNodeStore((state) => state.status);
  const [massaWalletInstalled, setMassaWalletInstalled] = useState<
    boolean | null
  >(null);
  const navigate = useNavigate();

  // Fetch staking addresses and start listening to changes
  useStakingListener();

  useEffect(() => {
    isMassaWalletInstalled()
      .then((isInstalled) => {
        setMassaWalletInstalled(isInstalled);
      })
      .catch((error) => {
        console.error('Error checking if Massa Wallet is installed:', error);
        goToErrorPage(
          navigate,
          Intl.t('errors.massa-wallet-check.title'),
          Intl.t('errors.massa-wallet-check.description', {
            error: getErrorMessage(error),
          }),
        );
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
