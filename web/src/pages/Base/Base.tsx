import { Toast, SideMenu, MassaLogo } from '@massalabs/react-ui-kit';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import {
  FiHome,
} from 'react-icons/fi';
import { AiOutlineDashboard } from 'react-icons/ai';
import { GrMoney } from "react-icons/gr";

import Intl from '@/i18n/i18n';
import { Path, routeFor } from '@/utils/routes';
import { useTheme } from '@/hooks/useTheme';
import { NodeStatusDisplay } from '@/components/NodeStatusDisplay';
import { useNodeStatus } from '@/hooks/useNodeStatus';

function isActive(path: Path) {
  const location = useLocation();
  return location.pathname.endsWith(path);
}


function Base() {
  const { theme, themeIcon, themeLabel, handleSetTheme } = useTheme();
  const navigate = useNavigate();
  const context = { themeLabel, themeIcon, theme, handleSetTheme };

  useNodeStatus().startListeningStatus();

  let menuConf = {
    title: 'Massa Node Manager',
    logo: <MassaLogo/>,
    fullMode: true,
  };

  let menuItems = [
    {
      label: Intl.t('menu.home'),
      icon: <FiHome data-testid="side-menu-home-icon" />,
      active: isActive(Path.home),
      footer: false,
      onClickItem: () => navigate(routeFor(Path.home))
    },
    {
      label: Intl.t('menu.dashboard'),
      icon: <AiOutlineDashboard data-testid="side-menu-dashboard-icon" />,
      active: isActive(Path.dashboard),
      footer: false,
      onClickItem: () =>
        navigate(routeFor(Path.dashboard)),
    },
    {
      label: Intl.t('menu.stacking'),
      icon: <GrMoney data-testid="side-menu-stacking-icon" />,
      active: isActive(Path.stacking),
      footer: false,
      onClickItem: () =>
        navigate(routeFor(Path.stacking)),
    },
    {
      label: themeLabel,
      icon: themeIcon,
      active: false,
      footer: true,
      onClickItem: () => handleSetTheme(),
    },
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
          <div className="flex justify-center p-4">
            <NodeStatusDisplay />
          </div>
          <div className="flex flex-1 justify-center items-center">
            <Outlet context={context} />
          </div>
        </div>
        <div className="absolute top-0 right-0 p-6">
          <div className="w-64">
            VERSION TODO
          </div>
        </div>
      </div>

      <Toast />
    </div>
  );
}

export default Base;
