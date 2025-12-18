import React, { useState, useEffect } from 'react';
import { X, Save } from 'lucide-react';
import { Company } from '../types';
import { useFormValidation } from '../hooks/useFormValidation';
import { validateRequired, validateMaxLength, combineValidations } from '../utils/validation';
import { ErrorMessage } from './ErrorMessage';

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onSave: (name: string) => void;
  company?: Company | null;
}

export const CompanyModal: React.FC<Props> = ({ isOpen, onClose, onSave, company }) => {
  const [name, setName] = useState('');
  const { validateFields, getError, hasError, hasErrors, clearAllErrors } = useFormValidation();

  useEffect(() => {
    if (company) {
      setName(company.name);
    } else {
      setName('');
    }
    clearAllErrors();
  }, [company, isOpen, clearAllErrors]);

  if (!isOpen) return null;

  const validateForm = (): boolean => {
    return validateFields({
      name: () => combineValidations(
        validateRequired(name, 'Nome'),
        validateMaxLength(name, 100, 'Nome')
      )
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    if (!validateForm()) {
      return;
    }

    onSave(name.trim());
    onClose();
  };

  const handleClose = () => {
    clearAllErrors();
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-0 md:p-4 bg-stone-900/50 backdrop-blur-sm">
      <div className="bg-white border-0 md:border border-stone-200 rounded-none md:rounded-2xl w-full h-full md:h-auto md:max-w-md shadow-2xl transform transition-all overflow-y-auto">
        <div className="flex justify-between items-center p-4 md:p-6 border-b border-stone-200 bg-stone-50 sticky top-0 z-10">
          <h3 className="text-lg md:text-xl font-bold text-stone-900">
            {company ? 'Editar Ambiente' : 'Novo Ambiente'}
          </h3>
          <button onClick={handleClose} className="text-stone-400 active:text-stone-700 md:hover:text-stone-700 transition-colors p-2 -mr-2 rounded-lg active:bg-stone-100">
            <X className="w-5 h-5" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-4 md:p-6 space-y-4">
          <div>
            <label className="block text-xs md:text-sm font-medium text-stone-600 mb-1.5">
              Nome do Ambiente <span className="text-red-500">*</span>
            </label>
            <input
              type="text"
              className={`w-full px-3 md:px-4 py-3 md:py-2.5 bg-stone-50 border ${
                hasError('name') ? 'border-red-500 focus:ring-red-400' : 'border-stone-200 focus:ring-stone-400'
              } rounded-lg md:rounded-xl text-sm md:text-base text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 transition-colors min-h-[44px]`}
              placeholder="Ex: Empresa Principal, Filial..."
              value={name}
              onChange={(e) => setName(e.target.value)}
              autoFocus
            />
            <ErrorMessage error={getError('name')} />
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
            {company ? 'Salvar Alterações' : 'Criar Ambiente'}
          </button>
        </form>
      </div>
    </div>
  );
};
