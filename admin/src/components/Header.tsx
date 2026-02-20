import { useAuth } from '../contexts/AuthContext'

export function Header() {
  const { user } = useAuth()

  return (
    <header className="bg-white shadow-sm px-6 py-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold text-gray-800">
          Painel Administrativo
        </h2>
        <div className="flex items-center gap-4">
          <span className="text-gray-600">
            Bem-vindo, <strong>{user?.name}</strong>
          </span>
        </div>
      </div>
    </header>
  )
}
