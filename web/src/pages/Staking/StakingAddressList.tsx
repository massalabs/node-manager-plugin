import React, { useState } from 'react';
import { FiPlus } from 'react-icons/fi';
import { useStakingStore } from '@/store/stakingStore';
import StakingAddressItem from './StakingAddressItem';
import AddStakingAddress from './addStakingAddress';

const StakingAddressList: React.FC = () => {
  const stakingAddresses = useStakingStore((state) => state.stakingAddresses);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);

  const handleAddClick = () => {
    setIsAddModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsAddModalOpen(false);
  };

  return (
    <>
      <div className="bg-secondary rounded-lg shadow p-6 h-full">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-semibold text-f-primary">
            Staking Address
          </h3>
          <button 
            onClick={handleAddClick}
            className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg flex items-center gap-2 transition-colors"
          >
            <FiPlus className="w-4 h-4" />
            Add
          </button>
        </div>
        
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-600">
            <thead className="bg-gray-700">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
                  Address
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
                  Final Rolls
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
                  Candidate Rolls
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
                  Final Balance
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
                  Candidate Balance
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
                  Thread
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
                  Target Rolls
                </th>
              </tr>
            </thead>
            <tbody className="bg-secondary divide-y divide-gray-600">
              {stakingAddresses.map((address) => (
                <StakingAddressItem key={address.address} address={address} />
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <AddStakingAddress 
        isOpen={isAddModalOpen} 
        onClose={handleCloseModal} 
      />
    </>
  );
};

export default StakingAddressList; 