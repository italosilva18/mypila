import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../services/api';
import { Quote, QuoteStatus, Category } from '../types';
import {
    Plus, Trash2, FileText, Loader2, Edit2, Copy, Download,
    Eye, Send, CheckCircle, XCircle, Play, Filter, Search
} from 'lucide-react';
import { QuoteModal } from '../components/QuoteModal';

const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value);
};

const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('pt-BR');
};

const statusConfig: Record<QuoteStatus, { label: string; color: string; bgColor: string; icon: React.ReactNode }> = {
    [QuoteStatus.DRAFT]: {
        label: 'Rascunho',
        color: 'text-stone-600',
        bgColor: 'bg-stone-100',
        icon: <Edit2 className="w-3 h-3" />
    },
    [QuoteStatus.SENT]: {
        label: 'Enviado',
        color: 'text-blue-600',
        bgColor: 'bg-blue-100',
        icon: <Send className="w-3 h-3" />
    },
    [QuoteStatus.APPROVED]: {
        label: 'Aprovado',
        color: 'text-green-600',
        bgColor: 'bg-green-100',
        icon: <CheckCircle className="w-3 h-3" />
    },
    [QuoteStatus.REJECTED]: {
        label: 'Rejeitado',
        color: 'text-red-600',
        bgColor: 'bg-red-100',
        icon: <XCircle className="w-3 h-3" />
    },
    [QuoteStatus.EXECUTED]: {
        label: 'Executado',
        color: 'text-purple-600',
        bgColor: 'bg-purple-100',
        icon: <Play className="w-3 h-3" />
    },
};

