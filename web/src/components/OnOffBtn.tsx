import React from 'react';
import { useNodeStore } from '@/store/nodeStore';
import { startNodeBody, startNodeReponse } from '@/models/nodeInfos';
import Intl from '@/i18n/i18n';
import { toast } from '@massalabs/react-ui-kit';
import { usePost } from '@/hooks/api/usePost';

const OnOffBtn: React.FC = () => {
    const { isRunning, setVersion } = useNodeStore();
    const nodeRunning = isRunning();
    const { mutate: startMutate, isLoading: isStarting } = usePost<startNodeReponse>('start') as ReturnType<typeof usePost<startNodeReponse>>;
    const { mutate: stopMutate, isLoading: isStopping } = usePost<unknown>('stop') as ReturnType<typeof usePost<unknown>>;

    const handleStart = () => {
        const payload: startNodeBody = { useBuildnet: true };
        startMutate(
            payload as unknown as startNodeReponse,
            {
                onSuccess: (data) => {
                    if (data && data.version) {
                        setVersion(data.version);
                    }
                    toast.success(Intl.t('home.button.startSuccess'));
                },
                onError: () => {
                    toast.error(Intl.t('home.button.startError'));
                },
            }
        );
    };

    const handleStop = () => {
        stopMutate(
            undefined,
            {
                onSuccess: () => {
                    toast.success(Intl.t('home.button.stopSuccess'));
                },
                onError: () => {
                    toast.error(Intl.t('home.button.stopError'));
                },
            }
        );
    };

    const handleClick = () => {
        if (!nodeRunning) {
            handleStart();
        } else {
            handleStop();
        }
    };

    return (
        <button
            className={`rounded-full px-6 py-2 text-white font-bold ${nodeRunning ? 'bg-red-500' : 'bg-green-500'}`}
            onClick={handleClick}
            disabled={isStarting || isStopping}
        >
            {nodeRunning ? Intl.t('home.button.off') : Intl.t('home.button.on')}
        </button>
    );
};

export default OnOffBtn;