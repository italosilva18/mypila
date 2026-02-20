import React, { useState, useEffect, useRef, useCallback } from 'react';
import { Bell, AlertCircle, X, Calendar } from 'lucide-react';
import { api } from '../services/api';
import { UpcomingTransaction } from '../types';
import { formatCurrency } from '../utils/currency';

interface Props {
  companyId: string;
}

export const NotificationBell: React.FC<Props> = ({ companyId }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [upcoming, setUpcoming] = useState<UpcomingTransaction[]>([]);
  const [loading, setLoading] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const fetchUpcoming = useCallback(async () => {
    if (!companyId) return;
    setLoading(true);
    try {
      const response = await api.getUpcomingTransactions(companyId, 7);
      setUpcoming(response.upcoming);
    } catch (error) {
      console.error('Failed to fetch upcoming transactions', error);
    } finally {
      setLoading(false);
    }
  }, [companyId]);

  useEffect(() => {
    fetchUpcoming();
    // Refresh every 5 minutes
    const interval = setInterval(fetchUpcoming, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, [fetchUpcoming]);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  const getDueDayLabel = (daysUntilDue: number): string => {
    if (daysUntilDue === 0) return 'Vence hoje!';
    if (daysUntilDue === 1) return 'Vence amanha';
    return `Vence em ${daysUntilDue} dias`;
  };

  const getDueDayColor = (daysUntilDue: number): string => {
    if (daysUntilDue === 0) return 'text-red-600 bg-red-100';
    if (daysUntilDue <= 2) return 'text-orange-600 bg-orange-100';
    return 'text-yellow-600 bg-yellow-100';
  };

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="relative p-2 rounded-lg hover:bg-primary-50 transition-colors text-muted hover:text-foreground"
        title="Contas a vencer"
      >
        <Bell className="w-5 h-5" />
        {upcoming.length > 0 && (
          <span className="absolute -top-0.5 -right-0.5 w-4 h-4 bg-red-500 text-white text-[10px] font-bold rounded-full flex items-center justify-center">
            {upcoming.length > 9 ? '9+' : upcoming.length}
          </span>
        )}
      </button>

      {isOpen && (
        <div className="absolute right-0 mt-2 w-[calc(100vw-24px)] sm:w-80 max-w-[320px] -right-2 sm:right-0 bg-card rounded-xl shadow-lg border border-border z-50 overflow-hidden animate-fadeIn">
          <div className="flex items-center justify-between p-3 border-b border-border bg-primary-50/50">
            <div className="flex items-center gap-2">
              <AlertCircle className="w-4 h-4 text-warning" />
              <h3 className="font-semibold text-foreground text-sm">Contas a Vencer</h3>
            </div>
            <button
              onClick={() => setIsOpen(false)}
              className="p-1 rounded hover:bg-primary-100 text-muted transition-colors"
            >
              <X className="w-4 h-4" />
            </button>
          </div>

          <div className="max-h-80 overflow-y-auto">
            {loading ? (
              <div className="p-4 text-center text-muted text-sm">
                Carregando...
              </div>
            ) : upcoming.length === 0 ? (
              <div className="p-6 text-center">
                <div className="inline-flex p-3 bg-success-light rounded-full mb-2">
                  <Calendar className="w-5 h-5 text-success" />
                </div>
                <p className="text-muted text-sm">Nenhuma conta a vencer nos proximos 7 dias</p>
              </div>
            ) : (
              <div className="divide-y divide-border">
                {upcoming.map((t) => (
                  <div key={t.id} className="p-3 hover:bg-primary-50/50 transition-colors">
                    <div className="flex items-start justify-between gap-2">
                      <div className="flex-1 min-w-0">
                        <p className="font-medium text-foreground text-sm truncate">
                          {t.description || 'Sem descricao'}
                        </p>
                        <p className="text-xs text-muted mt-0.5">
                          {t.category} - {t.dueDay}/{t.month}/{t.year}
                        </p>
                      </div>
                      <p className="font-semibold text-foreground text-sm shrink-0">
                        {formatCurrency(t.amount)}
                      </p>
                    </div>
                    <span className={`inline-block mt-1.5 text-[10px] font-medium px-2 py-0.5 rounded-full ${getDueDayColor(t.daysUntilDue)}`}>
                      {getDueDayLabel(t.daysUntilDue)}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>

          {upcoming.length > 0 && (
            <div className="p-2 border-t border-border bg-primary-50/30">
              <p className="text-[10px] text-muted text-center">
                {upcoming.length} conta{upcoming.length > 1 ? 's' : ''} a vencer nos proximos 7 dias
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
};
