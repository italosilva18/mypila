import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { api } from '../services/api';
import { QuoteComparison as QuoteComparisonType, Quote } from '../types';
import { ArrowLeft, TrendingUp, TrendingDown, Minus, Loader2, BarChart3 } from 'lucide-react';

const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value);
};

const formatPercent = (value: number) => {
    const sign = value > 0 ? '+' : '';
    return `${sign}${value.toFixed(1)}%`;
};

export const QuoteComparisonPage: React.FC = () => {
    const { companyId, quoteId } = useParams<{ companyId: string; quoteId: string }>();
    const navigate = useNavigate();
    const [loading, setLoading] = useState(true);
    const [quote, setQuote] = useState<Quote | null>(null);
    const [comparison, setComparison] = useState<QuoteComparisonType | null>(null);

    useEffect(() => {
        const loadData = async () => {
            if (!quoteId) return;
            try {
                setLoading(true);
                const [quoteData, comparisonData] = await Promise.all([
                    api.getQuote(quoteId),
                    api.getQuoteComparison(quoteId)
                ]);
                setQuote(quoteData);
                setComparison(comparisonData);
            } catch (err) {
                console.error('Failed to load comparison data', err);
            } finally {
                setLoading(false);
            }
        };
        loadData();
    }, [quoteId]);

    const getVarianceColor = (variance: number) => {
        if (variance > 0) return 'text-green-600';
        if (variance < 0) return 'text-red-600';
        return 'text-stone-500';
    };

    const getVarianceBg = (variance: number) => {
        if (variance > 0) return 'bg-green-50';
        if (variance < 0) return 'bg-red-50';
        return 'bg-stone-50';
    };

    const getVarianceIcon = (variance: number) => {
        if (variance > 0) return <TrendingUp className="w-4 h-4" />;
        if (variance < 0) return <TrendingDown className="w-4 h-4" />;
        return <Minus className="w-4 h-4" />;
    };

    if (loading) {
        return (
            <div className="flex items-center justify-center min-h-[400px]">
                <Loader2 className="w-6 h-6 text-stone-400 animate-spin" />
            </div>
        );
    }

    if (!quote || !comparison) {
        return (
            <div className="text-center py-12">
                <p className="text-stone-500">Orcamento nao encontrado</p>
            </div>
        );
    }

    return (
        <div className="space-y-6 mobile-content-padding">
            {/* Header */}
            <header className="px-1 md:px-0">
                <button
                    onClick={() => navigate(`/company/${companyId}/quotes`)}
                    className="flex items-center gap-2 text-stone-600 hover:text-stone-900 mb-4 text-sm"
                >
                    <ArrowLeft className="w-4 h-4" />
                    Voltar para Orcamentos
                </button>

                <h1 className="text-base md:text-2xl font-bold text-stone-900 flex items-center gap-2">
                    <BarChart3 className="w-5 h-5 md:w-6 md:h-6 text-stone-600" />
                    Comparativo: Orcado vs Realizado
                </h1>
                <p className="text-stone-500 text-xs md:text-sm mt-1">
                    {quote.number} - {quote.title}
                </p>
            </header>

            {/* Summary Cards */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="bg-white border border-stone-200 rounded-2xl p-5">
                    <p className="text-sm text-stone-500 mb-1">Valor Orcado</p>
                    <p className="text-2xl font-bold text-stone-900">{formatCurrency(comparison.quotedTotal)}</p>
                </div>

                <div className="bg-white border border-stone-200 rounded-2xl p-5">
                    <p className="text-sm text-stone-500 mb-1">Valor Executado</p>
                    <p className="text-2xl font-bold text-stone-900">{formatCurrency(comparison.executedTotal)}</p>
                </div>

                <div className={`border rounded-2xl p-5 ${getVarianceBg(comparison.variance)}`}>
                    <p className="text-sm text-stone-500 mb-1">Variacao</p>
                    <div className="flex items-center gap-2">
                        <span className={`text-2xl font-bold ${getVarianceColor(comparison.variance)}`}>
                            {formatCurrency(Math.abs(comparison.variance))}
                        </span>
                        <span className={`flex items-center gap-1 text-sm ${getVarianceColor(comparison.variance)}`}>
                            {getVarianceIcon(comparison.variance)}
                            {formatPercent(comparison.variancePercent)}
                        </span>
                    </div>
                    <p className="text-xs text-stone-500 mt-1">
                        {comparison.variance > 0 ? 'Economia' : comparison.variance < 0 ? 'Excedente' : 'Dentro do orcamento'}
                    </p>
                </div>
            </div>

            {/* Items Comparison */}
            <div className="bg-white border border-stone-200 rounded-2xl overflow-hidden">
                <div className="p-4 md:p-6 border-b border-stone-200 bg-stone-50">
                    <h2 className="font-semibold text-stone-900">Comparativo por Item</h2>
                </div>

                {/* Desktop Table */}
                <div className="hidden md:block">
                    <table className="w-full">
                        <thead className="bg-stone-50 text-stone-500 text-xs uppercase">
                            <tr>
                                <th className="px-6 py-3 text-left font-semibold">Item</th>
                                <th className="px-6 py-3 text-right font-semibold">Orcado</th>
                                <th className="px-6 py-3 text-right font-semibold">Executado</th>
                                <th className="px-6 py-3 text-right font-semibold">Variacao</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-stone-100">
                            {comparison.items.map((item, index) => (
                                <tr key={index} className="hover:bg-stone-50/50">
                                    <td className="px-6 py-4 text-stone-900">{item.description}</td>
                                    <td className="px-6 py-4 text-right text-stone-600">{formatCurrency(item.quoted)}</td>
                                    <td className="px-6 py-4 text-right text-stone-600">{formatCurrency(item.executed)}</td>
                                    <td className="px-6 py-4 text-right">
                                        <div className={`inline-flex items-center gap-1 ${getVarianceColor(item.variance)}`}>
                                            {getVarianceIcon(item.variance)}
                                            <span>{formatCurrency(Math.abs(item.variance))}</span>
                                        </div>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>

                {/* Mobile Cards */}
                <div className="md:hidden divide-y divide-stone-100">
                    {comparison.items.map((item, index) => (
                        <div key={index} className="p-4">
                            <p className="font-medium text-stone-900 mb-3">{item.description}</p>
                            <div className="grid grid-cols-3 gap-2 text-sm">
                                <div>
                                    <p className="text-stone-500 text-xs">Orcado</p>
                                    <p className="text-stone-900">{formatCurrency(item.quoted)}</p>
                                </div>
                                <div>
                                    <p className="text-stone-500 text-xs">Executado</p>
                                    <p className="text-stone-900">{formatCurrency(item.executed)}</p>
                                </div>
                                <div>
                                    <p className="text-stone-500 text-xs">Variacao</p>
                                    <div className={`flex items-center gap-1 ${getVarianceColor(item.variance)}`}>
                                        {getVarianceIcon(item.variance)}
                                        <span>{formatCurrency(Math.abs(item.variance))}</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>

            {/* Visual Chart - Simple Bar Representation */}
            <div className="bg-white border border-stone-200 rounded-2xl p-4 md:p-6">
                <h2 className="font-semibold text-stone-900 mb-4">Visualizacao</h2>
                <div className="space-y-4">
                    {comparison.items.map((item, index) => {
                        const maxValue = Math.max(item.quoted, item.executed, 1);
                        const quotedPercent = (item.quoted / maxValue) * 100;
                        const executedPercent = (item.executed / maxValue) * 100;

                        return (
                            <div key={index}>
                                <p className="text-sm text-stone-600 mb-2">{item.description}</p>
                                <div className="space-y-1">
                                    <div className="flex items-center gap-2">
                                        <span className="text-xs text-stone-500 w-16">Orcado</span>
                                        <div className="flex-1 bg-stone-100 rounded-full h-4 overflow-hidden">
                                            <div
                                                className="bg-blue-500 h-full rounded-full transition-all"
                                                style={{ width: `${quotedPercent}%` }}
                                            />
                                        </div>
                                        <span className="text-xs text-stone-600 w-24 text-right">{formatCurrency(item.quoted)}</span>
                                    </div>
                                    <div className="flex items-center gap-2">
                                        <span className="text-xs text-stone-500 w-16">Real</span>
                                        <div className="flex-1 bg-stone-100 rounded-full h-4 overflow-hidden">
                                            <div
                                                className={`h-full rounded-full transition-all ${item.executed <= item.quoted ? 'bg-green-500' : 'bg-red-500'}`}
                                                style={{ width: `${executedPercent}%` }}
                                            />
                                        </div>
                                        <span className="text-xs text-stone-600 w-24 text-right">{formatCurrency(item.executed)}</span>
                                    </div>
                                </div>
                            </div>
                        );
                    })}
                </div>
            </div>
        </div>
    );
};
