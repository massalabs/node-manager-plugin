import { useMemo } from 'react';

import { AccordionCategory, Balance, Tooltip } from '@massalabs/react-ui-kit';
import { FiInfo } from 'react-icons/fi';

import Intl from '@/i18n/i18n';
import { StakingAddress } from '@/models/staking';
import { DeferredCredit, Slot } from '@/models/staking';

interface DeferredCreditListProps {
  currentAddress: StakingAddress;
  lastSlot: Slot | undefined;
  t0: number | undefined;
}

const DeferredCreditList: React.FC<DeferredCreditListProps> = (
  props: DeferredCreditListProps,
) => {
  const getDeferredCreditReleaseDate = useMemo(() => {
    return (credit: DeferredCredit) => {
      if (!props.lastSlot) {
        console.error('Node info last slot is null');
      }
      const periodDiff = credit.slot.period - (props.lastSlot?.period || 0);
      const periodLength = props.t0 || 0;
      const releaseTime = periodDiff * periodLength;
      return new Date(Date.now() + releaseTime);
    };
  }, [props.t0, props.lastSlot]);

  const getDeferredCreditsTable = useMemo(() => {
    if (
      !props.currentAddress?.deferred_credits ||
      props.currentAddress?.deferred_credits.length === 0
    ) {
      return <p className="text-gray-400 text-center">No locked MAS</p>;
    }

    return (
      <table className="min-w-full divide-y divide-gray-600">
        <thead className="bg-gray-700">
          <tr>
            <th className="px-4 py-2 text-left text-xs font-medium text-gray-300 uppercase tracking-wider w-1/4">
              Amount
            </th>
            <th className="px-4 py-2 text-left text-xs font-medium text-gray-300 uppercase tracking-wider w-3/4">
              Approx Release Date
            </th>
          </tr>
        </thead>
        <tbody className="bg-secondary divide-y divide-gray-600">
          {props.currentAddress?.deferred_credits.map((credit, index) => {
            const releaseDate = getDeferredCreditReleaseDate(credit);

            return (
              <tr key={index} className="border-b border-gray-600">
                <td className="px-4 py-2 text-sm">
                  <Balance size="xs" amount={credit.amount.toFixed(2)} />
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
  }, [props.currentAddress?.deferred_credits, getDeferredCreditReleaseDate]);

  return (
    <div>
      <AccordionCategory
        categoryTitle={
          <div className="flex items-center justify-between w-full">
            <span className="flex items-center gap-1">
              Locked MAS
              <Tooltip
                body={Intl.t(
                  'staking.stakingAddressDetails.deferred-credits-tooltip',
                )}
              >
                <FiInfo className="w-3 h-3 text-gray-400" />
              </Tooltip>
            </span>
            <span className="text-sm text-gray-400 mr-5">
              ({props.currentAddress?.deferred_credits?.length || 0})
            </span>
          </div>
        }
      >
        <div className="mt-2 max-h-96 overflow-auto">
          {getDeferredCreditsTable}
        </div>
      </AccordionCategory>
    </div>
  );
};

export default DeferredCreditList;
