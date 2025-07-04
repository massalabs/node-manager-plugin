import React, { useState, useEffect } from 'react';
import { FiInfo } from 'react-icons/fi';
import {
  SidePanel,
  Balance,
  Tooltip,
  Input,
  Button,
  AccordionCategory,
} from '@massalabs/react-ui-kit';

import { StakingAddress } from '@/models/staking';
import { useStakingAddress } from '@/hooks/useStakingAddress';
import { useFetchNodeInfo } from '@/hooks/useFetchNodeInfo';
import ConfirmModal from '@/components/ConfirmModal';
import Intl from '@/i18n/i18n';

interface StakingAddressDetailsProps {
  isOpen: boolean;
  onClose: () => void;
  address: StakingAddress;
}

const StakingAddressDetails: React.FC<StakingAddressDetailsProps> = ({
  isOpen,
  onClose,
  address,
}) => {
  const [targetRolls, setTargetRolls] = useState(address.targetRolls);
  const [isConfirmModalOpen, setIsConfirmModalOpen] = useState(false);
  const [newTargetRolls, setNewTargetRolls] = useState(0);
  
  const { updateStakingAddress } = useStakingAddress();
  const { data: nodeInfo } = useFetchNodeInfo();

  // SidePanel component doesn't provide a way to handle the open/close state of the panel programmatically
  // so we need to simulate a click on the toggle dropdown button to open and closethe panel
  const clickSidePanelButton = () => {
    const sidePanel = document.querySelector('[data-panel-type="staking-address-details"]');
    if (sidePanel) {
      const button = sidePanel.querySelector('button') as HTMLButtonElement;
      if (button) {
        button.click();
      }
    }
  }

  // Effect to trigger SidePanel dropdown when isOpen becomes true
  useEffect(() => {
    if (isOpen) {
      clickSidePanelButton();
    }
  }, [isOpen]);

  const pannelClose = () => {
    clickSidePanelButton();
    onClose();
  }

  
  const handleValidateClick = () => {
    const newTarget = targetRolls;
    if (newTarget !== address.targetRolls) {
      setNewTargetRolls(targetRolls);
      setIsConfirmModalOpen(true);
    }
  };

  const handleConfirmUpdate = () => {
    const newTarget = newTargetRolls;
    updateStakingAddress.mutate({
      address: address.address,
      targetRolls: newTarget,
    });
    setIsConfirmModalOpen(false);
  };

  const handleCloseConfirmModal = () => {
    setIsConfirmModalOpen(false);
  };

  const getTargetChangeMessage = () => {
    const currentTarget = address.targetRolls;
    const newTarget = newTargetRolls;
    const rollPrice = Number(nodeInfo?.config?.rollPrice) || 100;
    const finalBalance = address.finalBalance;

    if (newTarget > currentTarget && Math.floor(finalBalance / rollPrice) > 0) {
      const maxRollsToBuy = Math.min(
        Math.floor(finalBalance / rollPrice), // number of rolls that can be bought with current MAS balance
        newTarget - currentTarget // number of rolls that are needed to reach the new target
      );
      return Intl.t('staking.updateRollTarget.confirmModal.rollBuy', { rollsToBuy: maxRollsToBuy.toString() });
    } else if (newTarget < currentTarget) {
      return Intl.t('staking.updateRollTarget.confirmModal.rollSell', { rollsToSell: (currentTarget - newTarget).toString() });
    }
    return '';
  };

  const getDeferredCreditsTable = () => {
    if (!address.deferredCredits || address.deferredCredits.length === 0) {
      return <p className="text-gray-400">No deferred credits</p>;
    }

    return (
      <table className="min-w-full divide-y divide-gray-600">
        <thead className="bg-gray-700">
          <tr>
            <th className="px-4 py-2 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
              Amount
            </th>
            <th className="px-4 py-2 text-left text-xs font-medium text-gray-300 uppercase tracking-wider">
              Approx Release Date
            </th>
          </tr>
        </thead>
        <tbody className="bg-secondary divide-y divide-gray-600">
          {address.deferredCredits.map((credit, index) => {
            const periodDiff = credit.slot.period - (nodeInfo?.executionStats?.activeCursor?.period || 0);
            const periodLength = nodeInfo?.config?.t0 || 0;
            const releaseTime = periodDiff * periodLength;
            const releaseDate = new Date(Date.now() + releaseTime * 1000);

            return (
              <tr key={index} className="border-b border-gray-600">
                <td className="px-4 py-2 text-sm">
                  <Balance amount={credit.amount.toString()} />
                </td>
                <td className="px-4 py-2 text-sm text-f-primary">
                  {releaseDate.toLocaleString()}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    );
  };

  return (
    <>
     {isOpen && (<SidePanel
        customClass="w-96"
        data-panel-type="staking-address-details"
        onClose={pannelClose}
      >
        <div className="flex flex-col gap-6 p-6">
          {/* Address */}
          <div>
            <h3 className="text-sm font-medium text-gray-300 mb-2">Address</h3>
            <p className="text-f-primary font-mono text-sm break-all">{address.address}</p>
          </div>

          {/* Thread */}
          <div>
            <h3 className="text-sm font-medium text-gray-300 mb-2">Thread</h3>
            <p className="text-f-primary">{address.thread}</p>
          </div>

          {/* Balances */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <h3 className="text-sm font-medium text-gray-300 mb-2">Candidate Balance</h3>
              <Balance amount={address.candidateBalance.toString()} />
            </div>
            <div>
              <h3 className="text-sm font-medium text-gray-300 mb-2">Final Balance</h3>
              <Balance amount={address.finalBalance.toString()} />
            </div>
          </div>

          {/* Rolls */}
          <div className="grid grid-cols-3 gap-4">
            <div>
              <h3 className="text-sm font-medium text-gray-300 mb-2 flex items-center gap-1">
                Active Rolls
                <Tooltip body="It takes 3 cycles (about 1h40min) for new rolls to become active and be used for staking">
                  <FiInfo className="w-3 h-3 text-gray-400" />
                </Tooltip>
              </h3>
              <p className="text-f-primary">{address.activeRolls}</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-gray-300 mb-2">Final Rolls</h3>
              <p className="text-f-primary">{address.finalRolls}</p>
            </div>
            <div>
              <h3 className="text-sm font-medium text-gray-300 mb-2">Candidate Rolls</h3>
              <p className="text-f-primary">{address.candidateRolls}</p>
            </div>
          </div>

          {/* Deferred Credits */}
          <div>
            <AccordionCategory
              categoryTitle={
                <div className="flex items-center justify-between w-full">
                  <span className="flex items-center gap-1">
                    Deferred Credits
                    <Tooltip body="When rolls are sold, staked MAS are frozen for a cycle before they can be spent">
                      <FiInfo className="w-3 h-3 text-gray-400" />
                    </Tooltip>
                  </span>
                  <span className="text-sm text-gray-400">
                    {address.deferredCredits?.length || 0}
                  </span>
                </div>
              }
            >
              <div className="mt-4">
                {getDeferredCreditsTable()}
              </div>
            </AccordionCategory>
          </div>

          {/* Set Roll Target */}
          <div className="border-t border-gray-600 pt-6">
            <h3 className="text-sm font-medium text-gray-300 mb-2">Set Roll Target</h3>
            <p className="text-sm text-gray-400 mb-4">
              Set the expected number of rolls for this address. Node manager will automatically sell or buy (within the limit of available MAS funds) rolls to match this target
            </p>
            
            <div className="flex items-center gap-2 mb-4">
              <Tooltip body={`1 roll = ${nodeInfo?.config?.rollPrice || 0} MAS`}>
                <FiInfo className="w-4 h-4 text-gray-400" />
              </Tooltip>
              <Input
                value={targetRolls}
                onChange={(e) => setTargetRolls(Number(e.target.value))}
                placeholder="Enter target rolls"
                type="number"
                min="0"
              />
            </div>

            <Button
              variant="primary"
              onClick={handleValidateClick}
              disabled={Number(targetRolls) === address.targetRolls}
            >
              Validate
            </Button>
            </div>
          </div>
        </SidePanel>
      )}

      <ConfirmModal
        isOpen={isConfirmModalOpen}
        onClose={handleCloseConfirmModal}
        onConfirm={handleConfirmUpdate}
        title={Intl.t('staking.updateRollTarget.confirmModal.title')}
      >
        <div className="flex flex-col gap-4">
          <p className="mas-body text-f-primary">
            {Intl.t('staking.updateRollTarget.confirmModal.body', { addressTargetRolls: address.targetRolls.toString(), newTargetRolls: newTargetRolls.toString() })}
            {getTargetChangeMessage()}
          </p>
        </div>
      </ConfirmModal>
    </>
  );
};

export default StakingAddressDetails; 