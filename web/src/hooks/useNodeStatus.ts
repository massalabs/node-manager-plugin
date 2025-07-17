import { useCallback, useEffect, useRef } from 'react';

import { useNavigate } from 'react-router-dom';

import intl from '@/i18n/i18n';
import { useNodeStore } from '@/store/nodeStore';
import { getErrorMessage, goToErrorPage, NodeStatus } from '@/utils';
import { getApiUrl } from '@/utils/utils';

export function useNodeStatus() {
  const eventSourceRef = useRef<EventSource | null>(null);
  const setStatus = useNodeStore((state) => state.setStatus);

  const navigate = useNavigate();

  /* use useCallback to avoid recreating a new function instance each time the hook is re-rendering
This function can be used in dependency array, so it needs to be a stable reference
*/
  const startListeningStatus = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const baseApi = getApiUrl() || '/api';
    const eventSource = new EventSource(`${baseApi}/status`);

    eventSource.onmessage = (event) => {
      console.log('Node status update received:', event.data);
      const status = event.data as NodeStatus;
      setStatus(status);
    };

    eventSource.onerror = (err) => {
      console.error('node status retrieving SSE error:', err);
      eventSource.close();
      goToErrorPage(
        navigate,
        intl.t('errors.node-status.title'),
        intl.t('errors.node-status.description', {
          error: getErrorMessage(err),
        }),
      );
    };

    eventSourceRef.current = eventSource;
  }, [navigate, setStatus]);

  useEffect(() => {
    // Cleanup on unmount
    return () => {
      console.log('Component unmounting, cleaning up EventSource connection');

      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
    };
  }, []);

  return { startListeningStatus, isListening: !!eventSourceRef.current };
}
