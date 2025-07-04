import React from 'react';

import { Spinner } from '@massalabs/react-ui-kit';

const Loading: React.FC<{ message?: string }> = ({ message }) => {
  return (
    <div className="flex items-center justify-center h-full">
      <div className="text-center">
        <div className="flex justify-center mb-4">
          <Spinner />
        </div>
        <p className="mas-body text-f-primary text-lg">{message}</p>
      </div>
    </div>
  );
};

export default Loading;
