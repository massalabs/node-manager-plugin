import React, { useState } from 'react';

import { FiPlus } from 'react-icons/fi';
import { FiLock } from 'react-icons/fi';
import { GrMoney } from 'react-icons/gr';

import AddStakingAddress from './addStakingAddress';
import StakingAddressDetails from './StakingAddressDetails';
import StakingAddressItem from './StakingAddressItem';
import Intl from '@/i18n/i18n';
import { StakingAddress } from '@/models/staking';
import { useNodeStore } from '@/store/nodeStore';
import { useStakingStore } from '@/store/stakingStore';
import { NodeStatus } from '@/utils';

// Reusable Add Button Component
const AddButton: React.FC<{ onClick: () => void }> = ({ onClick }) => (
  <button
    onClick={onClick}
    className="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded-lg flex 
    items-center gap-2 transition-colors"
  >
    <FiPlus className="w-4 h-4" />
    Add
  </button>
);

const StakingAddressList: React.FC = () => {
  const stakingAddresses = useStakingStore((state) => state.stakingAddresses);
  const status = useNodeStore((state) => state.status);
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [selectedAddress, setSelectedAddress] = useState<StakingAddress | null>(
    null,
  );
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);

  const hasAddresses = stakingAddresses.length > 0;

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

  // When node is not running, show only background, title and lock icon
  if (status !== NodeStatus.ON) {
    return (
      <div className="bg-secondary rounded-2xl p-8 h-full w-full relative border border-gray-700">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-semibold text-white">Staking Address</h3>
        </div>
        <div className="flex flex-col items-center justify-center h-full gap-4">
          <FiLock className="text-6xl text-gray-400" />
          <p className="text-gray-400 text-lg">
            {Intl.t('staking.node-not-running')}
          </p>
        </div>
      </div>
    );
  }

  return (
    <>
      <div className="bg-secondary rounded-2xl p-8 h-full w-full relative border border-gray-700">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-lg font-semibold text-white">Staking Address</h3>
          {/* Only show add button in header if there are addresses */}
          {hasAddresses && <AddButton onClick={handleAddClick} />}
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
              {hasAddresses ? (
                stakingAddresses.map((address) => (
                  <StakingAddressItem
                    key={address.address}
                    address={address}
                    onOpenDetails={handleOpenDetails}
                    isSelected={selectedAddress?.address === address.address}
                  />
                ))
              ) : (
                <tr>
                  <td colSpan={5} className="px-6 py-12 text-center">
                    <div className="flex flex-col items-center gap-4">
                      <GrMoney className="text-6xl text-gray-400" />
                      <p className="text-gray-400 text-lg">
                        {Intl.t('staking.empty-state.message')}
                      </p>
                      <AddButton onClick={handleAddClick} />
                    </div>
                  </td>
                </tr>
              )}
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
