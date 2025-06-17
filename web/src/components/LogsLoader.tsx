import axios from 'axios';
import { useState } from 'react';

import { useNodeStore } from '@/store/nodeStore';
import Intl from '@/i18n/i18n';
import { networks } from '@/utils/const';

export default function LogsLoader() {
  const [isLoading, setIsLoading] = useState(false);
  const network = useNodeStore((state) => state.network);

  const handleDownload = async () => {
    try {
      setIsLoading(true);
      const baseApi = import.meta.env.VITE_BASE_API || '/api';
      const response = await axios.get(`${baseApi}/nodeLogs?isMainnet=${network === networks.mainnet}`);

      const logs = response.data;

      if (logs) {
        // Create an invisible download link
        const blob = new Blob([logs], { type: 'text/plain' });
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `massa-node-${network}-${new Date().toISOString()}.log`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        window.URL.revokeObjectURL(url);
      }
    } catch (error) {
      console.error('Error downloading logs:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <button
      onClick={handleDownload}
      disabled={isLoading}
      className="w-full px-4 py-2 bg-primary text-white rounded-lg
       hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2"
    >
      {isLoading ? (
        <>
          <svg
            className="animate-spin h-5 w-5 text-white"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            ></circle>
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962
              7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            ></path>
          </svg>
          {Intl.t('home.logs.loading')}
        </>
      ) : (
        <>
          <svg
            className="h-5 w-5"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
            xmlns="http://www.w3.org/2000/svg"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"
            />
          </svg>
          {Intl.t('home.logs.download')}
        </>
      )}
    </button>
  );
}
