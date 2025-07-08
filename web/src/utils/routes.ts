import { NavigateFunction } from 'react-router-dom';

export enum Path {
  home = 'home',
  dashboard = 'dashboard',
  stacking = 'stacking',
  error = 'error',
}

export function routeFor(path: string) {
  return `${import.meta.env.VITE_BASE_APP}/${path}`;
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
