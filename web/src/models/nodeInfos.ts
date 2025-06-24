export interface startNodeReponse {
  version: string;
}

export interface startNodeBody {
  useBuildnet: boolean;
  password: string;
}

export interface nodeInfosResponse {
  autoRestart: boolean;
  version: string;
}

export interface autoRestartBody {
  autoRestart: boolean;
}
