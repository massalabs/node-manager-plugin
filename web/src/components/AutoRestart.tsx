import React, { useEffect } from 'react';

import { Toggle, Tooltip, toast } from '@massalabs/react-ui-kit';
import axios from 'axios';

import { usePost } from '@/hooks/api/usePost';
import Intl from '@/i18n/i18n';
import { autoRestartBody, nodeInfosResponse } from '@/models/nodeInfos';
import { useNodeStore } from '@/store/nodeStore';
import { networks } from '@/utils/const';

const AutoRestart: React.FC = () => {
  const autoRestart = useNodeStore((state) => state.autoRestart);
  const setAutoRestart = useNodeStore((state) => state.setAutoRestart);
  const setNetwork = useNodeStore((state) => state.setNetwork);

  const { mutate: setAutoRestartMutate } = usePost<autoRestartBody>(
    'autoRestart',
  ) as ReturnType<typeof usePost<autoRestartBody>>;

  /* retrieve AutoRestart from api when the component mount */
  useEffect(() => {
    axios
      .get<nodeInfosResponse>(`${import.meta.env.VITE_BASE_API}/nodeInfos`)
      .then(({ data }) => {
        setAutoRestart(data.autoRestart ?? false);
        setNetwork(getNetworkFromVersion(data.version));
      });
  }, [setAutoRestart, setNetwork]);

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
      onError: () => {
        toast.error(Intl.t('home.auto-restart.error'));
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

function getNetworkFromVersion(version: string): networks {
  if (version.includes('mainnet')) {
    return networks.mainnet;
  }
  return networks.buildnet;
}

export default AutoRestart;
