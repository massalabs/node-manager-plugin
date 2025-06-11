import React from 'react';

import { Toggle, Tooltip, toast } from '@massalabs/react-ui-kit';

import { usePost } from '@/hooks/api/usePost';
import Intl from '@/i18n/i18n';
import { autoRestartBody } from '@/models/nodeInfos';
import { useNodeStore } from '@/store/nodeStore';

const AutoRestart: React.FC = () => {
  const autoRestart = useNodeStore((state) => state.autoRestart);
  const setAutoRestart = useNodeStore((state) => state.setAutoRestart);
  const { mutate: setAutoRestartMutate } = usePost<unknown>(
    'autoRestart',
  ) as ReturnType<typeof usePost<unknown>>;

  const handleToggleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const checked = event.target.checked;
    setAutoRestartMutate({ autoRestart: checked } as autoRestartBody, {
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
