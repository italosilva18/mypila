import { Transaction } from '../types';

export const exportToCSV = (transactions: Transaction[], filename: string) => {
    // CSV Header
    const headers = ['Data', 'Ano', 'Descrição', 'Categoria', 'Valor', 'Status'];

    // Map data to rows
    const rows = transactions.map(t => [
        t.month,
        t.year.toString(),
        `"${t.description || ''}"`, // Wrap in quotes to handle commas in text
        t.category,
        t.amount.toFixed(2).replace('.', ','), // Format for Excel (in Pt-BR usually uses comma)
        t.status
    ]);

    // Combine
    const csvContent = [
        headers.join(';'),
        ...rows.map(r => r.join(';'))
    ].join('\n');

    // Create Blob and Download
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.setAttribute('href', url);
    link.setAttribute('download', `${filename}.csv`);
    link.style.visibility = 'hidden';
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
};
