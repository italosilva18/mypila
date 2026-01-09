import { Transaction, Company, AuthResponse, LoginRequest, RegisterRequest, Category, RecurringTransaction, Quote, CreateQuoteRequest, QuoteTemplate, CreateQuoteTemplateRequest, QuoteComparison, QuoteStatus } from '../types';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8081/api';

class ApiService {
  private token: string | null = localStorage.getItem('token');

  setToken(token: string | null) {
    this.token = token;
    if (token) {
      localStorage.setItem('token', token);
    } else {
      localStorage.removeItem('token');
    }
  }

  getToken(): string | null {
    return this.token;
  }

  private async request<T>(endpoint: string, options?: RequestInit): Promise<T> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };

    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    // Ensure we don't have double slashes if endpoint starts with /
    const cleanEndpoint = endpoint.startsWith('/') ? endpoint : `/${endpoint}`;

    // Fix: Only append / if not present in API_URL and not in endpoint - actually simpler logic:
    // API_URL usually has no trailing slash, so we append endpoint.

    const response = await fetch(`${API_URL}${cleanEndpoint}`, {
      headers: { ...headers, ...options?.headers },
      ...options,
    });

    if (!response.ok) {
      // Handle 401 Unauthorized - token expired or invalid
      if (response.status === 401) {
        this.setToken(null);
        // Redirect to login page
        window.location.href = '/login';
        throw new Error('Invalid or expired token');
      }

      const error = await response.json().catch(() => ({ error: 'Unknown error' }));
      throw new Error(error.error || `HTTP error! status: ${response.status}`);
    }

    return response.json();
  }

  // Auth
  async login(data: LoginRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    this.setToken(response.token);
    return response;
  }

  async register(data: RegisterRequest): Promise<AuthResponse> {
    const response = await this.request<AuthResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    this.setToken(response.token);
    return response;
  }

  async getMe(): Promise<{ user: { id: string; name: string; email: string } }> {
    return this.request<{ user: { id: string; name: string; email: string } }>('/auth/me');
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

  logout() {
    this.setToken(null);
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

  async updateCompany(id: string, name: string): Promise<Company> {
    return this.request<Company>(`/companies/${id}`, {
      method: 'PUT',
      body: JSON.stringify({ name }),
    });
  }

  async deleteCompany(id: string): Promise<void> {
    return this.request(`/companies/${id}`, {
      method: 'DELETE',
    });
  }

  // Transactions
  async getTransactions(companyId: string): Promise<Transaction[]> {
    // Backend returns paginated response {data: Transaction[], pagination: {...}}
    const response = await this.request<{ data: Transaction[]; pagination: { page: number; limit: number; total: number; totalPages: number } }>(`/transactions?companyId=${companyId}`);
    return response.data || [];
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
    if (this.token) {
      headers['Authorization'] = `Bearer ${this.token}`;
    }

    const response = await fetch(`${API_URL}/quotes/${id}/pdf`, { headers });
    if (!response.ok) {
      throw new Error('Falha ao gerar PDF');
    }
    return response.blob();
  }

  async getQuoteComparison(id: string): Promise<QuoteComparison> {
    return this.request<QuoteComparison>(`/quotes/${id}/comparison`);
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
