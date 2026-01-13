import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { X, Save, Plus, Trash2, Loader2 } from 'lucide-react';
import { Quote, Category, CreateQuoteRequest } from '../types';
import { api } from '../services/api';
import { useFormValidation } from '../hooks/useFormValidation';
import { useEscapeKey } from '../hooks/useEscapeKey';
import { validateRequired, validateMaxLength, validatePositiveNumber, combineValidations } from '../utils/validation';
import { ErrorMessage } from './ErrorMessage';
import { formatCurrency } from '../utils/currency';

// Fallback UUID generator for browsers that don't support crypto.randomUUID
const generateUUID = (): string => {
    if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
        return generateUUID();
    }
    // Fallback implementation
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
        const r = Math.random() * 16 | 0;
        const v = c === 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
};

interface QuoteItemForm {
    id: string;
    description: string;
    quantity: string;
    unitPrice: string;
    categoryId: string;
}

interface Props {
    isOpen: boolean;
    onClose: () => void;
    onSave: () => void;
    quote: Quote | null;
    categories: Category[];
    companyId: string;
}

export const QuoteModal: React.FC<Props> = ({ isOpen, onClose, onSave, quote, categories, companyId }) => {
    const [isSubmitting, setIsSubmitting] = useState(false);
    const { validateFields, getError, hasError, clearAllErrors } = useFormValidation();

    // Handle Escape key to close modal
    const handleClose = useCallback(() => {
        clearAllErrors();
        onClose();
    }, [clearAllErrors, onClose]);

    useEscapeKey(handleClose, isOpen);

    // Client data
    const [clientName, setClientName] = useState('');
    const [clientEmail, setClientEmail] = useState('');
    const [clientPhone, setClientPhone] = useState('');
    const [clientDocument, setClientDocument] = useState('');
    const [clientAddress, setClientAddress] = useState('');
    const [clientCity, setClientCity] = useState('');
    const [clientState, setClientState] = useState('');
    const [clientZipCode, setClientZipCode] = useState('');

    // Quote data
    const [title, setTitle] = useState('');
    const [description, setDescription] = useState('');
    const [items, setItems] = useState<QuoteItemForm[]>([]);
    const [discount, setDiscount] = useState('0');
    const [discountType, setDiscountType] = useState<'PERCENT' | 'VALUE'>('VALUE');
    const [validUntil, setValidUntil] = useState('');
    const [notes, setNotes] = useState('');

    useEffect(() => {
        if (quote) {
            setClientName(quote.clientName);
            setClientEmail(quote.clientEmail || '');
            setClientPhone(quote.clientPhone || '');
            setClientDocument(quote.clientDocument || '');
            setClientAddress(quote.clientAddress || '');
            setClientCity(quote.clientCity || '');
            setClientState(quote.clientState || '');
            setClientZipCode(quote.clientZipCode || '');
            setTitle(quote.title);
            setDescription(quote.description || '');
            setItems(quote.items.map(item => ({
                id: item.id || generateUUID(),
                description: item.description,
                quantity: item.quantity.toString(),
                unitPrice: item.unitPrice.toString(),
                categoryId: item.categoryId || ''
            })));
            setDiscount(quote.discount.toString());
            setDiscountType(quote.discountType);
            setValidUntil(quote.validUntil.split('T')[0]);
            setNotes(quote.notes || '');
        } else {
            resetForm();
        }
    }, [quote, isOpen]);

    const resetForm = useCallback(() => {
        setClientName('');
        setClientEmail('');
        setClientPhone('');
        setClientDocument('');
        setClientAddress('');
        setClientCity('');
        setClientState('');
        setClientZipCode('');
        setTitle('');
        setDescription('');
        setItems([{ id: generateUUID(), description: '', quantity: '1', unitPrice: '', categoryId: '' }]);
        setDiscount('0');
        setDiscountType('VALUE');
        const defaultDate = new Date();
        defaultDate.setDate(defaultDate.getDate() + 30);
        setValidUntil(defaultDate.toISOString().split('T')[0]);
        setNotes('');
        clearAllErrors();
    }, [clearAllErrors]);

    const addItem = useCallback(() => {
        setItems(prev => [...prev, { id: generateUUID(), description: '', quantity: '1', unitPrice: '', categoryId: '' }]);
    }, []);

    const removeItem = useCallback((id: string) => {
        setItems(prev => prev.length > 1 ? prev.filter(item => item.id !== id) : prev);
    }, []);

    const updateItem = useCallback((id: string, field: keyof QuoteItemForm, value: string) => {
        setItems(prev => prev.map(item => item.id === id ? { ...item, [field]: value } : item));
    }, []);

    const { subtotal, total, discountValue } = useMemo(() => {
        const sub = items.reduce((acc, item) => {
            const qty = parseFloat(item.quantity) || 0;
            const price = parseFloat(item.unitPrice) || 0;
            return acc + (qty * price);
        }, 0);

        const disc = parseFloat(discount) || 0;
        let discVal = 0;
        let tot = sub;

        if (discountType === 'PERCENT') {
            discVal = sub * (disc / 100);
            tot = sub - discVal;
        } else {
            discVal = disc;
            tot = sub - disc;
        }

        return { subtotal: sub, total: Math.max(0, tot), discountValue: discVal };
    }, [items, discount, discountType]);

    const validateForm = useCallback((): boolean => {
        const validations: Record<string, () => { isValid: boolean; error?: string }> = {
            clientName: () => combineValidations(
                validateRequired(clientName, 'Nome do cliente'),
                validateMaxLength(clientName, 100, 'Nome do cliente')
            ),
            title: () => combineValidations(
                validateRequired(title, 'Titulo'),
                validateMaxLength(title, 200, 'Titulo')
            ),
            validUntil: () => validateRequired(validUntil, 'Data de validade'),
        };

        // Validate items
        items.forEach((item, index) => {
            validations[`item_${index}_description`] = () =>
                validateRequired(item.description, `Descricao do item ${index + 1}`);
            validations[`item_${index}_quantity`] = () =>
                validatePositiveNumber(item.quantity, `Quantidade do item ${index + 1}`);
            validations[`item_${index}_unitPrice`] = () =>
                validatePositiveNumber(item.unitPrice, `Preco do item ${index + 1}`);
        });

        return validateFields(validations);
    }, [clientName, title, validUntil, items, validateFields]);

    const handleSubmit = useCallback(async (e: React.FormEvent) => {
        e.preventDefault();
        if (!validateForm()) return;

        try {
            setIsSubmitting(true);

            const data: CreateQuoteRequest = {
                clientName,
                clientEmail: clientEmail || undefined,
                clientPhone: clientPhone || undefined,
                clientDocument: clientDocument || undefined,
                clientAddress: clientAddress || undefined,
                clientCity: clientCity || undefined,
                clientState: clientState || undefined,
                clientZipCode: clientZipCode || undefined,
                title,
                description: description || undefined,
                items: items.map(item => ({
                    description: item.description,
                    quantity: parseFloat(item.quantity) || 0,
                    unitPrice: parseFloat(item.unitPrice) || 0,
                    categoryId: item.categoryId || undefined,
                })),
                discount: parseFloat(discount) || 0,
                discountType,
                validUntil,
                notes: notes || undefined,
            };

            if (quote) {
                await api.updateQuote(quote.id, data);
            } else {
                await api.createQuote(companyId, data);
            }

            onSave();
        } catch (err) {
            console.error('Failed to save quote', err);
        } finally {
            setIsSubmitting(false);
        }
    }, [validateForm, clientName, clientEmail, clientPhone, clientDocument, clientAddress, clientCity, clientState, clientZipCode, title, description, items, discount, discountType, validUntil, notes, quote, companyId, onSave]);

    if (!isOpen) return null;

    const states = ['AC', 'AL', 'AP', 'AM', 'BA', 'CE', 'DF', 'ES', 'GO', 'MA', 'MT', 'MS', 'MG', 'PA', 'PB', 'PR', 'PE', 'PI', 'RJ', 'RN', 'RS', 'RO', 'RR', 'SC', 'SP', 'SE', 'TO'];

    return (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-0 md:p-4 bg-stone-900/50 backdrop-blur-sm">
            <div
                role="dialog"
                aria-modal="true"
                aria-labelledby="quote-modal-title"
                className="bg-white border-0 md:border border-stone-200 rounded-none md:rounded-2xl w-full h-full md:h-auto md:max-w-4xl md:max-h-[90vh] shadow-2xl overflow-hidden flex flex-col"
            >
                {/* Header */}
                <div className="flex justify-between items-center p-4 md:p-6 border-b border-stone-200 bg-stone-50 sticky top-0 z-10">
                    <h3 id="quote-modal-title" className="text-lg md:text-xl font-bold text-stone-900">
                        {quote ? 'Editar Orcamento' : 'Novo Orcamento'}
                    </h3>
                    <button onClick={handleClose} className="text-stone-400 hover:text-stone-700 transition-colors p-2 -mr-2 rounded-lg" aria-label="Fechar modal">
                        <X className="w-5 h-5" />
                    </button>
                </div>

                {/* Form */}
                <form onSubmit={handleSubmit} className="flex-1 overflow-y-auto p-4 md:p-6 space-y-6">
                    {/* Client Section */}
                    <section>
                        <h4 className="text-sm font-semibold text-stone-700 mb-3 uppercase tracking-wide">Dados do Cliente</h4>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div className="md:col-span-2">
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                                    Nome <span className="text-red-500">*</span>
                                </label>
                                <input
                                    type="text"
                                    value={clientName}
                                    onChange={(e) => setClientName(e.target.value)}
                                    placeholder="Nome do cliente ou empresa"
                                    className={`w-full px-3 py-2.5 bg-stone-50 border ${hasError('clientName') ? 'border-red-500' : 'border-stone-200'} rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400`}
                                />
                                <ErrorMessage error={getError('clientName')} />
                            </div>

                            <div>
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">Email</label>
                                <input
                                    type="email"
                                    value={clientEmail}
                                    onChange={(e) => setClientEmail(e.target.value)}
                                    placeholder="email@exemplo.com"
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                />
                            </div>

                            <div>
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">Telefone</label>
                                <input
                                    type="text"
                                    value={clientPhone}
                                    onChange={(e) => setClientPhone(e.target.value)}
                                    placeholder="(00) 00000-0000"
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                />
                            </div>

                            <div>
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">CPF/CNPJ</label>
                                <input
                                    type="text"
                                    value={clientDocument}
                                    onChange={(e) => setClientDocument(e.target.value)}
                                    placeholder="000.000.000-00"
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                />
                            </div>

                            <div>
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">CEP</label>
                                <input
                                    type="text"
                                    value={clientZipCode}
                                    onChange={(e) => setClientZipCode(e.target.value)}
                                    placeholder="00000-000"
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                />
                            </div>

                            <div className="md:col-span-2">
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">Endereco</label>
                                <input
                                    type="text"
                                    value={clientAddress}
                                    onChange={(e) => setClientAddress(e.target.value)}
                                    placeholder="Rua, numero, bairro"
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                />
                            </div>

                            <div>
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">Cidade</label>
                                <input
                                    type="text"
                                    value={clientCity}
                                    onChange={(e) => setClientCity(e.target.value)}
                                    placeholder="Cidade"
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                />
                            </div>

                            <div>
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">Estado</label>
                                <select
                                    value={clientState}
                                    onChange={(e) => setClientState(e.target.value)}
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                >
                                    <option value="">Selecione</option>
                                    {states.map(state => (
                                        <option key={state} value={state}>{state}</option>
                                    ))}
                                </select>
                            </div>
                        </div>
                    </section>

                    {/* Quote Info Section */}
                    <section>
                        <h4 className="text-sm font-semibold text-stone-700 mb-3 uppercase tracking-wide">Dados do Orcamento</h4>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            <div className="md:col-span-2">
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                                    Titulo <span className="text-red-500">*</span>
                                </label>
                                <input
                                    type="text"
                                    value={title}
                                    onChange={(e) => setTitle(e.target.value)}
                                    placeholder="Ex: Desenvolvimento de Website"
                                    className={`w-full px-3 py-2.5 bg-stone-50 border ${hasError('title') ? 'border-red-500' : 'border-stone-200'} rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400`}
                                />
                                <ErrorMessage error={getError('title')} />
                            </div>

                            <div className="md:col-span-2">
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">Descricao</label>
                                <textarea
                                    value={description}
                                    onChange={(e) => setDescription(e.target.value)}
                                    placeholder="Descricao detalhada do servico ou produto"
                                    rows={3}
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400 resize-none"
                                />
                            </div>

                            <div>
                                <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                                    Validade <span className="text-red-500">*</span>
                                </label>
                                <input
                                    type="date"
                                    value={validUntil}
                                    onChange={(e) => setValidUntil(e.target.value)}
                                    className={`w-full px-3 py-2.5 bg-stone-50 border ${hasError('validUntil') ? 'border-red-500' : 'border-stone-200'} rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400`}
                                />
                                <ErrorMessage error={getError('validUntil')} />
                            </div>
                        </div>
                    </section>

                    {/* Items Section */}
                    <section>
                        <div className="flex items-center justify-between mb-3">
                            <h4 className="text-sm font-semibold text-stone-700 uppercase tracking-wide">Itens</h4>
                            <button
                                type="button"
                                onClick={addItem}
                                className="flex items-center gap-1 text-sm text-stone-600 hover:text-stone-900"
                            >
                                <Plus className="w-4 h-4" />
                                Adicionar Item
                            </button>
                        </div>

                        <div className="space-y-3">
                            {items.map((item, index) => (
                                <div key={item.id} className="p-4 bg-stone-50 rounded-xl border border-stone-200">
                                    <div className="grid grid-cols-12 gap-3">
                                        <div className="col-span-12 md:col-span-5">
                                            <label className="block text-xs font-medium text-stone-500 mb-1">Descricao</label>
                                            <input
                                                type="text"
                                                value={item.description}
                                                onChange={(e) => updateItem(item.id, 'description', e.target.value)}
                                                placeholder="Descricao do item"
                                                className={`w-full px-3 py-2 bg-white border ${hasError(`item_${index}_description`) ? 'border-red-500' : 'border-stone-200'} rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-stone-400`}
                                            />
                                        </div>

                                        <div className="col-span-4 md:col-span-2">
                                            <label className="block text-xs font-medium text-stone-500 mb-1">Qtd</label>
                                            <input
                                                type="number"
                                                step="0.01"
                                                value={item.quantity}
                                                onChange={(e) => updateItem(item.id, 'quantity', e.target.value)}
                                                className={`w-full px-3 py-2 bg-white border ${hasError(`item_${index}_quantity`) ? 'border-red-500' : 'border-stone-200'} rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-stone-400`}
                                            />
                                        </div>

                                        <div className="col-span-5 md:col-span-2">
                                            <label className="block text-xs font-medium text-stone-500 mb-1">Preco Unit.</label>
                                            <input
                                                type="number"
                                                step="0.01"
                                                value={item.unitPrice}
                                                onChange={(e) => updateItem(item.id, 'unitPrice', e.target.value)}
                                                placeholder="0,00"
                                                className={`w-full px-3 py-2 bg-white border ${hasError(`item_${index}_unitPrice`) ? 'border-red-500' : 'border-stone-200'} rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-stone-400`}
                                            />
                                        </div>

                                        <div className="col-span-12 md:col-span-2">
                                            <label className="block text-xs font-medium text-stone-500 mb-1">Categoria</label>
                                            <select
                                                value={item.categoryId}
                                                onChange={(e) => updateItem(item.id, 'categoryId', e.target.value)}
                                                className="w-full px-3 py-2 bg-white border border-stone-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                            >
                                                <option value="">Nenhuma</option>
                                                {categories.map(cat => (
                                                    <option key={cat.id} value={cat.id}>{cat.name}</option>
                                                ))}
                                            </select>
                                        </div>

                                        <div className="col-span-3 md:col-span-1 flex items-end">
                                            <button
                                                type="button"
                                                onClick={() => removeItem(item.id)}
                                                disabled={items.length === 1}
                                                className="w-full p-2 text-red-500 hover:bg-red-50 rounded-lg disabled:opacity-30 disabled:cursor-not-allowed"
                                            >
                                                <Trash2 className="w-4 h-4 mx-auto" />
                                            </button>
                                        </div>
                                    </div>

                                    <div className="mt-2 text-right text-sm text-stone-600">
                                        Total: <span className="font-medium text-stone-900">
                                            {formatCurrency((parseFloat(item.quantity) || 0) * (parseFloat(item.unitPrice) || 0))}
                                        </span>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </section>

                    {/* Discount Section */}
                    <section>
                        <h4 className="text-sm font-semibold text-stone-700 mb-3 uppercase tracking-wide">Desconto</h4>
                        <div className="flex gap-3">
                            <div className="flex-1">
                                <input
                                    type="number"
                                    step="0.01"
                                    value={discount}
                                    onChange={(e) => setDiscount(e.target.value)}
                                    placeholder="0"
                                    className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                                />
                            </div>
                            <select
                                value={discountType}
                                onChange={(e) => setDiscountType(e.target.value as 'PERCENT' | 'VALUE')}
                                className="px-4 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400"
                            >
                                <option value="VALUE">R$ (Valor)</option>
                                <option value="PERCENT">% (Percentual)</option>
                            </select>
                        </div>
                    </section>

                    {/* Notes Section */}
                    <section>
                        <h4 className="text-sm font-semibold text-stone-700 mb-3 uppercase tracking-wide">Observacoes</h4>
                        <textarea
                            value={notes}
                            onChange={(e) => setNotes(e.target.value)}
                            placeholder="Observacoes adicionais que aparecera no orcamento"
                            rows={3}
                            className="w-full px-3 py-2.5 bg-stone-50 border border-stone-200 rounded-xl text-sm focus:outline-none focus:ring-2 focus:ring-stone-400 resize-none"
                        />
                    </section>

                    {/* Totals */}
                    <section className="bg-stone-100 rounded-xl p-4 space-y-2">
                        <div className="flex justify-between text-sm">
                            <span className="text-stone-600">Subtotal:</span>
                            <span className="text-stone-900">{formatCurrency(subtotal)}</span>
                        </div>
                        {discountValue > 0 && (
                            <div className="flex justify-between text-sm">
                                <span className="text-stone-600">
                                    Desconto{discountType === 'PERCENT' ? ` (${discount}%)` : ''}:
                                </span>
                                <span className="text-red-600">-{formatCurrency(discountValue)}</span>
                            </div>
                        )}
                        <div className="flex justify-between text-lg font-bold pt-2 border-t border-stone-200">
                            <span className="text-stone-900">Total:</span>
                            <span className="text-stone-900">{formatCurrency(total)}</span>
                        </div>
                    </section>
                </form>

                {/* Footer */}
                <div className="p-4 md:p-6 border-t border-stone-200 bg-stone-50">
                    <div className="flex gap-3">
                        <button
                            type="button"
                            onClick={handleClose}
                            className="flex-1 py-3 text-stone-600 bg-stone-200 hover:bg-stone-300 rounded-xl font-medium transition-colors"
                        >
                            Cancelar
                        </button>
                        <button
                            onClick={handleSubmit}
                            disabled={isSubmitting}
                            className="flex-1 py-3 text-white bg-stone-800 hover:bg-stone-700 rounded-xl font-medium transition-colors flex items-center justify-center gap-2 disabled:opacity-50"
                        >
                            {isSubmitting ? (
                                <Loader2 className="w-4 h-4 animate-spin" />
                            ) : (
                                <Save className="w-4 h-4" />
                            )}
                            {quote ? 'Salvar Alteracoes' : 'Criar Orcamento'}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    );
};
