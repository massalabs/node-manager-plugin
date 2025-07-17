import { useEffect, useRef, useCallback } from 'react';

import {
  JsonRpcPublicProvider,
  PublicAPI,
  NodeStatusInfo,
} from '@massalabs/massa-web3';
import { useQuery } from '@tanstack/react-query';

export const useFetchNodeInfo = (fetchInterval = 10000) => {
  const providerRef = useRef<JsonRpcPublicProvider | null>(null);

  useEffect(() => {
    providerRef.current = new JsonRpcPublicProvider(
      new PublicAPI('http://localhost:33035'),
    );
  }, []);

  const fetchNodeInfo = useCallback(async (): Promise<NodeStatusInfo> => {
    if (providerRef.current === null) {
      return {} as NodeStatusInfo;
    }
    return await providerRef.current.getNodeStatus();
  }, []);

  return useQuery({
    queryKey: ['fetchNodeInfo'],
    queryFn: fetchNodeInfo,
    refetchInterval: fetchInterval,
  });
};
