
// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-nocheck
// Jest global mocks below
import { TextEncoder, TextDecoder } from 'util';

if (typeof global.TextEncoder === 'undefined') {
  global.TextEncoder = TextEncoder;
}
if (typeof global.TextDecoder === 'undefined') {
  global.TextDecoder = TextDecoder;
}

// Mock utility functions that use import.meta.env
jest.mock('@/utils/utils', () => ({
  getApiUrl: () => '/TestApi',
  getBaseAppUrl: () => '/TestApp',
}));
