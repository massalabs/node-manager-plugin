import React, { useState } from 'react';

import { Password } from '@massalabs/react-ui-kit';

import ConfirmModal from '@/components/ConfirmModal';
import { useStartNode } from '@/hooks/node-manager/useStartNode';
import { useStopNode } from '@/hooks/node-manager/useStopNode';
import Intl from '@/i18n/i18n';
import { useNodeStore } from '@/store/nodeStore';
import { isRunning, NodeStatus } from '@/utils';

export const OnOffBtn: React.FC = () => {
  const [isPasswordModalOpen, setIsPasswordModalOpen] = useState(false);
  const [password, setPassword] = useState('');
  const [passwordError, setPasswordError] = useState('');

  const status = useNodeStore((state) => state.status);
  const getHasPwd = useNodeStore((state) => state.getHasPwd);

  const { isStarting, startNode } = useStartNode();
  const { isStopping, stopNode } = useStopNode();

  const nodeRunning = isRunning(status);
  const isDisabled =
    isStarting ||
    isStopping ||
    status === NodeStatus.STARTING ||
    status === NodeStatus.STOPPING;

  const handleClick = () => {
    if (!nodeRunning) {
      if (!getHasPwd()) {
        setIsPasswordModalOpen(true);
      } else {
        startNode('');
      }
    } else {
      stopNode();
    }
  };

  const handleSubmitPassword = () => {
    startNode(password);
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

  const buttonText = nodeRunning
    ? Intl.t('node.button.close')
    : Intl.t('node.button.run');

  return (
    <>
      <button
        onClick={handleClick}
        disabled={isDisabled}
        className={
          'w-full bg-white text-gray-800 font-medium py-3 px-6 rounded-lg hover:bg-gray-100' +
          'disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center gap-2'
        }
      >
        {isStarting && (
          <div className="w-2 h-2 bg-yellow-400 rounded-full animate-pulse" />
        )}
        {buttonText}
      </button>

      <ConfirmModal
        isOpen={isPasswordModalOpen}
        onClose={handleClosePasswordModal}
        onConfirm={handleSubmitPassword}
        title={Intl.t('node.password.title')}
      >
        <div className="flex flex-col gap-4">
          <p className="mas-body">{Intl.t('node.password.description')}</p>

          {/* Warning Zone */}
          <div className="bg-yellow-500/20 border border-yellow-500/50 rounded-lg p-3 text-center">
            <p className="mas-body text-yellow-300 text-sm">
              {Intl.t('node.password.warning')}
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
