import React, { createContext, useContext, useState, useCallback, ReactNode, useRef, useEffect } from 'react';

export interface Toast {
  id: string;
  type: 'success' | 'error' | 'warning' | 'info';
  message: string;
}

interface ToastContextType {
  toasts: Toast[];
  addToast: (type: Toast['type'], message: string) => void;
  removeToast: (id: string) => void;
}

const ToastContext = createContext<ToastContextType | undefined>(undefined);

let toastCounter = 0;
const MAX_TOASTS = 5; // Maximum number of toasts to show at once

export const ToastProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [toasts, setToasts] = useState<Toast[]>([]);
  const timeoutRefs = useRef<Map<string, NodeJS.Timeout>>(new Map());

  // Cleanup all timeouts on unmount
  useEffect(() => {
    return () => {
      timeoutRefs.current.forEach(timeout => clearTimeout(timeout));
      timeoutRefs.current.clear();
    };
  }, []);

  const removeToast = useCallback((id: string) => {
    // Clear timeout if exists
    const timeout = timeoutRefs.current.get(id);
    if (timeout) {
      clearTimeout(timeout);
      timeoutRefs.current.delete(id);
    }
    setToasts(prev => prev.filter(toast => toast.id !== id));
  }, []);

  const addToast = useCallback((type: Toast['type'], message: string) => {
    const id = `toast-${Date.now()}-${++toastCounter}`;
    const newToast: Toast = { id, type, message };

    setToasts(prev => {
      // If we have MAX_TOASTS, remove the oldest one
      const updatedToasts = prev.length >= MAX_TOASTS ? prev.slice(1) : prev;
      return [...updatedToasts, newToast];
    });

    // Auto-dismiss after 5 seconds with proper cleanup
    const timeout = setTimeout(() => {
      setToasts(current => current.filter(toast => toast.id !== id));
      timeoutRefs.current.delete(id);
    }, 5000);

    // Store timeout reference for cleanup
    timeoutRefs.current.set(id, timeout);
  }, []);

  return (
    <ToastContext.Provider value={{ toasts, addToast, removeToast }}>
      {children}
    </ToastContext.Provider>
  );
};

export const useToast = () => {
  const context = useContext(ToastContext);
  if (context === undefined) {
    throw new Error('useToast must be used within a ToastProvider');
  }
  return context;
};
