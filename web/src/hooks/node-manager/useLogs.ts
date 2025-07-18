import { useState } from 'react';

import { toast } from '@massalabs/react-ui-kit/src/components/Toast';
import axios from 'axios';

import Intl from '@/i18n/i18n';
import { networks } from '@/utils/const';
import { getApiUrl } from '@/utils/utils';

export const useLogs = () => {
  const [isLoading, setIsLoading] = useState(false);

  const downloadLogs = async (networkToUse: networks) => {
    try {
      setIsLoading(true);
      const baseApi = getApiUrl();
      const response = await axios.get(
        `${baseApi}/nodeLogs?isMainnet=${networkToUse === networks.mainnet}`,
      );

      const logs = response.data;

      if (logs) {
        // Create an invisible download link
        const blob = new Blob([logs], { type: 'text/plain' });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `massa-node-${networkToUse}-${new Date().toISOString()}.log`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
      } else {
        toast(Intl.t('node.logs.noLogs'));
      }
    } catch (error) {
      console.error('Error downloading logs:', error);
      toast.error(Intl.t('node.logs.error'));
    } finally {
      setIsLoading(false);
    }
  };

  return {
    isLoading,
    downloadLogs,
  };
};
