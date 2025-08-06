import React, { useState, useEffect, useMemo } from 'react';

import {
  SidePanel,
  Balance,
  Tooltip,
  Input,
  Button,
  Clipboard,
  maskAddress,
} from '@massalabs/react-ui-kit';
import { FiInfo, FiX } from 'react-icons/fi';

import DeferredCreditList from './DeferredCreditList';
import RollsOpList from './RollsOpList';
import ConfirmModal from '@/components/ConfirmModal';
import { useError } from '@/contexts/ErrorContext';
import { useStakingAddress } from '@/hooks/staking-manager/useStakingAddress';
import { useFetchNodeInfo } from '@/hooks/useFetchNodeInfo';
import Intl from '@/i18n/i18n';
import { useStakingStore } from '@/store/stakingStore';

interface StakingAddressDetailsProps {
  isOpen: boolean;
  onClose: () => void;
  address: string;
}

const StakingAddressDetails: React.FC<StakingAddressDetailsProps> = ({
  isOpen,
  onClose,
  address,
}) => {
  const [targetRolls, setTargetRolls] = useState(0);
  const [isConfirmModalOpen, setIsConfirmModalOpen] = useState(false);
  const [targetRollChangeMsg, setTargetRollChangeMsg] = useState('');

  const { setError } = useError();

  const { updateStakingAddress } = useStakingAddress();
  const { data: nodeInfo, isError: isNodeInfoError } = useFetchNodeInfo(
    1000 * 60 * 3,
  ); // fetch node status every 3 min
  const currentAddress = useStakingStore((state) =>
    state.stakingAddresses.find((addr) => addr.address === address),
  );

  // Effect to trigger SidePanel dropdown when isOpen becomes true
  useEffect(() => {
    if (isOpen) {
      clickSidePanelButton();
    }
  }, [isOpen]);

  // handle error when fetching node info
  useEffect(() => {
    if (isNodeInfoError) {
      setError({
        title: 'Error fetching node info in staking address details',
        message: 'Please try again later',
      });
    }
  }, [isNodeInfoError, setError]);

  // Effect to update the current address and target rolls when the staking addresses list changes
  useEffect(() => {
    if (!currentAddress) {
      return;
    }

    setTargetRolls(currentAddress.target_rolls ?? 0);
  }, [currentAddress]);

  // Helper function to format MAS with 2 decimal places
  const formatMas = useMemo(() => {
    return (masAmount: number): string => {
      return masAmount.toFixed(2);
    };
  }, []);

  const getTargetChangeMessage = useMemo(() => {
    return () => {
      if (!currentAddress) {
        return '';
      }
      const currentTarget = currentAddress.target_rolls;
      const newTarget = targetRolls;
      const rollPrice = Number(nodeInfo?.config?.rollPrice) || 100;
      const finalBalance = currentAddress.final_balance;

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
  }, [currentAddress, targetRolls, nodeInfo?.config?.rollPrice]);

  if (!currentAddress) {
    setError({
      title: 'Staking address not found',
      message: 'Address ' + address + ' not found in staking addresses list',
    });
    return;
  }

  // SidePanel component doesn't provide a way to handle the open/close state of the panel programmatically
  // so we need to simulate a click on the toggle dropdown button to open and closethe panel
  const clickSidePanelButton = () => {
    const sidePanel = document.querySelector(
      '[data-panel-type="staking-address-details"]',
    );
    if (sidePanel) {
      const button = sidePanel.querySelector('button') as HTMLButtonElement;
      if (button) {
        button.click();
      }
    }
  };

  const pannelClose = () => {
    clickSidePanelButton();
    onClose();
  };

  const handleValidateClick = () => {
    const newTarget = targetRolls;
    if (newTarget !== currentAddress?.target_rolls) {
      // setNewTargetRolls(targetRolls);
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
      address: currentAddress?.address,
      target_rolls: newTarget,
    });
    setIsConfirmModalOpen(false);
  };

  const handleCloseConfirmModal = () => {
    setIsConfirmModalOpen(false);
  };

  return (
    <>
      {isOpen && (
        <SidePanel
          customClass="!w-[550px]"
          data-panel-type="staking-address-details"
          onClose={pannelClose}
        >
          <div className="flex flex-col gap-6 p-10 relative">
            {/* Close Button */}
            <button
              onClick={pannelClose}
              className="absolute top-2 right-2 p-1 text-gray-400 hover:text-white transition-colors z-10"
              title="Close"
            >
              <FiX className="w-5 h-5" />
            </button>

            {/* Address and Thread */}
            <div className="flex items-start gap-4">
              {/* Address */}
              <div className="w-1/2">
                <div className="flex items-center gap-2 h-8">
                  <h3 className="text-sm font-medium text-gray-300 whitespace-nowrap">
                    Address:
                  </h3>
                  <div className="flex-1 min-w-0">
                    <Clipboard
                      rawContent={currentAddress?.address ?? ''}
                      displayedContent={maskAddress(
                        currentAddress?.address ?? '',
                      )}
                    />
                  </div>
                </div>
              </div>

              {/* Thread */}
              <div className="w-1/2">
                <div className="flex items-center justify-end gap-2 h-8">
                  <h3 className="text-sm font-medium text-gray-300 whitespace-nowrap">
                    Thread:
                  </h3>
                  <p className="text-f-primary text-sm">
                    {currentAddress?.thread}
                  </p>
                </div>
              </div>
            </div>

            <hr className="h-1 border-t-0 bg-gradient-to-r from-transparent via-gray-500 to-transparent opacity-60" />

            {/* Balances */}
            <div className="grid grid-cols-2 gap-3">
              <div>
                <h3 className="text-sm font-medium text-gray-300 mb-1">
                  Final Balance
                </h3>
                <Balance
                  size="xs"
                  amount={formatMas(currentAddress?.final_balance ?? 0)}
                />
              </div>
              <div>
                <h3 className="text-sm font-medium text-gray-300 mb-1">
                  Candidate Balance
                </h3>
                <Balance
                  size="xs"
                  amount={formatMas(currentAddress?.candidate_balance ?? 0)}
                />
              </div>
            </div>

            <hr className="h-1 border-t-0 bg-gradient-to-r from-transparent via-gray-500 to-transparent opacity-60" />

            {/* Rolls */}
            <div className="grid grid-cols-3 gap-3">
              <div>
                <h3 className="text-sm font-medium text-gray-300 mb-1 flex items-center gap-1">
                  Active Rolls
                  <Tooltip
                    body={Intl.t(
                      'staking.stakingAddressDetails.active-rolls-tooltip',
                    )}
                  >
                    <FiInfo className="w-3 h-3 text-gray-400" />
                  </Tooltip>
                </h3>
                <p className="text-f-primary">
                  {currentAddress?.active_roll_count ?? 0}
                </p>
              </div>
              <div>
                <h3 className="text-sm font-medium text-gray-300 mb-1">
                  Final Rolls
                </h3>
                <p className="text-f-primary">
                  {currentAddress?.final_roll_count ?? 0}
                </p>
              </div>
              <div>
                <h3 className="text-sm font-medium text-gray-300 mb-1">
                  Candidate Rolls
                </h3>
                <p className="text-f-primary">
                  {currentAddress?.candidate_roll_count ?? 0}
                </p>
              </div>
            </div>

            {/* Deferred Credits */}
            <DeferredCreditList
              currentAddress={currentAddress}
              lastSlot={nodeInfo?.lastSlot}
              t0={nodeInfo?.config?.t0}
            />

            {/* Set Roll Target */}
            <div className="border-t border-gray-600 pt-4">
              <h3 className="text-sm font-medium text-gray-300 mb-1">
                Set Roll Target
              </h3>
              <p className="text-sm text-gray-400 mb-3">
                Set the expected number of rolls for this address. Node manager
                will automatically sell or buy (within the limit of available
                MAS funds) rolls to match this target
              </p>

              <div className="flex items-center gap-2">
                <Tooltip
                  body={`1 roll = ${nodeInfo?.config?.rollPrice || 0} MAS`}
                >
                  <FiInfo className="w-4 h-4 text-gray-400" />
                </Tooltip>
                <Input
                  value={targetRolls}
                  onChange={(e) => setTargetRolls(Number(e.target.value))}
                  placeholder="Enter target rolls"
                  type="number"
                  min="0"
                />
                <Button
                  variant="primary"
                  onClick={handleValidateClick}
                  disabled={
                    Number(targetRolls) === currentAddress?.target_rolls
                  }
                  customClass={`px-3 py-1 text-sm ${
                    Number(targetRolls) === currentAddress?.target_rolls
                      ? 'bg-gray-500 hover:bg-gray-500 opacity-75 cursor-not-allowed'
                      : 'bg-green-500 hover:bg-green-600'
                  }`}
                >
                  Validate
                </Button>
              </div>
            </div>

            {/* Roll Operation History */}
            <RollsOpList address={address} />
          </div>
        </SidePanel>
      )}

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
                  currentAddress?.target_rolls?.toString() ?? '0',
                newTargetRolls: targetRolls?.toString() ?? '0',
              },
            )}
            {targetRollChangeMsg}
          </p>
        </div>
      </ConfirmModal>
    </>
  );
};

export default StakingAddressDetails;
