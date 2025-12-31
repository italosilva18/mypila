import React from 'react';
import { ChevronLeft, Bell } from 'lucide-react';

interface Props {
  title: string;
  showBack?: boolean;
  onBack?: () => void;
  showNotifications?: boolean;
}

export const MobileHeader: React.FC<Props> = ({ title, showBack, onBack, showNotifications = true }) => {
  return (
    <header className="sticky top-0 z-40 md:hidden">
      <div className="bg-white/80 backdrop-blur-xl border-b border-stone-100 px-3 py-2">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            {showBack && (
              <button
                onClick={onBack}
                className="p-1 -ml-1 rounded-md active:bg-stone-100 transition-colors"
              >
                <ChevronLeft size={18} className="text-stone-600" />
              </button>
            )}
            <h1 className="text-sm font-semibold text-stone-900">{title}</h1>
          </div>
          {showNotifications && (
            <button className="p-1.5 rounded-md active:bg-stone-100 transition-colors">
              <Bell size={16} className="text-stone-500" />
            </button>
          )}
        </div>
      </div>
    </header>
  );
};
