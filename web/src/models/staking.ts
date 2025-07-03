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
  finalRolls: number;
  candidateRolls: number;
  finalBalance: number;
  candidateBalance: number;
  thread: number;
  deferredCredits: DeferredCredit[];
  targetRolls: number;
}

export interface StakingAddressesResponse {
  addresses: StakingAddress[];
}

export interface AddStakingAddressBody {
  nickname: string;
}

export interface UpdateStakingAddressBody {
  address: string;
  targetRolls: number;
}

export interface RemoveStakingAddressBody {
  address: string;
} 