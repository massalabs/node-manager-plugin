export enum NodeStatus {
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

// If the node is running and has finished bootstrapping
export function isReady(status: NodeStatus): boolean {
  return status != NodeStatus.BOOTSTRAPPING && isRunning(status);
}
