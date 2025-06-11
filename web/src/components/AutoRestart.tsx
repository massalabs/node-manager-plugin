import React, { useEffect } from 'react';

import { Toggle, Tooltip, toast } from '@massalabs/react-ui-kit';
import axios from 'axios';

import { usePost } from '@/hooks/api/usePost';
import Intl from '@/i18n/i18n';
import { configBody } from '@/models/nodeInfos';
import { useNodeStore } from '@/store/nodeStore';

const AutoRestart: React.FC = () => {
  const autoRestart = useNodeStore((state) => state.autoRestart);
  const setAutoRestart = useNodeStore((state) => state.setAutoRestart);
  const { mutate: setAutoRestartMutate } = usePost<unknown>(
    'config',
  ) as ReturnType<typeof usePost<unknown>>;

  /* retrieve AutoRestart from api when the component mount */
  useEffect(() => {
    axios
      .get<configBody>(`${import.meta.env.VITE_BASE_API}/config`)
      .then(({ data }) => {
        setAutoRestart(data.autoRestart ?? false);
      });
  }, [setAutoRestart]);

  const handleToggleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const checked = event.target.checked;
    setAutoRestartMutate({ autoRestart: checked } as configBody, {
      onSuccess: () => {
        setAutoRestart(checked);
        if (checked) {
          toast.success(Intl.t('home.auto-restart.enabled'));
        } else {
          toast.success(Intl.t('home.auto-restart.disabled'));
        }
      },
    });
  };

  return (
    <div className="flex justify-center gap-5">
      <label htmlFor="auto-restart-toggle" className="flex items-center gap-2">
        <Tooltip body={Intl.t('home.auto-restart.description')} />
        {Intl.t('home.auto-restart.label')}
      </label>

      <Toggle
        id="auto-restart-toggle"
        checked={autoRestart}
        onChange={handleToggleChange}
      />
    </div>
  );
};

export default AutoRestart;
