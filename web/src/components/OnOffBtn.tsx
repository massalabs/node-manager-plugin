import React, { useState } from 'react';

import { toast, Password } from '@massalabs/react-ui-kit';
import { AxiosError } from 'axios';
import { useNavigate } from 'react-router-dom';

import ConfirmModal from '@/components/ConfirmModal';
import { usePost } from '@/hooks/usePost';
import Intl from '@/i18n/i18n';
import { startNodeBody, startNodeReponse } from '@/models/nodeInfos';
import { useNodeStore } from '@/store/nodeStore';
import { getErrorMessage, isRunning, networks } from '@/utils';
import { goToErrorPage } from '@/utils/routes';

const OnOffBtn: React.FC = () => {
  const [isPasswordModalOpen, setIsPasswordModalOpen] = useState(false);
  const [password, setPassword] = useState('');
  const [passwordError, setPasswordError] = useState('');
  const navigate = useNavigate();
  const setVersion = useNodeStore((state) => state.setVersion);
  const status = useNodeStore((state) => state.status);
  const network = useNodeStore((state) => state.network);
  const getHasPwd = useNodeStore((state) => state.getHasPwd);
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

    // network may change between the start of the request and the result, so we need to save the current network
    const currentNetwork = network;

    startMutate(payload as unknown as startNodeReponse, {
      onSuccess: (data) => {
        if (data && data.version) {
          setVersion(data.version);
        }
        setHasPwd(true, currentNetwork);
        setPasswordError('');
        toast.success(Intl.t('home.startSuccess'));
      },
      onError: (err: AxiosError) => {
        console.error('Error starting node:', err);
        goToErrorPage(
          navigate,
          Intl.t('errors.start-node.title'),
          Intl.t('errors.start-node.description', {
            error: getErrorMessage(err),
          }),
        );
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
      if (!getHasPwd()) {
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
    setPasswordError('');
  };

  const handlePasswordChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setPassword(e.target.value);
    if (passwordError) {
      setPasswordError('');
    }
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
          <p className="mas-body">{Intl.t('home.nodePassword.description')}</p>

          {/* Warning Zone */}
          <div className="bg-yellow-500/20 border border-yellow-500/50 rounded-lg p-3 text-center">
            <p className="mas-body text-yellow-300 text-sm">
              {Intl.t('home.nodePassword.warning')}
            </p>
          </div>

          <Password
            value={password}
            onChange={handlePasswordChange}
            error={passwordError}
          />
        </div>
      </ConfirmModal>
    </>
  );
};

export default OnOffBtn;
