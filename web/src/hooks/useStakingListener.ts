import { useEffect, useRef } from 'react';

import { useNavigate } from 'react-router-dom';

import intl from '@/i18n/i18n';
import { StakingAddress } from '@/models/staking';
import { useNodeStore } from '@/store/nodeStore';
import { useStakingStore } from '@/store/stakingStore';
import { goToErrorPage, isStopStakingMonitoring, NodeStatus } from '@/utils';

export function useStakingListener() {
  const eventSourceRef = useRef<EventSource | null>(null);
  const setStakingAddresses = useStakingStore(
    (state) => state.setStakingAddresses,
  );
  const status = useNodeStore((state) => state.status);

  const navigate = useNavigate();

  const startListeningStakingAddresses = () => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    const baseApi = import.meta.env.VITE_BASE_API || '/api';
    const eventSource = new EventSource(`${baseApi}/staking-addresses`);

    eventSource.onmessage = (event) => {
      console.log('Staking addresses update received:', event.data);
      try {
        const stakingAddresses: StakingAddress[] = JSON.parse(event.data);
        setStakingAddresses(stakingAddresses);
      } catch (error) {
        console.error('Failed to parse staking addresses data:', error);
      }
    };

    eventSource.onerror = (err) => {
      console.error('Staking addresses retrieving SSE error:', err);
      eventSource.close();
      goToErrorPage(
        navigate,
        intl.t('errors.get-staking-addresses.title'),
        intl.t('errors.get-staking-addresses.description', {
          error: err instanceof Error ? err.message : String(err),
        }),
      );
    };

    eventSourceRef.current = eventSource;
  };

  useEffect(() => {
    if (status === NodeStatus.ON) {
      startListeningStakingAddresses();
    }

    if (eventSourceRef.current?.OPEN && isStopStakingMonitoring(status)) {
      eventSourceRef.current?.close();
    }

    // Cleanup on unmount
    return () => {
      console.log(
        'Component unmounting, cleaning up staking addresses EventSource connection',
      );

      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
    };
  }, [status, startListeningStakingAddresses]);
}
