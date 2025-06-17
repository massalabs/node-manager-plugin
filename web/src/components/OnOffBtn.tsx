import React, { useState } from 'react';

import { toast, Password, PopupModal, PopupModalContent, PopupModalHeader, PopupModalFooter } from '@massalabs/react-ui-kit';

import { usePost } from '@/hooks/api/usePost';
import Intl from '@/i18n/i18n';
import { startNodeBody, startNodeReponse } from '@/models/nodeInfos';
import { useNodeStore } from '@/store/nodeStore';
import { isRunning, networks } from '@/utils';

const OnOffBtn: React.FC = () => {
  // TODO: password handling is not ready for use yet
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [password, setPassword] = useState('');

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

  const handleStart = (password: string) => {
    const payload: startNodeBody = {
      useBuildnet: network === networks.buildnet,
      password: password,
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
      //setIsModalOpen(true);
      handleStart(password);
    } else {
      handleStop();
    }
  };

  const handleSubmitPassword = () => {
    handleStart(password);
    setIsModalOpen(false);
    setPassword('');
  };

  return (
    <>
     <button
      className={`rounded-full px-6 py-2 text-white font-bold ${
        nodeRunning ? 'bg-red-500' : 'bg-green-500'
      }`}
      onClick={handleClick}
      disabled={isStarting || isStopping}
    >
      {nodeRunning ? Intl.t('home.button.off') : Intl.t('home.button.on')}
    </button>

    {isModalOpen && <PopupModal
      fullMode={true}
      customClass="w-[520px] h-[200px]"
      onClose={() => {
        setIsModalOpen(false);
      }}
    >
      <PopupModalHeader>
      <p className="mas-title mb-6">
        {Intl.t('home.nodePassword.title')}
      </p>
      </PopupModalHeader>
      <PopupModalContent>
        <div className="flex flex-col gap-4">
          <p className="mas-body">
            {Intl.t('home.nodePassword.description')}
          </p>
          <Password 
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
         
        </div>
      </PopupModalContent>
      <PopupModalFooter>
        <div className="flex justify-end w-full mt-4">
          <button 
            className="bg-green-500 text-white font-bold px-4 py-2 rounded"
            onClick={handleSubmitPassword}
            >
              {Intl.t('home.nodePassword.submit')}
          </button>
        </div>
      </PopupModalFooter>
    </PopupModal>}
    </>
   
  );
};

export default OnOffBtn;
