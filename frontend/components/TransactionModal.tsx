import React, { useState, useEffect } from 'react';
import { X, Save } from 'lucide-react';
import { Category, Status, Transaction } from '../types';
import { api } from '../services/api';
import { useFormValidation } from '../hooks/useFormValidation';
import { validateRequired, validateMaxLength, validatePositiveNumber, combineValidations } from '../utils/validation';
import { ErrorMessage } from './ErrorMessage';

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onSave: (transaction: Omit<Transaction, 'id'>) => void;
  transaction?: Transaction | null;
  categories: Category[];
  companyId: string;
}

const MONTHS = [
  'Janeiro', 'Fevereiro', 'Março', 'Abril', 'Maio', 'Junho',
  'Julho', 'Agosto', 'Setembro', 'Outubro', 'Novembro', 'Dezembro', 'Acumulado'
];

export const TransactionModal: React.FC<Props> = ({ isOpen, onClose, onSave, transaction, categories, companyId }) => {
  const [formData, setFormData] = useState({
    month: 'Janeiro',
    year: new Date().getFullYear(),
    amount: 0,
    category: '',
    status: Status.OPEN,
    description: ''
  });

  const { validateFields, getError, hasError, hasErrors, clearAllErrors } = useFormValidation();

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
      // Reset defaults for new entry
      setFormData({
        month: 'Janeiro',
        year: new Date().getFullYear(),
        amount: 0,
        category: categories.length > 0 ? categories[0].name : '',
        status: Status.OPEN,
        description: ''
      });
    }
    clearAllErrors();
  }, [transaction, isOpen, categories, clearAllErrors]);

  if (!isOpen) return null;

  const validateForm = (): boolean => {
    return validateFields({
      description: () => combineValidations(
        validateRequired(formData.description, 'Descrição'),
        validateMaxLength(formData.description, 200, 'Descrição')
      ),
      amount: () => validatePositiveNumber(formData.amount, 'Valor'),
      category: () => validateRequired(formData.category, 'Categoria'),
      month: () => validateRequired(formData.month, 'Mês'),
      year: () => validateRequired(formData.year, 'Ano')
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    onSave(formData);
    onClose();
  };

  const handleClose = () => {
    clearAllErrors();
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-0 md:p-4 bg-stone-900/50 backdrop-blur-sm" role="dialog" aria-modal="true" aria-labelledby="modal-title">
      <div className="bg-white border-0 md:border border-stone-200 rounded-none md:rounded-2xl w-full h-full md:h-auto md:max-w-md shadow-2xl transform transition-all overflow-y-auto">
        <div className="flex justify-between items-center p-4 md:p-6 border-b border-stone-200 bg-stone-50 sticky top-0 z-10">
          <h3 id="modal-title" className="text-lg md:text-xl font-bold text-stone-900">
            {transaction ? 'Editar Transação' : 'Nova Transação'}
          </h3>
          <button onClick={handleClose} className="text-stone-400 active:text-stone-700 md:hover:text-stone-700 transition-colors p-2 -mr-2 rounded-lg active:bg-stone-100" aria-label="Fechar modal">
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-4 md:p-6 space-y-4">
          <div>
            <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
              Descrição <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              className={`w-full px-3 md:px-4 py-3 md:py-2.5 bg-stone-50 border ${
                hasError('description') ? 'border-red-500 focus:ring-red-400' : 'border-stone-200 focus:ring-stone-400'
              } rounded-lg md:rounded-xl text-sm md:text-base text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 transition-colors min-h-[44px]`}
              placeholder="Ex: Salário, Aluguel..."
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            />
            <ErrorMessage error={getError('description')} />
          </div>

          <div className="grid grid-cols-2 gap-3 md:gap-4">
            <div>
              <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                Valor <span className="text-red-500">*</span>
              </label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-stone-400 text-sm">R$</span>
                <input
                  type="number"
                  step="0.01"
                  className={`w-full pl-10 pr-3 md:pr-4 py-3 md:py-2.5 bg-stone-50 border ${
                    hasError('amount') ? 'border-red-500 focus:ring-red-400' : 'border-stone-200 focus:ring-stone-400'
                  } rounded-lg md:rounded-xl text-sm md:text-base text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 transition-colors min-h-[44px]`}
                  placeholder="0,00"
                  value={formData.amount || ''}
                  onChange={(e) => setFormData({ ...formData, amount: parseFloat(e.target.value) || 0 })}
                />
              </div>
              <ErrorMessage error={getError('amount')} />
            </div>
            <div>
              <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                Categoria <span className="text-red-500">*</span>
              </label>
              <select
                className={`w-full px-3 md:px-4 py-3 md:py-2.5 bg-stone-50 border ${
                  hasError('category') ? 'border-red-500 focus:ring-red-400' : 'border-stone-200 focus:ring-stone-400'
                } rounded-lg md:rounded-xl text-sm md:text-base text-stone-900 focus:outline-none focus:ring-2 transition-colors min-h-[44px]`}
                value={formData.category}
                onChange={(e) => setFormData({ ...formData, category: e.target.value })}
              >
                <option value="" disabled>Selecione...</option>
                {categories.map(c => (
                  <option key={c.id} value={c.name}>{c.name}</option>
                ))}
              </select>
              <ErrorMessage error={getError('category')} />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3 md:gap-4">
            <div>
              <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                Mês <span className="text-red-500">*</span>
              </label>
              <select
                className={`w-full px-3 md:px-4 py-3 md:py-2.5 bg-stone-50 border ${
                  hasError('month') ? 'border-red-500 focus:ring-red-400' : 'border-stone-200 focus:ring-stone-400'
                } rounded-lg md:rounded-xl text-sm md:text-base text-stone-900 focus:outline-none focus:ring-2 transition-colors min-h-[44px]`}
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
              <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
                Ano <span className="text-red-500">*</span>
              </label>
              <input
                type="number"
                className={`w-full px-3 md:px-4 py-3 md:py-2.5 bg-stone-50 border ${
                  hasError('year') ? 'border-red-500 focus:ring-red-400' : 'border-stone-200 focus:ring-stone-400'
                } rounded-lg md:rounded-xl text-sm md:text-base text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 transition-colors min-h-[44px]`}
                value={formData.year}
                onChange={(e) => setFormData({ ...formData, year: parseInt(e.target.value) || new Date().getFullYear() })}
              />
              <ErrorMessage error={getError('year')} />
            </div>
          </div>

          <div>
            <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">Status</label>
            <select
              value={formData.status}
              onChange={(e) => setFormData({ ...formData, status: e.target.value as Status })}
              className="w-full px-3 md:px-4 py-3 md:py-2.5 bg-stone-50 border border-stone-200 rounded-lg md:rounded-xl text-sm md:text-base text-stone-900 focus:outline-none focus:ring-2 focus:ring-stone-400 min-h-[44px]"
            >
              <option value={Status.PAID}>PAGO</option>
              <option value={Status.OPEN}>ABERTO</option>
            </select>
          </div>

          <button
            type="submit"
            disabled={hasErrors()}
            className={`w-full mt-4 font-medium py-3.5 md:py-3 rounded-lg md:rounded-xl transition-all shadow-lg flex items-center justify-center gap-2 text-sm md:text-base min-h-[48px] ${
              hasErrors()
                ? 'bg-stone-300 text-stone-500 cursor-not-allowed'
                : 'bg-stone-800 active:bg-stone-700 md:hover:bg-stone-700 text-white shadow-stone-900/20 active:scale-[0.98]'
            }`}
          >
            <Save className="w-4 h-4" />
            {transaction ? 'Salvar Alterações' : 'Salvar Transação'}
          </button>
        </form>
      </div>
    </div>
  );
};
