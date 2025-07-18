import React, { useState } from 'react';

import { Clipboard, maskAddress, Tag } from '@massalabs/react-ui-kit';
import { FiTrash2, FiEdit3 } from 'react-icons/fi';

import ConfirmModal from '@/components/ConfirmModal';
import { useStakingAddress } from '@/hooks/useStakingAddress';
import Intl from '@/i18n/i18n';
import { StakingAddress } from '@/models/staking';

interface StakingAddressItemProps {
  address: StakingAddress;
  onOpenDetails: (address: StakingAddress) => void;
  isSelected: boolean;
}

const StakingAddressItem: React.FC<StakingAddressItemProps> = ({
  address,
  onOpenDetails,
  isSelected,
}) => {
  const [isDeleteModalOpen, setIsDeleteModalOpen] = useState(false);
  const { removeStakingAddress } = useStakingAddress();

  const handleDeleteClick = () => {
    setIsDeleteModalOpen(true);
  };

  const handleDeleteModalClose = () => {
    setIsDeleteModalOpen(false);
  };

  const handleConfirmDelete = () => {
    removeStakingAddress.mutate({
      address: address.address,
    });
    handleDeleteModalClose();
  };

  const handleDetailsClick = () => {
    onOpenDetails(address);
  };

  const getStakingStatusBadge = () => {
    const isStaking = address.active_roll_count > 0;

    return (
      <Tag type={isStaking ? 'success' : 'warning'} customClass="text-xs">
        {isStaking
          ? Intl.t('staking.status.staking')
          : Intl.t('staking.status.not-staking')}
      </Tag>
    );
  };

  return (
    <>
      <tr
        className={`border-b border-gray-600 hover:bg-gray-700 ${
          isSelected ? 'bg-gray-700' : ''
        }`}
      >
        <td className="px-6 py-4 whitespace-nowrap text-sm w-1/10">
          <Clipboard
            rawContent={address.address}
            displayedContent={maskAddress(address.address)}
            customClass="max-w-[200px]"
          />
        </td>
        <td className="px-6 py-4 whitespace-nowrap text-sm text-f-primary">
          {address.final_balance.toFixed(2)} MAS
        </td>
        <td className="px-6 py-4 whitespace-nowrap text-sm text-f-primary">
          {address.active_roll_count}
        </td>
        <td className="px-6 py-4 whitespace-nowrap text-sm">
          {getStakingStatusBadge()}
        </td>
        <td className="px-6 py-4 whitespace-nowrap text-sm">
          <div className="flex items-center justify-between">
            <button
              onClick={handleDetailsClick}
              className={
                'bg-white hover:bg-gray-100 text-gray-700 border border-gray-300 rounded-lg ' +
                'px-3 py-2 transition-colors flex items-center gap-2'
              }
              title="View details"
            >
              <FiEdit3 className="w-4 h-4" />
              <span className="text-sm font-medium">Details</span>
            </button>
            <button
              onClick={handleDeleteClick}
              className={
                'bg-red-500 hover:bg-red-600 text-white border border-red-600 rounded-lg ' +
                'px-3 py-2 transition-colors flex items-center gap-2'
              }
              title="Delete staking address"
            >
              <FiTrash2 className="w-4 h-4" />
              <span className="text-sm font-medium">Delete</span>
            </button>
          </div>
        </td>
      </tr>

      <ConfirmModal
        isOpen={isDeleteModalOpen}
        onClose={handleDeleteModalClose}
        onConfirm={handleConfirmDelete}
        title={Intl.t('staking.delete-address.title')}
      >
        <div className="flex flex-col gap-4">
          <p className="mas-body text-f-primary">
            {Intl.t('staking.delete-address.message', {
              address: maskAddress(address.address),
            })}
          </p>
        </div>
      </ConfirmModal>
    </>
  );
};

export default StakingAddressItem;
