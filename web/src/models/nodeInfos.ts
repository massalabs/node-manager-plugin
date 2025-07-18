export interface startNodeBody {
  useBuildnet: boolean;
  password: string;
}

export interface networkData {
  version: string;
  hasPwd: boolean;
}

export interface nodeInfosResponse {
  autoRestart: boolean;
  networks: networkData[];
  isMainnet: boolean;
  pluginVersion: string;
}

export interface autoRestartBody {
  autoRestart: boolean;
}
