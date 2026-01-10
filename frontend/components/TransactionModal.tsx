import React, { useState, useEffect, useCallback } from 'react';
import { X, Save, Loader2 } from 'lucide-react';
import { Category, CategoryType, Status, Transaction } from '../types';
import { useFormValidation } from '../hooks/useFormValidation';
import { useEscapeKey } from '../hooks/useEscapeKey';
import { validateRequired, validateMaxLength, validatePositiveNumber, combineValidations } from '../utils/validation';
import { ErrorMessage } from './ErrorMessage';
import { MONTHS as BASE_MONTHS } from '../utils/constants';

const NEW_CATEGORY_VALUE = '__NEW_CATEGORY__';

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onSave: (transaction: Omit<Transaction, 'id'>) => void;
  onCreateCategory?: (name: string, type: CategoryType) => Promise<Category>;
  transaction?: Transaction | null;
  categories: Category[];
  companyId: string;
}

const MONTHS = [...BASE_MONTHS, 'Acumulado'];

export const TransactionModal: React.FC<Props> = ({ isOpen, onClose, onSave, onCreateCategory, transaction, categories, companyId }) => {
  const [formData, setFormData] = useState({
    month: 'Janeiro',
    year: new Date().getFullYear(),
    amount: 0,
    category: '',
    status: Status.OPEN,
    description: ''
  });
  const [isCreatingCategory, setIsCreatingCategory] = useState(false);
  const [newCategoryName, setNewCategoryName] = useState('');
  const [newCategoryType, setNewCategoryType] = useState<CategoryType>(CategoryType.EXPENSE);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const { validateFields, getError, hasError, hasErrors, clearAllErrors } = useFormValidation();

  const handleClose = useCallback(() => {
    clearAllErrors();
    setIsCreatingCategory(false);
    setNewCategoryName('');
    setNewCategoryType(CategoryType.EXPENSE);
    onClose();
  }, [clearAllErrors, onClose]);

  useEscapeKey(handleClose, isOpen);

  useEffect(() => {
    if (transaction) {
      setFormData({
        month: transaction.month,
        year: transaction.year,
        amount: transaction.amount,
        category: transaction.category,
        status: transaction.status,
        description: transaction.description || ''
      });
    } else {
      setFormData({
        month: 'Janeiro',
        year: new Date().getFullYear(),
        amount: 0,
        category: categories.length > 0 ? categories[0].name : '',
        status: Status.OPEN,
        description: ''
      });
    }
    setIsCreatingCategory(false);
    setNewCategoryName('');
    setNewCategoryType(CategoryType.EXPENSE);
    clearAllErrors();
  }, [transaction, isOpen, categories, clearAllErrors]);

  const validateForm = useCallback((): boolean => {
    const validations: Record<string, () => { isValid: boolean; error?: string }> = {
      description: () => combineValidations(
        validateRequired(formData.description, 'Descricao'),
        validateMaxLength(formData.description, 200, 'Descricao')
      ),
      amount: () => validatePositiveNumber(formData.amount, 'Valor'),
      month: () => validateRequired(formData.month, 'Mes'),
      year: () => validateRequired(formData.year, 'Ano')
    };

    if (isCreatingCategory) {
      validations.newCategoryName = () => combineValidations(
        validateRequired(newCategoryName, 'Nome da categoria'),
        validateMaxLength(newCategoryName, 100, 'Nome da categoria')
      );
    } else {
      validations.category = () => validateRequired(formData.category, 'Categoria');
    }

    return validateFields(validations);
  }, [validateFields, formData, isCreatingCategory, newCategoryName]);

  const handleSubmit = useCallback(async (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    setIsSubmitting(true);
    try {
      let categoryName = formData.category;

      if (isCreatingCategory && onCreateCategory) {
        const newCategory = await onCreateCategory(newCategoryName.trim(), newCategoryType);
        categoryName = newCategory.name;
      }

      onSave({ ...formData, category: categoryName, companyId });
      onClose();
    } catch (error) {
      console.error('Failed to save transaction', error);
    } finally {
      setIsSubmitting(false);
    }
  }, [validateForm, formData, companyId, onSave, onClose, isCreatingCategory, onCreateCategory, newCategoryName, newCategoryType]);

  if (!isOpen) return null;

  return (
    <div className="modal-overlay" role="dialog" aria-modal="true" aria-labelledby="modal-title">
      <div className="modal-content max-h-[90vh] overflow-y-auto animate-slideUp">
        <div className="flex justify-between items-center p-4 md:p-6 border-b border-border bg-background/50 sticky top-0 z-10">
          <h3 id="modal-title" className="text-lg md:text-xl font-bold text-foreground">
            {transaction ? 'Editar Transacao' : 'Nova Transacao'}
          </h3>
          <button onClick={handleClose} className="text-muted hover:text-foreground transition-colors p-2 -mr-2 rounded-lg hover:bg-primary-50">
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-4 md:p-6 space-y-4">
          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">
              Descricao <span className="text-destructive">*</span>
            </label>
            <input
              type="text"
              className={`input ${hasError('description') ? 'input-error' : ''}`}
              placeholder="Ex: Salario, Aluguel..."
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            />
            <ErrorMessage error={getError('description')} />
          </div>

          <div className="grid grid-cols-2 gap-3 md:gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">
                Valor <span className="text-destructive">*</span>
              </label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-muted text-sm">R$</span>
                <input
                  type="number"
                  step="0.01"
                  className={`input pl-10 ${hasError('amount') ? 'input-error' : ''}`}
                  placeholder="0,00"
                  value={formData.amount || ''}
                  onChange={(e) => setFormData({ ...formData, amount: parseFloat(e.target.value) || 0 })}
                />
              </div>
              <ErrorMessage error={getError('amount')} />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">
                Categoria <span className="text-destructive">*</span>
              </label>
              {!isCreatingCategory ? (
                <>
                  <select
                    className={`select ${hasError('category') ? 'input-error' : ''}`}
                    value={formData.category}
                    onChange={(e) => {
                      if (e.target.value === NEW_CATEGORY_VALUE) {
                        setIsCreatingCategory(true);
                        setFormData({ ...formData, category: '' });
                      } else {
                        setFormData({ ...formData, category: e.target.value });
                      }
                    }}
                  >
                    <option value="" disabled>Selecione...</option>
                    {categories.map(c => (
                      <option key={c.id} value={c.name}>{c.name}</option>
                    ))}
                    {onCreateCategory && (
                      <option value={NEW_CATEGORY_VALUE} className="font-medium">+ Nova categoria</option>
                    )}
                  </select>
                  <ErrorMessage error={getError('category')} />
                </>
              ) : (
                <div className="space-y-2">
                  <div className="flex gap-2">
                    <input
                      type="text"
                      className={`input flex-1 ${hasError('newCategoryName') ? 'input-error' : ''}`}
                      placeholder="Nome da categoria"
                      value={newCategoryName}
                      onChange={(e) => setNewCategoryName(e.target.value)}
                    />
                    <button
                      type="button"
                      onClick={() => {
                        setIsCreatingCategory(false);
                        setNewCategoryName('');
                        setFormData({ ...formData, category: categories.length > 0 ? categories[0].name : '' });
                      }}
                      className="px-3 py-2 text-muted hover:text-foreground bg-card hover:bg-primary-50 rounded-lg transition-colors border border-border"
                      title="Cancelar"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                  <select
                    className="select text-sm"
                    value={newCategoryType}
                    onChange={(e) => setNewCategoryType(e.target.value as CategoryType)}
                  >
                    <option value={CategoryType.EXPENSE}>Despesa</option>
                    <option value={CategoryType.INCOME}>Receita</option>
                  </select>
                  <ErrorMessage error={getError('newCategoryName')} />
                </div>
              )}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3 md:gap-4">
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">
                Mes <span className="text-destructive">*</span>
              </label>
              <select
                className={`select ${hasError('month') ? 'input-error' : ''}`}
                value={formData.month}
                onChange={(e) => setFormData({ ...formData, month: e.target.value })}
              >
                {MONTHS.map(m => (
                  <option key={m} value={m}>{m}</option>
                ))}
              </select>
              <ErrorMessage error={getError('month')} />
            </div>
            <div>
              <label className="block text-sm font-medium text-foreground mb-1.5">
                Ano <span className="text-destructive">*</span>
              </label>
              <input
                type="number"
                className={`input ${hasError('year') ? 'input-error' : ''}`}
                value={formData.year}
                onChange={(e) => setFormData({ ...formData, year: parseInt(e.target.value) || new Date().getFullYear() })}
              />
              <ErrorMessage error={getError('year')} />
            </div>
          </div>

          <div>
            <label className="block text-sm font-medium text-foreground mb-1.5">Status</label>
            <select
              value={formData.status}
              onChange={(e) => setFormData({ ...formData, status: e.target.value as Status })}
              className="select"
            >
              <option value={Status.PAID}>PAGO</option>
              <option value={Status.OPEN}>ABERTO</option>
            </select>
          </div>

          <button
            type="submit"
            disabled={hasErrors() || isSubmitting}
            className={`w-full mt-4 flex items-center justify-center gap-2 ${
              hasErrors() || isSubmitting
                ? 'bg-muted/20 text-muted cursor-not-allowed py-3 rounded-xl border border-border'
                : 'btn-primary'
            }`}
          >
            {isSubmitting ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <Save className="w-4 h-4" />
            )}
            {isSubmitting ? 'Salvando...' : (transaction ? 'Salvar Alteracoes' : 'Salvar Transacao')}
          </button>
        </form>
      </div>
    </div>
  );
};
