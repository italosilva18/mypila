/**
 * EXEMPLO DE USO DO SISTEMA DE VALIDAÇÃO
 *
 * Este arquivo demonstra como implementar validação em um formulário customizado
 */

import React, { useState } from 'react';
import { useFormValidation } from '../hooks/useFormValidation';
import {
  validateRequired,
  validateMaxLength,
  validatePositiveNumber,
  validateRange,
  combineValidations
} from '../utils/validation';
import { ErrorMessage } from '../components/ErrorMessage';
import { Save } from 'lucide-react';

interface FormData {
  name: string;
  description: string;
  price: number;
  quantity: number;
  category: string;
}

export const FormValidationExample: React.FC = () => {
  const [formData, setFormData] = useState<FormData>({
    name: '',
    description: '',
    price: 0,
    quantity: 1,
    category: ''
  });

  const [submitted, setSubmitted] = useState(false);

  const { validateFields, getError, hasError, hasErrors, clearAllErrors } = useFormValidation();

  // Define as regras de validação para cada campo
  const validateForm = (): boolean => {
    return validateFields({
      // Nome: obrigatório e máximo 50 caracteres
      name: () => combineValidations(
        validateRequired(formData.name, 'Nome'),
        validateMaxLength(formData.name, 50, 'Nome')
      ),

      // Descrição: opcional, mas se preenchida, máximo 200 caracteres
      description: () => validateMaxLength(formData.description, 200, 'Descrição'),

      // Preço: obrigatório e deve ser positivo
      price: () => validatePositiveNumber(formData.price, 'Preço'),

      // Quantidade: obrigatória, entre 1 e 100
      quantity: () => validateRange(formData.quantity, 1, 100, 'Quantidade'),

      // Categoria: obrigatória
      category: () => validateRequired(formData.category, 'Categoria')
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();

    // Valida o formulário antes de submeter
    if (!validateForm()) {
      console.log('Formulário inválido - corriga os erros antes de continuar');
      return;
    }

    // Se chegou aqui, o formulário está válido
    console.log('Formulário válido! Dados:', formData);
    setSubmitted(true);

    // Simula envio para API
    setTimeout(() => {
      alert('Dados enviados com sucesso!');
      setSubmitted(false);
      // Reset form
      setFormData({ name: '', description: '', price: 0, quantity: 1, category: '' });
      clearAllErrors();
    }, 1000);
  };

  const handleReset = () => {
    setFormData({ name: '', description: '', price: 0, quantity: 1, category: '' });
    clearAllErrors();
  };

  return (
    <div className="max-w-2xl mx-auto p-6">
      <h1 className="text-3xl font-bold text-stone-900 mb-2">Exemplo de Validação</h1>
      <p className="text-stone-500 mb-6">
        Demonstração do sistema de validação de formulários
      </p>

      <form onSubmit={handleSubmit} className="bg-white border border-stone-200 rounded-2xl p-6 space-y-6">
        {/* Campo: Nome */}
        <div>
          <label className="block text-sm font-medium text-stone-600 mb-1.5">
            Nome do Produto <span className="text-red-500">*</span>
          </label>
          <input
            type="text"
            value={formData.name}
            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
            placeholder="Ex: Notebook Dell"
            className={`w-full px-4 py-2.5 bg-stone-50 border ${
              hasError('name')
                ? 'border-red-500 focus:ring-red-400'
                : 'border-stone-200 focus:ring-stone-400'
            } rounded-xl text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 transition-colors`}
          />
          <ErrorMessage error={getError('name')} />
          <p className="text-xs text-stone-400 mt-1">Máximo 50 caracteres</p>
        </div>

        {/* Campo: Descrição */}
        <div>
          <label className="block text-sm font-medium text-stone-600 mb-1.5">
            Descrição (opcional)
          </label>
          <textarea
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            placeholder="Descreva o produto..."
            rows={3}
            className={`w-full px-4 py-2.5 bg-stone-50 border ${
              hasError('description')
                ? 'border-red-500 focus:ring-red-400'
                : 'border-stone-200 focus:ring-stone-400'
            } rounded-xl text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 transition-colors resize-none`}
          />
          <ErrorMessage error={getError('description')} />
          <p className="text-xs text-stone-400 mt-1">
            {formData.description.length}/200 caracteres
          </p>
        </div>

        {/* Campos: Preço e Quantidade */}
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="block text-sm font-medium text-stone-600 mb-1.5">
              Preço <span className="text-red-500">*</span>
            </label>
            <div className="relative">
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-stone-400">R$</span>
              <input
                type="number"
                step="0.01"
                value={formData.price || ''}
                onChange={(e) => setFormData({ ...formData, price: parseFloat(e.target.value) || 0 })}
                placeholder="0.00"
                className={`w-full pl-10 pr-4 py-2.5 bg-stone-50 border ${
                  hasError('price')
                    ? 'border-red-500 focus:ring-red-400'
                    : 'border-stone-200 focus:ring-stone-400'
                } rounded-xl text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 transition-colors`}
              />
            </div>
            <ErrorMessage error={getError('price')} />
            <p className="text-xs text-stone-400 mt-1">Deve ser maior que zero</p>
          </div>

          <div>
            <label className="block text-sm font-medium text-stone-600 mb-1.5">
              Quantidade <span className="text-red-500">*</span>
            </label>
            <input
              type="number"
              value={formData.quantity}
              onChange={(e) => setFormData({ ...formData, quantity: parseInt(e.target.value) || 0 })}
              placeholder="1"
              className={`w-full px-4 py-2.5 bg-stone-50 border ${
                hasError('quantity')
                  ? 'border-red-500 focus:ring-red-400'
                  : 'border-stone-200 focus:ring-stone-400'
              } rounded-xl text-stone-900 placeholder-stone-400 focus:outline-none focus:ring-2 transition-colors`}
            />
            <ErrorMessage error={getError('quantity')} />
            <p className="text-xs text-stone-400 mt-1">Entre 1 e 100</p>
          </div>
        </div>

        {/* Campo: Categoria */}
        <div>
          <label className="block text-sm font-medium text-stone-600 mb-1.5">
            Categoria <span className="text-red-500">*</span>
          </label>
          <select
            value={formData.category}
            onChange={(e) => setFormData({ ...formData, category: e.target.value })}
            className={`w-full px-4 py-2.5 bg-stone-50 border ${
              hasError('category')
                ? 'border-red-500 focus:ring-red-400'
                : 'border-stone-200 focus:ring-stone-400'
            } rounded-xl text-stone-900 focus:outline-none focus:ring-2 transition-colors`}
          >
            <option value="">Selecione uma categoria...</option>
            <option value="eletronicos">Eletrônicos</option>
            <option value="moveis">Móveis</option>
            <option value="acessorios">Acessórios</option>
            <option value="outros">Outros</option>
          </select>
          <ErrorMessage error={getError('category')} />
        </div>

        {/* Botões */}
        <div className="flex gap-3 pt-4">
          <button
            type="submit"
            disabled={hasErrors() || submitted}
            className={`flex-1 py-3 rounded-xl font-medium transition-all shadow-lg flex items-center justify-center gap-2 ${
              hasErrors() || submitted
                ? 'bg-stone-300 text-stone-500 cursor-not-allowed'
                : 'bg-stone-800 hover:bg-stone-700 text-white shadow-stone-900/20 active:scale-[0.98]'
            }`}
          >
            <Save className="w-4 h-4" />
            {submitted ? 'Enviando...' : 'Salvar Produto'}
          </button>

          <button
            type="button"
            onClick={handleReset}
            className="px-6 py-3 bg-stone-100 hover:bg-stone-200 text-stone-700 font-medium rounded-xl transition-all"
          >
            Limpar
          </button>
        </div>

        {/* Indicador de estado do formulário */}
        <div className="pt-4 border-t border-stone-100">
          <div className="flex items-center justify-between text-xs">
            <span className="text-stone-500">
              Estado do formulário:
            </span>
            <span className={`font-medium ${hasErrors() ? 'text-red-500' : 'text-green-600'}`}>
              {hasErrors() ? 'Contém erros' : 'Válido'}
            </span>
          </div>
        </div>
      </form>

      {/* Seção de ajuda */}
      <div className="mt-6 bg-blue-50 border border-blue-200 rounded-xl p-4">
        <h3 className="text-sm font-bold text-blue-900 mb-2">Como testar</h3>
        <ul className="text-xs text-blue-700 space-y-1">
          <li>• Tente submeter o formulário vazio</li>
          <li>• Digite um nome com mais de 50 caracteres</li>
          <li>• Digite uma descrição com mais de 200 caracteres</li>
          <li>• Insira um preço zero ou negativo</li>
          <li>• Insira uma quantidade fora do intervalo 1-100</li>
          <li>• Não selecione uma categoria</li>
        </ul>
      </div>
    </div>
  );
};
