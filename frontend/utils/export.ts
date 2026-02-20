import { Transaction, Status } from '../types';

export const exportToCSV = (transactions: Transaction[], filename: string) => {
    // CSV Header - includes paidAmount and remaining
    const headers = ['Data', 'Ano', 'Descrição', 'Categoria', 'Valor Total', 'Valor Pago', 'Valor Restante', 'Status'];

    // Map data to rows
    const rows = transactions.map(t => {
        const paidAmount = t.paidAmount || 0;
        const remainingAmount = t.amount - paidAmount;
        const isPartial = paidAmount > 0 && paidAmount < t.amount;
        const isPaid = t.status === Status.PAID || paidAmount >= t.amount;
        const statusText = isPaid ? 'PAGO' : isPartial ? 'PARCIAL' : 'ABERTO';

        return [
            t.month,
            t.year.toString(),
            `"${t.description || ''}"`, // Wrap in quotes to handle commas in text
            t.category,
            t.amount.toFixed(2).replace('.', ','), // Format for Excel (in Pt-BR usually uses comma)
            paidAmount.toFixed(2).replace('.', ','),
            (remainingAmount > 0 ? remainingAmount : 0).toFixed(2).replace('.', ','),
            statusText
        ];
    });

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
