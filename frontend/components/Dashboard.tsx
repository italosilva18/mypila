import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useParams, Link } from 'react-router-dom';
import {
    Building2,
    ArrowLeft,
    Plus,
    Search,
    Wallet,
    CheckCircle2,
    AlertCircle,
    Edit2
} from 'lucide-react';
import { useTransactions } from '../hooks/useTransactions';
import { useDebounce } from '../hooks/useDebounce';
import { Transaction, Category, CategoryType, Status } from '../types';

import { api } from '../services/api';
import { FinancialChart } from './FinancialChart';
import { useDateFilter } from '../contexts/DateFilterContext';
import { DateSelector } from './DateSelector';
import { TrendChart } from './TrendChart';
import { TransactionModal } from './TransactionModal';
import { formatCurrency } from '../utils/currency';

export const Dashboard: React.FC = () => {
    const { companyId } = useParams<{ companyId: string }>();
    const { month, year } = useDateFilter();
    const { transactions, loading, createTransaction, updateTransaction, toggleStatus, deleteTransaction } = useTransactions(companyId!);
    const [searchTerm, setSearchTerm] = useState('');
    const debouncedSearchTerm = useDebounce(searchTerm, 300);
    const [showModal, setShowModal] = useState(false);
    const [editingTransaction, setEditingTransaction] = useState<Transaction | null>(null);
    const [companyName, setCompanyName] = useState('Carregando...');
    const [categories, setCategories] = useState<Category[]>([]);

    // Memoized filtered transactions - only recalculates when dependencies change
    const filteredTransactions = useMemo(() => {
        return transactions.filter(t => {
            const matchesSearch = t.description?.toLowerCase().includes(debouncedSearchTerm.toLowerCase()) ||
                t.category.toLowerCase().includes(debouncedSearchTerm.toLowerCase());

            const matchesDate = t.month === month && t.year === year;

            return matchesSearch && matchesDate;
        });
    }, [transactions, debouncedSearchTerm, month, year]);

    // Memoized stats - only recalculates when filtered transactions change
    const currentStats = useMemo(() => {
        return {
            paid: filteredTransactions.filter(t => t.status === Status.PAID).reduce((acc, t) => acc + t.amount, 0),
            open: filteredTransactions.filter(t => t.status === Status.OPEN).reduce((acc, t) => acc + t.amount, 0),
            total: filteredTransactions.reduce((acc, t) => acc + t.amount, 0)
        };
    }, [filteredTransactions]);

    useEffect(() => {
        if (companyId) {
            api.getCompanies().then(companies => {
                const current = companies.find(c => c.id === companyId);
                if (current) setCompanyName(current.name);
            });

            api.getCategories(companyId).then(setCategories);
        }
    }, [companyId]);

    // Memoized handlers - prevent unnecessary re-renders of child components
    const handleSaveTransaction = useCallback(async (data: Omit<Transaction, 'id'>) => {
        try {
            if (editingTransaction) {
                await updateTransaction(editingTransaction.id, data);
            } else {
                await createTransaction(data);
            }
            setShowModal(false);
            setEditingTransaction(null);
        } catch (err) {
            console.error('Failed to save transaction', err);
        }
    }, [editingTransaction, updateTransaction, createTransaction]);

    const handleEditClick = useCallback((transaction: Transaction) => {
        setEditingTransaction(transaction);
        setShowModal(true);
    }, []);

    const handleModalClose = useCallback(() => {
        setShowModal(false);
        setEditingTransaction(null);
    }, []);

    const handleCreateCategory = useCallback(async (name: string, type: CategoryType): Promise<Category> => {
        if (!companyId) throw new Error('Company ID is required');
        const newCategory = await api.createCategory(companyId, name, type, '#78716c', 0);
        setCategories(prev => [...prev, newCategory]);
        return newCategory;
    }, [companyId]);

    if (loading) return (
        <div className="min-h-screen bg-paper flex items-center justify-center text-stone-900">
            <div className="flex flex-col items-center gap-4">
                <div className="w-10 h-10 border-4 border-stone-600 border-t-transparent rounded-full animate-spin"></div>
                <p className="text-stone-500">Carregando dados...</p>
            </div>
        </div>
    );

    return (
        <div className="min-h-screen bg-paper text-stone-900 font-sans selection:bg-stone-300/50">

            {/* Header - Desktop */}
            <header className="bg-white/70 backdrop-blur-xl border-b border-stone-200 sticky top-0 z-40 hidden md:block">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-4">
                            <Link to="/" className="p-2 hover:bg-stone-100 rounded-xl transition-colors text-stone-500 hover:text-stone-800 group" aria-label="Voltar para lista de empresas">
                                <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
                            </Link>
                            <div>
                                <h1 className="text-xl font-bold text-stone-900 flex items-center gap-2">
                                    <Building2 className="w-5 h-5 text-stone-600" aria-hidden="true" />
                                    {companyName}
                                </h1>
                            </div>
                            <div className="h-8 w-px bg-stone-200 mx-2" aria-hidden="true"></div>
                            <DateSelector />
                        </div>

                        <button
                            onClick={() => setShowModal(true)}
                            className="bg-stone-800 hover:bg-stone-700 text-white px-4 py-2.5 rounded-xl text-sm font-medium flex items-center gap-2 transition-all shadow-lg shadow-stone-900/20 active:scale-95"
                            aria-label="Adicionar nova transação"
                        >
                            <Plus className="w-4 h-4" aria-hidden="true" />
                            Nova Transação
                        </button>
                    </div>
                </div>
            </header>

            {/* Header - Mobile - Compacto */}
            <header className="bg-white/90 backdrop-blur-xl border-b border-stone-100 sticky top-0 z-40 md:hidden">
                <div className="px-3 py-2">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-1.5">
                            <Link to="/" className="p-1 -ml-1 rounded-md active:bg-stone-100 transition-colors" aria-label="Voltar para lista de empresas">
                                <ArrowLeft size={16} className="text-stone-500" />
                            </Link>
                            <h1 className="text-sm font-semibold text-stone-800 truncate max-w-[120px]">{companyName}</h1>
                        </div>
                        <div className="flex items-center gap-2">
                            <DateSelector />
                            <button
                                onClick={() => setShowModal(true)}
                                className="bg-stone-800 active:bg-stone-700 text-white p-1.5 rounded-lg transition-all shadow-sm active:scale-95"
                                aria-label="Adicionar nova transação"
                            >
                                <Plus size={16} />
                            </button>
                        </div>
                    </div>
                </div>
            </header>

            <main className="max-w-7xl mx-auto px-2 md:px-4 sm:px-6 lg:px-8 py-2 md:py-8 mobile-content-padding">
                {/* Stats Grid - Compacto Mobile */}
                <div className="grid grid-cols-3 md:grid-cols-3 gap-1.5 md:gap-6 mb-3 md:mb-8">
                    <div className="bg-white/80 backdrop-blur-sm border border-stone-100 rounded-lg md:rounded-2xl p-2 md:p-6 relative overflow-hidden group hover:shadow-card transition-all">
                        <div className="hidden md:block absolute top-0 right-0 w-24 h-24 bg-green-500/10 rounded-full blur-2xl -mr-6 -mt-6 transition-all group-hover:bg-green-500/20"></div>
                        <div className="flex items-center gap-1 mb-1 md:mb-4">
                            <div className="p-0.5 md:p-2 bg-green-50 md:bg-green-100 rounded md:rounded-lg text-green-600">
                                <CheckCircle2 className="w-3 h-3 md:w-5 md:h-5" />
                            </div>
                            <p className="text-[8px] md:text-sm font-medium text-stone-400 md:text-stone-500">Pago</p>
                        </div>
                        <p className="text-xs md:text-3xl font-bold text-stone-800">{formatCurrency(currentStats.paid)}</p>
                    </div>

                    <div className="bg-white/80 backdrop-blur-sm border border-stone-100 rounded-lg md:rounded-2xl p-2 md:p-6 relative overflow-hidden group hover:shadow-card transition-all">
                        <div className="hidden md:block absolute top-0 right-0 w-24 h-24 bg-amber-500/10 rounded-full blur-2xl -mr-6 -mt-6 transition-all group-hover:bg-amber-500/20"></div>
                        <div className="flex items-center gap-1 mb-1 md:mb-4">
                            <div className="p-0.5 md:p-2 bg-amber-50 md:bg-amber-100 rounded md:rounded-lg text-amber-600">
                                <AlertCircle className="w-3 h-3 md:w-5 md:h-5" />
                            </div>
                            <p className="text-[8px] md:text-sm font-medium text-stone-400 md:text-stone-500">Aberto</p>
                        </div>
                        <p className="text-xs md:text-3xl font-bold text-stone-800">{formatCurrency(currentStats.open)}</p>
                    </div>

                    <div className="bg-white/80 backdrop-blur-sm border border-stone-100 rounded-lg md:rounded-2xl p-2 md:p-6 relative overflow-hidden group hover:shadow-card transition-all">
                        <div className="hidden md:block absolute top-0 right-0 w-24 h-24 bg-stone-500/10 rounded-full blur-2xl -mr-6 -mt-6 transition-all group-hover:bg-stone-500/20"></div>
                        <div className="flex items-center gap-1 mb-1 md:mb-4">
                            <div className="p-0.5 md:p-2 bg-stone-50 md:bg-stone-100 rounded md:rounded-lg text-stone-600">
                                <Wallet className="w-3 h-3 md:w-5 md:h-5" />
                            </div>
                            <p className="text-[8px] md:text-sm font-medium text-stone-400 md:text-stone-500">Total</p>
                        </div>
                        <p className="text-xs md:text-3xl font-bold text-stone-800">{formatCurrency(currentStats.total)}</p>
                    </div>
                </div>

                {/* Charts Section */}
                <FinancialChart transactions={transactions} categories={categories.map(c => ({ ...c, color: c.color || '#78716c' }))} />
                <TrendChart transactions={transactions} year={year} />

                {/* Transactions Section */}
                <div className="bg-white/80 backdrop-blur-sm border border-stone-100 md:border-stone-200 rounded-xl md:rounded-3xl overflow-hidden">
                    <div className="p-2 md:p-6 border-b border-stone-100 md:border-stone-200 flex flex-row gap-2 md:gap-4 justify-between items-center">
                        <h2 className="text-xs md:text-lg font-semibold md:font-bold text-stone-700 md:text-stone-900">Transações</h2>
                        <div className="relative flex-1 max-w-[180px] md:max-w-[256px]">
                            <Search className="absolute left-2 md:left-3 top-1/2 -translate-y-1/2 w-3 h-3 md:w-4 md:h-4 text-stone-400" aria-hidden="true" />
                            <input
                                type="text"
                                placeholder="Buscar..."
                                value={searchTerm}
                                onChange={(e) => setSearchTerm(e.target.value)}
                                className="w-full bg-stone-50 border border-stone-100 md:border-stone-200 rounded-md md:rounded-xl py-1.5 md:py-2 pl-7 md:pl-10 pr-2 md:pr-4 text-[10px] md:text-sm text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-1 md:focus:ring-2 focus:ring-stone-300 md:focus:ring-stone-400"
                                aria-label="Buscar transações"
                            />
                        </div>
                    </div>

                    {/* Desktop Table View */}
                    <div className="overflow-x-auto hidden md:block">
                        <table className="w-full text-left border-collapse">
                            <thead>
                                <tr className="border-b border-stone-200 bg-stone-50/50">
                                    <th className="py-4 px-6 text-xs font-semibold uppercase tracking-wider text-stone-500">Descrição</th>
                                    <th className="py-4 px-6 text-xs font-semibold uppercase tracking-wider text-stone-500">Categoria</th>
                                    <th className="py-4 px-6 text-xs font-semibold uppercase tracking-wider text-stone-500">Data</th>
                                    <th className="py-4 px-6 text-xs font-semibold uppercase tracking-wider text-stone-500">Valor</th>
                                    <th className="py-4 px-6 text-xs font-semibold uppercase tracking-wider text-stone-500 text-center">Status</th>
                                    <th className="py-4 px-6 text-xs font-semibold uppercase tracking-wider text-stone-500 text-right">Ações</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-stone-100">
                                {filteredTransactions.map((t) => (
                                    <tr key={t.id} className="group hover:bg-stone-50/50 transition-colors">
                                        <td className="py-4 px-6">
                                            <p className="font-medium text-stone-900">{t.description || 'Sem descrição'}</p>
                                        </td>
                                        <td className="py-4 px-6">
                                            <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-stone-100 text-stone-700 border border-stone-200">
                                                {t.category}
                                            </span>
                                        </td>
                                        <td className="py-4 px-6 text-sm text-stone-500">
                                            {t.month} / {t.year}
                                        </td>
                                        <td className="py-4 px-6 font-medium text-stone-900">
                                            {formatCurrency(t.amount)}
                                        </td>
                                        <td className="py-4 px-6 text-center">
                                            <button
                                                onClick={() => toggleStatus(t.id)}
                                                className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-medium cursor-pointer transition-all ${t.status === Status.PAID
                                                    ? 'bg-green-100 text-green-700 border border-green-200 hover:bg-green-200'
                                                    : 'bg-amber-100 text-amber-700 border border-amber-200 hover:bg-amber-200'
                                                    }`}
                                                aria-label={`Mudar status de ${t.description} para ${t.status === Status.PAID ? 'aberto' : 'pago'}`}
                                            >
                                                {t.status === Status.PAID ? <CheckCircle2 className="w-3 h-3" aria-hidden="true" /> : <AlertCircle className="w-3 h-3" aria-hidden="true" />}
                                                {t.status === Status.PAID ? 'Pago' : 'Aberto'}
                                            </button>
                                        </td>
                                        <td className="py-4 px-6 text-right">
                                            <div className="flex items-center justify-end gap-2">
                                                <button
                                                    onClick={() => handleEditClick(t)}
                                                    className="text-stone-400 hover:text-blue-600 p-2 rounded-lg hover:bg-blue-50 transition-colors"
                                                    aria-label={`Editar transação ${t.description}`}
                                                >
                                                    <Edit2 className="w-4 h-4" />
                                                </button>
                                                <button
                                                    onClick={() => deleteTransaction(t.id)}
                                                    className="text-stone-400 hover:text-red-600 p-2 rounded-lg hover:bg-red-50 transition-colors"
                                                    aria-label={`Excluir transação ${t.description}`}
                                                >
                                                    <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>

                    {/* Mobile Card View - Ultra Compacto */}
                    <div className="md:hidden divide-y divide-stone-50">
                        {filteredTransactions.map((t) => (
                            <div key={t.id} className="px-2.5 py-2 active:bg-stone-50/50 transition-colors">
                                <div className="flex items-center gap-2">
                                    {/* Status indicator */}
                                    <button
                                        onClick={() => toggleStatus(t.id)}
                                        className={`shrink-0 p-1 rounded-md transition-all active:scale-95 ${t.status === Status.PAID
                                            ? 'bg-green-50 text-green-600'
                                            : 'bg-amber-50 text-amber-600'
                                            }`}
                                        aria-label={`Mudar status de ${t.description} para ${t.status === Status.PAID ? 'aberto' : 'pago'}`}
                                    >
                                        {t.status === Status.PAID ? <CheckCircle2 size={14} /> : <AlertCircle size={14} />}
                                    </button>

                                    {/* Info */}
                                    <div className="flex-1 min-w-0">
                                        <div className="flex items-center justify-between gap-1">
                                            <p className="font-medium text-xs text-stone-800 truncate">{t.description || 'Sem descrição'}</p>
                                            <p className="text-xs font-semibold text-stone-900 shrink-0">
                                                {formatCurrency(t.amount)}
                                            </p>
                                        </div>
                                        <div className="flex items-center gap-1.5 mt-0.5">
                                            <span className="text-[9px] px-1.5 py-0.5 rounded bg-stone-100 text-stone-600 font-medium">
                                                {t.category}
                                            </span>
                                            <span className="text-[9px] text-stone-400">
                                                {t.month}/{t.year}
                                            </span>
                                        </div>
                                    </div>

                                    {/* Actions */}
                                    <div className="flex items-center shrink-0">
                                        <button
                                            onClick={() => handleEditClick(t)}
                                            className="text-stone-400 active:text-blue-600 p-1.5 rounded-md active:bg-blue-50 transition-colors"
                                            aria-label={`Editar transação ${t.description}`}
                                        >
                                            <Edit2 size={14} />
                                        </button>
                                        <button
                                            onClick={() => deleteTransaction(t.id)}
                                            className="text-stone-400 active:text-red-600 p-1.5 rounded-md active:bg-red-50 transition-colors"
                                            aria-label={`Excluir transação ${t.description}`}
                                        >
                                            <svg width="14" height="14" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
                                        </button>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </main>

            {/* Modal */}
            <TransactionModal
                isOpen={showModal}
                onClose={handleModalClose}
                onSave={handleSaveTransaction}
                onCreateCategory={handleCreateCategory}
                transaction={editingTransaction}
                categories={categories}
                companyId={companyId!}
            />
        </div>
    );
};
