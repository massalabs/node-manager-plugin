import { NavigateFunction } from 'react-router-dom';

export enum Path {
  home = '/',
  dashboard = '/dashboard',
  stacking = '/stacking',
}

export function routeFor(path: string) {
  return `${import.meta.env.VITE_BASE_APP}/${path}`;
}

export function goToErrorPage(navigate: NavigateFunction) {
  navigate(routeFor('error'));
}
