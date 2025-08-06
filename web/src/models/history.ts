export type ValueHistoryPoint = {
  timestamp: string;
  value: number | null;
};

export enum SinceFetch {
  H1 = '1H',
  D1 = '1D',
  W1 = '1W',
  M1 = '1M',
  Y1 = '1Y',
  DEFAULT = '',
}

export type ValueHistorySamplesResponse = {
  samples: ValueHistoryPoint[];
  emptyDataPointNum: number;
};
