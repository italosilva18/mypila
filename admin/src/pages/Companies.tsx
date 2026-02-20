import { useEffect, useState } from 'react'
import { toast } from 'react-toastify'
import { Search, Plus, Building2 } from 'lucide-react'
import { api } from '../services/api'
import { Company } from '../types'

export function Companies() {
  const [companies, setCompanies] = useState<Company[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [search, setSearch] = useState('')

  useEffect(() => {
    loadCompanies()
  }, [])

  const loadCompanies = async () => {
    try {
      setIsLoading(true)
      const { data } = await api.get('/admin/companies')
      // API returns { data: [...], pagination: {...} }
      setCompanies(data.data || [])
    } catch (error) {
      toast.error('Erro ao carregar empresas')
    } finally {
      setIsLoading(false)
    }
  }

  const filteredCompanies = companies.filter(company =>
    company.name.toLowerCase().includes(search.toLowerCase())
  )

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Gerenciar Empresas</h1>
        <button className="btn-primary flex items-center gap-2">
          <Plus size={20} />
          Nova Empresa
        </button>
      </div>

      <div className="bg-white rounded-lg shadow-md">
        <div className="p-4 border-b border-gray-200">
          <div className="relative">
            <Search className="absolute left-3 top-3 text-gray-400" size={20} />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Buscar empresas..."
              className="input pl-10"
            />
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 p-6">
          {filteredCompanies.map((company) => (
            <div key={company.id} className="bg-white border border-gray-200 rounded-lg p-6 hover:shadow-lg transition-shadow">
              <div className="flex items-start justify-between">
                <div className="p-3 bg-blue-100 rounded-lg">
                  <Building2 className="text-blue-600" size={24} />
                </div>
              </div>
              <h3 className="mt-4 text-lg font-semibold text-gray-900">{company.name}</h3>
              {company.cnpj && (
                <p className="text-sm text-gray-500 mt-1">CNPJ: {company.cnpj}</p>
              )}
              <p className="text-sm text-gray-400 mt-2">
                Criada em: {new Date(company.createdAt).toLocaleDateString('pt-BR')}
              </p>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
