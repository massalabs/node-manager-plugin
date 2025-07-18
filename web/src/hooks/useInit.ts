import { useEffect, useState } from 'react';

import { useNodeStatus } from './useNodeStatus';
import { useStakingListener } from './useStakingListener';
import Intl from '@/i18n/i18n';
import { useNodeStore } from '@/store/nodeStore';
import { getErrorMessage } from '@/utils';
import { networks } from '@/utils/const';
import { ErrorData } from '@/utils/error';
import { isMassaWalletInstalled } from '@/utils/station';
import { getPluginInfos } from '@/utils/utils';

export const useInit = (setError: (error: ErrorData | null) => void) => {
  const [massaWalletInstalled, setMassaWalletInstalled] = useState<
    boolean | null
  >(null);
  const { startListeningStatus } = useNodeStatus();
  const { initInfos } = useNodeStore();

  // Fetch staking addresses and start listening to changes
  useStakingListener();

  useEffect(() => {
    // Initialize plugin info on mount
    const initializePlugin = async () => {
      try {
        const pluginInfo = await getPluginInfos();
        initInfos(
          pluginInfo.autoRestart ?? false,
          pluginInfo.networks,
          pluginInfo.isMainnet ? networks.mainnet : networks.buildnet,
          pluginInfo.pluginVersion ?? '',
        );
      } catch (error) {
        console.error('Error initializing plugin info:', error);
        setError({
          title: Intl.t('errors.plugin-init.title'),
          message: Intl.t('errors.plugin-init.description', {
            error: getErrorMessage(error),
          }),
        });
      }
    };

    initializePlugin();
  }, [initInfos, setError]);

  useEffect(() => {
    // Setup node status listening
    startListeningStatus();
  }, [startListeningStatus]);

  useEffect(() => {
    // Check if Massa Wallet is installed
    isMassaWalletInstalled()
      .then((isInstalled) => {
        setMassaWalletInstalled(isInstalled);
      })
      .catch((error) => {
        console.error('Error checking if Massa Wallet is installed:', error);
        setError({
          title: Intl.t('errors.massa-wallet-check.title'),
          message: Intl.t('errors.massa-wallet-check.description', {
            error: getErrorMessage(error),
          }),
        });
      });
  }, [setError]);

  return { massaWalletInstalled };
};
