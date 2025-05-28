import { Link, useLocation } from 'react-router-dom';

import Intl from '@/i18n/i18n';
import { routeFor, ErrorData } from '@/utils';

export default function Error() {
  const location = useLocation();
  const errorData: ErrorData = location.state?.error
  return (
      <div
        id="error-page"
        className="flex flex-col justify-center items-center h-screen text-f-primary"
      >
        <h1 className="mas-banner">
          {errorData?.title ? errorData.title : Intl.t('errors.unexpected-error.title')}
        </h1>
        <p className="mas-bod">
          {errorData ? errorData.message : Intl.t('errors.unexpected-error.description')}
        </p>
        <Link to={routeFor('index')} className="underline">
          {Intl.t('errors.unexpected-error.link')}
        </Link>
      </div>
  );
}
