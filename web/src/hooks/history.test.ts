import { getTotalValue } from './useTotValueHistory';
import { StakingAddress } from '@/models/staking';

// Mock data
const mockStakingAddresses: StakingAddress[] = [
  {
    address: 'address1',
    thread: 1,
    final_balance: 1000,
    candidate_balance: 500,
    active_roll_count: 5,
    final_roll_count: 10,
    candidate_roll_count: 8,
    target_rolls: 12,
    deferred_credits: [
      { amount: 100, slot: { period: 1, thread: 1 } },
      { amount: 200, slot: { period: 2, thread: 1 } },
    ],
  },
  {
    address: 'address2',
    thread: 2,
    final_balance: 2000,
    candidate_balance: 1000,
    active_roll_count: 3,
    final_roll_count: 6,
    candidate_roll_count: 4,
    target_rolls: 8,
    deferred_credits: [{ amount: 150, slot: { period: 3, thread: 2 } }],
  },
];

describe('getTotalValue function', () => {
  it('should calculate total value correctly for multiple addresses', () => {
    const result = getTotalValue(mockStakingAddresses);

    // Expected calculation:
    // Address 1: 1000 + (10 * 100) + (100 + 200) = 1000 + 1000 + 300 = 2300
    // Address 2: 2000 + (6 * 100) + 150 = 2000 + 600 + 150 = 2750
    // Total: 2300 + 2750 = 5050
    expect(result).toBe(5050);
  });

  it('should handle empty addresses array', () => {
    const result = getTotalValue([]);
    expect(result).toBe(0);
  });

  it('should handle addresses with no deferred credits', () => {
    const addressesWithoutDeferredCredits: StakingAddress[] = [
      {
        ...mockStakingAddresses[0],
        deferred_credits: [],
      },
    ];

    const result = getTotalValue(addressesWithoutDeferredCredits);
    // Expected: 1000 + (10 * 100) + 0 = 2000
    expect(result).toBe(2000);
  });

  it('should handle addresses with zero balances and rolls', () => {
    const zeroAddresses: StakingAddress[] = [
      {
        address: 'zero',
        thread: 1,
        final_balance: 0,
        candidate_balance: 0,
        active_roll_count: 0,
        final_roll_count: 0,
        candidate_roll_count: 0,
        target_rolls: 0,
        deferred_credits: [],
      },
    ];

    const result = getTotalValue(zeroAddresses);
    expect(result).toBe(0);
  });

  it('should handle addresses with only deferred credits', () => {
    const onlyDeferredCredits: StakingAddress[] = [
      {
        address: 'deferred',
        thread: 1,
        final_balance: 0,
        candidate_balance: 0,
        active_roll_count: 0,
        final_roll_count: 0,
        candidate_roll_count: 0,
        target_rolls: 0,
        deferred_credits: [
          { amount: 500, slot: { period: 1, thread: 1 } },
          { amount: 300, slot: { period: 2, thread: 1 } },
        ],
      },
    ];

    const result = getTotalValue(onlyDeferredCredits);
    expect(result).toBe(800);
  });
});
