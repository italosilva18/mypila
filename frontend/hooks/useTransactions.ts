import { useState, useEffect, useCallback } from 'react';
import { Transaction, Status, PaginationInfo } from '../types';
import { api } from '../services/api';
import { useToast } from '../contexts/ToastContext';

interface Stats {
  paid: number;
  open: number;
  total: number;
}

const DEFAULT_PAGE_SIZE = 50;

export function useTransactions(companyId: string) {
  const { addToast } = useToast();
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [stats, setStats] = useState<Stats>({ paid: 0, open: 0, total: 0 });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pagination, setPagination] = useState<PaginationInfo>({
    page: 1,
    limit: DEFAULT_PAGE_SIZE,
    total: 0,
    totalPages: 0
  });

  const fetchTransactions = useCallback(async (page: number = 1, limit: number = DEFAULT_PAGE_SIZE) => {
    if (!companyId) {
      setLoading(false);
      return;
    }
    try {
      setLoading(true);
      setError(null);
      const response = await api.getTransactions(companyId, page, limit);
      setTransactions(response.data);
      setPagination(response.pagination);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch transactions');
    } finally {
      setLoading(false);
    }
  }, [companyId]);

  const fetchStats = useCallback(async () => {
    if (!companyId) return;
    try {
      const data = await api.getStats(companyId);
      setStats(data);
    } catch (err) {
      console.error('Failed to fetch stats:', err);
    }
  }, [companyId]);

  const refreshData = useCallback(async () => {
    await Promise.all([fetchTransactions(pagination.page, pagination.limit), fetchStats()]);
  }, [fetchTransactions, fetchStats, pagination.page, pagination.limit]);

  const goToPage = useCallback(async (page: number) => {
    await fetchTransactions(page, pagination.limit);
  }, [fetchTransactions, pagination.limit]);

  const nextPage = useCallback(async () => {
    if (pagination.page < pagination.totalPages) {
      await goToPage(pagination.page + 1);
    }
  }, [goToPage, pagination.page, pagination.totalPages]);

  const prevPage = useCallback(async () => {
    if (pagination.page > 1) {
      await goToPage(pagination.page - 1);
    }
  }, [goToPage, pagination.page]);

  const setPageSize = useCallback(async (limit: number) => {
    await fetchTransactions(1, limit);
  }, [fetchTransactions]);

  useEffect(() => {
    fetchTransactions(1, DEFAULT_PAGE_SIZE);
    fetchStats();
  }, [companyId]); // eslint-disable-line react-hooks/exhaustive-deps

  const createTransaction = useCallback(async (data: Omit<Transaction, 'id' | 'companyId'>) => {
    if (!companyId) return;
    try {
      setError(null);
      const transactionData = { ...data, companyId };
      const newTransaction = await api.createTransaction(transactionData);
      setTransactions(prev => [...prev, newTransaction]);
      await fetchStats();
      addToast('success', 'Transacao criada com sucesso!');
      return newTransaction;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to create transaction';
      setError(message);
      addToast('error', `Erro ao criar transacao: ${message}`);
      throw err;
    }
  }, [companyId, fetchStats, addToast]);

  const updateTransaction = useCallback(async (id: string, data: Omit<Transaction, 'id' | 'companyId'>) => {
    if (!companyId) return;
    try {
      setError(null);
      const transactionData = { ...data, companyId };
      const updated = await api.updateTransaction(id, transactionData);
      setTransactions(prev => prev.map(t => t.id === id ? updated : t));
      await fetchStats();
      addToast('success', 'Transacao atualizada com sucesso!');
      return updated;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to update transaction';
      setError(message);
      addToast('error', `Erro ao atualizar transacao: ${message}`);
      throw err;
    }
  }, [companyId, fetchStats, addToast]);

  const deleteTransaction = useCallback(async (id: string) => {
    try {
      setError(null);
      await api.deleteTransaction(id);
      setTransactions(prev => prev.filter(t => t.id !== id));
      await fetchStats();
      addToast('success', 'Transacao excluida com sucesso!');
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to delete transaction';
      setError(message);
      addToast('error', `Erro ao excluir transacao: ${message}`);
      throw err;
    }
  }, [fetchStats, addToast]);

  const toggleStatus = useCallback(async (id: string) => {
    try {
      setError(null);
      const updated = await api.toggleStatus(id);
      setTransactions(prev => prev.map(t => t.id === id ? updated : t));
      await fetchStats();
      addToast('info', `Status alterado para ${updated.status === Status.PAID ? 'pago' : 'aberto'}`);
      return updated;
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to toggle status';
      setError(message);
      addToast('error', `Erro ao alterar status: ${message}`);
      throw err;
    }
  }, [fetchStats, addToast]);

  const seedData = async () => {
    try {
      setError(null);
      await api.seedData();
      await refreshData();
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to seed data';
      setError(message);
      throw err;
    }
  };

  return {
    transactions,
    stats,
    loading,
    error,
    pagination,
    createTransaction,
    updateTransaction,
    deleteTransaction,
    toggleStatus,
    refreshData,
    seedData,
    goToPage,
    nextPage,
    prevPage,
    setPageSize,
  };
}
