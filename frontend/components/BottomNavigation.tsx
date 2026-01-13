import React from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { LayoutDashboard, Tags, RefreshCw, FileText, ClipboardList, Settings } from 'lucide-react';

interface NavItem {
  path: string;
  icon: React.ReactNode;
  label: string;
  isAction?: boolean;
  onClick?: () => void;
}

interface BottomNavigationProps {
  onSettingsClick?: () => void;
}

export const BottomNavigation: React.FC<BottomNavigationProps> = ({ onSettingsClick }) => {
  const location = useLocation();
  const navigate = useNavigate();

  const navItems: NavItem[] = [
    { path: 'dashboard', icon: <LayoutDashboard size={18} />, label: 'Home' },
    { path: 'categories', icon: <Tags size={18} />, label: 'Categ.' },
    { path: 'recurring', icon: <RefreshCw size={18} />, label: 'Fixas' },
    { path: 'quotes', icon: <ClipboardList size={18} />, label: 'Orcam.' },
    { path: 'reports', icon: <FileText size={18} />, label: 'Relat.' },
    ...(onSettingsClick ? [{ path: 'settings', icon: <Settings size={18} />, label: 'Config.', isAction: true, onClick: onSettingsClick }] : []),
  ];

  const isActive = (path: string) => location.pathname.endsWith(path);

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 md:hidden" role="navigation" aria-label="Navegacao principal">
      {/* Background */}
      <div className="absolute inset-0 bg-card border-t border-border shadow-soft" />

      {/* Safe area padding */}
      <div className="relative flex items-center justify-around px-1 pt-1.5 pb-safe">
        {navItems.map((item) => (
          <button
            key={item.path}
            onClick={() => item.isAction && item.onClick ? item.onClick() : navigate(item.path)}
            aria-label={item.label}
            aria-current={!item.isAction && isActive(item.path) ? 'page' : undefined}
            className={`flex flex-col items-center justify-center py-1 px-2 rounded-lg transition-all duration-200 min-w-[48px] ${
              !item.isAction && isActive(item.path)
                ? 'text-primary-600'
                : 'text-muted active:scale-95'
            }`}
          >
            {/* Active indicator */}
            {!item.isAction && isActive(item.path) && (
              <div className="absolute -top-0 w-6 h-0.5 bg-gradient-primary rounded-full" />
            )}

            {/* Icon container */}
            <div className={`p-1 rounded-md transition-all duration-200 ${
              !item.isAction && isActive(item.path) ? 'bg-primary-100' : ''
            }`}>
              {item.icon}
            </div>

            {/* Label */}
            <span className={`text-[9px] mt-0.5 font-medium transition-all ${
              !item.isAction && isActive(item.path) ? 'text-primary-600' : 'text-muted'
            }`}>
              {item.label}
            </span>
          </button>
        ))}
      </div>
    </nav>
  );
};
