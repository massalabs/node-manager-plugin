export enum NodeStatus {
  ON = 'on',
  OFF = 'off',
  BOOTSTRAPPING = 'bootstrapping',
  STOPPING = 'stopping',
  CRASHED = 'crashed',
  DESYNCED = 'desynced',
  RESTARTING = 'restarting',
  PLUGINERROR = 'pluginError',
}

// When the node process is running
export function isRunning(status: NodeStatus): boolean {
  return status !== NodeStatus.OFF && status !== NodeStatus.CRASHED;
}

// If the node is running and has finished bootstrapping
export function isReady(status: NodeStatus): boolean {
  return status != NodeStatus.BOOTSTRAPPING && isRunning(status);
}
