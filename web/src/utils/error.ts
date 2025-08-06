import { AxiosError } from 'axios';

export type ErrorData = {
  message: string;
  title: string;
};

export function getErrorMessage(error: unknown): string {
  if (error instanceof AxiosError) {
    return (error.response?.data as Error).message || error.message;
  }
  return error instanceof Error ? error.message : String(error);
}
