import { toast } from '@massalabs/react-ui-kit';
import { AxiosError } from 'axios';

import { useError } from '@/contexts/ErrorContext';
import { usePost } from '@/hooks/usePost';
import Intl from '@/i18n/i18n';
import { startNodeBody } from '@/models/nodeInfos';
import { useNodeStore } from '@/store/nodeStore';
import { getErrorMessage, networks } from '@/utils';

export const useStartNode = () => {
  const { setError } = useError();
  const network = useNodeStore((state) => state.currentNetwork);
  const setHasPwd = useNodeStore((state) => state.setHasPwd);

  const { mutate: startMutate, isLoading: isStarting } = usePost<unknown>(
    'start',
  ) as ReturnType<typeof usePost<unknown>>;

  const startNode = (password: string) => {
    const payload: startNodeBody = {
      useBuildnet: network === networks.buildnet,
      password: password,
    };

    // network may change between the start of the request and the result, so we need to save the current network
    const currentNetwork = network;

    startMutate(payload, {
      onSuccess: () => {
        setHasPwd(true, currentNetwork);
        toast.success(Intl.t('node.startSuccess'));
      },
      onError: (err: AxiosError) => {
        console.error('Error starting node:', err);
        setError({
          title: Intl.t('errors.start-node.title'),
          message: Intl.t('errors.start-node.description', {
            error: getErrorMessage(err),
          }),
        });
      },
    });
  };

  return {
    isStarting,
    startNode,
  };
};
