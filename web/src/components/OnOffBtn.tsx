import React from 'react';
import { useNodeStore } from '@/store/nodeStore';
import Intl from '@/i18n/i18n';

const OnOffBtn: React.FC = () => {
    const { isRunning } = useNodeStore();

    const buttonState = isRunning();

    return (
        <button
            className={`rounded-full px-6 py-2 text-white font-bold ${buttonState ? 'bg-red-500' : 'bg-green-500'}`}
        >
            {buttonState ? Intl.t('home.button.off') : Intl.t('home.button.on')}
        </button>
    );
};

export default OnOffBtn;