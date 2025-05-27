import React from 'react';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import ReactDOM from 'react-dom/client';
import {
  RouterProvider,
  createBrowserRouter,
  createRoutesFromElements,
  Route,
} from 'react-router-dom';

import '@massalabs/react-ui-kit/src/global.css';
import './index.css';
import Base from '@/pages/Base/Base.tsx';
import Error from '@/pages/Error.tsx';
import Home from '@/pages/Home/Home.tsx';
import { Path } from '@/utils/routes.ts';

const baseURL = import.meta.env.VITE_BASE_APP;

const queryClient = new QueryClient();

const router = createBrowserRouter(
  createRoutesFromElements(
    <Route path={baseURL} element={<Base />}>
      <Route path={baseURL+Path.home} element={<Home />} />
      <Route path={baseURL+Path.dashboard} element={<Home />} />
      <Route path={baseURL+Path.stacking} element={<Home />} />

      {/* routes for errors */}
      <Route path={baseURL+"error"} element={<Error />} />
      <Route path="*" element={<Error />} />
    </Route>,
  ),
);

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} fallbackElement={<Error />} />
    </QueryClientProvider>
  </React.StrictMode>,
);
