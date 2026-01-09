import React, { createContext, useContext, useState, useMemo, useCallback, ReactNode } from 'react';
import { MONTHS } from '../utils/constants';

interface DateFilterContextData {
    month: string;
    year: number;
    setMonth: (month: string) => void;
    setYear: (year: number) => void;
    nextMonth: () => void;
    prevMonth: () => void;
}

const DateFilterContext = createContext<DateFilterContextData>({} as DateFilterContextData);

export const DateFilterProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
    const [month, setMonth] = useState(MONTHS[new Date().getMonth()]);
    const [year, setYear] = useState(new Date().getFullYear());

    const nextMonth = useCallback(() => {
        setMonth(currentMonth => {
            const currentIndex = MONTHS.indexOf(currentMonth);
            if (currentIndex === 11) {
                setYear(y => y + 1);
                return MONTHS[0];
            } else {
                return MONTHS[currentIndex + 1];
            }
        });
    }, []);

    const prevMonth = useCallback(() => {
        setMonth(currentMonth => {
            const currentIndex = MONTHS.indexOf(currentMonth);
            if (currentIndex === 0) {
                setYear(y => y - 1);
                return MONTHS[11];
            } else {
                return MONTHS[currentIndex - 1];
            }
        });
    }, []);

    const contextValue = useMemo(() => ({
        month,
        year,
        setMonth,
        setYear,
        nextMonth,
        prevMonth
    }), [month, year, nextMonth, prevMonth]);

    return (
        <DateFilterContext.Provider value={contextValue}>
            {children}
        </DateFilterContext.Provider>
    );
};

export const useDateFilter = () => {
    const context = useContext(DateFilterContext);
    if (!context || Object.keys(context).length === 0) {
        throw new Error('useDateFilter must be used within a DateFilterProvider');
    }
    return context;
};
