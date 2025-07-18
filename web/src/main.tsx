import React from 'react';

import { Toast } from '@massalabs/react-ui-kit';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import ReactDOM from 'react-dom/client';

import '@massalabs/react-ui-kit/src/global.css';
import './index.css';
import StakingAddressList from './staking/StakingAddressList';
import DownloadMassaWallet from '@/components/DownloadMassaWallet';
import { ErrorProvider, useError } from '@/contexts/ErrorContext';
import Error from '@/error';
import HistoryGraph from '@/graphHistory/HistoryGraph';
import { Header } from '@/header';
import { useInit } from '@/hooks/useInit';
import { NodeManager } from '@/node';

const queryClient = new QueryClient();

const MainContent: React.FC = () => {
  const { error, setError } = useError();
  const { massaWalletInstalled } = useInit(setError);

  // If there's an error, show error page
  if (error) {
    return (
      <div className="min-h-screen theme-dark bg-primary text-f-primary">
        <Header />
        <Error errorData={error} onReturn={() => setError(null)} />
        <Toast durationMs={1000} />
      </div>
    );
  }

  // If massaWalletInstalled is null, show loading
  if (massaWalletInstalled === null) {
    return (
      <div className="flex items-center justify-center h-screen theme-dark bg-primary text-f-primary">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-f-primary">Loading...</p>
        </div>
      </div>
    );
  }

  // If massaWalletInstalled is false, show download wallet
  if (massaWalletInstalled === false) {
    return (
      <div className="min-h-screen theme-dark bg-primary text-f-primary">
        <Header />
        <DownloadMassaWallet />
        <Toast durationMs={1000} />
      </div>
    );
  }

  // Main layout when Massa Wallet is installed
  return (
    <div className="min-h-screen theme-dark bg-primary text-f-primary">
      <div>
        <Header />
        <div className="p-[5%]">
          {/* First row: NodeManager and HistoryGraph */}
          <div className="flex gap-[5%] mb-[5%] h-[400px]">
            <div className="w-1/3">
              <NodeManager />
            </div>
            <div className="w-2/3">
              <HistoryGraph />
            </div>
          </div>

          {/* Second row: StakingAddressList */}
          <div className="h-[400px]">
            <StakingAddressList />
          </div>
        </div>
      </div>
      <Toast durationMs={1000} />
    </div>
  );
};

const Main: React.FC = () => {
  return (
    <ErrorProvider>
      <MainContent />
    </ErrorProvider>
  );
};

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <QueryClientProvider client={queryClient}>
    <Main />
  </QueryClientProvider>,
);
