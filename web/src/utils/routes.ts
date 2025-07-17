import { NavigateFunction } from 'react-router-dom';

import { getBaseAppUrl } from './utils';

export enum Path {
  home = 'home',
  dashboard = 'dashboard',
  stacking = 'stacking',
  error = 'error',
}

export function routeFor(path: string) {
  return `${getBaseAppUrl()}/${path}`;
}

export function goToErrorPage(
  navigate: NavigateFunction,
  title: string,
  message: string,
) {
  navigate(routeFor('error'), {
    state: {
      error: {
        title,
        message,
      },
    },
  });
}
