import React from 'react';
import { NavLink, useNavigate } from 'react-router-dom';
import { Wallet, LogOut, Tags, FileText, ArrowLeft, RefreshCw, ClipboardList, Sparkles } from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';

export const Sidebar: React.FC = () => {
    const { logout, user } = useAuth();
    const navigate = useNavigate();

    const handleLogout = () => {
        logout();
        navigate('/login');
    };

    return (
        <aside className="hidden md:flex flex-col w-64 bg-card border-r border-border h-screen fixed left-0 top-0 z-50 shadow-soft">
            <div className="p-6">
                <div className="flex items-center gap-3 mb-8">
                    <div className="p-2.5 bg-gradient-primary rounded-xl shadow-soft">
                        <Sparkles className="w-5 h-5 text-white" />
                    </div>
                    <span className="text-xl font-bold text-foreground tracking-tight">
                        MyPila<span className="text-primary-500">Pro</span>
                    </span>
                </div>

                <nav className="flex-1 space-y-2">
                    <NavLink
                        to="/"
                        end
                        className={({ isActive }) =>
                            `nav-item ${isActive ? 'nav-item-active' : ''}`
                        }
                    >
                        <ArrowLeft className="w-5 h-5" />
                        <span>Voltar</span>
                    </NavLink>

                    <div className="pt-4 pb-2">
                        <p className="px-4 text-xs font-semibold text-muted uppercase tracking-wider">Menu Principal</p>
                    </div>

                    <NavLink
                        to="dashboard"
                        className={({ isActive }) =>
                            `nav-item ${isActive ? 'nav-item-active' : ''}`
                        }
                    >
                        <Wallet className="w-5 h-5" />
                        <span>Dashboard</span>
                    </NavLink>

                    <NavLink
                        to="categories"
                        className={({ isActive }) =>
                            `nav-item ${isActive ? 'nav-item-active' : ''}`
                        }
                    >
                        <Tags className="w-5 h-5" />
                        <span>Categorias</span>
                    </NavLink>

                    <NavLink
                        to="reports"
                        className={({ isActive }) =>
                            `nav-item ${isActive ? 'nav-item-active' : ''}`
                        }
                    >
                        <FileText className="w-5 h-5" />
                        <span>Relatorios</span>
                    </NavLink>

                    <NavLink
                        to="recurring"
                        className={({ isActive }) =>
                            `nav-item ${isActive ? 'nav-item-active' : ''}`
                        }
                    >
                        <RefreshCw className="w-5 h-5" />
                        <span>Recorrencias</span>
                    </NavLink>

                    <NavLink
                        to="quotes"
                        className={({ isActive }) =>
                            `nav-item ${isActive ? 'nav-item-active' : ''}`
                        }
                    >
                        <ClipboardList className="w-5 h-5" />
                        <span>Orcamentos</span>
                    </NavLink>
                </nav>
            </div>

            <div className="p-4 border-t border-border mt-auto">
                <div className="flex items-center gap-3 px-4 py-3">
                    <div className="w-8 h-8 rounded-full bg-gradient-primary flex items-center justify-center text-white font-bold text-sm shadow-soft">
                        {user?.name?.charAt(0).toUpperCase()}
                    </div>
                    <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-foreground truncate">{user?.name}</p>
                        <p className="text-xs text-muted truncate">{user?.email}</p>
                    </div>
                    <button
                        onClick={handleLogout}
                        className="text-muted hover:text-foreground transition-colors p-2 rounded-lg hover:bg-primary-50"
                        title="Sair"
                        aria-label="Sair da conta"
                    >
                        <LogOut className="w-5 h-5" />
                    </button>
                </div>
            </div>
        </aside>
    );
};
