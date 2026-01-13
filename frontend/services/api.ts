import { Transaction, Company, AuthResponse, LoginRequest, RegisterRequest, Category, RecurringTransaction, Quote, CreateQuoteRequest, QuoteTemplate, CreateQuoteTemplateRequest, QuoteComparison, QuoteStatus, PaginationInfo, UpcomingResponse, CNPJData, UpdateCompanyRequest } from '../types';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081/api';

// Default timeout in milliseconds (30 seconds)
const DEFAULT_TIMEOUT = 30000;

// Timeout configuration for different operation types
const TIMEOUT_CONFIG = {
  default: DEFAULT_TIMEOUT,
  upload: 120000,    // 2 minutes for uploads
  download: 120000,  // 2 minutes for downloads
  auth: 15000,       // 15 seconds for auth operations
};

class ApiService {
  private accessToken: string | null = localStorage.getItem('accessToken');
  private refreshToken: string | null = localStorage.getItem('refreshToken');
  private isRefreshing = false;
  private refreshPromise: Promise<boolean> | null = null;
  private timeout: number = DEFAULT_TIMEOUT;

  setRequestTimeout(ms: number) {
    this.timeout = ms;
  }

  getRequestTimeout(): number {
    return this.timeout;
  }

  setTokens(accessToken: string | null, refreshToken: string | null) {
    this.accessToken = accessToken;
    this.refreshToken = refreshToken;
    if (accessToken) {
      localStorage.setItem('accessToken', accessToken);
    } else {
      localStorage.removeItem('accessToken');
    }
    if (refreshToken) {
      localStorage.setItem('refreshToken', refreshToken);
    } else {
      localStorage.removeItem('refreshToken');
    }
  }

  getToken(): string | null {
    return this.accessToken;
  }

