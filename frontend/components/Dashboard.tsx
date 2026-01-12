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
    Edit2,
    Trash2,
    Filter,
    X
} from 'lucide-react';
import { useTransactions } from '../hooks/useTransactions';
import { useDebounce } from '../hooks/useDebounce';
import { Transaction, Category, CategoryType, Status, PaginationInfo } from '../types';

import { api } from '../services/api';
import { FinancialChart } from './FinancialChart';
import { useDateFilter } from '../contexts/DateFilterContext';
import { DateSelector } from './DateSelector';
import { TrendChart } from './TrendChart';
import { TransactionModal } from './TransactionModal';
import { Pagination } from './Pagination';
import { formatCurrency } from '../utils/currency';

const DEFAULT_PAGE_SIZE = 20;

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
    const [clientPage, setClientPage] = useState(1);
    const [clientPageSize, setClientPageSize] = useState(DEFAULT_PAGE_SIZE);

    // Filter states
    const [statusFilter, setStatusFilter] = useState<Status | 'ALL'>('ALL');
    const [categoryFilter, setCategoryFilter] = useState<string>('ALL');
    const [typeFilter, setTypeFilter] = useState<CategoryType | 'ALL'>('ALL');
    const [showFilters, setShowFilters] = useState(false);

    // Get category type by name
    const getCategoryType = useCallback((categoryName: string): CategoryType | null => {
        const category = categories.find(c => c.name === categoryName);
        return category ? category.type : null;
    }, [categories]);

    // Filter transactions by date, search, status, category, and type
    const allFilteredTransactions = useMemo(() => {
        return transactions.filter(t => {
            const matchesSearch = t.description?.toLowerCase().includes(debouncedSearchTerm.toLowerCase()) ||
                t.category.toLowerCase().includes(debouncedSearchTerm.toLowerCase());
            const matchesDate = t.month === month && t.year === year;
            const matchesStatus = statusFilter === 'ALL' || t.status === statusFilter;
            const matchesCategory = categoryFilter === 'ALL' || t.category === categoryFilter;
            const matchesType = typeFilter === 'ALL' || getCategoryType(t.category) === typeFilter;
            return matchesSearch && matchesDate && matchesStatus && matchesCategory && matchesType;
        });
    }, [transactions, debouncedSearchTerm, month, year, statusFilter, categoryFilter, typeFilter, getCategoryType]);

    // Reset to page 1 when filters change
    useEffect(() => {
        setClientPage(1);
    }, [debouncedSearchTerm, month, year, statusFilter, categoryFilter, typeFilter]);

    // Clear all filters
    const clearFilters = useCallback(() => {
        setStatusFilter('ALL');
        setCategoryFilter('ALL');
        setTypeFilter('ALL');
        setSearchTerm('');
    }, []);

    // Check if any filter is active
    const hasActiveFilters = statusFilter !== 'ALL' || categoryFilter !== 'ALL' || typeFilter !== 'ALL' || searchTerm !== '';

    // Calculate client-side pagination
    const clientPagination: PaginationInfo = useMemo(() => ({
        page: clientPage,
        limit: clientPageSize,
        total: allFilteredTransactions.length,
        totalPages: Math.ceil(allFilteredTransactions.length / clientPageSize)
    }), [clientPage, clientPageSize, allFilteredTransactions.length]);

    // Get paginated transactions for current page
    const filteredTransactions = useMemo(() => {
        const startIndex = (clientPage - 1) * clientPageSize;
        return allFilteredTransactions.slice(startIndex, startIndex + clientPageSize);
    }, [allFilteredTransactions, clientPage, clientPageSize]);

    // Stats should be calculated from ALL filtered transactions, not just the current page
    const currentStats = useMemo(() => {
        return {
            paid: allFilteredTransactions.filter(t => t.status === Status.PAID).reduce((acc, t) => acc + t.amount, 0),
            open: allFilteredTransactions.filter(t => t.status === Status.OPEN).reduce((acc, t) => acc + t.amount, 0),
            total: allFilteredTransactions.reduce((acc, t) => acc + t.amount, 0)
        };
    }, [allFilteredTransactions]);

    const handlePageChange = useCallback((page: number) => {
        setClientPage(page);
    }, []);

    const handlePageSizeChange = useCallback((pageSize: number) => {
        setClientPageSize(pageSize);
        setClientPage(1);
    }, []);

    useEffect(() => {
        if (companyId) {
            api.getCompanies().then(companies => {
                const current = companies.find(c => c.id === companyId);
                if (current) setCompanyName(current.name);
            });
            api.getCategories(companyId).then(setCategories);
        }
    }, [companyId]);

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
        <div className="min-h-screen flex items-center justify-center bg-background">
            <div className="card p-8 flex flex-col items-center gap-4">
                <div className="w-10 h-10 border-4 border-primary-500 border-t-transparent rounded-full animate-spin"></div>
                <p className="text-muted">Carregando dados...</p>
            </div>
        </div>
    );

    return (
        <div className="min-h-screen bg-background">
            {/* Header - Desktop */}
            <header className="bg-card border-b border-border shadow-soft sticky top-0 z-40 hidden md:block">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-4">
                            <Link to="/" className="p-2 hover:bg-primary-50 rounded-xl transition-colors text-muted hover:text-foreground group">
                                <ArrowLeft className="w-5 h-5 group-hover:-translate-x-1 transition-transform" />
                            </Link>
                            <div className="flex items-center gap-3">
                                <div className="p-2 bg-primary-100 rounded-xl border border-primary-200">
                                    <Building2 className="w-5 h-5 text-primary-600" />
                                </div>
                                <h1 className="text-xl font-bold text-foreground">{companyName}</h1>
                            </div>
                            <div className="h-8 w-px bg-border mx-2"></div>
                            <DateSelector />
                        </div>

                        <button
                            onClick={() => setShowModal(true)}
                            className="btn-primary flex items-center gap-2"
                        >
                            <Plus className="w-4 h-4" />
                            Nova Transacao
                        </button>
                    </div>
                </div>
            </header>

            {/* Header - Mobile */}
            <header className="bg-card border-b border-border shadow-soft sticky top-0 z-40 md:hidden">
                <div className="px-3 py-2">
                    <div className="flex items-center justify-between">
                        <div className="flex items-center gap-1.5">
                            <Link to="/" className="p-1 -ml-1 rounded-md active:bg-primary-50 transition-colors">
                                <ArrowLeft size={16} className="text-muted" />
                            </Link>
                            <h1 className="text-sm font-semibold text-foreground truncate max-w-[120px]">{companyName}</h1>
                        </div>
                        <div className="flex items-center gap-2">
                            <DateSelector />
                            <button
                                onClick={() => setShowModal(true)}
                                className="bg-gradient-primary text-white p-1.5 rounded-lg shadow-soft active:scale-95 transition-all"
                            >
                                <Plus size={16} />
                            </button>
                        </div>
                    </div>
                </div>
            </header>

            <main className="max-w-7xl mx-auto px-2 md:px-4 sm:px-6 lg:px-8 py-2 md:py-8 mobile-content-padding">
                {/* Stats Grid */}
                <div className="grid grid-cols-3 md:grid-cols-3 gap-1.5 md:gap-6 mb-3 md:mb-8">
                    {/* Paid Card */}
                    <div className="stat-card group">
                        <div className="flex items-center gap-1 mb-1 md:mb-4">
                            <div className="p-0.5 md:p-2 bg-success-light rounded md:rounded-lg text-success border border-success/30">
                                <CheckCircle2 className="w-3 h-3 md:w-5 md:h-5" />
                            </div>
                            <p className="text-[8px] md:text-sm font-medium text-muted">Pago</p>
                        </div>
                        <p className="text-xs md:text-3xl font-bold text-foreground">{formatCurrency(currentStats.paid)}</p>
                    </div>

                    {/* Open Card */}
                    <div className="stat-card group">
                        <div className="flex items-center gap-1 mb-1 md:mb-4">
                            <div className="p-0.5 md:p-2 bg-warning-light rounded md:rounded-lg text-warning border border-warning/30">
                                <AlertCircle className="w-3 h-3 md:w-5 md:h-5" />
                            </div>
                            <p className="text-[8px] md:text-sm font-medium text-muted">Aberto</p>
                        </div>
                        <p className="text-xs md:text-3xl font-bold text-foreground">{formatCurrency(currentStats.open)}</p>
                    </div>

                    {/* Total Card */}
                    <div className="stat-card group">
                        <div className="flex items-center gap-1 mb-1 md:mb-4">
                            <div className="p-0.5 md:p-2 bg-primary-100 rounded md:rounded-lg text-primary-600 border border-primary-200">
                                <Wallet className="w-3 h-3 md:w-5 md:h-5" />
                            </div>
                            <p className="text-[8px] md:text-sm font-medium text-muted">Total</p>
                        </div>
                        <p className="text-xs md:text-3xl font-bold text-foreground">{formatCurrency(currentStats.total)}</p>
                    </div>
                </div>

                {/* Charts Section */}
                <FinancialChart transactions={transactions} categories={categories.map(c => ({ ...c, color: c.color || '#78716c' }))} />
                <TrendChart transactions={transactions} year={year} />

                {/* Transactions Section */}
                <div className="card overflow-hidden">
                    <div className="p-2 md:p-6 border-b border-border">
                        <div className="flex flex-row gap-2 md:gap-4 justify-between items-center">
                            <h2 className="text-xs md:text-lg font-semibold md:font-bold text-foreground">Transacoes</h2>
                            <div className="flex items-center gap-2">
                                <div className="relative flex-1 max-w-[140px] md:max-w-[200px]">
                                    <Search className="absolute left-2 md:left-3 top-1/2 -translate-y-1/2 w-3 h-3 md:w-4 md:h-4 text-muted" />
                                    <input
                                        type="text"
                                        placeholder="Buscar..."
                                        value={searchTerm}
                                        onChange={(e) => setSearchTerm(e.target.value)}
                                        className="w-full input py-1.5 md:py-2 pl-7 md:pl-10 pr-2 md:pr-4 text-[10px] md:text-sm"
                                    />
                                </div>
                                <button
                                    onClick={() => setShowFilters(!showFilters)}
                                    className={`p-1.5 md:p-2 rounded-lg transition-colors ${showFilters || hasActiveFilters ? 'bg-primary-100 text-primary-600' : 'hover:bg-primary-50 text-muted'}`}
                                    title="Filtros"
                                >
                                    <Filter className="w-3.5 h-3.5 md:w-4 md:h-4" />
                                </button>
                                {hasActiveFilters && (
                                    <button
                                        onClick={clearFilters}
                                        className="p-1.5 md:p-2 rounded-lg hover:bg-destructive-light text-muted hover:text-destructive transition-colors"
                                        title="Limpar filtros"
                                    >
                                        <X className="w-3.5 h-3.5 md:w-4 md:h-4" />
                                    </button>
                                )}
                            </div>
                        </div>

                        {/* Filter Panel */}
                        {showFilters && (
                            <div className="mt-3 pt-3 border-t border-border grid grid-cols-3 gap-2 md:gap-4">
                                {/* Status Filter */}
                                <div>
                                    <label className="block text-[9px] md:text-xs font-medium text-muted mb-1">Status</label>
                                    <select
                                        value={statusFilter}
                                        onChange={(e) => setStatusFilter(e.target.value as Status | 'ALL')}
                                        className="w-full input py-1 md:py-2 text-[10px] md:text-sm"
                                    >
                                        <option value="ALL">Todos</option>
                                        <option value={Status.PAID}>Pago</option>
                                        <option value={Status.OPEN}>Aberto</option>
                                    </select>
                                </div>

                                {/* Category Filter */}
                                <div>
                                    <label className="block text-[9px] md:text-xs font-medium text-muted mb-1">Categoria</label>
                                    <select
                                        value={categoryFilter}
                                        onChange={(e) => setCategoryFilter(e.target.value)}
                                        className="w-full input py-1 md:py-2 text-[10px] md:text-sm"
                                    >
                                        <option value="ALL">Todas</option>
                                        {categories.map(cat => (
                                            <option key={cat.id} value={cat.name}>{cat.name}</option>
                                        ))}
                                    </select>
                                </div>

                                {/* Type Filter */}
                                <div>
                                    <label className="block text-[9px] md:text-xs font-medium text-muted mb-1">Tipo</label>
                                    <select
                                        value={typeFilter}
                                        onChange={(e) => setTypeFilter(e.target.value as CategoryType | 'ALL')}
                                        className="w-full input py-1 md:py-2 text-[10px] md:text-sm"
                                    >
                                        <option value="ALL">Todos</option>
                                        <option value={CategoryType.EXPENSE}>Despesa</option>
                                        <option value={CategoryType.INCOME}>Receita</option>
                                    </select>
                                </div>
                            </div>
                        )}

                        {/* Active Filters Badge */}
                        {hasActiveFilters && !showFilters && (
                            <div className="mt-2 flex flex-wrap gap-1.5">
                                {statusFilter !== 'ALL' && (
                                    <span className="text-[9px] md:text-xs px-2 py-0.5 bg-primary-100 text-primary-700 rounded-full">
                                        Status: {statusFilter === Status.PAID ? 'Pago' : 'Aberto'}
                                    </span>
                                )}
                                {categoryFilter !== 'ALL' && (
                                    <span className="text-[9px] md:text-xs px-2 py-0.5 bg-primary-100 text-primary-700 rounded-full">
                                        Categoria: {categoryFilter}
                                    </span>
                                )}
                                {typeFilter !== 'ALL' && (
                                    <span className="text-[9px] md:text-xs px-2 py-0.5 bg-primary-100 text-primary-700 rounded-full">
                                        Tipo: {typeFilter === CategoryType.EXPENSE ? 'Despesa' : 'Receita'}
                                    </span>
                                )}
                                {searchTerm && (
                                    <span className="text-[9px] md:text-xs px-2 py-0.5 bg-primary-100 text-primary-700 rounded-full">
                                        Busca: {searchTerm}
                                    </span>
                                )}
                            </div>
                        )}
                    </div>

                    {/* Desktop Table View */}
                    <div className="overflow-x-auto hidden md:block">
                        <table className="w-full text-left border-collapse">
                            <thead>
                                <tr className="table-header">
                                    <th className="py-4 px-6">Descricao</th>
                                    <th className="py-4 px-6">Categoria</th>
                                    <th className="py-4 px-6">Data</th>
                                    <th className="py-4 px-6">Valor</th>
                                    <th className="py-4 px-6 text-center">Status</th>
                                    <th className="py-4 px-6 text-right">Acoes</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-border">
                                {filteredTransactions.map((t) => (
                                    <tr key={t.id} className="table-row group">
                                        <td className="py-4 px-6">
                                            <p className="font-medium text-foreground">{t.description || 'Sem descricao'}</p>
                                        </td>
                                        <td className="py-4 px-6">
                                            <span className="badge badge-primary">{t.category}</span>
                                        </td>
                                        <td className="py-4 px-6 text-sm text-muted">{t.month} / {t.year}</td>
                                        <td className="py-4 px-6 font-medium text-foreground">{formatCurrency(t.amount)}</td>
                                        <td className="py-4 px-6 text-center">
                                            <button
                                                onClick={() => toggleStatus(t.id)}
                                                className={`badge cursor-pointer transition-all ${
                                                    t.status === Status.PAID ? 'badge-success' : 'badge-warning'
                                                }`}
                                            >
                                                {t.status === Status.PAID ? <CheckCircle2 className="w-3 h-3 mr-1" /> : <AlertCircle className="w-3 h-3 mr-1" />}
                                                {t.status === Status.PAID ? 'Pago' : 'Aberto'}
                                            </button>
                                        </td>
                                        <td className="py-4 px-6 text-right">
                                            <div className="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                                                <button
                                                    onClick={() => handleEditClick(t)}
                                                    className="text-muted hover:text-primary-600 p-2 rounded-lg hover:bg-primary-50 transition-colors"
                                                >
                                                    <Edit2 className="w-4 h-4" />
                                                </button>
                                                <button
                                                    onClick={() => deleteTransaction(t.id)}
                                                    className="text-muted hover:text-destructive p-2 rounded-lg hover:bg-destructive-light transition-colors"
                                                >
                                                    <Trash2 className="w-4 h-4" />
                                                </button>
                                            </div>
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>

                    {/* Mobile Card View */}
                    <div className="md:hidden divide-y divide-border">
                        {filteredTransactions.map((t) => (
                            <div key={t.id} className="px-2.5 py-2 active:bg-primary-50/50 transition-colors">
                                <div className="flex items-center gap-2">
                                    <button
                                        onClick={() => toggleStatus(t.id)}
                                        className={`shrink-0 p-1 rounded-md transition-all active:scale-95 ${
                                            t.status === Status.PAID
                                                ? 'bg-success-light text-success border border-success/30'
                                                : 'bg-warning-light text-warning border border-warning/30'
                                        }`}
                                    >
                                        {t.status === Status.PAID ? <CheckCircle2 size={14} /> : <AlertCircle size={14} />}
                                    </button>

                                    <div className="flex-1 min-w-0">
                                        <div className="flex items-center justify-between gap-1">
                                            <p className="font-medium text-xs text-foreground truncate">{t.description || 'Sem descricao'}</p>
                                            <p className="text-xs font-semibold text-foreground shrink-0">{formatCurrency(t.amount)}</p>
                                        </div>
                                        <div className="flex items-center gap-1.5 mt-0.5">
                                            <span className="text-[9px] px-1.5 py-0.5 rounded bg-primary-100 text-primary-700 font-medium border border-primary-200">
                                                {t.category}
                                            </span>
                                            <span className="text-[9px] text-muted">{t.month}/{t.year}</span>
                                        </div>
                                    </div>

                                    <div className="flex items-center shrink-0">
                                        <button
                                            onClick={() => handleEditClick(t)}
                                            className="text-muted active:text-primary-600 p-1.5 rounded-md active:bg-primary-50 transition-colors"
                                        >
                                            <Edit2 size={14} />
                                        </button>
                                        <button
                                            onClick={() => deleteTransaction(t.id)}
                                            className="text-muted active:text-destructive p-1.5 rounded-md active:bg-destructive-light transition-colors"
                                        >
                                            <Trash2 size={14} />
                                        </button>
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>

                    {/* Empty State */}
                    {allFilteredTransactions.length === 0 && (
                        <div className="py-12 text-center">
                            <div className="inline-flex p-4 bg-primary-50 rounded-2xl border border-primary-100 mb-4">
                                <Wallet className="w-8 h-8 text-primary-400" />
                            </div>
                            <p className="text-muted text-sm">Nenhuma transacao encontrada</p>
                            <button
                                onClick={() => setShowModal(true)}
                                className="mt-4 text-primary-600 hover:text-primary-700 text-sm font-medium transition-colors"
                            >
                                + Adicionar primeira transacao
                            </button>
                        </div>
                    )}

                    {/* Pagination */}
                    {allFilteredTransactions.length > 0 && (
                        <div className="border-t border-border">
                            <Pagination
                                pagination={clientPagination}
                                onPageChange={handlePageChange}
                                onPageSizeChange={handlePageSizeChange}
                                pageSizeOptions={[10, 20, 50, 100]}
                                disabled={loading}
                            />
                        </div>
                    )}
                </div>
            </main>

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
