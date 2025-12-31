import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { api } from '../services/api';
import { Category, CategoryType } from '../types';
import { Plus, Trash2, Tag, Loader2, ArrowUpCircle, ArrowDownCircle, X, Save } from 'lucide-react';
import { useFormValidation } from '../hooks/useFormValidation';
import { validateRequired, validateMaxLength, validatePositiveNumber, combineValidations } from '../utils/validation';
import { ErrorMessage } from '../components/ErrorMessage';

const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('pt-BR', { style: 'currency', currency: 'BRL' }).format(value);
};

export const Categories: React.FC = () => {
    const { companyId } = useParams<{ companyId: string }>();
    const [categories, setCategories] = useState<Category[]>([]);
    const [loading, setLoading] = useState(true);

    // Form State
    const [editingId, setEditingId] = useState<string | null>(null);
    const [name, setName] = useState('');
    const [type, setType] = useState<CategoryType>(CategoryType.EXPENSE);
    const [color, setColor] = useState('#78716c');
    const [budget, setBudget] = useState('');

    const [isSubmitting, setIsSubmitting] = useState(false);
    const { validateFields, getError, hasError, clearAllErrors } = useFormValidation();

    useEffect(() => {
        if (companyId) {
            loadCategories();
        }
    }, [companyId]);

    const loadCategories = async () => {
        try {
            setLoading(true);
            const data = await api.getCategories(companyId!);
            setCategories(data);
        } catch (err) {
            console.error('Failed to load categories', err);
        } finally {
            setLoading(false);
        }
    };

    const validateForm = (): boolean => {
        const validations: any = {
            name: () => combineValidations(
                validateRequired(name, 'Nome'),
                validateMaxLength(name, 100, 'Nome')
            )
        };

        // Only validate budget if it's filled
        if (budget.trim()) {
            validations.budget = () => validatePositiveNumber(budget, 'Orçamento');
        }

        return validateFields(validations);
    };

    const resetForm = () => {
        setEditingId(null);
        setName('');
        setType(CategoryType.EXPENSE);
        setColor('#78716c');
        setBudget('');
        clearAllErrors();
    };

    const handleEdit = (category: Category) => {
        setEditingId(category.id);
        setName(category.name);
        setType(category.type);
        setColor(category.color || '#78716c');
        setBudget(category.budget ? category.budget.toString() : '');
    };

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        if (!companyId) return;

        if (!validateForm()) return;

        try {
            setIsSubmitting(true);
            const budgetValue = budget ? parseFloat(budget) : 0;

            if (editingId) {
                const updated = await api.updateCategory(editingId, {
                    name,
                    type,
                    color,
                    budget: budgetValue
                });
                setCategories(categories.map(c => c.id === editingId ? updated : c));
            } else {
                const newCategory = await api.createCategory(companyId, name, type, color, budgetValue);
                setCategories([...categories, newCategory]);
            }
            resetForm();
        } catch (err) {
            console.error('Failed to save category', err);
        } finally {
            setIsSubmitting(false);
        }
    };

    const handleDelete = async (id: string) => {
        if (!confirm('Tem certeza que deseja excluir esta categoria?')) return;
        try {
            await api.deleteCategory(id);
            setCategories(categories.filter(c => c.id !== id));
        } catch (err) {
            console.error('Failed to delete category', err);
        }
    };

    return (
        <div className="space-y-3 md:space-y-6 mobile-content-padding">
            <header className="px-1 md:px-0">
                <h1 className="text-base md:text-2xl font-bold text-stone-900 flex items-center gap-1.5 md:gap-2">
                    <Tag className="w-4 h-4 md:w-6 md:h-6 text-stone-600" />
                    Gerenciar Categorias
                </h1>
                <p className="text-stone-500 text-xs md:text-sm">Crie e organize as categorias das suas transações.</p>
            </header>

            {/* Create Form */}
            <div className="bg-white/70 border border-stone-100 md:border-stone-200 rounded-lg md:rounded-2xl p-3 md:p-6">
                <form onSubmit={handleSubmit} className="flex flex-col md:flex-row gap-2 md:gap-4 items-end">
                    <div className="flex-1 w-full">
                        <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1 md:mb-1.5">Nome</label>
                        <input
                            type="text"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            placeholder="Ex: Marketing..."
                            className={`w-full px-3 md:px-4 py-2 md:py-2.5 bg-stone-50 border rounded-lg md:rounded-xl text-sm md:text-base text-stone-900 focus:outline-none focus:ring-2 ${
                                hasError('name')
                                    ? 'border-red-500 focus:ring-red-400'
                                    : 'border-stone-200 focus:ring-stone-400'
                            }`}
                        />
                        <ErrorMessage error={getError('name')} />
                    </div>

                    <div className="grid grid-cols-3 md:flex gap-2 w-full md:w-auto">
                        <div className="w-full md:w-20">
                            <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1 md:mb-1.5">Cor</label>
                            <input
                                type="color"
                                value={color}
                                onChange={(e) => setColor(e.target.value)}
                                className="h-9 md:h-11 w-full bg-transparent cursor-pointer rounded-lg md:rounded-xl"
                            />
                        </div>

                        <div className="w-full md:w-32">
                            <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1 md:mb-1.5">Meta</label>
                            <input
                                type="number"
                                value={budget}
                                onChange={(e) => setBudget(e.target.value)}
                                placeholder="R$ 0"
                                className={`w-full px-2 md:px-4 py-2 md:py-2.5 bg-stone-50 border rounded-lg md:rounded-xl text-sm md:text-base text-stone-900 focus:outline-none focus:ring-2 ${
                                    hasError('budget')
                                        ? 'border-red-500 focus:ring-red-400'
                                        : 'border-stone-200 focus:ring-stone-400'
                                }`}
                            />
                            <ErrorMessage error={getError('budget')} />
                        </div>

                        <div className="w-full md:w-32">
                            <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1 md:mb-1.5">Tipo</label>
                            <select
                                value={type}
                                onChange={(e) => setType(e.target.value as CategoryType)}
                                className="w-full px-2 md:px-4 py-2 md:py-2.5 bg-stone-50 border border-stone-200 rounded-lg md:rounded-xl text-sm md:text-base text-stone-900 focus:outline-none focus:ring-2 focus:ring-stone-400"
                            >
                                <option value={CategoryType.EXPENSE}>Despesa</option>
                                <option value={CategoryType.INCOME}>Receita</option>
                            </select>
                        </div>
                    </div>

                    <div className="flex gap-2 w-full md:w-auto">
                        {editingId && (
                            <button
                                type="button"
                                onClick={resetForm}
                                className="bg-stone-200 active:bg-stone-300 text-stone-700 px-3 md:px-4 py-2 md:py-2.5 rounded-lg md:rounded-xl font-medium transition-all flex items-center gap-2"
                            >
                                <X className="w-4 h-4" />
                            </button>
                        )}
                        <button
                            type="submit"
                            disabled={isSubmitting || !name.trim()}
                            className="flex-1 bg-stone-800 active:bg-stone-700 text-white px-4 md:px-6 py-2 md:py-2.5 rounded-lg md:rounded-xl text-sm md:text-base font-medium transition-all shadow-lg shadow-stone-900/20 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2 justify-center"
                        >
                            {isSubmitting ? <Loader2 className="w-4 h-4 animate-spin" /> : (editingId ? <Save className="w-4 h-4" /> : <Plus className="w-4 h-4" />)}
                            {editingId ? 'Salvar' : 'Adicionar'}
                        </button>
                    </div>
                </form>
            </div>

            {/* List - Desktop Table */}
            <div className="bg-white/70 border border-stone-100 md:border-stone-200 rounded-lg md:rounded-2xl overflow-hidden hidden md:block">
                {loading ? (
                    <div className="p-8 text-center text-stone-500">Carregando...</div>
                ) : categories.length === 0 ? (
                    <div className="p-8 text-center text-stone-500">Nenhuma categoria cadastrada.</div>
                ) : (
                    <table className="w-full text-left">
                        <thead className="bg-stone-50 text-stone-500 text-xs uppercase font-semibold">
                            <tr>
                                <th className="px-6 py-4">Cor</th>
                                <th className="px-6 py-4">Nome</th>
                                <th className="px-6 py-4">Orçamento</th>
                                <th className="px-6 py-4">Tipo</th>
                                <th className="px-6 py-4 text-right">Ações</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-stone-100">
                            {categories.map((category) => (
                                <tr
                                    key={category.id}
                                    className="group hover:bg-stone-50/50 cursor-pointer transition-colors"
                                    onClick={() => handleEdit(category)}
                                >
                                    <td className="px-6 py-4">
                                        <div
                                            className="w-6 h-6 rounded-full border border-stone-200 shadow-sm"
                                            style={{ backgroundColor: category.color || '#78716c' }}
                                        />
                                    </td>
                                    <td className="px-6 py-4 text-stone-900 font-medium">{category.name}</td>
                                    <td className="px-6 py-4 text-stone-500">
                                        {category.budget ? formatCurrency(category.budget) : '-'}
                                    </td>
                                    <td className="px-6 py-4">
                                        <span className={`inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full text-xs font-medium border ${category.type === CategoryType.INCOME
                                            ? 'bg-green-100 text-green-700 border-green-200'
                                            : 'bg-red-100 text-red-700 border-red-200'
                                            }`}>
                                            {category.type === CategoryType.INCOME ? <ArrowUpCircle className="w-3 h-3" /> : <ArrowDownCircle className="w-3 h-3" />}
                                            {category.type === CategoryType.INCOME ? 'Receita' : 'Despesa'}
                                        </span>
                                    </td>
                                    <td className="px-6 py-4 text-right">
                                        <button
                                            onClick={(e) => {
                                                e.stopPropagation();
                                                handleDelete(category.id);
                                            }}
                                            className="text-stone-400 hover:text-red-600 transition-colors p-2 hover:bg-red-50 rounded-lg"
                                            title="Excluir"
                                        >
                                            <Trash2 className="w-4 h-4" />
                                        </button>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>

            {/* List - Mobile Cards */}
            <div className="md:hidden space-y-2">
                {loading ? (
                    <div className="p-6 text-center text-stone-500 text-sm">Carregando...</div>
                ) : categories.length === 0 ? (
                    <div className="p-6 text-center text-stone-500 text-sm">Nenhuma categoria cadastrada.</div>
                ) : (
                    categories.map((category) => (
                        <div
                            key={category.id}
                            className="bg-white/70 border border-stone-100 rounded-lg p-3 active:bg-stone-50 transition-colors"
                            onClick={() => handleEdit(category)}
                        >
                            <div className="flex items-center gap-2.5">
                                <div
                                    className="w-5 h-5 rounded-full border border-stone-200 shadow-sm shrink-0"
                                    style={{ backgroundColor: category.color || '#78716c' }}
                                />
                                <div className="flex-1 min-w-0">
                                    <div className="flex items-center justify-between gap-2">
                                        <p className="text-sm font-medium text-stone-900 truncate">{category.name}</p>
                                        <span className={`shrink-0 inline-flex items-center gap-0.5 px-1.5 py-0.5 rounded text-[9px] font-medium ${category.type === CategoryType.INCOME
                                            ? 'bg-green-100 text-green-700'
                                            : 'bg-red-100 text-red-700'
                                            }`}>
                                            {category.type === CategoryType.INCOME ? <ArrowUpCircle className="w-2.5 h-2.5" /> : <ArrowDownCircle className="w-2.5 h-2.5" />}
                                            {category.type === CategoryType.INCOME ? 'Receita' : 'Despesa'}
                                        </span>
                                    </div>
                                    <p className="text-xs text-stone-500 mt-0.5">
                                        Meta: {category.budget ? formatCurrency(category.budget) : '-'}
                                    </p>
                                </div>
                                <button
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        handleDelete(category.id);
                                    }}
                                    className="text-stone-400 active:text-red-600 p-1.5 rounded-md active:bg-red-50 transition-colors shrink-0"
                                >
                                    <Trash2 className="w-4 h-4" />
                                </button>
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};
