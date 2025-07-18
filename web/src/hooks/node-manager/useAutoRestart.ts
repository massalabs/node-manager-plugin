import { toast } from '@massalabs/react-ui-kit';

import { usePost } from '@/hooks/usePost';
import Intl from '@/i18n/i18n';
import { autoRestartBody } from '@/models/nodeInfos';
import { useNodeStore } from '@/store/nodeStore';

export const useAutoRestart = () => {
  const autoRestart = useNodeStore((state) => state.autoRestart);
  const setAutoRestart = useNodeStore((state) => state.setAutoRestart);

  const { mutate: setAutoRestartMutate } = usePost<autoRestartBody>(
    'autoRestart',
  ) as ReturnType<typeof usePost<autoRestartBody>>;

  const handleToggleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const checked = event.target.checked;
    setAutoRestartMutate({ autoRestart: checked } as autoRestartBody, {
      onSuccess: () => {
        setAutoRestart(checked);
        if (checked) {
          toast.success(Intl.t('node.autoRestart.enabled'));
        } else {
          toast.success(Intl.t('node.autoRestart.disabled'));
        }
      },
      onError: () => {
        toast.error(Intl.t('node.autoRestart.error'));
      },
    });
  };

  return {
    autoRestart,
    handleToggleChange,
  };
};
