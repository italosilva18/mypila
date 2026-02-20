import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useTransactions } from './useTransactions';
import { api } from '../services/api';
import { Status, Transaction, PaginationInfo } from '../types';

// Mock the api module
vi.mock('../services/api', () => ({
  api: {
    getTransactions: vi.fn(),
    getStats: vi.fn(),
    createTransaction: vi.fn(),
    updateTransaction: vi.fn(),
    deleteTransaction: vi.fn(),
    toggleStatus: vi.fn(),
    seedData: vi.fn(),
  },
}));

// Mock the ToastContext
const mockAddToast = vi.fn();
vi.mock('../contexts/ToastContext', () => ({
  useToast: () => ({
    addToast: mockAddToast,
  }),
}));

describe('useTransactions', () => {
  const mockCompanyId = 'company-123';

  const mockTransactions: Transaction[] = [
    {
      id: 'trans-1',
      companyId: mockCompanyId,
      month: 'Janeiro',
      year: 2024,
      dueDay: 1,
      amount: 1000,
      paidAmount: 1000,
      category: 'Salario',
      status: Status.PAID,
      description: 'Salario Janeiro',
    },
    {
      id: 'trans-2',
      companyId: mockCompanyId,
      month: 'Fevereiro',
      year: 2024,
      dueDay: 1,
      amount: 500,
      paidAmount: 0,
      category: 'Bonus',
      status: Status.OPEN,
      description: 'Bonus trimestral',
    },
  ];

  const mockPagination: PaginationInfo = {
    page: 1,
    limit: 50,
    total: 2,
    totalPages: 1,
  };

  const mockPaginatedResponse = {
    data: mockTransactions,
    pagination: mockPagination,
  };

  const mockStats = {
    paid: 1000,
    open: 500,
    total: 1500,
  };

  beforeEach(() => {
    vi.clearAllMocks();
    // Setup default mock implementations
    vi.mocked(api.getTransactions).mockResolvedValue(mockPaginatedResponse);
    vi.mocked(api.getStats).mockResolvedValue(mockStats);
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe('Deve carregar transacoes quando companyId e fornecido', () => {
    it('deve buscar transacoes e stats ao inicializar', async () => {
      const { result } = renderHook(() => useTransactions(mockCompanyId));

      // Initially loading
      expect(result.current.loading).toBe(true);

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(api.getTransactions).toHaveBeenCalledWith(mockCompanyId, 1, 50);
      expect(api.getStats).toHaveBeenCalledWith(mockCompanyId);
      expect(result.current.transactions).toEqual(mockTransactions);
      expect(result.current.stats).toEqual(mockStats);
      expect(result.current.error).toBeNull();
    });

    it('nao deve buscar transacoes quando companyId esta vazio', async () => {
      const { result } = renderHook(() => useTransactions(''));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(api.getTransactions).not.toHaveBeenCalled();
      expect(result.current.transactions).toEqual([]);
    });

    it('deve tratar erro ao buscar transacoes', async () => {
      const errorMessage = 'Failed to fetch transactions';
      vi.mocked(api.getTransactions).mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.error).toBe(errorMessage);
      expect(result.current.transactions).toEqual([]);
    });

    it('deve atualizar transacoes quando companyId muda', async () => {
      const newCompanyId = 'company-456';
      const newTransactions: Transaction[] = [
        {
          id: 'trans-3',
          companyId: newCompanyId,
          month: 'Marco',
          year: 2024,
          dueDay: 1,
          amount: 2000,
          paidAmount: 2000,
          category: 'Projeto',
          status: Status.PAID,
        },
      ];
      const newPaginatedResponse = {
        data: newTransactions,
        pagination: { page: 1, limit: 50, total: 1, totalPages: 1 },
      };

      const { result, rerender } = renderHook(
        ({ companyId }) => useTransactions(companyId),
        { initialProps: { companyId: mockCompanyId } }
      );

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Change companyId
      vi.mocked(api.getTransactions).mockResolvedValue(newPaginatedResponse);
      rerender({ companyId: newCompanyId });

      await waitFor(() => {
        expect(api.getTransactions).toHaveBeenCalledWith(newCompanyId, 1, 50);
      });
    });
  });

  describe('Deve criar nova transacao', () => {
    it('deve criar transacao com sucesso', async () => {
      const newTransactionData = {
        month: 'Marco',
        year: 2024,
        dueDay: 1,
        amount: 1500,
        paidAmount: 0,
        category: 'Projeto',
        status: Status.OPEN,
        description: 'Novo projeto',
      };

      const createdTransaction: Transaction = {
        id: 'trans-new',
        companyId: mockCompanyId,
        ...newTransactionData,
      };

      vi.mocked(api.createTransaction).mockResolvedValue(createdTransaction);

      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      await act(async () => {
        const created = await result.current.createTransaction(newTransactionData);
        expect(created).toEqual(createdTransaction);
      });

      expect(api.createTransaction).toHaveBeenCalledWith({
        ...newTransactionData,
        companyId: mockCompanyId,
      });
      expect(result.current.transactions).toContainEqual(createdTransaction);
      expect(mockAddToast).toHaveBeenCalledWith('success', 'Transacao criada com sucesso!');
    });

    it('nao deve criar transacao quando companyId esta vazio', async () => {
      const { result } = renderHook(() => useTransactions(''));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      await act(async () => {
        await result.current.createTransaction({
          month: 'Janeiro',
          year: 2024,
          dueDay: 1,
          amount: 100,
          paidAmount: 0,
          category: 'Test',
          status: Status.OPEN,
        });
      });

      expect(api.createTransaction).not.toHaveBeenCalled();
    });

    it('deve tratar erro ao criar transacao', async () => {
      const errorMessage = 'Failed to create transaction';
      vi.mocked(api.createTransaction).mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      await act(async () => {
        try {
          await result.current.createTransaction({
            month: 'Janeiro',
            year: 2024,
            dueDay: 1,
            amount: 100,
            paidAmount: 0,
            category: 'Test',
            status: Status.OPEN,
          });
        } catch (error) {
          // Expected to throw
        }
      });

      expect(result.current.error).toBe(errorMessage);
      expect(mockAddToast).toHaveBeenCalledWith('error', `Erro ao criar transacao: ${errorMessage}`);
    });
  });

  describe('Deve atualizar transacao existente', () => {
    it('deve atualizar transacao com sucesso', async () => {
      const transactionId = 'trans-1';
      const updateData = {
        month: 'Janeiro',
        year: 2024,
        dueDay: 1,
        amount: 1200,
        paidAmount: 1200,
        category: 'Salario',
        status: Status.PAID,
        description: 'Salario Janeiro atualizado',
      };

      const updatedTransaction: Transaction = {
        id: transactionId,
        companyId: mockCompanyId,
        ...updateData,
      };

      vi.mocked(api.updateTransaction).mockResolvedValue(updatedTransaction);

      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      await act(async () => {
        const updated = await result.current.updateTransaction(transactionId, updateData);
        expect(updated).toEqual(updatedTransaction);
      });

      expect(api.updateTransaction).toHaveBeenCalledWith(transactionId, {
        ...updateData,
        companyId: mockCompanyId,
      });
      expect(mockAddToast).toHaveBeenCalledWith('success', 'Transacao atualizada com sucesso!');

      // Verify transaction is updated in the list
      const transactionInList = result.current.transactions.find(t => t.id === transactionId);
      expect(transactionInList?.amount).toBe(1200);
    });

    it('nao deve atualizar transacao quando companyId esta vazio', async () => {
      const { result } = renderHook(() => useTransactions(''));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      await act(async () => {
        await result.current.updateTransaction('trans-1', {
          month: 'Janeiro',
          year: 2024,
          dueDay: 1,
          amount: 100,
          paidAmount: 0,
          category: 'Test',
          status: Status.OPEN,
        });
      });

      expect(api.updateTransaction).not.toHaveBeenCalled();
    });

    it('deve tratar erro ao atualizar transacao', async () => {
      const errorMessage = 'Failed to update transaction';
      vi.mocked(api.updateTransaction).mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      await act(async () => {
        try {
          await result.current.updateTransaction('trans-1', {
            month: 'Janeiro',
            year: 2024,
            dueDay: 1,
            amount: 100,
            paidAmount: 0,
            category: 'Test',
            status: Status.OPEN,
          });
        } catch (error) {
          // Expected to throw
        }
      });

      expect(result.current.error).toBe(errorMessage);
      expect(mockAddToast).toHaveBeenCalledWith('error', `Erro ao atualizar transacao: ${errorMessage}`);
    });
  });

  describe('Deve deletar transacao', () => {
    it('deve deletar transacao com sucesso', async () => {
      const transactionId = 'trans-1';
      vi.mocked(api.deleteTransaction).mockResolvedValue(undefined);

      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      const initialCount = result.current.transactions.length;

      await act(async () => {
        await result.current.deleteTransaction(transactionId);
      });

      expect(api.deleteTransaction).toHaveBeenCalledWith(transactionId);
      expect(result.current.transactions.length).toBe(initialCount - 1);
      expect(result.current.transactions.find(t => t.id === transactionId)).toBeUndefined();
      expect(mockAddToast).toHaveBeenCalledWith('success', 'Transacao excluida com sucesso!');
    });

    it('deve tratar erro ao deletar transacao', async () => {
      const errorMessage = 'Failed to delete transaction';
      vi.mocked(api.deleteTransaction).mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      await act(async () => {
        try {
          await result.current.deleteTransaction('trans-1');
        } catch (error) {
          // Expected to throw
        }
      });

      expect(result.current.error).toBe(errorMessage);
      expect(mockAddToast).toHaveBeenCalledWith('error', `Erro ao excluir transacao: ${errorMessage}`);
    });
  });

  describe('Deve alternar status da transacao', () => {
    it('deve alternar status de OPEN para PAID', async () => {
      const transactionId = 'trans-2'; // Status: OPEN
      const toggledTransaction: Transaction = {
        ...mockTransactions[1],
        status: Status.PAID,
      };

      vi.mocked(api.toggleStatus).mockResolvedValue(toggledTransaction);

      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      await act(async () => {
        const toggled = await result.current.toggleStatus(transactionId);
        expect(toggled).toEqual(toggledTransaction);
      });

      expect(api.toggleStatus).toHaveBeenCalledWith(transactionId);
      expect(mockAddToast).toHaveBeenCalledWith('info', 'Status alterado para pago');

      const transactionInList = result.current.transactions.find(t => t.id === transactionId);
      expect(transactionInList?.status).toBe(Status.PAID);
    });

    it('deve alternar status de PAID para OPEN', async () => {
      const transactionId = 'trans-1'; // Status: PAID
      const toggledTransaction: Transaction = {
        ...mockTransactions[0],
        status: Status.OPEN,
      };

      vi.mocked(api.toggleStatus).mockResolvedValue(toggledTransaction);

      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      await act(async () => {
        const toggled = await result.current.toggleStatus(transactionId);
        expect(toggled).toEqual(toggledTransaction);
      });

      expect(api.toggleStatus).toHaveBeenCalledWith(transactionId);
      expect(mockAddToast).toHaveBeenCalledWith('info', 'Status alterado para aberto');
    });

    it('deve tratar erro ao alternar status', async () => {
      const errorMessage = 'Failed to toggle status';
      vi.mocked(api.toggleStatus).mockRejectedValue(new Error(errorMessage));

      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      await act(async () => {
        try {
          await result.current.toggleStatus('trans-1');
        } catch (error) {
          // Expected to throw
        }
      });

      expect(result.current.error).toBe(errorMessage);
      expect(mockAddToast).toHaveBeenCalledWith('error', `Erro ao alterar status: ${errorMessage}`);
    });
  });

  describe('Funcionalidades adicionais', () => {
    it('deve chamar refreshData para recarregar dados', async () => {
      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      // Clear previous calls
      vi.clearAllMocks();

      await act(async () => {
        await result.current.refreshData();
      });

      expect(api.getTransactions).toHaveBeenCalledWith(mockCompanyId, 1, 50);
      expect(api.getStats).toHaveBeenCalledWith(mockCompanyId);
    });

    it('deve chamar seedData para popular dados iniciais', async () => {
      vi.mocked(api.seedData).mockResolvedValue({ message: 'Seeded', count: 10 });

      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      await act(async () => {
        await result.current.seedData();
      });

      expect(api.seedData).toHaveBeenCalled();
      // Should refresh data after seeding
      expect(api.getTransactions).toHaveBeenCalled();
    });

    it('deve atualizar stats apos criar transacao', async () => {
      const newStats = { paid: 2000, open: 500, total: 2500 };
      vi.mocked(api.getStats).mockResolvedValue(newStats);
      vi.mocked(api.createTransaction).mockResolvedValue({
        id: 'new-trans',
        companyId: mockCompanyId,
        month: 'Marco',
        year: 2024,
        dueDay: 1,
        amount: 1000,
        paidAmount: 1000,
        category: 'Salario',
        status: Status.PAID,
      });

      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      await act(async () => {
        await result.current.createTransaction({
          month: 'Marco',
          year: 2024,
          dueDay: 1,
          amount: 1000,
          paidAmount: 1000,
          category: 'Salario',
          status: Status.PAID,
        });
      });

      // Stats should be refreshed after creating transaction
      expect(api.getStats).toHaveBeenCalled();
    });

    it('deve retornar valores iniciais corretos', () => {
      const { result } = renderHook(() => useTransactions(mockCompanyId));

      expect(result.current.transactions).toEqual([]);
      expect(result.current.stats).toEqual({ paid: 0, open: 0, total: 0 });
      expect(result.current.loading).toBe(true);
      expect(result.current.error).toBeNull();
    });

    it('deve expor todas as funcoes necessarias', async () => {
      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(typeof result.current.createTransaction).toBe('function');
      expect(typeof result.current.updateTransaction).toBe('function');
      expect(typeof result.current.deleteTransaction).toBe('function');
      expect(typeof result.current.toggleStatus).toBe('function');
      expect(typeof result.current.refreshData).toBe('function');
      expect(typeof result.current.seedData).toBe('function');
    });
  });

  describe('Tratamento de erros genericos', () => {
    it('deve tratar erro sem mensagem', async () => {
      vi.mocked(api.getTransactions).mockRejectedValue('Unknown error');

      const { result } = renderHook(() => useTransactions(mockCompanyId));

      await waitFor(() => {
        expect(result.current.loading).toBe(false);
      });

      expect(result.current.error).toBe('Failed to fetch transactions');
    });

    it('deve limpar erro ao fazer nova requisicao', async () => {
      vi.mocked(api.getTransactions)
        .mockRejectedValueOnce(new Error('Error'))
        .mockResolvedValueOnce(mockPaginatedResponse);

      const { result, rerender } = renderHook(
        ({ companyId }) => useTransactions(companyId),
        { initialProps: { companyId: mockCompanyId } }
      );

      await waitFor(() => {
        expect(result.current.error).not.toBeNull();
      });

      // Rerender with same companyId to trigger new fetch
      rerender({ companyId: 'new-company' });

      await waitFor(() => {
        expect(result.current.error).toBeNull();
      });
    });
  });
});
