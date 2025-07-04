import React from 'react';

import { FiBarChart2 } from 'react-icons/fi';

const HistoryGraph: React.FC = () => {
  return (
    <div className="bg-secondary rounded-lg shadow p-6 h-full">
      <h3 className="text-lg font-semibold text-f-primary mb-4">
        $MAS history
      </h3>
      <div className="flex flex-col items-center justify-center h-full">
        <FiBarChart2 className="text-6xl text-gray-400 mb-4" />
        <p className="text-gray-400 text-sm">Not enough data</p>
      </div>
    </div>
  );
};

export default HistoryGraph;
