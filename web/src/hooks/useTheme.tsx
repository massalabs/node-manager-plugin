import { useLocalStorage } from '@massalabs/react-ui-kit/src/lib/util/hooks/useLocalStorage';
import { FiSun, FiMoon } from 'react-icons/fi';
import {Theme} from '@massalabs/react-ui-kit'

const THEME_STORAGE_KEY = 'massa-node-manager-theme';

type ThemeSettings = {
  [key: string]: {
    icon: JSX.Element;
    label: string;
  };
};

const themeSettings: ThemeSettings = {
  'theme-dark': {
    icon: <FiSun />,
    label: 'light theme',
  },
  'theme-light': {
    icon: <FiMoon />,
    label: 'dark theme',
  },
};

export function useTheme() {
  const [theme, setTheme] = useLocalStorage<Theme>(
    THEME_STORAGE_KEY,
    'theme-dark',
  );

  const themeIcon = themeSettings[theme].icon;
  const themeLabel = themeSettings[theme].label;

  function handleSetTheme() {
    setTheme(theme === 'theme-dark' ? 'theme-light' : 'theme-dark');
  }

  return {
    theme,
    themeIcon,
    themeLabel,
    handleSetTheme,
  };
}