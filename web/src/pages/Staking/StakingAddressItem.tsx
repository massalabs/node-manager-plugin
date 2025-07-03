import React from 'react';
import { StakingAddress } from '@/models/staking';

interface StakingAddressItemProps {
  address: StakingAddress;
}

const StakingAddressItem: React.FC<StakingAddressItemProps> = ({ address }) => {
  return (
    <tr className="border-b border-gray-600 hover:bg-gray-700">
      <td className="px-6 py-4 whitespace-nowrap text-sm text-f-primary">
        {address.address}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-f-primary">
        {address.finalRolls}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-f-primary">
        {address.candidateRolls}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-f-primary">
        {address.finalBalance.toFixed(2)} MAS
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-f-primary">
        {address.candidateBalance.toFixed(2)} MAS
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-f-primary">
        {address.thread}
      </td>
      <td className="px-6 py-4 whitespace-nowrap text-sm text-f-primary">
        {address.targetRolls}
      </td>
    </tr>
  );
};

export default StakingAddressItem; 