import { useRef } from 'react';
import { useNodeStore, NodeStatus } from '@/store/nodeStore';
import { useNavigate } from 'react-router-dom';
import intl from '@/i18n/i18n';

export function useNodeStatus() {
    const eventSourceRef = useRef<EventSource | null>(null);
    const setStatus = useNodeStore((state) => state.setStatus);
    const navigate = useNavigate();

    const startListeningStatus = () => {
        if (eventSourceRef.current) {
            eventSourceRef.current.close();
        }

        const baseApi = import.meta.env.VITE_BASE_API || '/api';
        const eventSource = new EventSource(`${baseApi}/status`);

        eventSource.onmessage = (event) => {
            console.log('\n\n Node status update received:', event.data);
            const status = event.data as NodeStatus;
            setStatus(status);
        };

        eventSource.onerror = (err) => {
            eventSource.close();
            navigate('/error', {
                state: {
                    error: {
                        title: intl.t('errors.node-status.title'),
                        message: intl.t('errors.node-status.description', {
                            error: err instanceof Error ? err.message : String(err)
                        }),
                    },
                },
            });
        };
        eventSourceRef.current = eventSource;
    };

    return { startListeningStatus, isListening: !!eventSourceRef.current};
}
