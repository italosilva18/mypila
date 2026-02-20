import { NavLink } from 'react-router-dom'
import { 
  LayoutDashboard, 
  Users, 
  Building2, 
  Receipt, 
  Settings,
  LogOut 
} from 'lucide-react'
import { useAuth } from '../contexts/AuthContext'

export function Sidebar() {
  const { logout } = useAuth()

  const menuItems = [
    { path: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
    { path: '/users', label: 'Usuários', icon: Users },
    { path: '/companies', label: 'Empresas', icon: Building2 },
    { path: '/transactions', label: 'Transações', icon: Receipt },
    { path: '/settings', label: 'Configurações', icon: Settings },
  ]

  return (
    <aside className="fixed left-0 top-0 h-full w-64 bg-slate-800 text-white">
      <div className="p-6">
        <h1 className="text-2xl font-bold">MyPila Admin</h1>
        <p className="text-slate-400 text-sm mt-1">Painel Administrativo</p>
      </div>
      
      <nav className="mt-6">
        {menuItems.map((item) => (
          <NavLink
            key={item.path}
            to={item.path}
            className={({ isActive }) =>
              `flex items-center gap-3 px-6 py-3 transition-colors ${
                isActive 
                  ? 'bg-blue-600 text-white' 
                  : 'text-slate-300 hover:bg-slate-700'
              }`
            }
          >
            <item.icon size={20} />
            {item.label}
          </NavLink>
        ))}
      </nav>

      <div className="absolute bottom-0 w-full p-6">
        <button
          onClick={logout}
          className="flex items-center gap-3 text-slate-300 hover:text-white transition-colors"
        >
          <LogOut size={20} />
          Sair
        </button>
      </div>
    </aside>
  )
}
