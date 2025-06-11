import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import ReactDOM from 'react-dom/client';
import {
  RouterProvider,
  createBrowserRouter,
  createRoutesFromElements,
  Route,
  Navigate,
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
      <Route index element={<Navigate to={Path.home} />} />
      <Route
        path="index"
        element={<Navigate to={baseURL + '/' + Path.home} />}
      />
      <Route path={Path.home} element={<Home />} />

      <Route path={Path.dashboard} element={<Home />} />
      <Route path={Path.stacking} element={<Home />} />

      {/* routes for errors */}
      <Route path={'error'} element={<Error />} />
      <Route
        path="*"
        element={
          <Navigate
            to={baseURL + '/' + Path.error}
            state={{
              error: {
                title: '404 page not found',
                message: "This page doesn't exist",
              },
            }}
          />
        }
      />
    </Route>,
  ),
);

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <QueryClientProvider client={queryClient}>
    <RouterProvider router={router} />
  </QueryClientProvider>,
);
