import React, { useState, useMemo } from 'react';
import { useParams } from 'react-router-dom';
import { useTransactions } from '../hooks/useTransactions';
import { Status } from '../types';
import { FileText, Printer, CheckCircle2, AlertCircle, Wallet, Download } from 'lucide-react';
import { exportToCSV } from '../utils/export';
import { formatCurrency } from '../utils/currency';

export const Reports: React.FC = () => {
    const { companyId } = useParams<{ companyId: string }>();
    const { transactions, loading } = useTransactions(companyId!);

    const [statusFilter, setStatusFilter] = useState<'ALL' | Status>('ALL');
    const [monthFilter, setMonthFilter] = useState<string>('ALL');

    const filteredTransactions = useMemo(() => {
        return transactions.filter(t => {
            const statusMatch = statusFilter === 'ALL' || t.status === statusFilter;
            const monthMatch = monthFilter === 'ALL' || t.month === monthFilter;
            return statusMatch && monthMatch;
        });
    }, [transactions, statusFilter, monthFilter]);

    const totalValue = useMemo(() => {
        return filteredTransactions.reduce((acc, t) => acc + t.amount, 0);
    }, [filteredTransactions]);

    const handlePrint = () => {
        window.print();
    };

    if (loading) return <div className="text-stone-900">Carregando relatórios...</div>;

    return (
        <div className="space-y-3 md:space-y-6 mobile-content-padding">
            <header className="flex flex-col md:flex-row md:items-center justify-between gap-3 no-print px-1 md:px-0">
                <div>
                    <h1 className="text-base md:text-2xl font-bold text-stone-900 flex items-center gap-1.5 md:gap-2">
                        <FileText className="w-4 h-4 md:w-6 md:h-6 text-stone-600" />
                        Relatórios
                    </h1>
                    <p className="text-stone-500 text-xs md:text-sm">Relatórios de pagamentos e pendências.</p>
                </div>
                <div className="flex gap-2">
                    <button
                        onClick={handlePrint}
                        className="flex items-center gap-1.5 bg-stone-100 active:bg-stone-200 text-stone-700 px-3 md:px-4 py-2 rounded-lg md:rounded-xl text-xs md:text-sm transition-all"
                    >
                        <Printer className="w-3.5 h-3.5 md:w-4 md:h-4" />
                        <span className="hidden sm:inline">Imprimir</span>
                    </button>
                    <button
                        onClick={() => exportToCSV(filteredTransactions, `relatorio_financeiro_${new Date().toISOString().split('T')[0]}`)}
                        className="flex items-center gap-1.5 bg-stone-800 active:bg-stone-700 text-white px-3 md:px-4 py-2 rounded-lg md:rounded-xl text-xs md:text-sm transition-all shadow-lg shadow-stone-900/20"
                    >
                        <Download className="w-3.5 h-3.5 md:w-4 md:h-4" />
                        CSV
                    </button>
                </div>
            </header>

            {/* Filters */}
            <div className="grid grid-cols-2 md:grid-cols-3 gap-2 md:gap-4 no-print">
                <div className="bg-white/70 border border-stone-100 md:border-stone-200 rounded-lg md:rounded-xl p-2.5 md:p-4">
                    <label className="block text-[9px] md:text-xs font-semibold text-stone-500 uppercase tracking-wider mb-1.5 md:mb-2">Status</label>
                    <div className="flex gap-1 md:gap-2">
                        <button
                            onClick={() => setStatusFilter('ALL')}
                            className={`flex-1 py-1.5 md:py-2 px-2 md:px-3 rounded-md md:rounded-lg text-[10px] md:text-sm font-medium transition-all ${statusFilter === 'ALL' ? 'bg-stone-800 text-white' : 'bg-stone-100 text-stone-600'}`}
                        >
                            Total
                        </button>
                        <button
                            onClick={() => setStatusFilter(Status.OPEN)}
                            className={`flex-1 py-1.5 md:py-2 px-2 md:px-3 rounded-md md:rounded-lg text-[10px] md:text-sm font-medium transition-all ${statusFilter === Status.OPEN ? 'bg-amber-600 text-white' : 'bg-stone-100 text-stone-600'}`}
                        >
                            Aberto
                        </button>
                        <button
                            onClick={() => setStatusFilter(Status.PAID)}
                            className={`flex-1 py-1.5 md:py-2 px-2 md:px-3 rounded-md md:rounded-lg text-[10px] md:text-sm font-medium transition-all ${statusFilter === Status.PAID ? 'bg-green-600 text-white' : 'bg-stone-100 text-stone-600'}`}
                        >
                            Pago
                        </button>
                    </div>
                </div>

                <div className="bg-white/70 border border-stone-100 md:border-stone-200 rounded-lg md:rounded-xl p-2.5 md:p-4">
                    <label className="block text-[9px] md:text-xs font-semibold text-stone-500 uppercase tracking-wider mb-1.5 md:mb-2">Mês</label>
                    <select
                        value={monthFilter}
                        onChange={(e) => setMonthFilter(e.target.value)}
                        className="w-full px-2 md:px-4 py-1.5 md:py-2 bg-stone-50 border border-stone-200 rounded-md md:rounded-lg text-xs md:text-sm text-stone-900 focus:outline-none focus:ring-2 focus:ring-stone-400"
                    >
                        <option value="ALL">Todos</option>
                        {['Janeiro', 'Fevereiro', 'Março', 'Abril', 'Maio', 'Junho', 'Julho', 'Agosto', 'Setembro', 'Outubro', 'Novembro', 'Dezembro'].map(m => (
                            <option key={m} value={m} className="bg-white">{m}</option>
                        ))}
                    </select>
                </div>
            </div>

            {/* Summary Card */}
            <div className="bg-gradient-to-br from-stone-100 to-stone-200 border border-stone-200 rounded-xl md:rounded-2xl p-3 md:p-6 flex items-center justify-between">
                <div>
                    <p className="text-stone-500 text-xs md:text-sm font-medium mb-0.5 md:mb-1">Total</p>
                    <h2 className="text-xl md:text-4xl font-bold text-stone-900 tracking-tight">{formatCurrency(totalValue)}</h2>
                </div>
                <div className="h-10 w-10 md:h-14 md:w-14 rounded-full bg-white/50 flex items-center justify-center">
                    <Wallet className="w-5 h-5 md:w-8 md:h-8 text-stone-600" />
                </div>
            </div>

            {/* Print Only Header */}
            < div className="hidden print:block mb-8" >
                <h1 className="text-2xl font-bold text-black mb-2">Relatório Financeiro</h1>
                <p className="text-gray-600">
                    Filtro: {statusFilter === 'ALL' ? 'Todas as Transações' : statusFilter === Status.OPEN ? 'Pendências (Em Aberto)' : 'Pagamentos Realizados'}
                </p>
                <p className="text-gray-600">Gerado em: {new Date().toLocaleDateString()}</p>
            </div >

            {/* Table - Desktop */}
            <div className="bg-white/70 border border-stone-200 rounded-2xl overflow-hidden print:bg-white print:border-gray-200 hidden md:block">
                <table className="w-full text-left">
                    <thead className="bg-stone-50 text-stone-500 text-xs uppercase font-semibold print:bg-gray-100 print:text-gray-700">
                        <tr>
                            <th className="px-6 py-4">Descrição</th>
                            <th className="px-6 py-4">Categoria</th>
                            <th className="px-6 py-4">Mês/Ano</th>
                            <th className="px-6 py-4">Valor</th>
                            <th className="px-6 py-4 text-center">Status</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-stone-100 print:divide-gray-200">
                        {filteredTransactions.map((t) => (
                            <tr key={t.id} className="group hover:bg-stone-50/50 print:text-black">
                                <td className="px-6 py-4 font-medium text-stone-900 print:text-black">{t.description || 'Sem descrição'}</td>
                                <td className="px-6 py-4 text-stone-600 print:text-black">{t.category}</td>
                                <td className="px-6 py-4 text-stone-500 print:text-black">{t.month}/{t.year}</td>
                                <td className="px-6 py-4 font-bold text-stone-900 print:text-black">{formatCurrency(t.amount)}</td>
                                <td className="px-6 py-4 text-center">
                                    <span className={`inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-xs font-bold border ${t.status === Status.PAID
                                        ? 'bg-green-100 text-green-700 border-green-200 print:bg-green-100 print:text-green-800'
                                        : 'bg-amber-100 text-amber-700 border-amber-200 print:bg-yellow-100 print:text-yellow-800'
                                        }`}>
                                        {t.status === Status.PAID ? 'PAGO' : 'ABERTO'}
                                    </span>
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            {/* Mobile Cards */}
            <div className="md:hidden bg-white/70 border border-stone-100 rounded-lg overflow-hidden">
                <div className="divide-y divide-stone-50">
                    {filteredTransactions.map((t) => (
                        <div key={t.id} className="px-3 py-2.5">
                            <div className="flex items-center gap-2">
                                <span className={`shrink-0 p-1 rounded-md ${t.status === Status.PAID ? 'bg-green-50 text-green-600' : 'bg-amber-50 text-amber-600'}`}>
                                    {t.status === Status.PAID ? <CheckCircle2 size={14} /> : <AlertCircle size={14} />}
                                </span>
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center justify-between gap-1">
                                        <p className="font-medium text-xs text-stone-800 truncate">{t.description || 'Sem descrição'}</p>
                                        <p className="text-xs font-semibold text-stone-900 shrink-0">{formatCurrency(t.amount)}</p>
                                    </div>
                                    <div className="flex items-center gap-1.5 mt-0.5">
                                        <span className="text-[9px] px-1.5 py-0.5 rounded bg-stone-100 text-stone-600 font-medium">{t.category}</span>
                                        <span className="text-[9px] text-stone-400">{t.month}/{t.year}</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>

            <style>{`
                @media print {
                    .no-print { display: none !important; }
                    body { background-color: white !important; color: black !important; }
                    .bg-paper { background-color: white !important; }
                }
            `}</style>
        </div >
    );
};