  private async tryRefreshToken(): Promise<boolean> {
    if (!this.refreshToken) {
      return false;
    }

    // If already refreshing, wait for the existing promise
    if (this.isRefreshing && this.refreshPromise) {
      return this.refreshPromise;
    }

    this.isRefreshing = true;
    this.refreshPromise = (async () => {
      // Setup timeout with AbortController for token refresh
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), TIMEOUT_CONFIG.auth);

      try {
        const response = await fetch(`${API_URL}/auth/refresh`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refreshToken: this.refreshToken }),
          signal: controller.signal,
        });

        if (!response.ok) {
          this.setTokens(null, null);
          return false;
        }

        const data = await response.json();
        this.setTokens(data.accessToken, data.refreshToken);
        return true;
      } catch {
        this.setTokens(null, null);
        return false;
      } finally {
        clearTimeout(timeoutId);
        this.isRefreshing = false;
        this.refreshPromise = null;
      }
    })();

    return this.refreshPromise;
  }

  private async request<T>(
    endpoint: string,
    options?: RequestInit & { timeout?: number },
    retryOnUnauthorized = true
  ): Promise<T> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };

    if (this.accessToken) {
      headers['Authorization'] = `Bearer ${this.accessToken}`;
    }

    // Ensure we don't have double slashes if endpoint starts with /
    const cleanEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;

    // Setup timeout with AbortController
    const controller = new AbortController();
    const timeoutMs = options?.timeout ?? this.timeout;
    const timeoutId = setTimeout(() => controller.abort(), timeoutMs);

    try {
      const response = await fetch(`${API_URL}${cleanEndpoint}`, {
        headers: { ...headers, ...options?.headers },
        ...options,
        signal: controller.signal,
      });

      if (!response.ok) {
        // Handle 401 Unauthorized - try to refresh token
        if (response.status === 401 && retryOnUnauthorized) {
          const refreshed = await this.tryRefreshToken();
          if (refreshed) {
            // Retry the request with new token
            return this.request<T>(endpoint, options, false);
          }
          // Refresh failed, redirect to login
          window.location.href = '/login';
          throw new Error('Session expired');
        }

        const error = await response.json().catch(() => ({ error: 'Unknown error' }));
        throw new Error(error.error?.message || error.error || `HTTP error! status: ${response.status}`);
      }

      return response.json();
    } catch (error) {
      if (error instanceof Error && error.name === 'AbortError') {
        throw new Error(`Request timeout after ${timeoutMs}ms`);
      }
      throw error;
    } finally {
      clearTimeout(timeoutId);
    }
  }

  // Auth
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
      timeout: TIMEOUT_CONFIG.auth,
    });
    this.setTokens(response.accessToken, response.refreshToken);
    return response;
  }

  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
      timeout: TIMEOUT_CONFIG.auth,
    });
    this.setTokens(response.accessToken, response.refreshToken);
    return response;
  }

  async getMe(): Promise<{ user: { id: string; name: string; email: string } }> {
    return this.request<{ user: { id: string; name: string; email: string } }>('/auth/me');
  }

  async forgotPassword(email: string): Promise<{ message: string }> {
    return this.request<{ message: string }>('/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify({ email }),
      timeout: TIMEOUT_CONFIG.auth,
    });
  }

  async resetPassword(token: string, newPassword: string): Promise<{ message: string }> {
    return this.request<{ message: string }>('/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ token, newPassword }),
      timeout: TIMEOUT_CONFIG.auth,
    });
  }

  // Categories
  async getCategories(companyId: string): Promise<Category[]> {
    return this.request<Category[]>(`/categories?companyId=${companyId}`);
  }

  async createCategory(companyId: string, name: string, type: string = 'EXPENSE', color?: string, budget?: number): Promise<Category> {
    return this.request<Category>(`/categories?companyId=${companyId}`, {
      method: 'POST',
      body: JSON.stringify({ name, type, color, budget })
    });
  }

  async updateCategory(id: string, data: Partial<Category>): Promise<Category> {
    return this.request<Category>(`/categories/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data)
    });
  }

  async deleteCategory(id: string): Promise<void> {
    return this.request(`/categories/${id}`, {
      method: 'DELETE'
    });
  }

  async logout(): Promise<void> {
    // Call backend to invalidate refresh token
    if (this.refreshToken) {
      try {
        await fetch(`${API_URL}/auth/logout`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${this.accessToken}`,
          },
          body: JSON.stringify({ refreshToken: this.refreshToken }),
        });
      } catch {
        // Ignore errors - we're logging out anyway
      }
    }
    this.setTokens(null, null);
  }

  // Companies
  async getCompanies(): Promise<Company[]> {
    return this.request<Company[]>('/companies');
  }

  async createCompany(name: string): Promise<Company> {
    return this.request<Company>('/companies', {
      method: 'POST',
      body: JSON.stringify({ name }),
    });
  }

  async updateCompany(id: string, data: UpdateCompanyRequest): Promise<Company> {
    return this.request<Company>(`/companies/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteCompany(id: string): Promise<void> {
    return this.request(`/companies/${id}`, {
      method: 'DELETE',
    });
  }

  // CNPJ Lookup
  async lookupCNPJ(cnpj: string): Promise<CNPJData> {
    const cleanCNPJ = cnpj.replace(/[^\d]/g, '');
    return this.request<CNPJData>(`/cnpj/${cleanCNPJ}`);
  }

  // Transactions
  async getTransactions(companyId: string, page: number = 1, limit: number = 50): Promise<{ data: Transaction[]; pagination: PaginationInfo }> {
    // Backend returns paginated response {data: Transaction[], pagination: {...}}
    const response = await this.request<{ data: Transaction[]; pagination: PaginationInfo }>(`/transactions?companyId=${companyId}&page=${page}&limit=${limit}`);
    return {
      data: response.data || [],
      pagination: response.pagination || { page: 1, limit, total: 0, totalPages: 0 }
    };
  }

  // Helper for backwards compatibility - returns just the data array
  async getAllTransactions(companyId: string): Promise<Transaction[]> {
    const response = await this.getTransactions(companyId, 1, 1000);
    return response.data;
  }

  async getUpcomingTransactions(companyId?: string, days: number = 7): Promise<UpcomingResponse> {
    const params = new URLSearchParams();
    if (companyId) params.append('companyId', companyId);
    params.append('days', days.toString());
    return this.request<UpcomingResponse>(`/transactions/upcoming?${params.toString()}`);
  }

  async getTransaction(id: string): Promise<Transaction> {
    return this.request<Transaction>(`/transactions/${id}`);
  }

  async createTransaction(data: Omit<Transaction, 'id'>): Promise<Transaction> {
    // companyId é obrigatório na criação
    if (!data.companyId) {
      throw new Error('companyId é obrigatório para criar transação');
    }
    return this.request<Transaction>('/transactions', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateTransaction(id: string, data: Omit<Transaction, 'id'>): Promise<Transaction> {
    return this.request<Transaction>(`/transactions/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteTransaction(id: string): Promise<void> {
    await this.request(`/transactions/${id}`, {
      method: 'DELETE',
    });
  }

  async toggleStatus(id: string): Promise<Transaction> {
    return this.request<Transaction>(`/transactions/${id}/toggle-status`, {
      method: 'PATCH',
    });
  }

  // Stats
  async getStats(companyId: string): Promise<{ paid: number; open: number; total: number }> {
    return this.request(`/stats?companyId=${companyId}`);
  }

  // Recurring Transactions
  async getRecurring(companyId: string): Promise<RecurringTransaction[]> {
    return this.request<RecurringTransaction[]>(`/recurring?companyId=${companyId}`);
  }

  async createRecurring(data: Omit<RecurringTransaction, 'id' | 'createdAt'>): Promise<RecurringTransaction> {
    return this.request<RecurringTransaction>('/recurring', {
      method: 'POST',
      body: JSON.stringify(data)
    });
  }

  async deleteRecurring(id: string): Promise<void> {
    return this.request(`/recurring/${id}`, {
      method: 'DELETE'
    });
  }

  async processRecurring(companyId: string, month: string, year: number): Promise<{ message: string; created: number }> {
    return this.request<{ message: string; created: number }>(`/recurring/process?companyId=${companyId}&month=${month}&year=${year}`, {
      method: 'POST'
    });
  }

  // Seed initial data
  async seedData(): Promise<{ message: string; count: number }> {
    return this.request('/seed', {
      method: 'POST',
    });
  }

  // ===== Quotes =====
  async getQuotes(companyId: string, status?: QuoteStatus): Promise<Quote[]> {
    const params = new URLSearchParams({ companyId });
    if (status) params.append('status', status);
    return this.request<Quote[]>(`/quotes?${params.toString()}`);
  }

  async getQuote(id: string): Promise<Quote> {
    return this.request<Quote>(`/quotes/${id}`);
  }

  async createQuote(companyId: string, data: CreateQuoteRequest): Promise<Quote> {
    return this.request<Quote>(`/quotes?companyId=${companyId}`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateQuote(id: string, data: Partial<CreateQuoteRequest>): Promise<Quote> {
    return this.request<Quote>(`/quotes/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteQuote(id: string): Promise<void> {
    return this.request(`/quotes/${id}`, {
      method: 'DELETE',
    });
  }

  async duplicateQuote(id: string): Promise<Quote> {
    return this.request<Quote>(`/quotes/${id}/duplicate`, {
      method: 'POST',
    });
  }

  async updateQuoteStatus(id: string, status: QuoteStatus): Promise<Quote> {
    return this.request<Quote>(`/quotes/${id}/status`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    });
  }

  async downloadQuotePDF(id: string): Promise<Blob> {
    const headers: HeadersInit = {};
    if (this.accessToken) {
      headers['Authorization'] = `Bearer ${this.accessToken}`;
    }

    // Setup timeout with AbortController for download
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), TIMEOUT_CONFIG.download);

    try {
      const response = await fetch(`${API_URL}/quotes/${id}/pdf`, {
        headers,
        signal: controller.signal,
      });
      if (!response.ok) {
        throw new Error('Falha ao gerar PDF');
      }
      return response.blob();
    } catch (error) {
      if (error instanceof Error && error.name === 'AbortError') {
        throw new Error(`Download timeout after ${TIMEOUT_CONFIG.download}ms`);
      }
      throw error;
    } finally {
      clearTimeout(timeoutId);
    }
  }

  async getQuoteComparison(id: string): Promise<QuoteComparison> {
    return this.request<QuoteComparison>(`/quotes/${id}/comparison`);
  }

  async generateTransactionFromQuote(id: string): Promise<Transaction> {
    return this.request<Transaction>(`/quotes/${id}/generate-transaction`, {
      method: 'POST',
    });
  }

  // ===== Quote Templates =====
  async getQuoteTemplates(companyId: string): Promise<QuoteTemplate[]> {
    return this.request<QuoteTemplate[]>(`/quote-templates?companyId=${companyId}`);
  }

  async getQuoteTemplate(id: string): Promise<QuoteTemplate> {
    return this.request<QuoteTemplate>(`/quote-templates/${id}`);
  }

  async createQuoteTemplate(companyId: string, data: CreateQuoteTemplateRequest): Promise<QuoteTemplate> {
    return this.request<QuoteTemplate>(`/quote-templates?companyId=${companyId}`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateQuoteTemplate(id: string, data: Partial<CreateQuoteTemplateRequest>): Promise<QuoteTemplate> {
    return this.request<QuoteTemplate>(`/quote-templates/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteQuoteTemplate(id: string): Promise<void> {
    return this.request(`/quote-templates/${id}`, {
      method: 'DELETE',
    });
  }
}

export const api = new ApiService();
