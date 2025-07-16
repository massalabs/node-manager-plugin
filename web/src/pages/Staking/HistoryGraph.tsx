import React, { useEffect, useState } from 'react';

import { FiBarChart2 } from 'react-icons/fi';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  Tooltip,
  ResponsiveContainer,
  CartesianGrid,
} from 'recharts';

import { useTotValueHistory } from '../../hooks/useTotValueHistory';
import { SinceFetch } from '@/models/history';
import { useNodeStore } from '@/store/nodeStore';

const SINCE_OPTIONS = [
  SinceFetch.H1,
  SinceFetch.D1,
  SinceFetch.W1,
  SinceFetch.M1,
  SinceFetch.Y1,
];

const HistoryGraph: React.FC = () => {
  const [selectedSince, setSelectedSince] = useState<SinceFetch>(SinceFetch.D1);
  const { valueHistory, fetchValueHistory } = useTotValueHistory();
  const nodeStatus = useNodeStore((state) => state.status);

  // when the node is up, fetch the value history for 1 month
  useEffect(() => {
    if (nodeStatus === 'on') {
      fetchValueHistory(SinceFetch.M1);
      setSelectedSince(SinceFetch.M1);
    }
  }, [nodeStatus, fetchValueHistory]);

  const handleSinceClick = (since: SinceFetch) => {
    setSelectedSince(since);
    fetchValueHistory(since);
  };

  const tickFormatter = (str: string) => {
    return selectedSince === SinceFetch.D1 || selectedSince === SinceFetch.H1
      ? str.slice(11, 16)
      : str.slice(5, 16);
  };

  return (
    <div className="bg-secondary rounded-lg shadow p-6 h-full w-4/5 mx-auto relative">
      <div className="flex justify-between items-center mb-4">
        <h3 className="text-lg font-semibold text-f-primary">$MAS history</h3>
        <div className="flex gap-2">
          {SINCE_OPTIONS.map((since) => (
            <button
              key={since}
              className={`w-8 h-8 rounded border flex items-center justify-center text-xs font-bold transition-colors ${
                selectedSince === since
                  ? 'bg-primary text-white'
                  : 'bg-gray-200 text-gray-700'
              }`}
              onClick={() => handleSinceClick(since)}
            >
              {since}
            </button>
          ))}
        </div>
      </div>
      {valueHistory.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-full">
          <FiBarChart2 className="text-6xl text-gray-400 mb-4" />
          <p className="text-gray-400 text-sm">Not enough data</p>
        </div>
      ) : (
        <ResponsiveContainer width="100%" height={300}>
          <AreaChart
            data={valueHistory}
            margin={{ top: 10, right: 30, left: 0, bottom: 0 }}
          >
            <defs>
              <linearGradient id="colorValue" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#8884d8" stopOpacity={0.8} />
                <stop offset="95%" stopColor="#8884d8" stopOpacity={0} />
              </linearGradient>
            </defs>
            <XAxis
              dataKey="timestamp"
              tickFormatter={tickFormatter}
              minTickGap={20}
            />
            <YAxis dataKey="value" domain={['dataMin - 10', 'dataMax + 10']} />
            <CartesianGrid strokeDasharray="3 3" />
            <Tooltip />
            <Area
              type="monotone"
              dataKey="value"
              stroke="#8884d8"
              fillOpacity={1}
              fill="url(#colorValue)"
            />
          </AreaChart>
        </ResponsiveContainer>
      )}
    </div>
  );
};

export default HistoryGraph;