export const Quotes: React.FC = () => {
    const { companyId } = useParams<{ companyId: string }>();
    const navigate = useNavigate();
    const [quotes, setQuotes] = useState<Quote[]>([]);
    const [categories, setCategories] = useState<Category[]>([]);
    const [loading, setLoading] = useState(true);
    const [isModalOpen, setIsModalOpen] = useState(false);
    const [editingQuote, setEditingQuote] = useState<Quote | null>(null);
    const [filterStatus, setFilterStatus] = useState<QuoteStatus | ''>('');
    const [searchTerm, setSearchTerm] = useState('');

    const loadData = useCallback(async () => {
        if (!companyId) return;
        try {
            setLoading(true);
            const [quotesData, categoriesData] = await Promise.all([
                api.getQuotes(companyId),
                api.getCategories(companyId)
            ]);
            setQuotes(quotesData);
            setCategories(categoriesData);
        } catch (err) {
            console.error('Failed to load data', err);
        } finally {
            setLoading(false);
        }
    }, [companyId]);

    useEffect(() => {
        loadData();
    }, [loadData]);

    const handleCreate = () => {
        setEditingQuote(null);
        setIsModalOpen(true);
    };

    const handleEdit = (quote: Quote) => {
        setEditingQuote(quote);
        setIsModalOpen(true);
    };

    const handleModalClose = () => {
        setIsModalOpen(false);
        setEditingQuote(null);
    };

    const handleModalSave = async () => {
        await loadData();
        handleModalClose();
    };

    const handleDelete = async (id: string) => {
        if (!confirm('Tem certeza que deseja excluir este orcamento?')) return;
        try {
            await api.deleteQuote(id);
            setQuotes(quotes.filter(q => q.id !== id));
        } catch (err) {
            console.error('Failed to delete quote', err);
        }
    };

    const handleDuplicate = async (id: string) => {
        try {
            const duplicated = await api.duplicateQuote(id);
            setQuotes([duplicated, ...quotes]);
        } catch (err) {
            console.error('Failed to duplicate quote', err);
        }
    };

    const handleDownloadPDF = async (id: string, number: string) => {
        try {
            const blob = await api.downloadQuotePDF(id);
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `${number}.pdf`;
            document.body.appendChild(a);
            a.click();
            window.URL.revokeObjectURL(url);
            document.body.removeChild(a);
        } catch (err) {
            console.error('Failed to download PDF', err);
        }
    };

    const _handleStatusChange = async (id: string, newStatus: QuoteStatus) => {
        try {
            const updated = await api.updateQuoteStatus(id, newStatus);
            setQuotes(quotes.map(q => q.id === id ? updated : q));
        } catch (err) {
            console.error('Failed to update status', err);
        }
    };
    void _handleStatusChange;

    const handleViewComparison = (quoteId: string) => {
        navigate(`/company/${companyId}/quotes/${quoteId}/comparison`);
    };

    const filteredQuotes = quotes.filter(quote => {
        const matchesStatus = !filterStatus || quote.status === filterStatus;
        const matchesSearch = !searchTerm ||
            quote.clientName.toLowerCase().includes(searchTerm.toLowerCase()) ||
            quote.number.toLowerCase().includes(searchTerm.toLowerCase()) ||
            quote.title.toLowerCase().includes(searchTerm.toLowerCase());
        return matchesStatus && matchesSearch;
    });

    const StatusBadge: React.FC<{ status: QuoteStatus }> = ({ status }) => {
        const config = statusConfig[status];
        return (
            <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${config.color} ${config.bgColor}`}>
                {config.icon}
                {config.label}
            </span>
        );
    };

    return (
        <div className="space-y-3 md:space-y-6 mobile-content-padding">
            {/* Header */}
            <header className="px-1 md:px-0 flex flex-col md:flex-row md:items-center md:justify-between gap-3">
                <div>
                    <h1 className="text-base md:text-2xl font-bold text-stone-900 flex items-center gap-1.5 md:gap-2">
                        <FileText className="w-4 h-4 md:w-6 md:h-6 text-stone-600" />
                        Orcamentos
                    </h1>
                    <p className="text-stone-500 text-xs md:text-sm">Crie e gerencie orcamentos para seus clientes.</p>
                </div>
                <button
                    onClick={handleCreate}
                    className="flex items-center gap-2 bg-stone-800 hover:bg-stone-700 text-white px-4 py-2.5 rounded-xl text-sm font-medium transition-all shadow-lg shadow-stone-900/20"
                >
                    <Plus className="w-4 h-4" />
                    Novo Orcamento
                </button>
            </header>

            {/* Filters */}
            <div className="bg-white/70 border border-stone-100 md:border-stone-200 rounded-lg md:rounded-2xl p-3 md:p-4">
                <div className="flex flex-col md:flex-row gap-3">
                    <div className="flex-1 relative">
                        <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-stone-400" />
                        <input
                            type="text"
                            placeholder="Buscar por cliente, numero ou titulo..."
                            value={searchTerm}
                            onChange={(e) => setSearchTerm(e.target.value)}
                            className="w-full pl-10 pr-4 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 focus:ring-stone-400"
                        />
                    </div>
                    <div className="flex items-center gap-2">
                        <Filter className="w-4 h-4 text-stone-400" />
                        <select
                            value={filterStatus}
                            onChange={(e) => setFilterStatus(e.target.value as QuoteStatus | '')}
                            className="px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm text-stone-900 focus:outline-none focus:ring-2 focus:ring-stone-400"
                        >
                            <option value="">Todos os status</option>
                            {Object.entries(statusConfig).map(([status, config]) => (
                                <option key={status} value={status}>{config.label}</option>
                            ))}
                        </select>
                    </div>
                </div>
            </div>

            {/* Content */}
            {loading ? (
                <div className="flex items-center justify-center py-12">
                    <Loader2 className="w-6 h-6 text-stone-400 animate-spin" />
                </div>
            ) : filteredQuotes.length === 0 ? (
                <div className="bg-white/70 border border-stone-100 md:border-stone-200 rounded-lg md:rounded-2xl p-8 text-center">
                    <FileText className="w-12 h-12 text-stone-300 mx-auto mb-3" />
                    <p className="text-stone-500">
                        {quotes.length === 0
                            ? 'Nenhum orcamento cadastrado ainda.'
                            : 'Nenhum orcamento encontrado com os filtros aplicados.'}
                    </p>
                    {quotes.length === 0 && (
                        <button
                            onClick={handleCreate}
                            className="mt-4 text-stone-600 hover:text-stone-900 text-sm font-medium"
                        >
                            Criar primeiro orcamento
                        </button>
                    )}
                </div>
            ) : (
                <>
                    {/* Desktop Table */}
                    <div className="hidden md:block bg-white/70 border border-stone-200 rounded-2xl overflow-hidden">
                        <table className="w-full text-left">
                            <thead className="bg-stone-50 text-stone-500 text-xs uppercase font-semibold">
                                <tr>
                                    <th className="px-6 py-4">Numero</th>
                                    <th className="px-6 py-4">Cliente</th>
                                    <th className="px-6 py-4">Titulo</th>
                                    <th className="px-6 py-4">Valor</th>
                                    <th className="px-6 py-4">Status</th>
                                    <th className="px-6 py-4">Validade</th>
                                    <th className="px-6 py-4 text-right">Acoes</th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-stone-100">
                                {filteredQuotes.map((quote) => (
                                    <tr key={quote.id} className="hover:bg-stone-50/50 transition-colors">
                                        <td className="px-6 py-4 font-mono text-sm text-stone-600">{quote.number}</td>
                                        <td className="px-6 py-4 text-stone-900 font-medium">{quote.clientName}</td>
                                        <td className="px-6 py-4 text-stone-600 max-w-xs truncate">{quote.title}</td>
                                        <td className="px-6 py-4 text-stone-900 font-medium">{formatCurrency(quote.total)}</td>
                                        <td className="px-6 py-4">
                                            <StatusBadge status={quote.status} />
                                        </td>
                                        <td className="px-6 py-4 text-stone-500 text-sm">{formatDate(quote.validUntil)}</td>
                                        <td className="px-6 py-4">
                                            <div className="flex items-center justify-end gap-1">
                                                <button
                                                    onClick={() => handleEdit(quote)}
                                                    className="p-2 text-stone-400 hover:text-stone-700 hover:bg-stone-100 rounded-lg transition-colors"
                                                    title="Editar"
                                                >
                                                    <Edit2 className="w-4 h-4" />
                                                </button>
                                                <button
                                                    onClick={() => handleDuplicate(quote.id)}
                                                    className="p-2 text-stone-400 hover:text-blue-600 hover:bg-blue-50 rounded-lg transition-colors"
                                                    title="Duplicar"
                                                >
                                                    <Copy className="w-4 h-4" />
                                                </button>
                                                <button
                                                    onClick={() => handleDownloadPDF(quote.id, quote.number)}
                                                    className="p-2 text-stone-400 hover:text-green-600 hover:bg-green-50 rounded-lg transition-colors"
                                                    title="Baixar PDF"
                                                >
                                                    <Download className="w-4 h-4" />
                                                </button>
                                                {quote.status === QuoteStatus.EXECUTED && (
                                                    <button
                                                        onClick={() => handleViewComparison(quote.id)}
                                                        className="p-2 text-stone-400 hover:text-purple-600 hover:bg-purple-50 rounded-lg transition-colors"
                                                        title="Ver Comparativo"
                                                    >
                                                        <Eye className="w-4 h-4" />
                                                    </button>
                                                )}
                                                <button
                                                    onClick={() => handleDelete(quote.id)}
                                                    className="p-2 text-stone-400 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                                                    title="Excluir"
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

                    {/* Mobile Cards */}
                    <div className="md:hidden space-y-3">
                        {filteredQuotes.map((quote) => (
                            <div key={quote.id} className="bg-white/70 border border-stone-100 rounded-xl p-4">
                                <div className="flex items-start justify-between mb-3">
                                    <div>
                                        <span className="font-mono text-xs text-stone-500">{quote.number}</span>
                                        <h3 className="font-medium text-stone-900">{quote.clientName}</h3>
                                        <p className="text-xs text-stone-500 truncate max-w-[200px]">{quote.title}</p>
                                    </div>
                                    <StatusBadge status={quote.status} />
                                </div>

                                <div className="flex items-center justify-between mb-3">
                                    <span className="text-lg font-bold text-stone-900">{formatCurrency(quote.total)}</span>
                                    <span className="text-xs text-stone-500">Valido ate {formatDate(quote.validUntil)}</span>
                                </div>

                                <div className="flex items-center gap-2 pt-3 border-t border-stone-100">
                                    <button
                                        onClick={() => handleEdit(quote)}
                                        className="flex-1 flex items-center justify-center gap-1 py-2 text-stone-600 bg-stone-100 rounded-lg text-sm"
                                    >
                                        <Edit2 className="w-3.5 h-3.5" />
                                        Editar
                                    </button>
                                    <button
                                        onClick={() => handleDownloadPDF(quote.id, quote.number)}
                                        className="flex-1 flex items-center justify-center gap-1 py-2 text-green-600 bg-green-50 rounded-lg text-sm"
                                    >
                                        <Download className="w-3.5 h-3.5" />
                                        PDF
                                    </button>
                                    <button
                                        onClick={() => handleDelete(quote.id)}
                                        className="p-2 text-red-600 bg-red-50 rounded-lg"
                                    >
                                        <Trash2 className="w-4 h-4" />
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                </>
            )}

            {/* Modal */}
            {isModalOpen && (
                <QuoteModal
                    isOpen={isModalOpen}
                    onClose={handleModalClose}
                    onSave={handleModalSave}
                    quote={editingQuote}
                    categories={categories}
                    companyId={companyId!}
                />
            )}
        </div>
    );
};
