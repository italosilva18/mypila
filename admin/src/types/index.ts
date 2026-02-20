export interface User {
  id: string
  name: string
  email: string
  companyCount?: number
  createdAt: string
  updatedAt: string
}

export interface Company {
  id: string
  userId: string
  name: string
  cnpj?: string
  createdAt: string
  updatedAt: string
}

export interface Transaction {
  id: string
  companyId: string
  description: string
  amount: number
  type: 'income' | 'expense'
  date: string
  categoryId?: string
  status: 'completed' | 'pending' | 'cancelled'
  createdAt: string
}

export interface DashboardStats {
  totalUsers: number
  totalCompanies: number
  totalTransactions: number
  totalRevenue: number
  recentUsers: User[]
  recentTransactions: Transaction[]
}

export interface AuthResponse {
  accessToken: string
  refreshToken: string
  user: User
}
