export enum NodeStatus {
  UNSET = 'unset',
  ON = 'on',
  OFF = 'off',
  STARTING = 'starting',
  BOOTSTRAPPING = 'bootstrapping',
  STOPPING = 'stopping',
  CRASHED = 'crashed',
  DESYNCED = 'desynced',
}

// When the node process is running
export function isRunning(status: NodeStatus): boolean {
  return status !== NodeStatus.OFF && status !== NodeStatus.CRASHED;
}

export function isStopStakingMonitoring(status: NodeStatus): boolean {
  return (
    status === NodeStatus.CRASHED ||
    status === NodeStatus.DESYNCED ||
    status === NodeStatus.STOPPING
  );
}

export function showRpcAddButton(status: NodeStatus): boolean {
  return status === NodeStatus.ON || status === NodeStatus.BOOTSTRAPPING;
}
