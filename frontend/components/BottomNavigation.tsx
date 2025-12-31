import React from 'react';
import { useLocation, useNavigate } from 'react-router-dom';
import { LayoutDashboard, Tags, RefreshCw, FileText, ClipboardList } from 'lucide-react';

interface NavItem {
  path: string;
  icon: React.ReactNode;
  label: string;
}

export const BottomNavigation: React.FC = () => {
  const location = useLocation();
  const navigate = useNavigate();

  const navItems: NavItem[] = [
    { path: 'dashboard', icon: <LayoutDashboard size={18} />, label: 'Home' },
    { path: 'categories', icon: <Tags size={18} />, label: 'Categ.' },
    { path: 'recurring', icon: <RefreshCw size={18} />, label: 'Fixas' },
    { path: 'quotes', icon: <ClipboardList size={18} />, label: 'Orcam.' },
    { path: 'reports', icon: <FileText size={18} />, label: 'Relat.' },
  ];

  const isActive = (path: string) => location.pathname.endsWith(path);

  return (
    <nav className="fixed bottom-0 left-0 right-0 z-50 md:hidden">
      {/* Background com blur */}
      <div className="absolute inset-0 bg-white/80 backdrop-blur-xl border-t border-stone-200/50" />

      {/* Safe area padding para dispositivos com notch */}
      <div className="relative flex items-center justify-around px-1 pt-1.5 pb-safe">
        {navItems.map((item) => (
          <button
            key={item.path}
            onClick={() => navigate(item.path)}
            className={`flex flex-col items-center justify-center py-1 px-2 rounded-lg transition-all duration-200 min-w-[56px] ${
              isActive(item.path)
                ? 'text-stone-900'
                : 'text-stone-400 active:scale-95'
            }`}
          >
            {/* Indicador ativo */}
            {isActive(item.path) && (
              <div className="absolute -top-0 w-6 h-0.5 bg-stone-900 rounded-full" />
            )}

            {/* Container do ícone */}
            <div className={`p-1 rounded-md transition-all duration-200 ${
              isActive(item.path) ? 'bg-stone-100' : ''
            }`}>
              {item.icon}
            </div>

            {/* Label */}
            <span className={`text-[9px] mt-0.5 font-medium transition-all ${
              isActive(item.path) ? 'text-stone-900' : 'text-stone-400'
            }`}>
              {item.label}
            </span>
          </button>
        ))}
      </div>
    </nav>
  );
};
