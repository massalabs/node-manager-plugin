export interface Slot {
  period: number;
  thread: number;
}

export interface DeferredCredit {
  slot: Slot;
  amount: number;
}

export interface StakingAddress {
  address: string;
  final_roll_count: number;
  candidate_roll_count: number;
  active_roll_count: number;
  final_balance: number;
  candidate_balance: number;
  thread: number;
  deferred_credits: DeferredCredit[];
  // if target rolls is negative, it means that the address is auto-compounding: buy as many rolls as possible
  target_rolls: number;
}

export interface StakingAddressesResponse {
  addresses: StakingAddress[];
}

export interface AddStakingAddressBody {
  password: string;
  nickname: string;
}

export interface UpdateStakingAddressBody {
  address: string;
  target_rolls: number;
}

export interface RemoveStakingAddressBody {
  address: string;
}

export interface RollOpHistory {
  op: 'BUY' | 'SELL';
  amount: number;
  timestamp: string;
  opId?: string; // Optional for backward compatibility
}

export interface RollOpHistoryResponse {
  operations: RollOpHistory[];
}
