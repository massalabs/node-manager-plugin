import React from 'react';
import { Toggle, Tooltip } from '@massalabs/react-ui-kit';
import { useNodeStore } from '@/store/nodeStore';
import Intl from '@/i18n/i18n';

const AutoRestart: React.FC = () => {
  const {autoRestart, setAutoRestart} = useNodeStore();

  const handleToggleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setAutoRestart(event.target.checked);
  };

  return (
    <div className="flex justify-center gap-5">
      <label htmlFor="auto-restart-toggle" className="flex items-center gap-2">
        <Tooltip body={Intl.t('home.auto-restart.description')}/>
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