

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
  amount: number;
  category: string;
  status: Status;
  description?: string;
}

export interface Company {
  id: string;
  userId: string;
  name: string;
  createdAt: string;
}

export interface User {
  id: string;
  name: string;
  email: string;
}

export interface AuthResponse {
  token: string;
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