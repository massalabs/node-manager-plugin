import { useEffect, useRef } from 'react';

import { useNavigate } from 'react-router-dom';

import intl from '@/i18n/i18n';
import { useNodeStore } from '@/store/nodeStore';
import { getErrorPath, NodeStatus } from '@/utils';

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
      console.error('node status retrieving SSE error:', err);
      eventSource.close();
      navigate(getErrorPath(), {
        state: {
          error: {
            title: intl.t('errors.node-status.title'),
            message: intl.t('errors.node-status.description', {
              error: err instanceof Error ? err.message : String(err),
            }),
          },
        },
      });
    };

    eventSourceRef.current = eventSource;
  };

  useEffect(() => {
    // Cleanup on unmount
    return () => {
      console.log('Component unmounting, cleaning up EventSource connection');

      if (eventSourceRef.current) {
        eventSourceRef.current.close();
        eventSourceRef.current = null;
      }
    };
  }, []);

  return { startListeningStatus, isListening: !!eventSourceRef.current };
}
