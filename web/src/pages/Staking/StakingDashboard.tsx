import React from 'react';
import HistoryGraph from './HistoryGraph';
import StakingAddressList from './StakingAddressList';

const StakingDashboard: React.FC = () => {
  return (
    <div className="flex flex-col gap-6 h-full p-6">
      <div className="h-3/5">
        <HistoryGraph />
      </div>
      <div className="h-2/5">
        <StakingAddressList />
      </div>
    </div>
  );
};

export default StakingDashboard; 