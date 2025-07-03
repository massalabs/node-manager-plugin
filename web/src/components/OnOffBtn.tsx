import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';

import {
  toast,
  Password,
} from '@massalabs/react-ui-kit';

import ConfirmModal from '@/components/ConfirmModal';
import { usePost } from '@/hooks/api/usePost';
import Intl from '@/i18n/i18n';
import { startNodeBody, startNodeReponse } from '@/models/nodeInfos';
import { useNodeStore } from '@/store/nodeStore';
import { isRunning, networks } from '@/utils';
import { getErrorPath } from '@/utils/error';

const OnOffBtn: React.FC = () => {
  const [isPasswordModalOpen, setIsPasswordModalOpen] = useState(false);
  const [password, setPassword] = useState('');
  const navigate = useNavigate();
  const setVersion = useNodeStore((state) => state.setVersion);
  const status = useNodeStore((state) => state.status);
  const network = useNodeStore((state) => state.network);
  const hasPwd = useNodeStore((state) => state.hasPwd);
  const setHasPwd = useNodeStore((state) => state.setHasPwd);

  const { mutate: startMutate, isLoading: isStarting } =
    usePost<startNodeReponse>('start') as ReturnType<
      typeof usePost<startNodeReponse>
    >;
  const { mutate: stopMutate, isLoading: isStopping } = usePost<unknown>(
    'stop',
  ) as ReturnType<typeof usePost<unknown>>;


  const nodeRunning = isRunning(status);
  
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
        setHasPwd(true);
        toast.success(Intl.t('home.startSuccess'));
      },
      onError: (err) => {
        console.error('Error starting node:', err);
        navigate(getErrorPath(), {
          state: {
            error: {
              title: Intl.t('errors.start-node.title'),
              message: Intl.t('errors.start-node.description', {
                error: err instanceof Error ? err.message : String(err),
              }),
            },
          },
        });
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
      if (!hasPwd) {
        setIsPasswordModalOpen(true);
      } else {
        handleStart('');
      }
    } else {
      handleStop();
    }
  };

  const handleSubmitPassword = () => {
    handleStart(password);
    setIsPasswordModalOpen(false);
    setPassword('');
  };

  const handleClosePasswordModal = () => {
    setIsPasswordModalOpen(false);
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

      <ConfirmModal
        isOpen={isPasswordModalOpen}
        onClose={handleClosePasswordModal}
        onConfirm={handleSubmitPassword}
        title={Intl.t('home.nodePassword.title')}
      >
        <div className="flex flex-col gap-4">
          <p className="mas-body">
            {Intl.t('home.nodePassword.description')}
          </p>
          
          {/* Warning Zone */}
          <div className="bg-yellow-500/20 border border-yellow-500/50 rounded-lg p-3 text-center">
            <p className="mas-body text-yellow-300 text-sm">
              {Intl.t('home.nodePassword.warning')}
            </p>
          </div>
          
          <Password
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
      </ConfirmModal>
    </>
  );
};

export default OnOffBtn;
