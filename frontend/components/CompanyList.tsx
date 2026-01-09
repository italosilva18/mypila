import React, { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { Company } from '../types';
import { api } from '../services/api';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';
import { Building2, Plus, ArrowRight, Loader2, LogOut, LayoutDashboard, Pencil, Trash2 } from 'lucide-react';
import { CompanyModal } from './CompanyModal';
import { useFormValidation } from '../hooks/useFormValidation';
import { validateRequired, validateMaxLength, combineValidations } from '../utils/validation';
import { ErrorMessage } from './ErrorMessage';

export const CompanyList: React.FC = () => {
    const [companies, setCompanies] = useState<Company[]>([]);
    const [loading, setLoading] = useState(true);
    const [newCompanyName, setNewCompanyName] = useState('');
    const [isCreating, setIsCreating] = useState(false);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingCompany, setEditingCompany] = useState<Company | null>(null);
    const { user, logout } = useAuth();
    const { addToast } = useToast();
    const { validateFields, getError, hasError, hasErrors, clearAllErrors } = useFormValidation();

    useEffect(() => {
        loadCompanies();
    }, []);

    const loadCompanies = async () => {
        try {
            const data = await api.getCompanies();
            setCompanies(data);
        } catch (err) {
            console.error('Failed to load companies', err);
        } finally {
            setLoading(false);
        }
    };

    const validateForm = (): boolean => {
        return validateFields({
            companyName: () => combineValidations(
                validateRequired(newCompanyName, 'Nome'),
                validateMaxLength(newCompanyName, 100, 'Nome')
            )
        });
    };

    const handleCreateCompany = async (e: React.FormEvent) => {
        e.preventDefault();

        if (!validateForm()) {
            return;
        }

        try {
            setIsCreating(true);
            const newCompany = await api.createCompany(newCompanyName);
            setCompanies([...companies, newCompany]);
            setNewCompanyName('');
            clearAllErrors();
            addToast('success', 'Ambiente criado com sucesso!');
        } catch (err) {
            console.error('Failed to create company', err);
            addToast('error', 'Erro ao criar ambiente. Tente novamente.');
        } finally {
            setIsCreating(false);
        }
    };

    const handleEditClick = useCallback((e: React.MouseEvent, company: Company) => {
        e.preventDefault();
        e.stopPropagation();
        setEditingCompany(company);
        setIsModalOpen(true);
    }, []);

    const handleDeleteClick = useCallback(async (e: React.MouseEvent, company: Company) => {
        e.preventDefault();
        e.stopPropagation();

        if (!window.confirm(`Tem certeza que deseja excluir o ambiente "${company.name}"? Esta ação não pode ser desfeita e todas as transações associadas serão excluídas.`)) {
            return;
        }

        try {
            await api.deleteCompany(company.id);
            setCompanies(prev => prev.filter(c => c.id !== company.id));
            addToast('success', 'Ambiente excluído com sucesso!');
        } catch (err) {
            console.error('Failed to delete company', err);
            addToast('error', 'Erro ao excluir ambiente. Tente novamente.');
        }
    }, [addToast]);

    const handleSaveCompany = async (name: string) => {
        if (!editingCompany) return;

        try {
            const updatedCompany = await api.updateCompany(editingCompany.id, name);
            setCompanies(companies.map(c => c.id === updatedCompany.id ? updatedCompany : c));
            addToast('success', 'Ambiente atualizado com sucesso!');
            setIsModalOpen(false);
            setEditingCompany(null);
        } catch (err) {
            console.error('Failed to update company', err);
            addToast('error', 'Erro ao atualizar ambiente. Tente novamente.');
        }
    };

    const handleCloseModal = () => {
        setIsModalOpen(false);
        setEditingCompany(null);
    };

    if (loading) {
        return (
            <div className="min-h-screen bg-paper flex items-center justify-center">
                <Loader2 className="w-8 h-8 text-stone-600 animate-spin" />
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-paper relative">
            {/* Background Effects - Vintage style */}
            <div className="absolute top-0 left-0 w-full h-full overflow-hidden z-0 pointer-events-none">
                <div className="absolute top-[10%] left-[20%] w-[50%] h-[50%] rounded-full bg-stone-300/20 blur-[100px]" />
                <div className="absolute bottom-[10%] right-[10%] w-[40%] h-[40%] rounded-full bg-stone-400/15 blur-[100px]" />
            </div>

            <nav className="relative z-10 bg-white/50 backdrop-blur-md border-b border-stone-200">
                <div className="max-w-7xl mx-auto px-3 md:px-4 sm:px-6 lg:px-8">
                    <div className="flex items-center justify-between h-12 md:h-16">
                        <div className="flex items-center gap-1.5 md:gap-2">
                            <div className="p-1.5 md:p-2 bg-gradient-to-tr from-stone-700 to-stone-800 rounded-md md:rounded-lg">
                                <LayoutDashboard className="w-4 h-4 md:w-5 md:h-5 text-white" />
                            </div>
                            <span className="text-base md:text-xl font-bold text-stone-900 tracking-tight">Financeiro<span className="text-stone-500">Pro</span></span>
                        </div>
                        <div className="flex items-center gap-2 md:gap-4">
                            <span className="text-xs md:text-sm text-stone-500 hidden sm:inline">Olá, <span className="text-stone-800 font-medium">{user?.name}</span></span>
                            <button
                                onClick={logout}
                                className="p-1.5 md:p-2 text-stone-500 hover:text-stone-800 active:text-stone-800 transition-colors rounded-md md:rounded-lg hover:bg-stone-100 active:bg-stone-100"
                                title="Sair"
                            >
                                <LogOut className="w-4 h-4 md:w-5 md:h-5" />
                            </button>
                        </div>
                    </div>
                </div>
            </nav>

            <main className="relative z-10 max-w-7xl mx-auto px-3 md:px-4 sm:px-6 lg:px-8 py-4 md:py-12">
                <div className="text-center mb-6 md:mb-12">
                    <h1 className="text-xl md:text-3xl font-bold text-stone-900 mb-1 md:mb-3">Seus Ambientes</h1>
                    <p className="text-xs md:text-base text-stone-500 max-w-2xl mx-auto">
                        Acesse seus painéis de gestão financeira.
                    </p>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 md:gap-6">
                    {companies.map((company) => (
                        <Link
                            key={company.id}
                            to={`/company/${company.id}`}
                            className="group relative bg-white/70 backdrop-blur-sm border border-stone-200 rounded-xl md:rounded-2xl p-3 md:p-6 hover:bg-white active:bg-white hover:border-stone-300 hover:shadow-card transition-all duration-300"
                        >
                            <div className="flex items-center justify-between mb-2 md:mb-4">
                                <div className="h-9 w-9 md:h-12 md:w-12 bg-gradient-to-br from-stone-100 to-stone-200 rounded-lg md:rounded-xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                                    <Building2 className="h-4 w-4 md:h-6 md:w-6 text-stone-600" />
                                </div>
                                <div className="flex items-center gap-1.5 md:gap-2">
                                    <button
                                        onClick={(e) => handleEditClick(e, company)}
                                        className="h-7 w-7 md:h-8 md:w-8 rounded-md md:rounded-lg bg-stone-100 hover:bg-blue-500 active:bg-blue-500 hover:text-white active:text-white flex items-center justify-center transition-all"
                                        title="Editar ambiente"
                                    >
                                        <Pencil className="h-3.5 w-3.5 md:h-4 md:w-4" />
                                    </button>
                                    <button
                                        onClick={(e) => handleDeleteClick(e, company)}
                                        className="h-7 w-7 md:h-8 md:w-8 rounded-md md:rounded-lg bg-stone-100 hover:bg-red-500 active:bg-red-500 hover:text-white active:text-white flex items-center justify-center transition-all"
                                        title="Excluir ambiente"
                                    >
                                        <Trash2 className="h-3.5 w-3.5 md:h-4 md:w-4" />
                                    </button>
                                    <div className="h-7 w-7 md:h-8 md:w-8 rounded-full bg-stone-100 flex items-center justify-center group-hover:bg-stone-800 transition-all">
                                        <ArrowRight className="h-3.5 w-3.5 md:h-4 md:w-4 text-stone-500 group-hover:text-white" />
                                    </div>
                                </div>
                            </div>
                            <h3 className="text-base md:text-xl font-bold text-stone-900 mb-0.5 md:mb-1 group-hover:text-stone-700 transition-colors">{company.name}</h3>
                            <p className="text-xs md:text-sm text-stone-500">
                                Criada em {new Date(company.createdAt).toLocaleDateString()}
                            </p>
                        </Link>
                    ))}

                    {/* New Company Card */}
                    <div className="bg-white/50 backdrop-blur-sm border border-stone-200 border-dashed rounded-xl md:rounded-2xl p-3 md:p-6 flex flex-col justify-center hover:bg-white/70 transition-all duration-300">
                        <form onSubmit={handleCreateCompany} className="space-y-3 md:space-y-4">
                            <h3 className="text-xs md:text-sm font-semibold text-stone-500 uppercase tracking-widest text-center">Novo Ambiente</h3>
                            <div>
                                <input
                                    type="text"
                                    value={newCompanyName}
                                    onChange={(e) => setNewCompanyName(e.target.value)}
                                    placeholder="Nome do ambiente..."
                                    className={`w-full px-3 md:px-4 py-2.5 md:py-3 bg-stone-50 border ${
                                        hasError('companyName') ? 'border-red-500 focus:ring-red-400' : 'border-stone-200 focus:ring-stone-400'
                                    } rounded-lg md:rounded-xl text-sm md:text-base text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:border-transparent transition-all`}
                                />
                                <ErrorMessage error={getError('companyName')} />
                            </div>
                            <button
                                type="submit"
                                disabled={isCreating || hasErrors()}
                                className={`w-full flex items-center justify-center gap-2 px-3 md:px-4 py-2.5 md:py-3 rounded-lg md:rounded-xl text-sm md:text-base font-medium transition-all shadow-lg ${
                                    isCreating || hasErrors()
                                        ? 'bg-stone-300 text-stone-500 cursor-not-allowed'
                                        : 'bg-stone-800 hover:bg-stone-700 active:bg-stone-700 text-white shadow-stone-900/20 active:scale-[0.98]'
                                }`}
                            >
                                {isCreating ? (
                                    <Loader2 className="w-4 h-4 animate-spin" />
                                ) : (
                                    <Plus className="w-4 h-4" />
                                )}
                                Criar
                            </button>
                        </form>
                    </div>
                </div>
            </main>

            <CompanyModal
                isOpen={isModalOpen}
                onClose={handleCloseModal}
                onSave={handleSaveCompany}
                company={editingCompany}
            />
        </div>
    );
};
