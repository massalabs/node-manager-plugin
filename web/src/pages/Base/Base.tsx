import { useEffect } from 'react';

import { Toast, SideMenu } from '@massalabs/react-ui-kit';
// import { AiOutlineDashboard } from 'react-icons/ai';
import { FiHome } from 'react-icons/fi';
import { GrMoney } from 'react-icons/gr';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';

import { NodeStatusDisplay } from '@/components/NodeStatusDisplay';
import { useNodeStatus } from '@/hooks/useNodeStatus';
import { useTheme } from '@/hooks/useTheme';
import Intl from '@/i18n/i18n';
import { useNodeStore } from '@/store/nodeStore';
import { Path, routeFor } from '@/utils/routes';

// Custom NodeLogo component to replace MassaLogo
const NodeLogo: React.FC<{ size?: number }> = ({ size = 32 }) => {
  return (
    <div className="bg-primary w-fit rounded-full p-1">
      <img
        src={import.meta.env.VITE_BASE_APP + '/favicon.svg'}
        alt="Node Logo"
        width={size}
        height={size}
        className="w-full h-full object-contain"
      />
    </div>
  );
};

function Base() {
  const { theme, themeIcon, themeLabel, handleSetTheme } = useTheme();
  const navigate = useNavigate();
  const context = { themeLabel, themeIcon, theme, handleSetTheme };

  const { startListeningStatus } = useNodeStatus();

  const nodeVersion = useNodeStore((state) => state.version);

  const location = useLocation();

  function isActive(path: Path) {
    return location.pathname.endsWith(path);
  }

  useEffect(() => {
    startListeningStatus();
  }, [startListeningStatus]);

  let menuConf = {
    title: 'Massa Node Manager',
    logo: <NodeLogo />,
    fullMode: true,
  };

  let menuItems = [
    {
      label: Intl.t('menu.home'),
      icon: <FiHome data-testid="side-menu-home-icon" />,
      active: isActive(Path.home),
      footer: false,
      onClickItem: () => navigate(routeFor(Path.home)),
    },
    // {
    //   label: Intl.t('menu.dashboard'),
    //   icon: <AiOutlineDashboard data-testid="side-menu-dashboard-icon" />,
    //   active: isActive(Path.dashboard),
    //   footer: false,
    //   onClickItem: () => navigate(routeFor(Path.dashboard)),
    // },
    {
      label: Intl.t('menu.stacking'),
      icon: <GrMoney data-testid="side-menu-stacking-icon" />,
      active: isActive(Path.stacking),
      footer: false,
      onClickItem: () => navigate(routeFor(Path.stacking)),
    },
    // {
    //   label: themeLabel,
    //   icon: themeIcon,
    //   active: false,
    //   footer: true,
    //   onClickItem: () => handleSetTheme(),
    // },
  ];

  return (
    <div className={theme}>
      <div className="bg-primary">
        <SideMenu
          conf={menuConf}
          items={menuItems}
          onClickLogo={() => navigate(routeFor(''))}
        />
        <div className="flex flex-col h-screen text-f-primary">
          <div className="flex justify-between items-center p-4">
            <div className="flex-1" />
            <div className="flex-1 flex justify-center">
              <NodeStatusDisplay />
            </div>
            {nodeVersion && (
              <div className="flex-1 flex justify-end mr-4">
                <span className="text-md text-white">{nodeVersion}</span>
              </div>
            )}
            {!nodeVersion && <div className="flex-1" />}
          </div>
          <div className="flex flex-1 justify-center items-center">
            <Outlet context={context} />
          </div>
        </div>
      </div>

      <Toast durationMs={1000} />
    </div>
  );
}

export default Base;
