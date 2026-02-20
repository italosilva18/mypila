

export enum CategoryType {
  EXPENSE = 'EXPENSE',
  INCOME = 'INCOME',
}

export interface Category {
  id: string;
  companyId: string;
  name: string;
  type: CategoryType;
  color?: string;
  budget?: number;
  createdAt: string;
}




export enum Status {
  PAID = 'PAGO',
  OPEN = 'ABERTO',
}

export interface Transaction {
  id: string;
  companyId: string;
  month: string; // e.g., "Janeiro", "Fevereiro"
  year: number;
  dueDay: number; // 1-31
  amount: number; // Valor Total
  paidAmount: number; // Valor Pago (parcial) - 0 se nao pago
  category: string;
  status: Status;
  description?: string;
}

export interface UpcomingTransaction extends Transaction {
  daysUntilDue: number;
}

export interface UpcomingResponse {
  upcoming: UpcomingTransaction[];
  count: number;
  days: number;
}

export interface Company {
  id: string;
  userId: string;
  name: string;
  cnpj?: string;
  legalName?: string;
  tradeName?: string;
  email?: string;
  phone?: string;
  address?: string;
  city?: string;
  state?: string;
  zipCode?: string;
  logoUrl?: string;
  createdAt: string;
}

export interface CNPJData {
  cnpj: string;
  razaoSocial: string;
  nomeFantasia: string;
  logradouro: string;
  numero: string;
  complemento: string;
  bairro: string;
  municipio: string;
  uf: string;
  cep: string;
  telefone: string;
  situacao: string;
  atividade: string;
}

export interface UpdateCompanyRequest {
  name?: string;
  cnpj?: string;
  legalName?: string;
  tradeName?: string;
  email?: string;
  phone?: string;
  address?: string;
  city?: string;
  state?: string;
  zipCode?: string;
  logoUrl?: string;
}

export interface User {
  id: string;
  name: string;
  email: string;
}

export interface AuthResponse {
  accessToken: string;
  refreshToken: string;
  expiresIn: number;
  user: User;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
}

export interface SummaryData {
  totalPaid: number;
  totalOpen: number;
  totalProjected: number;
}

export interface RecurringTransaction {
  id: string;
  companyId: string;
  description: string;
  amount: number;
  category: string;
  dayOfMonth: number;
  createdAt: string;
}

// ===== Quote Types =====

export enum QuoteStatus {
  DRAFT = 'DRAFT',
  SENT = 'SENT',
  APPROVED = 'APPROVED',
  REJECTED = 'REJECTED',
  EXECUTED = 'EXECUTED',
}

export interface QuoteItem {
  id?: string;
  description: string;
  quantity: number;
  unitPrice: number;
  total: number;
  categoryId?: string;
}

export interface Quote {
  id: string;
  companyId: string;
  number: string;
  // Dados do cliente
  clientName: string;
  clientEmail?: string;
  clientPhone?: string;
  clientDocument?: string;
  clientAddress?: string;
  clientCity?: string;
  clientState?: string;
  clientZipCode?: string;
  // Dados do orçamento
  title: string;
  description?: string;
  items: QuoteItem[];
  subtotal: number;
  discount: number;
  discountType: 'PERCENT' | 'VALUE';
  total: number;
  status: QuoteStatus;
  validUntil: string;
  notes?: string;
  templateId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateQuoteRequest {
  clientName: string;
  clientEmail?: string;
  clientPhone?: string;
  clientDocument?: string;
  clientAddress?: string;
  clientCity?: string;
  clientState?: string;
  clientZipCode?: string;
  title: string;
  description?: string;
  items: Omit<QuoteItem, 'id' | 'total'>[];
  discount: number;
  discountType: 'PERCENT' | 'VALUE';
  validUntil: string;
  notes?: string;
  templateId?: string;
}

export interface QuoteTemplate {
  id: string;
  companyId: string;
  name: string;
  headerText?: string;
  footerText?: string;
  termsText?: string;
  primaryColor: string;
  logoUrl?: string;
  isDefault: boolean;
  createdAt: string;
}

export interface CreateQuoteTemplateRequest {
  name: string;
  headerText?: string;
  footerText?: string;
  termsText?: string;
  primaryColor: string;
  logoUrl?: string;
  isDefault: boolean;
}

export interface QuoteComparisonItem {
  description: string;
  categoryId?: string;
  quoted: number;
  executed: number;
  variance: number;
}

export interface QuoteComparison {
  quoteId: string;
  quotedTotal: number;
  executedTotal: number;
  variance: number;
  variancePercent: number;
  items: QuoteComparisonItem[];
}

// ===== Pagination Types =====
export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  pagination: PaginationInfo;
}