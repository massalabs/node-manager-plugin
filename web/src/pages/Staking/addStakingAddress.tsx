import React, { useCallback, useEffect, useState } from 'react';

import {
  Clipboard,
  PopupModal,
  PopupModalHeader,
  PopupModalContent,
  Password,
  maskAddress,
} from '@massalabs/react-ui-kit';
import { useNavigate } from 'react-router-dom';

import { getStationAccounts } from './utils/station';
import ConfirmModal from '@/components/ConfirmModal';
import { useStakingAddress } from '@/hooks/useStakingAddress';
import Intl from '@/i18n/i18n';
import { useStakingStore } from '@/store/stakingStore';
import { goToErrorPage } from '@/utils';

type Account = {
  address: string;
  nickname: string;
  status: 'ok' | 'corrupted';
};

interface AddStakingAddressProps {
  isOpen: boolean;
  onClose: () => void;
}

const AddStakingAddress: React.FC<AddStakingAddressProps> = ({
  isOpen,
  onClose,
}) => {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [isConfirmModalOpen, setIsConfirmModalOpen] = useState(false);
  const [password, setPassword] = useState('');

  const { addStakingAddress } = useStakingAddress();
  const stakingAddresses = useStakingStore((state) => state.stakingAddresses);
  const navigate = useNavigate();

  const loadAccounts = useCallback(async () => {
    setLoading(true);
    try {
      const stationAccounts = await getStationAccounts();
      // Filter out accounts that are already in the stakingStore
      const existingAddresses = stakingAddresses.map((addr) => addr.address);
      const availableAccounts = stationAccounts.filter(
        (account) => !existingAddresses.includes(account.address),
      );
      setAccounts(availableAccounts);
    } catch (error) {
      console.error('Error loading accounts:', error);
      goToErrorPage(
        navigate,
        Intl.t('errors.load-massa-wallet-accounts.title'),
        Intl.t('errors.load-massa-wallet-accounts.description', {
          error: error instanceof Error ? error.message : String(error),
        }),
      );
    } finally {
      setLoading(false);
    }
  }, [navigate, stakingAddresses]);

  useEffect(() => {
    if (isOpen) {
      loadAccounts();
    }
  }, [isOpen, loadAccounts]);

  const handleAccountClick = (account: Account) => {
    setSelectedAccount(account);
    setIsConfirmModalOpen(true);
  };

  const handleConfirmModalClose = () => {
    setIsConfirmModalOpen(false);
    setSelectedAccount(null);
    setPassword('');
  };

  const handleConfirmAdd = () => {
    addStakingAddress.mutate({
      nickname: selectedAccount?.nickname || '',
      password: password,
    });
    handleConfirmModalClose();
    onClose();
  };

  return (
    <>
      {isOpen && (
        <PopupModal
          fullMode={true}
          customClass="border-2 border-black bg-secondary w-full md:w-3/4 lg:w-2/3"
          onClose={onClose}
        >
          <PopupModalHeader customClassHeader="bg-gray-700">
            <p className="mas-title mb-6">Add Staking Address</p>
          </PopupModalHeader>

          <PopupModalContent customClassContent="bg-secondary pb-5 pt-5">
            {loading ? (
              <div className="flex items-center justify-center h-32">
                <div className="text-f-primary">Loading accounts...</div>
              </div>
            ) : accounts.length === 0 ? (
              <div className="flex items-center justify-center h-32">
                <div className="text-gray-400">No available accounts found</div>
              </div>
            ) : (
              <div className="overflow-y-auto max-h-96">
                <table className="w-full">
                  <thead className="bg-gray-700">
                    <tr>
                      <th
                        className="w-4/5 px-4 py-3 text-left text-xs font-medium text-gray-300 
                      uppercase tracking-wider"
                      >
                        Nickname
                      </th>
                      <th
                        className="w-1/5 px-4 py-3 text-left text-xs font-medium text-gray-300 uppercase
                      tracking-wider"
                      >
                        Address
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-secondary divide-y divide-gray-600">
                    {accounts.map((account) => (
                      <tr
                        key={account.address}
                        className="hover:bg-gray-700 cursor-pointer"
                      >
                        <td
                          className="w-4/5 px-4 py-4 whitespace-nowrap text-sm text-f-primary"
                          onClick={() => handleAccountClick(account)}
                        >
                          {account.nickname}
                        </td>
                        <td className="w-1/5 px-4 py-4 whitespace-nowrap text-sm">
                          <Clipboard
                            rawContent={account.address}
                            displayedContent={maskAddress(account.address)}
                          />
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </PopupModalContent>
        </PopupModal>
      )}

      <ConfirmModal
        isOpen={isConfirmModalOpen}
        onClose={handleConfirmModalClose}
        onConfirm={handleConfirmAdd}
        title={Intl.t('staking.add-address.title')}
      >
        <div className="flex flex-col gap-4">
          <p className="mas-body text-f-primary">
            {selectedAccount &&
              Intl.t('staking.add-address.description', {
                nickname: selectedAccount.nickname,
              })}
          </p>

          {/* Warning Zone */}
          <div className="bg-yellow-500/20 border border-yellow-500/50 rounded-lg p-3 text-center">
            <p className="mas-body text-yellow-300 text-sm">
              {Intl.t('staking.add-address.warning')}
            </p>
          </div>

          <Password
            value={password}
            onChange={(e) => setPassword(e.target.value)}
          />
        </div>
      </ConfirmModal>
    </>
  );
};

export default AddStakingAddress;
