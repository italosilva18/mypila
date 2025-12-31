import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { Trash2, Plus, RefreshCw, Calendar, Check } from 'lucide-react';
import { api } from '../services/api';
import { RecurringTransaction, Category } from '../types';
import { useDateFilter } from '../contexts/DateFilterContext';
import { useFormValidation } from '../hooks/useFormValidation';
import { validateRequired, validateMaxLength, validatePositiveNumber, validateRange, combineValidations } from '../utils/validation';
import { ErrorMessage } from '../components/ErrorMessage';
import { formatCurrency } from '../utils/currency';

export const Recurring: React.FC = () => {
    const { companyId } = useParams<{ companyId: string }>();
    const { month, year } = useDateFilter();
    const [rules, setRules] = useState<RecurringTransaction[]>([]);
    const [categories, setCategories] = useState<Category[]>([]);
    const [loading, setLoading] = useState(true);
    const [showForm, setShowForm] = useState(false);
    const [processing, setProcessing] = useState(false);

    // Form State
    const [formData, setFormData] = useState({
        description: '',
        amount: '',
        category: '',
        dayOfMonth: 1
    });

    const { validateFields, getError, hasError, hasErrors, clearAllErrors } = useFormValidation();

    useEffect(() => {
        loadData();
    }, [companyId]);

    const loadData = async () => {
        if (!companyId) return;
        try {
            setLoading(true);
            const [fetchedRules, fetchedCategories] = await Promise.all([
                api.getRecurring(companyId),
                api.getCategories(companyId)
            ]);
            setRules(fetchedRules);
            setCategories(fetchedCategories);
            if (fetchedCategories.length > 0) {
                setFormData(prev => ({ ...prev, category: fetchedCategories[0].name }));
            }
        } catch (error) {
            console.error('Error loading recurring data', error);
        } finally {
            setLoading(false);
        }
    };

    const handleDelete = async (id: string) => {
        if (confirm('Tem certeza que deseja remover esta recorrência?')) {
            await api.deleteRecurring(id);
            setRules(rules.filter(r => r.id !== id));
        }
    };

    const validateForm = (): boolean => {
        return validateFields({
            description: () => combineValidations(
                validateRequired(formData.description, 'Descrição'),
                validateMaxLength(formData.description, 200, 'Descrição')
            ),
            amount: () => validatePositiveNumber(formData.amount, 'Valor'),
            category: () => validateRequired(formData.category, 'Categoria'),
            dayOfMonth: () => validateRange(formData.dayOfMonth, 1, 31, 'Dia do mês')
        });
    };

    const handleCreate = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!companyId) return;

        if (!validateForm()) {
            return;
        }

        try {
            const newRule = await api.createRecurring({
                companyId,
                description: formData.description,
                amount: Number(formData.amount),
                category: formData.category,
                dayOfMonth: Number(formData.dayOfMonth)
            });
            setRules([...rules, newRule]);
            setShowForm(false);
            setFormData({ description: '', amount: '', category: categories[0]?.name || '', dayOfMonth: 1 });
            clearAllErrors();
        } catch (error) {
            console.error('Failed to create rule', error);
        }
    };

    const handleProcess = async () => {
        if (!companyId) return;
        try {
            setProcessing(true);
            const result = await api.processRecurring(companyId, month, year);
            alert(`Processado! ${result.created} novas transações geradas para ${month}/${year}.`);
        } catch (error) {
            alert('Erro ao processar.');
        } finally {
            setProcessing(false);
        }
    };

    const handleFormToggle = () => {
        setShowForm(!showForm);
        if (showForm) {
            clearAllErrors();
        }
    };

    if (loading) return <div className="text-stone-900">Carregando...</div>;

    return (
        <div className="space-y-3 md:space-y-6 mobile-content-padding">
            <header className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2 md:gap-4 px-1 md:px-0">
                <div>
                    <h1 className="text-base md:text-2xl font-bold text-stone-900 flex items-center gap-1.5 md:gap-2">
                        <RefreshCw className="w-4 h-4 md:w-6 md:h-6 text-stone-600" />
                        Recorrentes
                    </h1>
                    <p className="text-stone-500 text-xs md:text-sm">Despesas fixas mensais.</p>
                </div>
                <div className="flex gap-2 w-full sm:w-auto">
                    <button
                        onClick={handleProcess}
                        disabled={processing}
                        className="flex-1 sm:flex-none flex items-center justify-center gap-1.5 bg-green-600 active:bg-green-500 text-white px-3 md:px-4 py-2 rounded-lg md:rounded-xl text-xs md:text-sm transition-all shadow-lg shadow-green-600/20 disabled:opacity-50"
                    >
                        <Check className="w-3.5 h-3.5 md:w-4 md:h-4" />
                        {processing ? 'Processando...' : 'Gerar'}
                    </button>
                    <button
                        onClick={handleFormToggle}
                        className="flex-1 sm:flex-none flex items-center justify-center gap-1.5 bg-stone-800 active:bg-stone-700 text-white px-3 md:px-4 py-2 rounded-lg md:rounded-xl text-xs md:text-sm transition-all shadow-lg shadow-stone-900/20"
                    >
                        <Plus className="w-3.5 h-3.5 md:w-4 md:h-4" />
                        Nova
                    </button>
                </div>
            </header>

            {/* Form */}
            {showForm && (
                <div className="bg-white border border-stone-100 md:border-stone-200 rounded-lg md:rounded-2xl p-3 md:p-6 animate-in fade-in slide-in-from-top-4">
                    <h3 className="text-sm md:text-lg font-bold text-stone-900 mb-3 md:mb-4">Nova Recorrência</h3>
                    <form onSubmit={handleCreate} className="space-y-3 md:space-y-4">
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-2 md:gap-4">
                            <div className="md:col-span-2">
                                <label className="text-xs md:text-sm font-medium text-stone-600 mb-1 md:mb-1.5 block">
                                    Descrição <span className="text-red-500">*</span>
                                </label>
                                <input
                                    type="text"
                                    placeholder="Ex: Aluguel"
                                    className={`w-full bg-stone-50 border ${
                                        hasError('description') ? 'border-red-500 focus:ring-red-400' : 'border-stone-200 focus:ring-stone-400'
                                    } rounded-lg px-3 py-2 text-sm md:text-base text-stone-900 focus:ring-2 outline-none transition-colors`}
                                    value={formData.description}
                                    onChange={e => setFormData({ ...formData, description: e.target.value })}
                                />
                                <ErrorMessage error={getError('description')} />
                            </div>

                            <div>
                                <label className="text-xs md:text-sm font-medium text-stone-600 mb-1 md:mb-1.5 block">
                                    Valor <span className="text-red-500">*</span>
                                </label>
                                <input
                                    type="number"
                                    step="0.01"
                                    placeholder="0.00"
                                    className={`w-full bg-stone-50 border ${
                                        hasError('amount') ? 'border-red-500 focus:ring-red-400' : 'border-stone-200 focus:ring-stone-400'
                                    } rounded-lg px-3 py-2 text-sm md:text-base text-stone-900 focus:ring-2 outline-none transition-colors`}
                                    value={formData.amount}
                                    onChange={e => setFormData({ ...formData, amount: e.target.value })}
                                />
                                <ErrorMessage error={getError('amount')} />
                            </div>

                            <div>
                                <label className="text-xs md:text-sm font-medium text-stone-600 mb-1 md:mb-1.5 block">
                                    Dia <span className="text-red-500">*</span>
                                </label>
                                <input
                                    type="number"
                                    min="1"
                                    max="31"
                                    className={`w-full bg-stone-50 border ${
                                        hasError('dayOfMonth') ? 'border-red-500 focus:ring-red-400' : 'border-stone-200 focus:ring-stone-400'
                                    } rounded-lg px-3 py-2 text-sm md:text-base text-stone-900 focus:ring-2 outline-none transition-colors`}
                                    value={formData.dayOfMonth}
                                    onChange={e => setFormData({ ...formData, dayOfMonth: Number(e.target.value) })}
                                />
                                <ErrorMessage error={getError('dayOfMonth')} />
                            </div>

                            <div className="md:col-span-2">
                                <label className="text-xs md:text-sm font-medium text-stone-600 mb-1 md:mb-1.5 block">
                                    Categoria <span className="text-red-500">*</span>
                                </label>
                                <select
                                    className={`w-full bg-stone-50 border ${
                                        hasError('category') ? 'border-red-500 focus:ring-red-400' : 'border-stone-200 focus:ring-stone-400'
                                    } rounded-lg px-3 py-2 text-sm md:text-base text-stone-900 focus:ring-2 outline-none transition-colors`}
                                    value={formData.category}
                                    onChange={e => setFormData({ ...formData, category: e.target.value })}
                                >
                                    {categories.map(c => <option key={c.id} value={c.name} className="bg-white">{c.name}</option>)}
                                </select>
                                <ErrorMessage error={getError('category')} />
                            </div>
                        </div>

                        <button
                            type="submit"
                            disabled={hasErrors()}
                            className={`w-full rounded-lg px-4 py-2.5 text-sm md:text-base font-medium transition-all ${
                                hasErrors()
                                    ? 'bg-stone-300 text-stone-500 cursor-not-allowed'
                                    : 'bg-stone-800 active:bg-stone-700 text-white'
                            }`}
                        >
                            Salvar
                        </button>
                    </form>
                </div>
            )}

            {/* List */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-2 md:gap-4">
                {rules.map(rule => (
                    <div key={rule.id} className="bg-white/70 border border-stone-100 md:border-stone-200 rounded-lg md:rounded-xl p-3 md:p-4 flex flex-col justify-between group hover:border-stone-300 hover:shadow-card transition-all">
                        <div className="flex justify-between items-start mb-2 md:mb-4">
                            <div className="min-w-0 flex-1">
                                <h3 className="font-bold text-stone-900 text-sm md:text-lg truncate">{rule.description}</h3>
                                <div className="flex items-center gap-1.5 md:gap-2 mt-0.5 md:mt-1">
                                    <span className="text-[9px] md:text-xs bg-stone-100 text-stone-600 px-1.5 md:px-2 py-0.5 rounded-full">{rule.category}</span>
                                    <span className="text-[9px] md:text-xs text-stone-400 flex items-center gap-0.5 md:gap-1">
                                        <Calendar className="w-2.5 h-2.5 md:w-3 md:h-3" /> Dia {rule.dayOfMonth}
                                    </span>
                                </div>
                            </div>
                            <span className="text-sm md:text-xl font-bold text-stone-900 shrink-0 ml-2">{formatCurrency(rule.amount)}</span>
                        </div>
                        <div className="flex justify-end pt-2 md:pt-4 border-t border-stone-100">
                            <button
                                onClick={() => handleDelete(rule.id)}
                                className="text-stone-400 active:text-red-600 transition-colors p-1.5 md:p-2 rounded-lg active:bg-red-50"
                            >
                                <Trash2 className="w-3.5 h-3.5 md:w-4 md:h-4" />
                            </button>
                        </div>
                    </div>
                ))}

                {rules.length === 0 && !loading && (
                    <div className="col-span-full text-center py-8 md:py-12 text-stone-500 text-sm">
                        Nenhuma regra recorrente cadastrada.
                    </div>
                )}
            </div>
        </div>
    );
};
