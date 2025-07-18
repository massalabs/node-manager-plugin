import { toast } from '@massalabs/react-ui-kit';

import { usePost } from '@/hooks/usePost';
import Intl from '@/i18n/i18n';

export const useStopNode = () => {
  const { mutate: stopMutate, isLoading: isStopping } = usePost<unknown>(
    'stop',
  ) as ReturnType<typeof usePost<unknown>>;

  const stopNode = () => {
    stopMutate(undefined, {
      onSuccess: () => {
        toast.success(Intl.t('home.stopSuccess'));
      },
      onError: () => {
        toast.error(Intl.t('home.stopError'));
      },
    });
  };

  return {
    isStopping,
    stopNode,
  };
};
