import React, { useMemo, useCallback, useState } from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
} from 'recharts';
import { ChevronDown } from 'lucide-react';
import { Transaction, Status } from '../types';
import { stringToColor } from '../utils/colors';

interface Props {
  transactions: Transaction[];
  categories?: { name: string; color: string }[];
}

const COLORS = {
  [Status.PAID]: '#16a34a', // green-600
  [Status.OPEN]: '#d97706', // amber-600
};

export const FinancialChart = React.memo<Props>(({ transactions, categories = [] }) => {
  // Get available years from transactions
  const availableYears = useMemo(() => {
    const years = [...new Set(transactions.map(t => t.year))].sort((a, b) => b - a);
    return years.length > 0 ? years : [new Date().getFullYear()];
  }, [transactions]);

  const [selectedYear, setSelectedYear] = useState(availableYears[0] || new Date().getFullYear());

  // Filter transactions by selected year
  const filteredTransactions = useMemo(() => {
    return transactions.filter(t => t.year === selectedYear);
  }, [transactions, selectedYear]);

  // Memoized helper to find category color - prevents recreation on every render
  const getCategoryColor = useCallback((categoryName: string) => {
    const cat = categories.find(c => c.name === categoryName);
    return cat?.color || stringToColor(categoryName);
  }, [categories]);

  // Aggregate data for Bar Chart (Monthly flow)
  const monthlyData = useMemo(() => {
    const data = filteredTransactions.reduce((acc, curr) => {
      // Skip accumulated items like Vacation for the timeline
      if (curr.month === 'Acumulado') return acc;

      const existing = acc.find(item => item.month === curr.month);
      if (existing) {
        if (curr.status === Status.PAID) existing.paid += curr.amount;
        else existing.open += curr.amount;
      } else {
        acc.push({
          month: curr.month,
          paid: curr.status === Status.PAID ? curr.amount : 0,
          open: curr.status === Status.OPEN ? curr.amount : 0,
        });
      }
      return acc;
    }, [] as { month: string; paid: number; open: number }[]);

    // Sort months logic (simplistic map for this specific dataset)
    const monthOrder: Record<string, number> = {
      'Janeiro': 1, 'Fevereiro': 2, 'Março': 3, 'Abril': 4, 'Maio': 5, 'Junho': 6,
      'Julho': 7, 'Agosto': 8, 'Setembro': 9, 'Outubro': 10, 'Novembro': 11, 'Dezembro': 12
    };
    data.sort((a, b) => (monthOrder[a.month] || 99) - (monthOrder[b.month] || 99));

    return data;
  }, [filteredTransactions]);

  // Aggregate data for Pie Chart (Outstanding by Category)
  const openData = useMemo(() => {
    return filteredTransactions
      .filter(t => t.status === Status.OPEN)
      .reduce((acc, curr) => {
        const existing = acc.find(item => item.name === curr.category);
        if (existing) {
          existing.value += curr.amount;
        } else {
          acc.push({ name: curr.category, value: curr.amount });
        }
        return acc;
      }, [] as { name: string; value: number }[]);
  }, [filteredTransactions]);

  return (
    <div className="grid grid-cols-1 lg:grid-cols-2 gap-3 md:gap-6 mb-3 md:mb-8">
      <div className="bg-white/70 backdrop-blur-sm p-3 md:p-6 rounded-lg md:rounded-xl shadow-soft border border-stone-100 md:border-stone-200">
        <div className="flex items-center justify-between mb-2 md:mb-4">
          <h3 className="text-xs md:text-lg font-semibold text-stone-900">Fluxo Mensal</h3>
          <div className="relative">
            <select
              value={selectedYear}
              onChange={(e) => setSelectedYear(Number(e.target.value))}
              className="appearance-none bg-stone-100 border border-stone-200 rounded-lg px-2 md:px-3 py-1 md:py-1.5 pr-6 md:pr-8 text-[10px] md:text-sm font-medium text-stone-700 cursor-pointer hover:bg-stone-200 transition-colors"
            >
              {availableYears.map(y => (
                <option key={y} value={y}>{y}</option>
              ))}
            </select>
            <ChevronDown className="absolute right-1.5 md:right-2 top-1/2 -translate-y-1/2 w-3 h-3 md:w-4 md:h-4 text-stone-500 pointer-events-none" />
          </div>
        </div>
        <div className="h-40 md:h-64">
          <ResponsiveContainer width="100%" height="100%" minWidth={100} minHeight={100}>
            <BarChart data={monthlyData} margin={{ top: 5, right: 5, left: -20, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#e7e5e4" />
              <XAxis dataKey="month" tick={{ fontSize: 9, fill: '#78716c' }} axisLine={false} tickLine={false} tickFormatter={(value) => value.substring(0, 3)} />
              <YAxis tick={{ fontSize: 9, fill: '#78716c' }} axisLine={false} tickLine={false} tickFormatter={(value) => `${value / 1000}k`} width={30} />
              <Tooltip
                formatter={(value) => new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(Number(value) || 0)}
                contentStyle={{ borderRadius: '8px', border: '1px solid #e7e5e4', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)', backgroundColor: '#faf8f5', fontSize: '11px' }}
              />
              <Legend wrapperStyle={{ fontSize: '10px' }} />
              <Bar dataKey="paid" name="Pago" stackId="a" fill={COLORS[Status.PAID]} radius={[0, 0, 4, 4]} />
              <Bar dataKey="open" name="Aberto" stackId="a" fill={COLORS[Status.OPEN]} radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      <div className="bg-white/70 backdrop-blur-sm p-3 md:p-6 rounded-lg md:rounded-xl shadow-soft border border-stone-100 md:border-stone-200">
        <div className="flex items-center justify-between mb-2 md:mb-4">
          <h3 className="text-xs md:text-lg font-semibold text-stone-900">Em Aberto por Categoria</h3>
          <span className="text-[9px] md:text-xs px-2 py-0.5 bg-stone-100 text-stone-600 rounded-full">{selectedYear}</span>
        </div>
        <div className="h-40 md:h-64">
          <ResponsiveContainer width="100%" height="100%" minWidth={100} minHeight={100}>
            <PieChart>
              <Pie
                data={openData}
                cx="50%"
                cy="50%"
                innerRadius={35}
                outerRadius={50}
                paddingAngle={5}
                dataKey="value"
              >
                {openData.map((entry, index) => (
                  <Cell key={`cell-${index}`} fill={getCategoryColor(entry.name)} />
                ))}
              </Pie>
              <Tooltip
                formatter={(value) => new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(Number(value) || 0)}
                contentStyle={{ borderRadius: '8px', border: '1px solid #e7e5e4', boxShadow: '0 4px 6px -1px rgb(0 0 0 / 0.1)', backgroundColor: '#faf8f5', fontSize: '11px' }}
              />
              <Legend verticalAlign="bottom" height={28} wrapperStyle={{ fontSize: '10px' }} />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </div>
    </div>
  );
});
FinancialChart.displayName = 'FinancialChart';
