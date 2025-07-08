import { AxiosError } from 'axios';
import { Path } from './routes';
export type ErrorData = {
  message: string;
  title: string;
};

export function getErrorPath() {
  return `${import.meta.env.VITE_BASE_APP}/${Path.error}`;
}

export function getErrorMessage(error: unknown): string {
  if (error instanceof AxiosError) {
    return (error.response?.data as Error).message || error.message;
  }
  return error instanceof Error ? error.message : String(error);
}