# Sistema de Validação de Formulários - Frontend

Sistema completo de validação client-side implementado em React/TypeScript para o projeto FinanceiroPro.

## Visão Geral

Este sistema fornece validação consistente, reutilizável e acessível para todos os formulários da aplicação, com feedback visual imediato e mensagens de erro claras.

## Estrutura de Arquivos

```
frontend/
├── utils/
│   ├── validation.ts                    # ✅ Funções utilitárias de validação
│   └── validation.test.example.ts       # 📘 Exemplos de testes
├── hooks/
│   └── useFormValidation.ts             # ✅ Hook customizado
├── components/
│   ├── ErrorMessage.tsx                 # ✅ Componente de mensagem de erro
│   ├── TransactionModal.tsx             # ✅ Validação implementada
│   ├── CompanyList.tsx                  # ✅ Validação implementada
│   └── CompanyModal.tsx                 # ✅ Validação implementada
├── pages/
│   └── Recurring.tsx                    # ✅ Validação implementada
├── examples/
│   └── FormValidationExample.tsx        # 📘 Exemplo completo
├── VALIDATION_SYSTEM.md                 # 📖 Documentação detalhada
├── VALIDATION_IMPLEMENTATION.md         # 📋 Resumo da implementação
├── VALIDATION_STYLE_GUIDE.md            # 🎨 Guia de estilo visual
└── README_VALIDATION.md                 # 📌 Este arquivo
```

## Instalação

Todos os arquivos necessários já foram criados. Não é necessária instalação adicional.

## Quick Start

### 1. Importar dependências

```typescript
import { useFormValidation } from '../hooks/useFormValidation';
import { validateRequired, validateMaxLength, validatePositiveNumber } from '../utils/validation';
import { ErrorMessage } from './ErrorMessage';
```

### 2. Usar no componente

```typescript
const MyForm = () => {
  const [formData, setFormData] = useState({ name: '', amount: 0 });
  const { validateFields, getError, hasError, hasErrors, clearAllErrors } = useFormValidation();

  const validateForm = () => {
    return validateFields({
      name: () => validateRequired(formData.name, 'Nome'),
      amount: () => validatePositiveNumber(formData.amount, 'Valor')
    });
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!validateForm()) return;
    // Processar formulário
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        value={formData.name}
        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
        className={hasError('name') ? 'border-red-500' : 'border-stone-200'}
      />
      <ErrorMessage error={getError('name')} />

      <button type="submit" disabled={hasErrors()}>
        Salvar
      </button>
    </form>
  );
};
```

## Funções de Validação

### validateRequired(value, fieldName)
Valida se o campo foi preenchido.

```typescript
validateRequired(formData.name, 'Nome')
// Erro se: '', null, undefined, ou apenas espaços
```

### validateMaxLength(value, max, fieldName)
Valida tamanho máximo de string.

```typescript
validateMaxLength(formData.description, 200, 'Descrição')
// Erro se: string.length > 200
```

### validatePositiveNumber(value, fieldName)
Valida se é número positivo.

```typescript
validatePositiveNumber(formData.amount, 'Valor')
// Erro se: <= 0 ou não é número
```

### validateRange(value, min, max, fieldName)
Valida se número está em intervalo.

```typescript
validateRange(formData.dayOfMonth, 1, 31, 'Dia do mês')
// Erro se: < 1 ou > 31
```

### combineValidations(...validations)
Combina múltiplas validações.

```typescript
combineValidations(
  validateRequired(formData.name, 'Nome'),
  validateMaxLength(formData.name, 100, 'Nome')
)
// Retorna o primeiro erro encontrado
```

## Hook useFormValidation

### Métodos Disponíveis

| Método | Descrição |
|--------|-----------|
| `validateField(name, fn)` | Valida um campo específico |
| `validateFields(validations)` | Valida múltiplos campos |
| `getError(name)` | Obtém mensagem de erro |
| `hasError(name)` | Verifica se campo tem erro |
| `hasErrors()` | Verifica se há erros no form |
| `clearError(name)` | Limpa erro de um campo |
| `clearAllErrors()` | Limpa todos os erros |

## Componentes Atualizados

### 1. TransactionModal.tsx

**Validações:**
- Descrição: obrigatória, máx 200 chars
- Valor: obrigatório, positivo
- Categoria: obrigatória
- Mês: obrigatório
- Ano: obrigatório

**Localização:** `D:\Sexto\frontend\components\TransactionModal.tsx`

### 2. Recurring.tsx

**Validações:**
- Descrição: obrigatória, máx 200 chars
- Valor: obrigatório, positivo
- Dia do mês: obrigatório, entre 1 e 31
- Categoria: obrigatória

**Localização:** `D:\Sexto\frontend\pages\Recurring.tsx`

### 3. CompanyList.tsx

**Validações:**
- Nome: obrigatório, máx 100 chars

**Localização:** `D:\Sexto\frontend\components\CompanyList.tsx`

### 4. CompanyModal.tsx

**Validações:**
- Nome: obrigatório, máx 100 chars

**Localização:** `D:\Sexto\frontend\components\CompanyModal.tsx`

## Estilo Visual

### Estados dos Campos

**Normal:**
```jsx
className="border border-stone-200 focus:ring-stone-400"
```

**Com Erro:**
```jsx
className="border border-red-500 focus:ring-red-400"
```

### Mensagens de Erro

```jsx
<ErrorMessage error={getError('fieldName')} />
```

- Cor: `text-red-500`
- Tamanho: `text-xs`
- Ícone: AlertCircle

### Botão Submit

