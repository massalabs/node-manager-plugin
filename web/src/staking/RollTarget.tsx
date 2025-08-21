import React, { useState, useMemo } from 'react';

import { Tooltip, Input, Button, Toggle } from '@massalabs/react-ui-kit';
import { FiInfo } from 'react-icons/fi';

import ConfirmModal from '@/components/ConfirmModal';
import { useStakingAddress } from '@/hooks/staking-manager/useStakingAddress';
import Intl from '@/i18n/i18n';
import { StakingAddress } from '@/models/staking';

interface RollTargetProps {
  currentAddress: StakingAddress;
  rollPrice: number;
}

const RollTarget: React.FC<RollTargetProps> = ({
  currentAddress,
  rollPrice,
}) => {
  const [targetRolls, setTargetRolls] = useState(currentAddress.target_rolls);
  const [isConfirmModalOpen, setIsConfirmModalOpen] = useState(false);
  const [targetRollChangeMsg, setTargetRollChangeMsg] = useState('');

  const { updateStakingAddress } = useStakingAddress();

  /*
   * Function that returns a message to display in the confirm modal. The message depend on current roll target and new one
   * If the roll target is set to maximum, return the a message indicating the number of rolls that can be bought with the current balance.
   * If the roll target is set to a value higher than the current target, return the a message indicating the number of rolls that are needed to be bought to reach the new target.
   * If the roll target is set to a value lower than the current target, return the a message indicating the number of rolls that are needed to be sold to reach the new target.
   */
  const getTargetChangeMessage = useMemo(() => {
    return () => {
      if (!currentAddress) {
        return '';
      }
      const currentTarget = currentAddress.target_rolls;
      const newTarget = targetRolls;
      const finalBalance = currentAddress.final_balance;

      // If the roll target is set to maximum
      if (newTarget === -1) {
        const rollsToBuy = Math.floor(finalBalance / rollPrice);
        return rollsToBuy === 0
          ? ''
          : ' ' +
              Intl.t(
                'staking.stakingAddressDetails.updateRollTarget.confirmModal.rollBuy',
                {
                  rollsToBuy: rollsToBuy.toString(),
                },
              );
      }

      // If the roll target is set to a value higher than the current target
      if (
        newTarget > currentAddress.candidate_roll_count &&
        newTarget > currentTarget &&
        Math.floor(finalBalance / rollPrice) > 0
      ) {
        const maxRollsToBuy = Math.min(
          Math.floor(finalBalance / rollPrice), // number of rolls that can be bought with current MAS balance
          newTarget - currentAddress.candidate_roll_count, // number of rolls that are needed to reach the new target
        );
        return (
          ' ' +
          Intl.t(
            'staking.stakingAddressDetails.updateRollTarget.confirmModal.rollBuy',
            {
              rollsToBuy: maxRollsToBuy.toString(),
            },
          )
        );
        // If the roll target is set to a value lower than the current target
      } else if (newTarget < currentAddress.candidate_roll_count) {
        return (
          ' ' +
          Intl.t(
            'staking.stakingAddressDetails.updateRollTarget.confirmModal.rollSell',
            {
              rollsToSell: (
                currentAddress.candidate_roll_count - newTarget
              ).toString(),
            },
          )
        );
      }
      return '';
    };
  }, [currentAddress, targetRolls, rollPrice]);

  const handleValidateClick = () => {
    const newTarget = targetRolls;
    if (newTarget !== currentAddress.target_rolls) {
      setIsConfirmModalOpen(true);
      setTargetRollChangeMsg(getTargetChangeMessage());
    }
  };

  const handleConfirmUpdate = () => {
    if (!currentAddress) {
      return;
    }
    const newTarget = targetRolls;
    updateStakingAddress.mutate({
      address: currentAddress.address,
      target_rolls: newTarget,
    });
    setIsConfirmModalOpen(false);
  };

  const handleCloseConfirmModal = () => {
    setIsConfirmModalOpen(false);
  };

  // If the maximum toggle is checked, set the target rolls to -1.
  // If the maximum toggle is unchecked, set the target rolls to the current candidate roll count.
  const handleMaximumToggle = (e: React.ChangeEvent<HTMLInputElement>) => {
    setTargetRolls(e.target.checked ? -1 : currentAddress.candidate_roll_count);
  };

  // If the input is changed, set the target rolls to the value of the input.
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = Number(e.target.value);
    setTargetRolls(value);
  };

  return (
    <>
      <div className="border-t border-gray-600 pt-4">
        <h3 className="text-sm font-medium text-gray-300 mb-1">
          Set Roll Target
        </h3>
        <p className="text-sm text-gray-400 mb-3">
          Set the expected number of rolls for this address. Node manager will
          automatically sell or buy (within the limit of available MAS funds: 1
          roll = {rollPrice || 0} MAS) rolls to match this target
        </p>

        {/* Maximum Toggle */}
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2">
            <label className="text-sm font-medium text-gray-300">
              {Intl.t('staking.stakingAddressDetails.updateRollTarget.maximum')}
            </label>
            <Tooltip
              body={Intl.t(
                'staking.stakingAddressDetails.updateRollTarget.maximumTooltip',
              )}
            >
              <FiInfo className="w-3 h-3 text-gray-400" />
            </Tooltip>
          </div>
          <Toggle checked={targetRolls === -1} onChange={handleMaximumToggle} />
        </div>

        {/* Target Rolls Input */}
        <div className="flex items-center justify-between mb-3">
          <label
            className={`text-sm font-medium text-gray-300 ${
              targetRolls === -1 ? 'opacity-50' : ''
            }`}
          >
            Set rolls target
          </label>
          <Input
            value={targetRolls === -1 ? '' : targetRolls}
            onChange={handleInputChange}
            placeholder="Enter target rolls"
            type="number"
            min="0"
            disable={targetRolls === -1}
            customClass="w-48 border border-gray-600"
          />
        </div>

        {/* Validate Button */}
        <div className="flex justify-start">
          <Button
            variant="primary"
            onClick={handleValidateClick}
            disabled={Number(targetRolls) === currentAddress.target_rolls}
            customClass={`px-3 py-1 text-sm ${
              Number(targetRolls) === currentAddress.target_rolls
                ? 'bg-gray-500 hover:bg-gray-500 opacity-75 cursor-not-allowed'
                : 'bg-green-500 hover:bg-green-600'
            }`}
          >
            Validate
          </Button>
        </div>
      </div>

      <ConfirmModal
        isOpen={isConfirmModalOpen}
        onClose={handleCloseConfirmModal}
        onConfirm={handleConfirmUpdate}
        title={Intl.t(
          'staking.stakingAddressDetails.updateRollTarget.confirmModal.title',
        )}
      >
        <div className="flex flex-col gap-4">
          <p className="mas-body text-f-primary">
            {Intl.t(
              'staking.stakingAddressDetails.updateRollTarget.confirmModal.body',
              {
                currentTargetRolls:
                  currentAddress.target_rolls === -1
                    ? 'MAX'
                    : currentAddress.target_rolls.toString(),
                newTargetRolls:
                  targetRolls === -1 ? 'MAX' : targetRolls.toString(),
              },
            )}
            {targetRollChangeMsg}
          </p>
        </div>
      </ConfirmModal>
    </>
  );
};

export default RollTarget;
