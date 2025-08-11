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
import { getApiUrl } from '@/utils/utils';

const ROLL_PRICE = 100.0;
/* The interval at which the value history list is updated with a new item in absence of new value from the backend*/
const INTERVAL_MS = 1000 * 60 * 10; // 10 minutes

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

function localTimezoneNow(): string {
  return new Date(
    Date.now() - new Date().getTimezoneOffset() * 60 * 1000,
  ).toISOString();
}

export function useTotValueHistory() {
  const [valueHistory, setValueHistory] = useState<ValueHistoryPoint[]>([]);
  const [nonEmptyDataPointRate, setNonEmptyDataPointRate] = useState<number>(0);
  const totValue = useRef<number>(0);

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

  /* When staking addresses total value changes, add this value to the value history
  and update the interval at which the value history is updated with a new item in absence of new value from the backend
  */
  useEffect(() => {
    const value = getTotalValue(stakingAddresses);

    // stakingAddresses could change without the total value to be changed (e.g. when a roll target is changed)
    if (value === totValue.current || valueHistory.length == 0) {
      return;
    }

    const incrementNonEmptyDataPointRate = () => {
      const currentNonEmptyDataPointNum =
        (nonEmptyDataPointRate * valueHistory.length) / 100;
      setNonEmptyDataPointRate(
        ((currentNonEmptyDataPointNum + 1) / (valueHistory.length + 1)) * 100,
      );
    };

    totValue.current = value;
    setValueHistory((prev) => [
      ...prev,
      {
        timestamp: localTimezoneNow(),
        value,
      },
    ]);
    incrementNonEmptyDataPointRate();

    // Clear previous interval if any
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }

    // If no new value is received from the backend, update the value history with the same value
    intervalRef.current = setInterval(() => {
      setValueHistory((prev) => [
        ...prev,
        { timestamp: localTimezoneNow(), value: totValue.current },
      ]);
      incrementNonEmptyDataPointRate();
    }, INTERVAL_MS);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [stakingAddresses]);

  const fetchValueHistory = useCallback(
    async (since: SinceFetch) => {
      const { sinceParam, sampleNum } = getParamsFromSince(since);

      if (pluginVersion === '') {
        // it means that the page has been reloaded and the network is not set yet
        return;
      }

      try {
        const res = await axios.get<ValueHistorySamplesResponse>(
          getApiUrl() + '/valueHistory',
          {
            params: {
              since: sinceParam,
              sampleNum,
              isMainnet: network == networks.mainnet,
            },
          },
        );
        if (!res.data.samples || res.data.samples.length == 0) {
          toast.error('Not enough data for graph');
          setValueHistory([]);
          return;
        }
        setValueHistory(res.data.samples);
        setNonEmptyDataPointRate(
          ((sampleNum - res.data.emptyDataPointNum) / sampleNum) * 100,
        );
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

  return { valueHistory, fetchValueHistory, nonEmptyDataPointRate };
}
