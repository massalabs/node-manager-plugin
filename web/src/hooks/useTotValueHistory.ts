import { useState, useEffect, useCallback, useRef } from 'react';

import { toast } from '@massalabs/react-ui-kit';
import axios from 'axios';

import {
  ValueHistoryPoint,
  SinceFetch,
  ValueHistorySamplesResponse,
} from '../models/history';
import { StakingAddress } from '../models/staking';
import { useError } from '@/contexts/ErrorContext';
import { useNodeStore } from '@/store/nodeStore';
import { useStakingStore } from '@/store/stakingStore';
import { networks } from '@/utils/const';
import { getErrorMessage } from '@/utils/error';

const ROLL_PRICE = 100.0;

export function getTotalValue(addresses: StakingAddress[]): number {
  let totalValue = 0;
  for (const addr of addresses) {
    let deferredCredits = 0;
    for (const defCredit of addr.deferred_credits) {
      deferredCredits += defCredit.amount;
    }
    totalValue +=
      addr.final_balance + addr.final_roll_count * ROLL_PRICE + deferredCredits;
  }
  return totalValue;
}

export function useTotValueHistory() {
  const [valueHistory, setValueHistory] = useState<ValueHistoryPoint[]>([]);
  const [since, setSince] = useState<SinceFetch>(SinceFetch.DEFAULT);

  const stakingAddresses = useStakingStore((state) => state.stakingAddresses);
  const network = useNodeStore((state) => state.currentNetwork);
  const pluginVersion = useNodeStore((state) => state.pluginVersion);
  const { setError } = useError();

  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  const getParamsFromSince = useCallback((since: SinceFetch) => {
    let sinceParam = '';
    let sampleNum = 0;
    const now = Date.now();
    switch (since) {
      case SinceFetch.H1:
        sinceParam = new Date(now - 1 * 60 * 60 * 1000).toISOString(); // 1 hour
        sampleNum = 20;
        break;
      case SinceFetch.D1:
        sinceParam = new Date(now - 24 * 60 * 60 * 1000).toISOString(); // 1 day
        sampleNum = 400;
        break;
      case SinceFetch.W1:
        sinceParam = new Date(now - 7 * 24 * 60 * 60 * 1000).toISOString(); // 1 week
        sampleNum = 700;
        break;
      case SinceFetch.M1:
        sinceParam = new Date(now - 30 * 24 * 60 * 60 * 1000).toISOString(); // 1 month (approx)
        sampleNum = 1000;
        break;
      case SinceFetch.Y1:
        sinceParam = new Date(now - 365 * 24 * 60 * 60 * 1000).toISOString(); // 1 year
        sampleNum = 1000;
        break;
      default:
        sinceParam = new Date(now - 24 * 60 * 60 * 1000).toISOString(); // 1 day
        sampleNum = 400;
        break;
    }
    return { sinceParam, sampleNum };
  }, []);

  const updateValueHistory = useCallback(() => {
    const value = getTotalValue(stakingAddresses);
    setValueHistory((prev) => [
      ...prev,
      { timestamp: new Date().toISOString(), value },
    ]);
  }, [stakingAddresses]);

  // allow to continusly update value history at the same interval it is updated in the backend.
  useEffect(() => {
    // Clear previous interval if any
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }

    if (valueHistory.length === 0) {
      return;
    }

    // Start interval for live value appending
    const { sinceParam, sampleNum } = getParamsFromSince(since);
    const sinceDate = new Date(sinceParam).getTime();

    const intervalMs = Math.floor((Date.now() - sinceDate) / sampleNum);
    intervalRef.current = setInterval(updateValueHistory, intervalMs);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [since, updateValueHistory, getParamsFromSince]);

  const fetchValueHistory = useCallback(
    async (since: SinceFetch) => {
      const { sinceParam, sampleNum } = getParamsFromSince(since);

      if (pluginVersion === '') {
        // it means that the page has been reloaded and the network is not set yet
        return;
      }

      try {
        const res = await axios.get<ValueHistorySamplesResponse>(
          '/api/valueHistory',
          {
            params: {
              since: sinceParam,
              sampleNum,
              isMainnet: network == networks.mainnet,
            },
          },
        );
        if (!res.data.samples || sampleNum - res.data.emptyDataPointNum < 5) {
          toast.error('Not enough data for graph');
          setValueHistory([]);
          return;
        }
        setValueHistory(res.data.samples);
        setSince(since);
      } catch (err) {
        setError({
          title: 'Error fetching value history',
          message: getErrorMessage(err),
        });
        return;
      }
    },
    [getParamsFromSince, network, pluginVersion, setError],
  );

  // Cleanup interval on unmount
  useEffect(() => {
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current);
      }
    };
  }, []);

  return { valueHistory, fetchValueHistory, SinceFetch };
}
