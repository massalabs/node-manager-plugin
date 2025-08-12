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
import { NodeStatus } from '@/utils/nodeStatus';
import { getApiUrl } from '@/utils/utils';

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

function localTimezoneNow(): string {
  return new Date(
    Date.now() - new Date().getTimezoneOffset() * 60 * 1000,
  ).toISOString();
}

interface SinceParams {
  timeMs: number;
  sampleNum: number;
}

const SINCE_PARAMS_MAP: Record<SinceFetch, SinceParams> = {
  [SinceFetch.H1]: {
    timeMs: 1 * 60 * 60 * 1000, // 1 hour
    sampleNum: 20,
  },
  [SinceFetch.D1]: {
    timeMs: 24 * 60 * 60 * 1000, // 1 day
    sampleNum: 400,
  },
  [SinceFetch.W1]: {
    timeMs: 7 * 24 * 60 * 60 * 1000, // 1 week
    sampleNum: 700,
  },
  [SinceFetch.M1]: {
    timeMs: 30 * 24 * 60 * 60 * 1000, // 1 month
    sampleNum: 1000,
  },
  [SinceFetch.Y1]: {
    timeMs: 365 * 24 * 60 * 60 * 1000, // 1 year
    sampleNum: 1000,
  },
  [SinceFetch.DEFAULT]: {
    timeMs: 24 * 60 * 60 * 1000, // 1 day
    sampleNum: 400,
  },
};

export function useTotValueHistory() {
  const [valueHistory, setValueHistory] = useState<ValueHistoryPoint[]>([]);
  const [nonEmptyDataPointRate, setNonEmptyDataPointRate] = useState<number>(0);
  const [since, setSince] = useState<SinceFetch>(SinceFetch.D1);

  const network = useNodeStore((state) => state.currentNetwork);
  const pluginVersion = useNodeStore((state) => state.pluginVersion);
  const status = useNodeStore((state) => state.status);
  const { setError } = useError();

  const intervalRef = useRef<NodeJS.Timeout | null>(null);

  const getParamsFromSince = useCallback((since: SinceFetch) => {
    const now = Date.now();
    const sinceParams = SINCE_PARAMS_MAP[since];
    return {
      sinceParam: new Date(now - sinceParams.timeMs).toISOString(),
      sampleNum: sinceParams.sampleNum,
    };
  }, []);

  const updateValueHistory = () => {
    setValueHistory((prevValueHistory) => {
      // Get current staking addresses from the store to ensure we have the latest data
      const currentStakingAddresses =
        useStakingStore.getState().stakingAddresses;

      const newValueHistory = [
        ...prevValueHistory,
        {
          timestamp: localTimezoneNow(),
          value: getTotalValue(currentStakingAddresses),
        },
      ];

      // a new non empty value is added to the value history, update the percentage of non empty data points variable
      setNonEmptyDataPointRate((prevRate) => {
        const currentNonEmptyDataPointNum =
          (prevRate * prevValueHistory.length) / 100;
        const newRate =
          ((currentNonEmptyDataPointNum + 1) / newValueHistory.length) * 100;
        return newRate;
      });

      return newValueHistory;
    });
  };

  const setUpdateValueHistoryInterval = () => {
    const interval =
      SINCE_PARAMS_MAP[since].timeMs / SINCE_PARAMS_MAP[since].sampleNum;

    // Clear previous interval if any
    if (intervalRef.current) {
      clearInterval(intervalRef.current);
      intervalRef.current = null;
    }

    // update the value history with the same value every 'interval' ms.
    intervalRef.current = setInterval(() => {
      updateValueHistory();
    }, interval);
  };

  /* When Since variable changes, update the interval at which the value history is updated with a new item */
  useEffect(() => {
    const fetchData = async () => {
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

        setUpdateValueHistoryInterval();
      } catch (err) {
        setError({
          title: 'Error fetching value history',
          message: getErrorMessage(err),
        });
        return;
      }
    };

    if (status === NodeStatus.ON) {
      fetchData();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [since, getParamsFromSince, network, pluginVersion, setError]);

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

  return {
    valueHistory,
    fetchValueHistory,
    nonEmptyDataPointRate,
    since,
    setSince,
  };
}
