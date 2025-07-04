import axios from 'axios';

const MASSA_STATION_URL = 'https://station.massa/';
const MASSA_WALLET_URL =
  MASSA_STATION_URL + 'plugin/massa-labs/massa-wallet/api/';

type PluginInfo = {
  name: string;
  status: string;
};

type Account = {
  address: string;
  nickname: string;
  status: 'ok' | 'corrupted';
};

export async function isMassaWalletInstalled(): Promise<boolean> {
  const response = await axios.get(MASSA_STATION_URL + 'plugin-manager');
  const plugins = response.data;

  // Check if Massa Wallet plugin exists in the list
  return plugins.some(
    (plugin: PluginInfo) =>
      plugin.name === 'Massa Wallet' && plugin.status === 'Up',
  );
}

export async function getStationAccounts(): Promise<Account[]> {
  const response = await axios.get(MASSA_WALLET_URL + 'accounts', {
    timeout: 5000,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
  });

  const accounts: Account[] = response.data;

  // Filter out corrupted accounts and return only OK ones
  return accounts.filter((account: Account) => account.status === 'ok');
}
