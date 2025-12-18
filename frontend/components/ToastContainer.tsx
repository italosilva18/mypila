import React, { useEffect, useState } from 'react';
import { useToast, Toast } from '../contexts/ToastContext';
import { X, CheckCircle2, XCircle, AlertTriangle, Info } from 'lucide-react';

const ToastItem: React.FC<{ toast: Toast }> = ({ toast }) => {
  const { removeToast } = useToast();
  const [isExiting, setIsExiting] = useState(false);

  const handleClose = () => {
    setIsExiting(true);
    setTimeout(() => {
      removeToast(toast.id);
    }, 300); // Match animation duration
  };

  useEffect(() => {
    // Trigger entrance animation
    const timer = setTimeout(() => {
      setIsExiting(false);
    }, 10);
    return () => clearTimeout(timer);
  }, []);

  const getToastStyles = () => {
    const baseStyles = 'flex items-start gap-3 p-4 rounded-xl shadow-card border-2 backdrop-blur-sm';
    const variants = {
      success: 'bg-emerald-50/95 border-emerald-200 text-emerald-900',
      error: 'bg-rose-50/95 border-rose-200 text-rose-900',
      warning: 'bg-amber-50/95 border-amber-200 text-amber-900',
      info: 'bg-sky-50/95 border-sky-200 text-sky-900',
    };
    return `${baseStyles} ${variants[toast.type]}`;
  };

  const getIcon = () => {
    const iconProps = { className: 'w-5 h-5 flex-shrink-0 mt-0.5', strokeWidth: 2.5 };
    const icons = {
      success: <CheckCircle2 {...iconProps} className="w-5 h-5 flex-shrink-0 mt-0.5 text-emerald-600" />,
      error: <XCircle {...iconProps} className="w-5 h-5 flex-shrink-0 mt-0.5 text-rose-600" />,
      warning: <AlertTriangle {...iconProps} className="w-5 h-5 flex-shrink-0 mt-0.5 text-amber-600" />,
      info: <Info {...iconProps} className="w-5 h-5 flex-shrink-0 mt-0.5 text-sky-600" />,
    };
    return icons[toast.type];
  };

  return (
    <div
      className={`${getToastStyles()} transform transition-all duration-300 ease-out ${
        isExiting
          ? 'translate-y-full md:translate-y-0 md:translate-x-full opacity-0'
          : 'translate-y-0 md:translate-x-0 opacity-100'
      }`}
      style={{
        minWidth: 'auto',
        maxWidth: '100%',
      }}
    >
      {getIcon()}
      <p className="flex-1 text-xs md:text-sm font-medium leading-relaxed">
        {toast.message}
      </p>
      <button
        onClick={handleClose}
        className="flex-shrink-0 p-1 rounded-lg hover:bg-black/5 active:bg-black/5 transition-colors"
        aria-label="Fechar notificacao"
      >
        <X className="w-4 h-4" strokeWidth={2.5} />
      </button>
    </div>
  );
};

export const ToastContainer: React.FC = () => {
  const { toasts } = useToast();

  return (
    <div
      className="fixed bottom-20 md:bottom-auto md:top-4 left-4 right-4 md:left-auto md:right-4 z-50 flex flex-col gap-2 md:gap-3 pointer-events-none"
      aria-live="polite"
      aria-atomic="true"
    >
      {toasts.map(toast => (
        <div key={toast.id} className="pointer-events-auto">
          <ToastItem toast={toast} />
        </div>
      ))}
    </div>
  );
};
