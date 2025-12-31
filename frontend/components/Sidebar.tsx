import React from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import { LayoutDashboard, Wallet, LogOut, Tags, FileText, ArrowLeft, RefreshCw, ClipboardList } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';

export const Sidebar: React.FC = () => {
    const { logout, user } = useAuth();
    const navigate = useNavigate();

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <aside className="hidden md:flex flex-col w-64 bg-stone-900 border-r border-stone-800 h-screen fixed left-0 top-0 z-50">
            <div className="p-6">
                <div className="flex items-center gap-2 mb-8">
                    <div className="p-2 bg-gradient-to-tr from-stone-700 to-stone-800 rounded-lg">
                        <LayoutDashboard className="w-5 h-5 text-stone-100" />
                    </div>
                    <span className="text-xl font-bold text-white tracking-tight">Financeiro<span className="text-stone-400">Pro</span></span>
                </div>

                <nav className="flex-1 space-y-2">
                    <NavLink
                        to="/"
                        end
                        className={({ isActive }) =>
                            `flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${isActive ? 'bg-stone-800 text-white shadow-lg shadow-stone-900/20' : 'text-stone-400 hover:bg-stone-800/50 hover:text-white'}`
                        }
                    >
                        <ArrowLeft className="w-5 h-5" />
                        <span>Voltar</span>
                    </NavLink>

                    <div className="pt-4 pb-2">
                        <p className="px-4 text-xs font-semibold text-stone-500 uppercase tracking-wider">Menu Principal</p>
                    </div>

                    <NavLink
                        to="dashboard"
                        className={({ isActive }) =>
                            `flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${isActive ? 'bg-stone-800 text-white shadow-lg shadow-stone-900/20' : 'text-stone-400 hover:bg-stone-800/50 hover:text-white'}`
                        }
                    >
                        <Wallet className="w-5 h-5" />
                        <span>Dashboard</span>
                    </NavLink>

                    <NavLink
                        to="categories"
                        className={({ isActive }) =>
                            `flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${isActive ? 'bg-stone-800 text-white shadow-lg shadow-stone-900/20' : 'text-stone-400 hover:bg-stone-800/50 hover:text-white'}`
                        }
                    >
                        <Tags className="w-5 h-5" />
                        <span>Categorias</span>
                    </NavLink>

                    <NavLink
                        to="reports"
                        className={({ isActive }) =>
                            `flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${isActive ? 'bg-stone-800 text-white shadow-lg shadow-stone-900/20' : 'text-stone-400 hover:bg-stone-800/50 hover:text-white'}`
                        }
                    >
                        <FileText className="w-5 h-5" />
                        <span>Relatórios</span>
                    </NavLink>

                    <NavLink
                        to="recurring"
                        className={({ isActive }) =>
                            `flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${isActive ? 'bg-stone-800 text-white shadow-lg shadow-stone-900/20' : 'text-stone-400 hover:bg-stone-800/50 hover:text-white'}`
                        }
                    >
                        <RefreshCw className="w-5 h-5" />
                        <span>Recorrências</span>
                    </NavLink>

                    <NavLink
                        to="quotes"
                        className={({ isActive }) =>
                            `flex items-center gap-3 px-4 py-3 rounded-xl transition-all ${isActive ? 'bg-stone-800 text-white shadow-lg shadow-stone-900/20' : 'text-stone-400 hover:bg-stone-800/50 hover:text-white'}`
                        }
                    >
                        <ClipboardList className="w-5 h-5" />
                        <span>Orçamentos</span>
                    </NavLink>
                </nav>
            </div>

            <div className="p-4 border-t border-stone-800 mt-auto">
                <div className="flex items-center gap-3 px-4 py-3">
                    <div className="w-8 h-8 rounded-full bg-gradient-to-tr from-stone-600 to-stone-700 flex items-center justify-center text-white font-bold text-sm">
                        {user?.name?.charAt(0).toUpperCase()}
                    </div>
                    <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-white truncate">{user?.name}</p>
                        <p className="text-xs text-stone-500 truncate">{user?.email}</p>
                    </div>
                    <button
                        onClick={handleLogout}
                        className="text-stone-400 hover:text-white transition-colors"
                        title="Sair"
                    >
                        <LogOut className="w-5 h-5" />
                    </button>
                </div>
            </div>
        </aside>
    );
};
