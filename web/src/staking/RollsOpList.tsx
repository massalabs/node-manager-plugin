import { useMemo } from 'react';

import {
  AccordionCategory,
  Clipboard,
  maskAddress,
  Tooltip,
} from '@massalabs/react-ui-kit';
import { FiInfo } from 'react-icons/fi';

import { useRollOpHistory } from '@/hooks/useRollOpHistory';
import Intl from '@/i18n/i18n';
import { useNodeStore } from '@/store/nodeStore';

const RollsOpList: React.FC<{ address: string }> = (props: {
  address: string;
}) => {
  const currentNetwork = useNodeStore((state) => state.currentNetwork);
  const isMainnet = currentNetwork === 'mainnet';

  const { data: rollOpHistory, isLoading: isLoadingRollOpHistory } =
    useRollOpHistory(props.address, isMainnet);

  const getRollOpHistoryTable = useMemo(() => {
    if (isLoadingRollOpHistory) {
      return (
        <p className="text-gray-400 text-center">Loading roll operations...</p>
      );
    }

    if (!rollOpHistory?.operations || rollOpHistory.operations.length === 0) {
      return (
        <p className="text-gray-400 text-center">No roll operations found</p>
      );
    }

    return (
      <table className="min-w-full divide-y divide-gray-900">
        <thead className="bg-gray-700">
          <tr>
            <th className="px-2 py-2 text-left text-xs font-medium text-gray-300 uppercase tracking-wider w-1/6">
              Operation
            </th>
            <th className="px-2 py-2 text-left text-xs font-medium text-gray-300 uppercase tracking-wider w-1/6">
              Amount
            </th>
            <th className="px-2 py-2 text-left text-xs font-medium text-gray-300 uppercase tracking-wider w-1/6">
              ID
            </th>
            <th className="px-2 py-2 text-left text-xs font-medium text-gray-300 uppercase tracking-wider w-1/6">
              Date
            </th>
          </tr>
        </thead>
        <tbody className="bg-secondary divide-y divide-gray-600">
          {rollOpHistory.operations.map((operation, index) => (
            <tr key={index} className="border-b border-gray-600">
              <td className="px-2 py-2 text-sm w-1/6 text-center">
                <span
                  className={`inline-flex items-center justify-center px-2 py-1 rounded-full text-xs font-medium ${
                    operation.op === 'BUY'
                      ? 'bg-green-100 text-green-800'
                      : 'bg-red-100 text-red-800'
                  }`}
                >
                  {operation.op}
                </span>
              </td>
              <td className="px-2 py-2 text-sm text-f-primary w-1/6 text-center">
                {operation.amount}
              </td>
              <td className="px-2 py-2 text-sm text-f-primary w-1/6 text-center">
                <Clipboard
                  rawContent={operation.opId ?? ''}
                  displayedContent={maskAddress(operation.opId ?? '')}
                />
              </td>
              <td className="px-2 py-2 text-sm text-f-primary text-xs text-center">
                {new Date(operation.timestamp).toLocaleString()}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    );
  }, [rollOpHistory, isLoadingRollOpHistory]);

  return (
    <div className="border-t border-gray-600 pt-4">
      <AccordionCategory
        categoryTitle={
          <div className="flex items-center justify-between w-full">
            <span className="flex items-center gap-1">
              Roll Operations
              <Tooltip
                body={Intl.t('stakingAddressDetails.roll-op-history-tooltip')}
              >
                <FiInfo className="w-3 h-3 text-gray-400" />
              </Tooltip>
            </span>
            <span className="text-sm text-gray-400 mr-5">
              ({rollOpHistory?.operations?.length || 0})
            </span>
          </div>
        }
        isChild={false}
      >
        <div className="mt-2 max-h-96 overflow-auto">
          {getRollOpHistoryTable}
        </div>
      </AccordionCategory>
    </div>
  );
};

export default RollsOpList;
