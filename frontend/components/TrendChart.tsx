import React, { useMemo } from 'react';
import {
    AreaChart,
    Area,
    XAxis,
    YAxis,
    CartesianGrid,
    Tooltip,
    ResponsiveContainer
} from 'recharts';
import { Transaction, Status } from '../types';

interface Props {
    transactions: Transaction[];
    year: number;
}

const MONTHS = ['Janeiro', 'Fevereiro', 'Março', 'Abril', 'Maio', 'Junho', 'Julho', 'Agosto', 'Setembro', 'Outubro', 'Novembro', 'Dezembro'];

export const TrendChart: React.FC<Props> = ({ transactions, year }) => {
    // Aggregate data for the whole year
    const data = useMemo(() => {
        return MONTHS.map(month => {
            const monthlyTransactions = transactions.filter(t => t.month === month && t.year === year);

            const paid = monthlyTransactions
                .filter(t => t.status === Status.PAID && t.category !== 'Salário' && t.category !== 'Receita') // Assuming Expense
                .reduce((acc, t) => acc + t.amount, 0);

            return {
                name: month.substring(0, 3), // Jan, Fev...
                despesas: paid,
                // receitas: income
            };
        });
    }, [transactions, year]);

    return (
        <div className="bg-white/70 backdrop-blur-sm p-3 md:p-6 rounded-lg md:rounded-xl shadow-soft border border-stone-100 md:border-stone-200 mb-3 md:mb-8">
            <h3 className="text-xs md:text-lg font-semibold text-stone-900 mb-2 md:mb-4">Evolução Anual ({year})</h3>
            <div className="h-36 md:h-64">
                <ResponsiveContainer width="100%" height="100%" minWidth={100} minHeight={100}>
                    <AreaChart data={data} margin={{ top: 5, right: 5, left: -20, bottom: 0 }}>
                        <defs>
                            <linearGradient id="colorDespesas" x1="0" y1="0" x2="0" y2="1">
                                <stop offset="5%" stopColor="#78716c" stopOpacity={0.2} />
                                <stop offset="95%" stopColor="#78716c" stopOpacity={0} />
                            </linearGradient>
                        </defs>
                        <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{ fontSize: 9, fill: '#78716c' }} />
                        <YAxis axisLine={false} tickLine={false} tickFormatter={(value) => `${value / 1000}k`} tick={{ fontSize: 9, fill: '#78716c' }} width={30} />
                        <Tooltip
                            formatter={(value) => new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(Number(value) || 0)}
                            contentStyle={{ borderRadius: '8px', border: '1px solid #e7e5e4', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)', backgroundColor: '#faf8f5', fontSize: '11px' }}
                        />
                        <CartesianGrid vertical={false} stroke="#e7e5e4" />
                        <Area
                            type="monotone"
                            dataKey="despesas"
                            stroke="#57534e"
                            strokeWidth={2}
                            fillOpacity={1}
                            fill="url(#colorDespesas)"
                            name="Despesas Pagas"
                        />
                    </AreaChart>
                </ResponsiveContainer>
            </div>
        </div>
    );
};
