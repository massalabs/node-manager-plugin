export interface startNodeReponse {
  version: string;
}

export interface startNodeBody {
  useBuildnet: boolean;
  password: string;
}

export interface configBody {
  autoRestart: boolean;
}
