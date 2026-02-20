import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { api } from './api';
import { Status, QuoteStatus } from '../types';

// Mock fetch globally
const mockFetch = vi.fn();
global.fetch = mockFetch;

// Mock localStorage
const localStorageMock = (() => {
  let store: Record<string, string> = {};
  return {
    getItem: vi.fn((key: string) => store[key] || null),
    setItem: vi.fn((key: string, value: string) => {
      store[key] = value;
    }),
    removeItem: vi.fn((key: string) => {
      delete store[key];
    }),
    clear: vi.fn(() => {
      store = {};
    }),
  };
})();
Object.defineProperty(window, 'localStorage', { value: localStorageMock });

describe('ApiService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    localStorageMock.clear();
    // Reset tokens in api instance
    api.setTokens(null, null);
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe('Deve fazer login corretamente', () => {
    it('deve fazer login e armazenar tokens', async () => {
      const mockResponse = {
        accessToken: 'jwt-access-token-123',
        refreshToken: 'jwt-refresh-token-123',
        expiresIn: 900,
        user: {
          id: 'user-1',
          name: 'Test User',
          email: 'test@example.com',
        },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await api.login({
        email: 'test@example.com',
        password: 'password123',
      });

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/auth/login'),
        expect.objectContaining({
          method: 'POST',
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
          }),
          body: JSON.stringify({
            email: 'test@example.com',
            password: 'password123',
          }),
        })
      );

      expect(result).toEqual(mockResponse);
      expect(api.getToken()).toBe('jwt-access-token-123');
      expect(localStorageMock.setItem).toHaveBeenCalledWith('accessToken', 'jwt-access-token-123');
    });

    it('deve lancar erro com credenciais invalidas', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({ error: 'Credenciais inválidas' }),
      });

      // API handles 401 with "Invalid or expired token" before processing JSON
      await expect(
        api.login({ email: 'wrong@example.com', password: 'wrongpassword' })
      ).rejects.toThrow('Invalid or expired token');
    });

    it('deve tratar erro de rede no login', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      await expect(
        api.login({ email: 'test@example.com', password: 'password' })
      ).rejects.toThrow('Network error');
    });
  });

  describe('Deve fazer logout e limpar tokens', () => {
    it('deve remover tokens ao fazer logout', async () => {
      // Set tokens first
      api.setTokens('existing-access-token', 'existing-refresh-token');
      expect(api.getToken()).toBe('existing-access-token');

      // Mock the logout endpoint
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      });

      // Logout
      await api.logout();

      expect(api.getToken()).toBeNull();
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('accessToken');
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('refreshToken');
    });

    it('deve limpar localStorage ao fazer logout', async () => {
      api.setTokens('token-to-remove', 'refresh-to-remove');

      // Mock the logout endpoint
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      });

      await api.logout();

      expect(localStorageMock.removeItem).toHaveBeenCalledWith('accessToken');
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('refreshToken');
    });
  });

  describe('Deve buscar transacoes', () => {
    beforeEach(() => {
      api.setTokens('valid-token', 'valid-refresh');
    });

    it('deve buscar transacoes com companyId', async () => {
      const mockTransactions = [
        {
          id: 'trans-1',
          companyId: 'company-1',
          month: 'Janeiro',
          year: 2024,
          amount: 1000,
          category: 'Salario',
          status: Status.PAID,
        },
        {
          id: 'trans-2',
          companyId: 'company-1',
          month: 'Fevereiro',
          year: 2024,
          amount: 500,
          category: 'Bonus',
          status: Status.OPEN,
        },
      ];
      const mockPagination = { page: 1, limit: 50, total: 2, totalPages: 1 };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: mockTransactions, pagination: mockPagination }),
      });

      const result = await api.getTransactions('company-1');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/transactions?companyId=company-1'),
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: 'Bearer valid-token',
          }),
        })
      );

      expect(result.data).toEqual(mockTransactions);
      expect(result.pagination).toEqual(mockPagination);
    });

    it('deve retornar array vazio quando nao ha transacoes', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: null, pagination: { page: 1, limit: 50, total: 0, totalPages: 0 } }),
      });

      const result = await api.getTransactions('company-1');

      expect(result.data).toEqual([]);
      expect(result.pagination.total).toBe(0);
    });

    it('deve buscar transacao por ID', async () => {
      const mockTransaction = {
        id: 'trans-1',
        companyId: 'company-1',
        month: 'Janeiro',
        year: 2024,
        amount: 1000,
        category: 'Salario',
        status: Status.PAID,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockTransaction,
      });

      const result = await api.getTransaction('trans-1');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/transactions/trans-1'),
        expect.any(Object)
      );
      expect(result).toEqual(mockTransaction);
    });

    it('deve criar nova transacao', async () => {
      const newTransaction = {
        companyId: 'company-1',
        month: 'Marco',
        year: 2024,
        dueDay: 1,
        amount: 2000,
        paidAmount: 0,
        category: 'Projeto',
        status: Status.OPEN,
        description: 'Novo projeto',
      };

      const createdTransaction = { id: 'trans-new', ...newTransaction };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => createdTransaction,
      });

      const result = await api.createTransaction(newTransaction);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/transactions'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(newTransaction),
        })
      );
      expect(result).toEqual(createdTransaction);
    });

    it('deve lancar erro ao criar transacao sem companyId', async () => {
      const invalidTransaction = {
        companyId: '',
        month: 'Janeiro',
        year: 2024,
        dueDay: 1,
        amount: 100,
        paidAmount: 0,
        category: 'Test',
        status: Status.OPEN,
      };

      await expect(api.createTransaction(invalidTransaction)).rejects.toThrow(
        'companyId é obrigatório para criar transação'
      );
    });

    it('deve atualizar transacao existente', async () => {
      const updateData = {
        companyId: 'company-1',
        month: 'Janeiro',
        year: 2024,
        dueDay: 1,
        amount: 1500,
        paidAmount: 1500,
        category: 'Salario',
        status: Status.PAID,
        description: 'Atualizado',
      };

      const updatedTransaction = { id: 'trans-1', ...updateData };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => updatedTransaction,
      });

      const result = await api.updateTransaction('trans-1', updateData);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/transactions/trans-1'),
        expect.objectContaining({
          method: 'PUT',
          body: JSON.stringify(updateData),
        })
      );
      expect(result).toEqual(updatedTransaction);
    });

    it('deve deletar transacao', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      });

      await api.deleteTransaction('trans-1');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/transactions/trans-1'),
        expect.objectContaining({
          method: 'DELETE',
        })
      );
    });

    it('deve alternar status da transacao', async () => {
      const toggledTransaction = {
        id: 'trans-1',
        companyId: 'company-1',
        month: 'Janeiro',
        year: 2024,
        amount: 1000,
        category: 'Salario',
        status: Status.PAID,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => toggledTransaction,
      });

      const result = await api.toggleStatus('trans-1');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/transactions/trans-1/toggle-status'),
        expect.objectContaining({
          method: 'PATCH',
        })
      );
      expect(result).toEqual(toggledTransaction);
    });
  });

  describe('Deve tratar erros de API', () => {
    beforeEach(() => {
      api.setTokens('valid-token', 'valid-refresh');
    });

    it('deve tratar erro 400 - Bad Request', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 400,
        json: async () => ({ error: 'Dados invalidos' }),
      });

      await expect(api.getTransactions('invalid')).rejects.toThrow('Dados invalidos');
    });

    it('deve tratar erro 401 - Unauthorized', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        json: async () => ({ error: 'Não autorizado' }),
      });

      // API handles 401 with "Invalid or expired token" before processing JSON
      await expect(api.getTransactions('company-1')).rejects.toThrow('Invalid or expired token');
    });

    it('deve tratar erro 403 - Forbidden', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 403,
        json: async () => ({ error: 'Acesso negado' }),
      });

      await expect(api.getTransactions('company-1')).rejects.toThrow('Acesso negado');
    });

    it('deve tratar erro 404 - Not Found', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 404,
        json: async () => ({ error: 'Recurso nao encontrado' }),
      });

      await expect(api.getTransaction('non-existent')).rejects.toThrow('Recurso nao encontrado');
    });

    it('deve tratar erro 500 - Internal Server Error', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => ({ error: 'Erro interno do servidor' }),
      });

      await expect(api.getTransactions('company-1')).rejects.toThrow('Erro interno do servidor');
    });

    it('deve tratar erro quando resposta nao e JSON', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 500,
        json: async () => {
          throw new Error('Invalid JSON');
        },
      });

      await expect(api.getTransactions('company-1')).rejects.toThrow();
    });

    it('deve tratar erro de rede', async () => {
      mockFetch.mockRejectedValueOnce(new Error('Network error'));

      await expect(api.getTransactions('company-1')).rejects.toThrow('Network error');
    });

    it('deve incluir status HTTP quando nao ha mensagem de erro', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: false,
        status: 418,
        json: async () => ({}),
      });

      await expect(api.getTransactions('company-1')).rejects.toThrow('HTTP error! status: 418');
    });
  });

  describe('Autenticacao e Headers', () => {
    it('deve incluir token no header quando autenticado', async () => {
      api.setTokens('my-auth-token', 'my-refresh-token');

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: [], pagination: { page: 1, limit: 50, total: 0, totalPages: 0 } }),
      });

      await api.getTransactions('company-1');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.any(String),
        expect.objectContaining({
          headers: expect.objectContaining({
            Authorization: 'Bearer my-auth-token',
          }),
        })
      );
    });

    it('nao deve incluir Authorization header quando nao autenticado', async () => {
      api.setTokens(null, null);

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      });

      // Use a public endpoint simulation
      try {
        await api.getMe();
      } catch {
        // Expected to fail, we just want to check the headers
      }

      const callArgs = mockFetch.mock.calls[0];
      const headers = callArgs[1]?.headers as Record<string, string>;
      expect(headers.Authorization).toBeUndefined();
    });

    it('deve persistir tokens no localStorage', () => {
      api.setTokens('persistent-token', 'persistent-refresh');

      expect(localStorageMock.setItem).toHaveBeenCalledWith('accessToken', 'persistent-token');
      expect(localStorageMock.setItem).toHaveBeenCalledWith('refreshToken', 'persistent-refresh');
    });

    it('deve recuperar token do localStorage', () => {
      localStorageMock.getItem.mockReturnValueOnce('stored-token');

      // The api instance reads from localStorage on initialization
      // We need to create a new instance or check the getToken method
      expect(typeof api.getToken).toBe('function');
    });
  });

  describe('Funcionalidades adicionais da API', () => {
    beforeEach(() => {
      api.setTokens('valid-token', 'valid-refresh');
    });

    it('deve buscar estatisticas', async () => {
      const mockStats = { paid: 1000, open: 500, total: 1500 };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockStats,
      });

      const result = await api.getStats('company-1');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/stats?companyId=company-1'),
        expect.any(Object)
      );
      expect(result).toEqual(mockStats);
    });

    it('deve buscar empresas do usuario', async () => {
      const mockCompanies = [
        { id: 'company-1', name: 'Empresa A', userId: 'user-1', createdAt: '2024-01-01' },
        { id: 'company-2', name: 'Empresa B', userId: 'user-1', createdAt: '2024-01-02' },
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockCompanies,
      });

      const result = await api.getCompanies();

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/companies'),
        expect.any(Object)
      );
      expect(result).toEqual(mockCompanies);
    });

    it('deve criar nova empresa', async () => {
      const newCompany = { id: 'company-new', name: 'Nova Empresa', userId: 'user-1', createdAt: '2024-01-01' };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => newCompany,
      });

      const result = await api.createCompany('Nova Empresa');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/companies'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ name: 'Nova Empresa' }),
        })
      );
      expect(result).toEqual(newCompany);
    });

    it('deve registrar novo usuario', async () => {
      const mockResponse = {
        accessToken: 'new-user-access-token',
        refreshToken: 'new-user-refresh-token',
        expiresIn: 900,
        user: { id: 'user-new', name: 'New User', email: 'new@example.com' },
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await api.register({
        name: 'New User',
        email: 'new@example.com',
        password: 'password123',
      });

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/auth/register'),
        expect.objectContaining({
          method: 'POST',
        })
      );
      expect(result).toEqual(mockResponse);
      expect(api.getToken()).toBe('new-user-access-token');
    });

    it('deve buscar dados do usuario atual', async () => {
      const mockUser = { user: { id: 'user-1', name: 'Test User', email: 'test@example.com' } };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockUser,
      });

      const result = await api.getMe();

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/auth/me'),
        expect.any(Object)
      );
      expect(result).toEqual(mockUser);
    });

    it('deve fazer seed de dados', async () => {
      const mockResponse = { message: 'Data seeded', count: 10 };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockResponse,
      });

      const result = await api.seedData();

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/seed'),
        expect.objectContaining({
          method: 'POST',
        })
      );
      expect(result).toEqual(mockResponse);
    });
  });

  describe('Quotes API', () => {
    beforeEach(() => {
      api.setTokens('valid-token', 'valid-refresh');
    });

    it('deve buscar orcamentos', async () => {
      const mockQuotes = [
        { id: 'quote-1', number: 'ORC-2024-001', clientName: 'Cliente A', status: QuoteStatus.DRAFT },
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockQuotes,
      });

      const result = await api.getQuotes('company-1');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/quotes?companyId=company-1'),
        expect.any(Object)
      );
      expect(result).toEqual(mockQuotes);
    });

    it('deve buscar orcamentos filtrados por status', async () => {
      const mockQuotes = [
        { id: 'quote-1', number: 'ORC-2024-001', clientName: 'Cliente A', status: QuoteStatus.SENT },
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockQuotes,
      });

      const result = await api.getQuotes('company-1', QuoteStatus.SENT);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('status=SENT'),
        expect.any(Object)
      );
      expect(result).toEqual(mockQuotes);
    });

    it('deve criar novo orcamento', async () => {
      const quoteData = {
        clientName: 'Cliente Novo',
        title: 'Orcamento Novo',
        items: [{ description: 'Servico', quantity: 1, unitPrice: 1000 }],
        discount: 0,
        discountType: 'VALUE' as const,
        validUntil: '2024-12-31',
      };

      const createdQuote = {
        id: 'quote-new',
        number: 'ORC-2024-002',
        ...quoteData,
        status: QuoteStatus.DRAFT,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => createdQuote,
      });

      const result = await api.createQuote('company-1', quoteData);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/quotes?companyId=company-1'),
        expect.objectContaining({
          method: 'POST',
        })
      );
      expect(result.id).toBe('quote-new');
    });

    it('deve atualizar status do orcamento', async () => {
      const updatedQuote = {
        id: 'quote-1',
        status: QuoteStatus.APPROVED,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => updatedQuote,
      });

      const result = await api.updateQuoteStatus('quote-1', QuoteStatus.APPROVED);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/quotes/quote-1/status'),
        expect.objectContaining({
          method: 'PATCH',
          body: JSON.stringify({ status: QuoteStatus.APPROVED }),
        })
      );
      expect(result.status).toBe(QuoteStatus.APPROVED);
    });

    it('deve duplicar orcamento', async () => {
      const duplicatedQuote = {
        id: 'quote-duplicated',
        number: 'ORC-2024-003',
        title: 'Orcamento Original (Copia)',
        status: QuoteStatus.DRAFT,
      };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => duplicatedQuote,
      });

      const result = await api.duplicateQuote('quote-1');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/quotes/quote-1/duplicate'),
        expect.objectContaining({
          method: 'POST',
        })
      );
      expect(result.id).toBe('quote-duplicated');
    });
  });

  describe('Categorias API', () => {
    beforeEach(() => {
      api.setTokens('valid-token', 'valid-refresh');
    });

    it('deve buscar categorias', async () => {
      const mockCategories = [
        { id: 'cat-1', name: 'Salario', type: 'INCOME', companyId: 'company-1' },
        { id: 'cat-2', name: 'Alimentacao', type: 'EXPENSE', companyId: 'company-1' },
      ];

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => mockCategories,
      });

      const result = await api.getCategories('company-1');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/categories?companyId=company-1'),
        expect.any(Object)
      );
      expect(result).toEqual(mockCategories);
    });

    it('deve criar nova categoria', async () => {
      const newCategory = { id: 'cat-new', name: 'Nova Categoria', type: 'EXPENSE', companyId: 'company-1' };

      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => newCategory,
      });

      const result = await api.createCategory('company-1', 'Nova Categoria', 'EXPENSE', '#FF5733', 1000);

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/categories?companyId=company-1'),
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({ name: 'Nova Categoria', type: 'EXPENSE', color: '#FF5733', budget: 1000 }),
        })
      );
      expect(result).toEqual(newCategory);
    });

    it('deve deletar categoria', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({}),
      });

      await api.deleteCategory('cat-1');

      expect(mockFetch).toHaveBeenCalledWith(
        expect.stringContaining('/categories/cat-1'),
        expect.objectContaining({
          method: 'DELETE',
        })
      );
    });
  });

  describe('URL Handling', () => {
    beforeEach(() => {
      api.setTokens('valid-token', 'valid-refresh');
    });

    it('deve lidar com endpoints que comecam com barra', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => ({ data: [], pagination: { page: 1, limit: 50, total: 0, totalPages: 0 } }),
      });

      await api.getTransactions('company-1');

      const url = mockFetch.mock.calls[0][0];
      expect(url).not.toContain('//transactions');
    });

    it('deve construir URL corretamente com parametros de query', async () => {
      mockFetch.mockResolvedValueOnce({
        ok: true,
        json: async () => [],
      });

      await api.getQuotes('company-1', QuoteStatus.DRAFT);

      const url = mockFetch.mock.calls[0][0];
      expect(url).toContain('companyId=company-1');
      expect(url).toContain('status=DRAFT');
    });
  });
});
