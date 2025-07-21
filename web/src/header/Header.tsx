import React from 'react';

import { SelectNetwork } from './SelectNetwork';
import { useNodeStore } from '@/store/nodeStore';
import { getBaseAppUrl } from '@/utils/utils';

// Custom NodeLogo component
const NodeLogo: React.FC<{ size?: number }> = ({ size = 32 }) => {
  return (
    <div className="bg-primary w-fit rounded-full p-1">
      <img
        src={getBaseAppUrl() + '/favicon.svg'}
        alt="Node Logo"
        width={size}
        height={size}
        className="w-full h-full object-contain"
      />
    </div>
  );
};

export const Header: React.FC = () => {
  const pluginVersion = useNodeStore((state) => state.pluginVersion);

  return (
    <div className="flex justify-between items-center p-4 bg-primary mb-10">
      {/* Left side - Logo and Plugin Version */}
      <div className="flex items-center gap-2">
        <NodeLogo />
        {pluginVersion && (
          <span className="text-sm text-white ml-2">{pluginVersion}</span>
        )}
      </div>

      {/* Right side - SelectNetwork */}
      <div className="flex items-center">
        <SelectNetwork />
      </div>
    </div>
  );
};
