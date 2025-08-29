import axios from 'axios';

const MASSA_WALLET_API = 'plugin/massa-labs/massa-wallet/api/';

export function getMassaStationUrl() {
  if (import.meta.env.VITE_ENV === 'standalone') {
    return 'https://station.massa/';
  }
  return '/';
}

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
  try {
    const response = await axios.get(getMassaStationUrl() + 'plugin-manager');
    const plugins = response.data;

    // Check if Massa Wallet plugin exists in the list
    return plugins.some(
      (plugin: PluginInfo) =>
        plugin.name === 'Massa Wallet' && plugin.status === 'Up',
    );
  } catch (error) {
    return false;
  }
}

export async function getStationAccounts(): Promise<Account[]> {
  const response = await axios.get(
    getMassaStationUrl() + MASSA_WALLET_API + 'accounts',
    {
      timeout: 5000,
      headers: {
        Accept: 'application/json',
        'Content-Type': 'application/json',
      },
    },
  );

  const accounts: Account[] = response.data;

  // Filter out corrupted accounts and return only OK ones
  return accounts.filter((account: Account) => account.status === 'ok');
}
