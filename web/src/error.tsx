import Intl from '@/i18n/i18n';
import { ErrorData } from '@/utils/error';

interface ErrorProps {
  errorData: ErrorData;
  onReturn: () => void;
}

export default function Error({ errorData, onReturn }: ErrorProps) {
  return (
    <div
      id="error-page"
      className="flex flex-col justify-center items-center h-screen text-f-primary"
    >
      <h1 className="mas-banner mb-5">
        {errorData?.title
          ? errorData.title
          : Intl.t('errors.unexpected-error.title')}
      </h1>
      <p
        className={
          'mas-body p-5 w-screen max-w-md overflow-y-auto border ' +
          'border-white bg-black max-h-screen/4 rounded-lg text-center text-white'
        }
      >
        {errorData
          ? errorData.message
          : Intl.t('errors.unexpected-error.description')}
      </p>
      <button onClick={onReturn} className="underline mt-5">
        {Intl.t('errors.back-to-home-link')}
      </button>
    </div>
  );
}
