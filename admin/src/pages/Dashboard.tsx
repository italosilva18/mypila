import { useEffect, useState } from 'react'
import { toast } from 'react-toastify'
import { Users, Building2, Receipt, TrendingUp } from 'lucide-react'
import { api } from '../services/api'
import { DashboardStats } from '../types'

export function Dashboard() {
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    loadStats()
  }, [])

  const loadStats = async () => {
    try {
      setIsLoading(true)
      const { data } = await api.get('/admin/stats')
      setStats(data)
    } catch (error) {
      toast.error('Erro ao carregar estatísticas')
    } finally {
      setIsLoading(false)
    }
  }

  const statCards = [
    { 
      title: 'Total de Usuários', 
      value: stats?.totalUsers || 0, 
      icon: Users, 
      color: 'bg-blue-500' 
    },
    { 
      title: 'Total de Empresas', 
      value: stats?.totalCompanies || 0, 
      icon: Building2, 
      color: 'bg-green-500' 
    },
    { 
      title: 'Total de Transações', 
      value: stats?.totalTransactions || 0, 
      icon: Receipt, 
      color: 'bg-purple-500' 
    },
    { 
      title: 'Receita Total', 
      value: `R$ ${(stats?.totalRevenue || 0).toLocaleString('pt-BR', { minimumFractionDigits: 2 })}`, 
      icon: TrendingUp, 
      color: 'bg-orange-500' 
    },
  ]

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  return (
    <div>
      <h1 className="text-2xl font-bold text-gray-900 mb-6">Dashboard</h1>
      
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {statCards.map((card) => (
          <div key={card.title} className="bg-white rounded-lg shadow-md p-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600">{card.title}</p>
                <p className="text-2xl font-bold text-gray-900 mt-1">{card.value}</p>
              </div>
              <div className={`${card.color} p-3 rounded-lg`}>
                <card.icon className="text-white" size={24} />
              </div>
            </div>
          </div>
        ))}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="bg-white rounded-lg shadow-md p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            Usuários Recentes
          </h2>
          {stats?.recentUsers && stats.recentUsers.length > 0 ? (
            <div className="space-y-3">
              {stats.recentUsers.map((user) => (
                <div key={user.id} className="flex items-center justify-between py-2 border-b border-gray-100 last:border-0">
                  <div>
                    <p className="font-medium text-gray-900">{user.name}</p>
                    <p className="text-sm text-gray-500">{user.email}</p>
                  </div>
                  <span className="text-sm text-gray-400">
                    {new Date(user.createdAt).toLocaleDateString('pt-BR')}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-gray-500">Nenhum usuário recente</p>
          )}
        </div>

        <div className="bg-white rounded-lg shadow-md p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">
            Transações Recentes
          </h2>
          {stats?.recentTransactions && stats.recentTransactions.length > 0 ? (
            <div className="space-y-3">
              {stats.recentTransactions.map((transaction) => (
                <div key={transaction.id} className="flex items-center justify-between py-2 border-b border-gray-100 last:border-0">
                  <div>
                    <p className="font-medium text-gray-900">{transaction.description}</p>
                    <p className={`text-sm ${transaction.type === 'income' ? 'text-green-600' : 'text-red-600'}`}>
                      {transaction.type === 'income' ? '+' : '-'} R$ {transaction.amount.toLocaleString('pt-BR', { minimumFractionDigits: 2 })}
                    </p>
                  </div>
                  <span className="text-sm text-gray-400">
                    {new Date(transaction.date).toLocaleDateString('pt-BR')}
                  </span>
                </div>
              ))}
            </div>
          ) : (
            <p className="text-gray-500">Nenhuma transação recente</p>
          )}
        </div>
      </div>
    </div>
  )
}
