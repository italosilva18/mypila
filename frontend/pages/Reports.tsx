import React, { useState, useMemo } from 'react';
import { useParams } from 'react-router-dom';
import { useTransactions } from '../hooks/useTransactions';
import { Status } from '../types';
import { FileText, Printer, CheckCircle2, AlertCircle, Wallet, Download, MessageCircle, Copy, Check } from 'lucide-react';
import { exportToCSV } from '../utils/export';
import { formatCurrency } from '../utils/currency';

export const Reports: React.FC = () => {
    const { companyId } = useParams<{ companyId: string }>();
    const { transactions, loading } = useTransactions(companyId!);

    const [statusFilter, setStatusFilter] = useState<'ALL' | Status>('ALL');
    const [monthFilter, setMonthFilter] = useState<string>('ALL');
    const [copied, setCopied] = useState(false);

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

    const totalPaid = useMemo(() => {
        return filteredTransactions.filter(t => t.status === Status.PAID).reduce((acc, t) => acc + t.amount, 0);
    }, [filteredTransactions]);

    const totalOpen = useMemo(() => {
        return filteredTransactions.filter(t => t.status === Status.OPEN).reduce((acc, t) => acc + t.amount, 0);
    }, [filteredTransactions]);

    const handlePrint = () => {
        window.print();
    };

    const generateWhatsAppReport = () => {
        const date = new Date().toLocaleDateString('pt-BR');
        const filterText = statusFilter === 'ALL' ? 'Todos' : statusFilter === Status.PAID ? 'Pagos' : 'Em Aberto';
        const monthText = monthFilter === 'ALL' ? 'Todos os meses' : monthFilter;

        let report = `*RELATORIO FINANCEIRO*\n`;
        report += `_${date}_\n\n`;
        report += `*Filtros:* ${filterText} | ${monthText}\n`;
        report += `━━━━━━━━━━━━━━━━━━━━\n\n`;

        // Summary
        report += `*RESUMO*\n`;
        report += `Total: *${formatCurrency(totalValue)}*\n`;
        if (statusFilter === 'ALL') {
            report += `Pago: ${formatCurrency(totalPaid)}\n`;
            report += `Aberto: ${formatCurrency(totalOpen)}\n`;
        }
        report += `\n━━━━━━━━━━━━━━━━━━━━\n\n`;

        // Group by category
        const byCategory = filteredTransactions.reduce((acc, t) => {
            if (!acc[t.category]) acc[t.category] = [];
            acc[t.category].push(t);
            return acc;
        }, {} as Record<string, typeof filteredTransactions>);

        report += `*DETALHAMENTO*\n\n`;

        Object.entries(byCategory).forEach(([category, items]) => {
            const categoryTotal = items.reduce((acc, t) => acc + t.amount, 0);
            report += `*${category}* (${formatCurrency(categoryTotal)})\n`;
            items.forEach(t => {
                const statusIcon = t.status === Status.PAID ? '✓' : '○';
                report += `  ${statusIcon} ${t.description || 'Sem descricao'} - ${formatCurrency(t.amount)}\n`;
            });
            report += `\n`;
        });

        report += `━━━━━━━━━━━━━━━━━━━━\n`;
        report += `_Gerado por MyPilaPro_`;

        return report;
    };

    const handleWhatsAppExport = () => {
        const report = generateWhatsAppReport();
        const encoded = encodeURIComponent(report);
        window.open(`https://wa.me/?text=${encoded}`, '_blank');
    };

    const handleCopyReport = async () => {
        const report = generateWhatsAppReport();
        try {
            await navigator.clipboard.writeText(report);
            setCopied(true);
            setTimeout(() => setCopied(false), 2000);
        } catch (err) {
            console.error('Failed to copy', err);
        }
    };

    if (loading) return (
        <div className="min-h-[50vh] flex items-center justify-center">
            <div className="card p-8 flex flex-col items-center gap-4">
                <div className="w-10 h-10 border-4 border-primary-500 border-t-transparent rounded-full animate-spin"></div>
                <p className="text-muted">Carregando relatorios...</p>
            </div>
        </div>
    );

    return (
        <div className="space-y-3 md:space-y-6 mobile-content-padding">
            <header className="flex flex-col md:flex-row md:items-center justify-between gap-3 no-print px-1 md:px-0">
                <div>
                    <h1 className="text-base md:text-2xl font-bold text-foreground flex items-center gap-1.5 md:gap-2">
                        <FileText className="w-4 h-4 md:w-6 md:h-6 text-primary-500" />
                        Relatorios
                    </h1>
                    <p className="text-muted text-xs md:text-sm">Relatorios de pagamentos e pendencias.</p>
                </div>
                <div className="flex gap-2 flex-wrap">
                    <button
                        onClick={handleCopyReport}
                        className="flex items-center gap-1.5 bg-card border border-border active:bg-primary-50 text-foreground px-3 md:px-4 py-2 rounded-xl text-xs md:text-sm transition-all"
                        title="Copiar relatorio"
                    >
                        {copied ? <Check className="w-3.5 h-3.5 md:w-4 md:h-4 text-success" /> : <Copy className="w-3.5 h-3.5 md:w-4 md:h-4" />}
                        <span className="hidden sm:inline">{copied ? 'Copiado!' : 'Copiar'}</span>
                    </button>
                    <button
                        onClick={handleWhatsAppExport}
                        className="flex items-center gap-1.5 bg-[#25D366] active:bg-[#1da851] text-white px-3 md:px-4 py-2 rounded-xl text-xs md:text-sm transition-all shadow-lg"
                        title="Enviar por WhatsApp"
                    >
                        <MessageCircle className="w-3.5 h-3.5 md:w-4 md:h-4" />
                        <span className="hidden sm:inline">WhatsApp</span>
                    </button>
                    <button
                        onClick={handlePrint}
                        className="flex items-center gap-1.5 bg-card border border-border active:bg-primary-50 text-foreground px-3 md:px-4 py-2 rounded-xl text-xs md:text-sm transition-all"
                    >
                        <Printer className="w-3.5 h-3.5 md:w-4 md:h-4" />
                        <span className="hidden sm:inline">Imprimir</span>
                    </button>
                    <button
                        onClick={() => exportToCSV(filteredTransactions, `relatorio_financeiro_${new Date().toISOString().split('T')[0]}`)}
                        className="btn-primary flex items-center gap-1.5 px-3 md:px-4 py-2 text-xs md:text-sm"
                    >
                        <Download className="w-3.5 h-3.5 md:w-4 md:h-4" />
                        CSV
                    </button>
                </div>
            </header>

            {/* Filters */}
            <div className="grid grid-cols-2 md:grid-cols-3 gap-2 md:gap-4 no-print">
                <div className="card p-2.5 md:p-4">
                    <label className="block text-[9px] md:text-xs font-semibold text-muted uppercase tracking-wider mb-1.5 md:mb-2">Status</label>
                    <div className="flex gap-1 md:gap-2">
                        <button
                            onClick={() => setStatusFilter('ALL')}
                            className={`flex-1 py-1.5 md:py-2 px-2 md:px-3 rounded-lg text-[10px] md:text-sm font-medium transition-all ${statusFilter === 'ALL' ? 'bg-primary-500 text-white' : 'bg-primary-50 text-primary-700'}`}
                        >
                            Total
                        </button>
                        <button
                            onClick={() => setStatusFilter(Status.OPEN)}
                            className={`flex-1 py-1.5 md:py-2 px-2 md:px-3 rounded-lg text-[10px] md:text-sm font-medium transition-all ${statusFilter === Status.OPEN ? 'bg-warning text-white' : 'bg-warning-light text-warning-dark'}`}
                        >
                            Aberto
                        </button>
                        <button
                            onClick={() => setStatusFilter(Status.PAID)}
                            className={`flex-1 py-1.5 md:py-2 px-2 md:px-3 rounded-lg text-[10px] md:text-sm font-medium transition-all ${statusFilter === Status.PAID ? 'bg-success text-white' : 'bg-success-light text-success-dark'}`}
                        >
                            Pago
                        </button>
                    </div>
                </div>

                <div className="card p-2.5 md:p-4">
                    <label className="block text-[9px] md:text-xs font-semibold text-muted uppercase tracking-wider mb-1.5 md:mb-2">Mes</label>
                    <select
                        value={monthFilter}
                        onChange={(e) => setMonthFilter(e.target.value)}
                        className="select text-xs md:text-sm py-1.5 md:py-2"
                    >
                        <option value="ALL">Todos</option>
                        {['Janeiro', 'Fevereiro', 'Marco', 'Abril', 'Maio', 'Junho', 'Julho', 'Agosto', 'Setembro', 'Outubro', 'Novembro', 'Dezembro'].map(m => (
                            <option key={m} value={m}>{m}</option>
                        ))}
                    </select>
                </div>
            </div>

            {/* Summary Card */}
            <div className="bg-gradient-primary rounded-xl md:rounded-2xl p-3 md:p-6 flex items-center justify-between text-white shadow-card">
                <div>
                    <p className="text-white/80 text-xs md:text-sm font-medium mb-0.5 md:mb-1">Total</p>
                    <h2 className="text-xl md:text-4xl font-bold tracking-tight">{formatCurrency(totalValue)}</h2>
                    {statusFilter === 'ALL' && (
                        <div className="flex gap-4 mt-2 text-xs md:text-sm">
                            <span className="text-white/80">Pago: <span className="font-semibold text-white">{formatCurrency(totalPaid)}</span></span>
                            <span className="text-white/80">Aberto: <span className="font-semibold text-white">{formatCurrency(totalOpen)}</span></span>
                        </div>
                    )}
                </div>
                <div className="h-10 w-10 md:h-14 md:w-14 rounded-full bg-white/20 flex items-center justify-center">
                    <Wallet className="w-5 h-5 md:w-8 md:h-8 text-white" />
                </div>
            </div>

            {/* Print Only Header */}
            <div className="hidden print:block mb-8">
                <h1 className="text-2xl font-bold text-black mb-2">Relatorio Financeiro</h1>
                <p className="text-gray-600">
                    Filtro: {statusFilter === 'ALL' ? 'Todas as Transacoes' : statusFilter === Status.OPEN ? 'Pendencias (Em Aberto)' : 'Pagamentos Realizados'}
                </p>
                <p className="text-gray-600">Gerado em: {new Date().toLocaleDateString()}</p>
            </div>

            {/* Table - Desktop */}
            <div className="card overflow-hidden print:bg-white print:border-gray-200 hidden md:block">
                <table className="w-full text-left">
                    <thead className="table-header print:bg-gray-100 print:text-gray-700">
                        <tr>
                            <th className="px-6 py-4">Descricao</th>
                            <th className="px-6 py-4">Categoria</th>
                            <th className="px-6 py-4">Mes/Ano</th>
                            <th className="px-6 py-4">Valor</th>
                            <th className="px-6 py-4 text-center">Status</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-border print:divide-gray-200">
                        {filteredTransactions.map((t) => (
                            <tr key={t.id} className="table-row print:text-black">
                                <td className="px-6 py-4 font-medium text-foreground print:text-black">{t.description || 'Sem descricao'}</td>
                                <td className="px-6 py-4 text-muted print:text-black">{t.category}</td>
                                <td className="px-6 py-4 text-muted print:text-black">{t.month}/{t.year}</td>
                                <td className="px-6 py-4 font-bold text-foreground print:text-black">{formatCurrency(t.amount)}</td>
                                <td className="px-6 py-4 text-center">
                                    <span className={`badge ${t.status === Status.PAID
                                        ? 'badge-success print:bg-green-100 print:text-green-800'
                                        : 'badge-warning print:bg-yellow-100 print:text-yellow-800'
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
            <div className="md:hidden card overflow-hidden">
                <div className="divide-y divide-border">
                    {filteredTransactions.map((t) => (
                        <div key={t.id} className="px-3 py-2.5">
                            <div className="flex items-center gap-2">
                                <span className={`shrink-0 p-1 rounded-md ${t.status === Status.PAID ? 'bg-success-light text-success' : 'bg-warning-light text-warning'}`}>
                                    {t.status === Status.PAID ? <CheckCircle2 size={14} /> : <AlertCircle size={14} />}
                                </span>
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center justify-between gap-1">
                                        <p className="font-medium text-xs text-foreground truncate">{t.description || 'Sem descricao'}</p>
                                        <p className="text-xs font-semibold text-foreground shrink-0">{formatCurrency(t.amount)}</p>
                                    </div>
                                    <div className="flex items-center gap-1.5 mt-0.5">
                                        <span className="text-[9px] px-1.5 py-0.5 rounded bg-primary-100 text-primary-700 font-medium">{t.category}</span>
                                        <span className="text-[9px] text-muted">{t.month}/{t.year}</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                    ))}
                </div>
            </div>

            {/* Empty State */}
            {filteredTransactions.length === 0 && (
                <div className="card py-12 text-center">
                    <div className="inline-flex p-4 bg-primary-50 rounded-2xl border border-primary-100 mb-4">
                        <FileText className="w-8 h-8 text-primary-400" />
                    </div>
                    <p className="text-muted text-sm">Nenhuma transacao encontrada com os filtros selecionados</p>
                </div>
            )}

            <style>{`
                @media print {
                    .no-print { display: none !important; }
                    body { background-color: white !important; color: black !important; }
                    .bg-paper { background-color: white !important; }
                }
            `}</style>
        </div>
    );
};
