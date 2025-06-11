import React from 'react';

import { toast } from '@massalabs/react-ui-kit';

import { usePost } from '@/hooks/api/usePost';
import Intl from '@/i18n/i18n';
import { startNodeBody, startNodeReponse } from '@/models/nodeInfos';
import { useNodeStore } from '@/store/nodeStore';
import { isRunning, networks } from '@/utils';

const OnOffBtn: React.FC = () => {
  const setVersion = useNodeStore((state) => state.setVersion);
  const status = useNodeStore((state) => state.status);
  const network = useNodeStore((state) => state.network);
  const nodeRunning = isRunning(status);
  const { mutate: startMutate, isLoading: isStarting } =
    usePost<startNodeReponse>('start') as ReturnType<
      typeof usePost<startNodeReponse>
    >;
  const { mutate: stopMutate, isLoading: isStopping } = usePost<unknown>(
    'stop',
  ) as ReturnType<typeof usePost<unknown>>;

  const handleStart = () => {
    const payload: startNodeBody = {
      useBuildnet: network === networks.buildnet,
    };
    startMutate(payload as unknown as startNodeReponse, {
      onSuccess: (data) => {
        if (data && data.version) {
          setVersion(data.version);
        }
        toast.success(Intl.t('home.startSuccess'));
      },
      onError: (err) => {
        console.error('Error starting node:', err);
        toast.error(Intl.t('home.startError'));
      },
    });
  };

  const handleStop = () => {
    stopMutate(undefined, {
      onSuccess: () => {
        toast.success(Intl.t('home.stopSuccess'));
      },
      onError: () => {
        toast.error(Intl.t('home.stopError'));
      },
    });
  };

  const handleClick = () => {
    if (!nodeRunning) {
      handleStart();
    } else {
      handleStop();
    }
  };

  return (
    <button
      className={`rounded-full px-6 py-2 text-white font-bold ${
        nodeRunning ? 'bg-red-500' : 'bg-green-500'
      }`}
      onClick={handleClick}
      disabled={isStarting || isStopping}
    >
      {nodeRunning ? Intl.t('home.button.off') : Intl.t('home.button.on')}
    </button>
  );
};

export default OnOffBtn;
