import React, { useEffect, useState, useCallback } from 'react';
import { Link } from 'react-router-dom';
import { Company } from '../types';
import { api } from '../services/api';
import { useAuth } from '../contexts/AuthContext';
import { useToast } from '../contexts/ToastContext';
import { Building2, Plus, ArrowRight, Loader2, LogOut, Sparkles, Pencil, Trash2 } from 'lucide-react';
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

        if (!window.confirm(`Tem certeza que deseja excluir "${company.name}"? Esta acao nao pode ser desfeita.`)) {
            return;
        }

        try {
            await api.deleteCompany(company.id);
            setCompanies(prev => prev.filter(c => c.id !== company.id));
            addToast('success', 'Ambiente excluido com sucesso!');
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
            <div className="min-h-screen flex items-center justify-center bg-background">
                <div className="card p-8 flex flex-col items-center gap-4">
                    <Loader2 className="w-10 h-10 text-primary-500 animate-spin" />
                    <p className="text-muted">Carregando...</p>
                </div>
            </div>
        );
    }

    return (
        <div className="min-h-screen bg-background">
            {/* Navigation */}
            <nav className="bg-card border-b border-border shadow-soft sticky top-0 z-40">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex items-center justify-between h-16">
                        <div className="flex items-center gap-3">
                            <div className="p-2.5 bg-gradient-primary rounded-xl shadow-soft">
                                <Sparkles className="w-5 h-5 text-white" />
                            </div>
                            <span className="text-xl font-bold text-foreground">
                                MyPila<span className="text-primary-500">Pro</span>
                            </span>
                        </div>
                        <div className="flex items-center gap-4">
                            <span className="text-sm text-muted hidden sm:inline">
                                Ola, <span className="text-foreground font-medium">{user?.name}</span>
                            </span>
                            <button
                                onClick={logout}
                                className="p-2 text-muted hover:text-foreground hover:bg-primary-50 transition-all rounded-xl"
                                title="Sair"
                            >
                                <LogOut className="w-5 h-5" />
                            </button>
                        </div>
                    </div>
                </div>
            </nav>

            {/* Main Content */}
            <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 md:py-16">
                {/* Header */}
                <div className="text-center mb-12">
                    <h1 className="text-3xl md:text-5xl font-bold text-foreground mb-4">
                        Seus <span className="gradient-text">Ambientes</span>
                    </h1>
                    <p className="text-muted text-lg max-w-2xl mx-auto">
                        Gerencie suas financas de forma inteligente e visual
                    </p>
                </div>

                {/* Company Grid */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {companies.map((company, index) => (
                        <Link
                            key={company.id}
                            to={`/company/${company.id}`}
                            className="group card-hover p-6 cursor-pointer animate-fadeIn"
                            style={{ animationDelay: `${index * 100}ms` }}
                        >
                            <div className="flex items-start justify-between mb-6">
                                <div className="p-3 bg-primary-100 rounded-xl border border-primary-200 group-hover:bg-primary-200 transition-colors">
                                    <Building2 className="h-6 w-6 text-primary-600" />
                                </div>
                                <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                    <button
                                        onClick={(e) => handleEditClick(e, company)}
                                        className="p-2 rounded-lg bg-primary-50 hover:bg-primary-100 text-primary-600 transition-all"
                                        title="Editar"
                                    >
                                        <Pencil className="h-4 w-4" />
                                    </button>
                                    <button
                                        onClick={(e) => handleDeleteClick(e, company)}
                                        className="p-2 rounded-lg bg-destructive-light hover:bg-destructive/20 text-destructive transition-all"
                                        title="Excluir"
                                    >
                                        <Trash2 className="h-4 w-4" />
                                    </button>
                                </div>
                            </div>

                            <h3 className="text-xl font-bold text-foreground mb-2 group-hover:text-primary-600 transition-colors">
                                {company.name}
                            </h3>
                            <p className="text-muted text-sm mb-6">
                                Criado em {new Date(company.createdAt).toLocaleDateString('pt-BR')}
                            </p>

                            <div className="flex items-center justify-between pt-4 border-t border-border">
                                <span className="text-sm text-primary-600 font-medium">Acessar</span>
                                <div className="p-2 rounded-full bg-primary-100 group-hover:bg-primary-500 transition-all">
                                    <ArrowRight className="h-4 w-4 text-primary-600 group-hover:text-white transition-colors" />
                                </div>
                            </div>
                        </Link>
                    ))}

                    {/* New Company Card */}
                    <div className="card p-6 border-dashed border-2 border-border hover:border-primary-300 transition-all">
                        <form onSubmit={handleCreateCompany} className="h-full flex flex-col justify-center">
                            <div className="text-center mb-6">
                                <div className="inline-flex p-3 bg-primary-100 rounded-xl border border-primary-200 mb-4">
                                    <Plus className="h-6 w-6 text-primary-600" />
                                </div>
                                <h3 className="text-lg font-semibold text-foreground">Novo Ambiente</h3>
                            </div>

                            <div className="space-y-4">
                                <div>
                                    <input
                                        type="text"
                                        value={newCompanyName}
                                        onChange={(e) => setNewCompanyName(e.target.value)}
                                        placeholder="Nome do ambiente..."
                                        className={`input ${hasError('companyName') ? 'input-error' : ''}`}
                                    />
                                    <ErrorMessage error={getError('companyName')} />
                                </div>

                                <button
                                    type="submit"
                                    disabled={isCreating || hasErrors()}
                                    className={`w-full py-3 rounded-xl font-medium transition-all ${
                                        isCreating || hasErrors()
                                            ? 'bg-muted/20 text-muted cursor-not-allowed'
                                            : 'btn-primary'
                                    }`}
                                >
                                    {isCreating ? (
                                        <Loader2 className="w-5 h-5 animate-spin mx-auto" />
                                    ) : (
                                        'Criar Ambiente'
                                    )}
                                </button>
                            </div>
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
