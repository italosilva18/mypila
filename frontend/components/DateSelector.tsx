import React from 'react';
import { ChevronLeft, ChevronRight, Calendar } from 'lucide-react';
import { useDateFilter } from '../contexts/DateFilterContext';

export const DateSelector: React.FC = () => {
    const { month, year, prevMonth, nextMonth } = useDateFilter();

    return (
        <div className="flex items-center bg-stone-50 border border-stone-200 rounded-lg md:rounded-xl p-0.5 md:p-1">
            <button
                onClick={prevMonth}
                aria-label="Mês anterior"
                className="p-1 md:p-2 text-stone-500 hover:text-stone-800 active:bg-stone-100 hover:bg-stone-100 rounded-md md:rounded-lg transition-colors"
            >
                <ChevronLeft className="w-4 h-4 md:w-5 md:h-5" />
            </button>

            <div className="flex items-center gap-1 md:gap-2 px-1.5 md:px-4 min-w-[80px] md:min-w-[160px] justify-center text-stone-900 font-medium select-none text-xs md:text-base">
                <Calendar className="w-3 h-3 md:w-4 md:h-4 text-stone-600 hidden md:block" />
                <span>{month}</span>
                <span className="text-stone-400">{year}</span>
            </div>

            <button
                onClick={nextMonth}
                aria-label="Próximo mês"
                className="p-1 md:p-2 text-stone-500 hover:text-stone-800 active:bg-stone-100 hover:bg-stone-100 rounded-md md:rounded-lg transition-colors"
            >
                <ChevronRight className="w-4 h-4 md:w-5 md:h-5" />
            </button>
        </div>
    );
};
