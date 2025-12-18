import React, { createContext, useContext, useState, useMemo, ReactNode } from 'react';

interface DateFilterContextData {
    month: string;
    year: number;
    setMonth: (month: string) => void;
    setYear: (year: number) => void;
    nextMonth: () => void;
    prevMonth: () => void;
}

const DateFilterContext = createContext<DateFilterContextData>({} as DateFilterContextData);

const MONTHS = ['Janeiro', 'Fevereiro', 'Março', 'Abril', 'Maio', 'Junho', 'Julho', 'Agosto', 'Setembro', 'Outubro', 'Novembro', 'Dezembro'];

export const DateFilterProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
    const [month, setMonth] = useState(MONTHS[new Date().getMonth()]);
    const [year, setYear] = useState(new Date().getFullYear());

    const nextMonth = () => {
        const currentIndex = MONTHS.indexOf(month);
        if (currentIndex === 11) {
            setMonth(MONTHS[0]);
            setYear(year + 1);
        } else {
            setMonth(MONTHS[currentIndex + 1]);
        }
    };

    const prevMonth = () => {
        const currentIndex = MONTHS.indexOf(month);
        if (currentIndex === 0) {
            setMonth(MONTHS[11]);
            setYear(year - 1);
        } else {
            setMonth(MONTHS[currentIndex - 1]);
        }
    };

    const contextValue = useMemo(() => ({
        month,
        year,
        setMonth,
        setYear,
        nextMonth,
        prevMonth
    }), [month, year]);

    return (
        <DateFilterContext.Provider value={contextValue}>
            {children}
        </DateFilterContext.Provider>
    );
};

export const useDateFilter = () => useContext(DateFilterContext);
