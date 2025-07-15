import React, { useState } from 'react';

import { FiPlus } from 'react-icons/fi';

import AddStakingAddress from './addStakingAddress';
import StakingAddressDetails from './StakingAddressDetails';
import StakingAddressItem from './StakingAddressItem';
import { StakingAddress } from '@/models/staking';
import { useStakingStore } from '@/store/stakingStore';

const StakingAddressList: React.FC = () => {
  const stakingAddresses = useStakingStore((state) => state.stakingAddresses);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [selectedAddress, setSelectedAddress] = useState<StakingAddress | null>(
    null,
  );
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);

  const handleAddClick = () => {
    setIsAddModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsAddModalOpen(false);
  };

  const handleOpenDetails = (address: StakingAddress) => {
    setSelectedAddress(address);
    setIsDetailsOpen(true);
  };

  const handleCloseDetails = () => {
    setIsDetailsOpen(false);
    setSelectedAddress(null);
  };

  return (
    <>
      <div className="bg-secondary rounded-lg shadow p-6 h-full w-4/5 mx-auto">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-semibold text-f-primary">
            Staking Address
          </h3>
          <button
            onClick={handleAddClick}
            className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg flex 
            items-center gap-2 transition-colors"
          >
            <FiPlus className="w-4 h-4" />
            Add
          </button>
        </div>

        <div className="overflow-x-auto w-full">
          <table className="w-full divide-y divide-gray-600">
            <thead className="bg-gray-700">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider w-1/10">
                  Address
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider w-1/6">
                  Balance
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider w-1/6">
                  Active Rolls
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider w-1/6">
                  Status
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-300 uppercase tracking-wider w-auto">
                  Actions
                </th>
              </tr>
            </thead>
            <tbody className="bg-secondary divide-y divide-gray-600">
              {stakingAddresses.map((address) => (
                <StakingAddressItem
                  key={address.address}
                  address={address}
                  onOpenDetails={handleOpenDetails}
                />
              ))}
            </tbody>
          </table>
        </div>
      </div>

      <AddStakingAddress isOpen={isAddModalOpen} onClose={handleCloseModal} />

      {selectedAddress && (
        <StakingAddressDetails
          isOpen={isDetailsOpen}
          onClose={handleCloseDetails}
          address={selectedAddress.address}
        />
      )}
    </>
  );
};

export default StakingAddressList;
