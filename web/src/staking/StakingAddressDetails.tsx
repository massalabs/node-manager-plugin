import React, { useEffect, useMemo } from 'react';

import {
  SidePanel,
  Balance,
  Tooltip,
  Clipboard,
  maskAddress,
} from '@massalabs/react-ui-kit';
import { FiX, FiInfo } from 'react-icons/fi';

import DeferredCreditList from './DeferredCreditList';
import RollsOpList from './RollsOpList';
import RollTarget from './RollTarget';
import { useError } from '@/contexts/ErrorContext';
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
  const { setError } = useError();

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

  // Helper function to format MAS with 2 decimal places
  const formatMas = useMemo(() => {
    return (masAmount: number): string => {
      return masAmount.toFixed(2);
    };
  }, []);

  if (!currentAddress) {
    setError({
      title: 'Staking address not found',
      message: 'Address ' + address + ' not found in staking addresses list',
    });
    return;
  }

  /* SidePanel component doesn't provide a way to handle the open/close state of the panel programmatically
   so we need to simulate a click on the toggle dropdown button to open and closethe panel */
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

            {/* Roll Target */}
            <RollTarget
              currentAddress={currentAddress}
              rollPrice={Number(nodeInfo?.config?.rollPrice || 0)}
            />

            {/* Roll Operation History */}
            <RollsOpList address={address} />
          </div>
        </SidePanel>
      )}
    </>
  );
};

export default StakingAddressDetails;
