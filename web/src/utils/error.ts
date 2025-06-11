import { Path } from './routes';
export type ErrorData = {
  message: string;
  title: string;
};

export function getErrorPath() {
  return `${import.meta.env.VITE_BASE_APP}/${Path.error}`;
}
