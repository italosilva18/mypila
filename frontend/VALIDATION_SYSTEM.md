# Sistema de Validação de Formulários

Sistema completo de validação client-side para formulários React/TypeScript, implementando boas práticas de UX e acessibilidade.

## Estrutura

### 1. Utilitários de Validação (`utils/validation.ts`)

Funções puras de validação que retornam `ValidationResult`:

```typescript
interface ValidationResult {
  isValid: boolean;
  error?: string;
}
```

#### Funções Disponíveis:

- **`validateRequired(value, fieldName)`**: Valida campos obrigatórios
- **`validateMaxLength(value, max, fieldName)`**: Valida tamanho máximo de string
- **`validatePositiveNumber(value, fieldName)`**: Valida número positivo
- **`validateRange(value, min, max, fieldName)`**: Valida número dentro de um intervalo
- **`combineValidations(...results)`**: Combina múltiplas validações

### 2. Hook `useFormValidation`

Hook customizado para gerenciar estado de validação de formulários.

```typescript
const {
  errors,              // Estado de todos os erros
  validateField,       // Valida um campo individual
  validateFields,      // Valida múltiplos campos
  clearError,         // Limpa erro de um campo
  clearAllErrors,     // Limpa todos os erros
  getError,           // Obtém mensagem de erro
  hasErrors,          // Verifica se há erros
  hasError            // Verifica se campo específico tem erro
} = useFormValidation();
```

### 3. Componente `ErrorMessage`

Componente reutilizável para exibir mensagens de erro com ícone e estilo consistente.

```typescript
<ErrorMessage error={getError('fieldName')} />
```

## Uso nos Componentes

### TransactionModal.tsx

Validações implementadas:
- Descrição: obrigatória, máximo 200 caracteres
- Valor: obrigatório, número positivo
- Categoria: obrigatória
- Mês: obrigatório
- Ano: obrigatório

```typescript
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
```

### Recurring.tsx

Validações implementadas:
- Descrição: obrigatória, máximo 200 caracteres
- Valor: obrigatório, número positivo
- Categoria: obrigatória
- Dia do mês: obrigatório, entre 1 e 31

### CompanyList.tsx

Validações implementadas:
- Nome: obrigatório, máximo 100 caracteres

## Padrão de Implementação

### 1. Importar dependências

```typescript
import { useFormValidation } from '../hooks/useFormValidation';
import { validateRequired, validateMaxLength, validatePositiveNumber, combineValidations } from '../utils/validation';
import { ErrorMessage } from './ErrorMessage';
```

### 2. Inicializar hook

```typescript
const { validateFields, getError, hasError, hasErrors, clearAllErrors } = useFormValidation();
```

### 3. Criar função de validação

```typescript
const validateForm = (): boolean => {
  return validateFields({
    fieldName: () => validateRequired(value, 'Label do Campo'),
    anotherField: () => combineValidations(
      validateRequired(value, 'Label'),
      validateMaxLength(value, 100, 'Label')
    )
  });
};
```

### 4. Chamar validação no submit

```typescript
const handleSubmit = (e: React.FormEvent) => {
  e.preventDefault();

  if (!validateForm()) {
    return; // Não prossegue se houver erros
  }

  // Continua com a lógica de submit
};
```

### 5. Aplicar estilo condicional nos inputs

```typescript
<input
  className={`base-classes ${
    hasError('fieldName')
      ? 'border-red-500 focus:ring-red-400'
      : 'border-stone-200 focus:ring-stone-400'
  }`}
/>
<ErrorMessage error={getError('fieldName')} />
```

### 6. Desabilitar botão de submit

```typescript
<button
  type="submit"
  disabled={hasErrors()}
  className={hasErrors()
    ? 'bg-stone-300 text-stone-500 cursor-not-allowed'
    : 'bg-stone-800 hover:bg-stone-700 text-white'
  }
>
  Salvar
</button>
```

### 7. Limpar erros ao fechar modal/form

```typescript
const handleClose = () => {
  clearAllErrors();
  onClose();
};
```

## Estilo Visual

### Campos com Erro
- Borda vermelha: `border-red-500`
- Ring vermelho no focus: `focus:ring-red-400`
- Transição suave: `transition-colors`

### Mensagens de Erro
- Cor: `text-red-500`
- Tamanho: `text-xs`
- Ícone: AlertCircle do lucide-react
- Margem superior: `mt-1`

### Botão Desabilitado
- Cor de fundo: `bg-stone-300`
- Cor do texto: `text-stone-500`
- Cursor: `cursor-not-allowed`

## Exemplo Completo

```typescript
import React, { useState } from 'react';
import { useFormValidation } from '../hooks/useFormValidation';
import { validateRequired, validateMaxLength, combineValidations } from '../utils/validation';
import { ErrorMessage } from './ErrorMessage';

export const MyForm: React.FC = () => {
  const [formData, setFormData] = useState({ name: '', email: '' });
  const { validateFields, getError, hasError, hasErrors, clearAllErrors } = useFormValidation();

  const validateForm = (): boolean => {
    return validateFields({
      name: () => combineValidations(
        validateRequired(formData.name, 'Nome'),
        validateMaxLength(formData.name, 100, 'Nome')
      ),
      email: () => validateRequired(formData.email, 'E-mail')
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;

    // Processar formulário
    console.log('Formulário válido:', formData);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-stone-600 mb-1.5">
          Nome <span className="text-red-500">*</span>
        </label>
        <input
          type="text"
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          className={`w-full px-4 py-2.5 bg-stone-50 border ${
            hasError('name') ? 'border-red-500 focus:ring-red-400' : 'border-stone-200 focus:ring-stone-400'
          } rounded-xl focus:outline-none focus:ring-2 transition-colors`}
        />
        <ErrorMessage error={getError('name')} />
      </div>

      <button
        type="submit"
        disabled={hasErrors()}
        className={`w-full py-3 rounded-xl font-medium transition-all ${
          hasErrors()
            ? 'bg-stone-300 text-stone-500 cursor-not-allowed'
            : 'bg-stone-800 hover:bg-stone-700 text-white'
        }`}
      >
        Salvar
      </button>
    </form>
  );
};
```

## Boas Práticas

1. **Sempre validar no submit**: Não confie apenas em atributos HTML como `required`
2. **Feedback visual imediato**: Inputs devem mudar de cor quando inválidos
3. **Mensagens claras**: Use mensagens descritivas sobre o que está errado
4. **Desabilitar submit**: Botão de submit deve estar desabilitado quando há erros
5. **Limpar erros**: Limpe os erros ao fechar modais ou resetar forms
6. **Asterisco para obrigatórios**: Use `<span className="text-red-500">*</span>` em labels
7. **Combinar validações**: Use `combineValidations` para múltiplas regras no mesmo campo

## Acessibilidade

- Labels descritivos para todos os campos
- Mensagens de erro associadas aos inputs
- Feedback visual claro (cor, borda)
- Estados de disabled claramente indicados
- Transições suaves entre estados

## Performance

- Validações são executadas apenas no submit (não on-change)
- Estados de erro são gerenciados de forma eficiente
- Re-renders minimizados com callbacks memoizados