**Ativo:**
```jsx
className="bg-stone-800 hover:bg-stone-700 text-white"
```

**Desabilitado:**
```jsx
disabled={hasErrors()}
className="bg-stone-300 text-stone-500 cursor-not-allowed"
```

## Padrão de Implementação

### Passo a Passo

1. **Importar** hook e validações
2. **Inicializar** hook no componente
3. **Criar** função de validação
4. **Chamar** validação no submit
5. **Aplicar** estilos condicionais
6. **Mostrar** mensagens de erro
7. **Desabilitar** botão quando inválido
8. **Limpar** erros ao fechar

### Template Completo

```typescript
import React, { useState } from 'react';
import { useFormValidation } from '../hooks/useFormValidation';
import { validateRequired, validateMaxLength, combineValidations } from '../utils/validation';
import { ErrorMessage } from './ErrorMessage';

export const MyForm = () => {
  const [formData, setFormData] = useState({ name: '', description: '' });
  const { validateFields, getError, hasError, hasErrors, clearAllErrors } = useFormValidation();

  const validateForm = () => {
    return validateFields({
      name: () => combineValidations(
        validateRequired(formData.name, 'Nome'),
        validateMaxLength(formData.name, 100, 'Nome')
      ),
      description: () => validateMaxLength(formData.description, 200, 'Descrição')
    });
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!validateForm()) return;
    console.log('Form válido:', formData);
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
          } rounded-xl text-stone-900 focus:outline-none focus:ring-2 transition-colors`}
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

## Testando o Sistema

### Cenários de Teste

#### TransactionModal
1. ❌ Submeter sem descrição → Erro: "Descrição é obrigatório"
2. ❌ Descrição com 201+ caracteres → Erro: "Descrição deve ter no máximo 200 caracteres"
3. ❌ Valor zero ou negativo → Erro: "Valor deve ser maior que zero"
4. ❌ Sem categoria selecionada → Erro: "Categoria é obrigatório"
5. ✅ Todos os campos válidos → Submit habilitado

#### Recurring
1. ❌ Submeter sem descrição → Erro
2. ❌ Valor negativo → Erro
3. ❌ Dia < 1 ou > 31 → Erro: "Dia do mês deve estar entre 1 e 31"
4. ✅ Todos os campos válidos → Submit habilitado

#### CompanyList/Modal
1. ❌ Nome vazio → Erro: "Nome é obrigatório"
2. ❌ Nome com 101+ caracteres → Erro: "Nome deve ter no máximo 100 caracteres"
3. ✅ Nome válido → Submit habilitado

## Documentação Adicional

### Arquivos de Referência

1. **VALIDATION_SYSTEM.md** - Documentação técnica completa do sistema
2. **VALIDATION_IMPLEMENTATION.md** - Resumo da implementação e tabelas de validação
3. **VALIDATION_STYLE_GUIDE.md** - Guia visual de estilos e padrões CSS
4. **examples/FormValidationExample.tsx** - Exemplo funcional completo
5. **utils/validation.test.example.ts** - Exemplos de testes unitários

### Links Úteis

- [React Hook Form](https://react-hook-form.com/) - Alternativa mais robusta (futuro)
- [Yup](https://github.com/jquense/yup) - Schema validation (futuro)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [WCAG Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)

## Troubleshooting

### Problema: Botão sempre desabilitado

**Causa:** `hasErrors()` retorna true mesmo sem erros
**Solução:** Certifique-se de chamar `clearAllErrors()` ao abrir o form

```typescript
useEffect(() => {
  if (isOpen) {
    clearAllErrors();
  }
}, [isOpen, clearAllErrors]);
```

### Problema: Mensagens de erro não aparecem

**Causa:** Validação não está sendo chamada
**Solução:** Certifique-se de chamar `validateForm()` no submit

```typescript
const handleSubmit = (e) => {
  e.preventDefault();
  if (!validateForm()) return; // ← Importante!
  // resto do código
};
```

### Problema: Erros persistem após fechar modal

**Causa:** Erros não foram limpos
**Solução:** Limpe erros ao fechar

```typescript
const handleClose = () => {
  clearAllErrors();
  onClose();
};
```

## Performance

- ✅ Validações executadas apenas no submit (não onChange)
- ✅ Callbacks memoizados no hook
- ✅ Re-renders minimizados
- ✅ Estado de erro gerenciado eficientemente

## Acessibilidade

- ✅ Labels descritivos
- ✅ Asteriscos para campos obrigatórios
- ✅ Mensagens de erro claras
- ✅ Feedback visual (cor, borda)
- ✅ Estados disabled visíveis
- ✅ Focus ring em todos os campos

## Próximos Passos (Opcional)

### Melhorias Futuras

1. **Validação em tempo real** - Validar onChange para feedback instantâneo
2. **Validação assíncrona** - Verificar duplicatas no backend
3. **Mais validadores** - Email, CPF, CNPJ, telefone, etc.
4. **i18n** - Internacionalizar mensagens de erro
5. **Testes unitários** - Implementar testes com Jest
6. **Documentação de API** - Gerar docs automáticas com TypeDoc

## Suporte

Para dúvidas ou problemas:
1. Consulte a documentação em `VALIDATION_SYSTEM.md`
2. Veja exemplos em `examples/FormValidationExample.tsx`
3. Revise o guia de estilo em `VALIDATION_STYLE_GUIDE.md`

## Licença

Este sistema de validação faz parte do projeto FinanceiroPro.

---

**Versão:** 1.0.0
**Última atualização:** 2024-12-15
**Autor:** Frontend Artisan
