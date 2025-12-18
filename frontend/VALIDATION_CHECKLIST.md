# Checklist de Implementação - Sistema de Validação

Use este checklist ao adicionar validação em novos formulários.

## Antes de Começar

- [ ] Identificar todos os campos do formulário
- [ ] Definir regras de validação para cada campo
- [ ] Determinar quais campos são obrigatórios
- [ ] Verificar limites de caracteres ou valores

## Implementação

### 1. Importações

- [ ] Importar `useFormValidation` hook
- [ ] Importar funções de validação necessárias (`validateRequired`, `validateMaxLength`, etc.)
- [ ] Importar componente `ErrorMessage`

```typescript
import { useFormValidation } from '../hooks/useFormValidation';
import { validateRequired, validateMaxLength, validatePositiveNumber, combineValidations } from '../utils/validation';
import { ErrorMessage } from './ErrorMessage';
```

### 2. Setup do Hook

- [ ] Inicializar hook no componente
- [ ] Desestruturar métodos necessários

```typescript
const { validateFields, getError, hasError, hasErrors, clearAllErrors } = useFormValidation();
```

### 3. Função de Validação

- [ ] Criar função `validateForm()`
- [ ] Adicionar validações para cada campo
- [ ] Usar `combineValidations` quando necessário
- [ ] Retornar booleano indicando se form é válido

```typescript
const validateForm = (): boolean => {
  return validateFields({
    fieldName: () => validateRequired(value, 'Label'),
    // adicionar mais campos...
  });
};
```

### 4. Handler de Submit

- [ ] Prevenir comportamento padrão (`e.preventDefault()`)
- [ ] Chamar `validateForm()`
- [ ] Interromper execução se inválido
- [ ] Processar dados se válido

```typescript
const handleSubmit = (e: React.FormEvent) => {
  e.preventDefault();
  if (!validateForm()) return;
  // processar formulário
};
```

### 5. Markup dos Campos

Para cada campo do formulário:

#### Label
- [ ] Adicionar label descritivo
- [ ] Classes: `block text-sm font-medium text-stone-600 mb-1.5`
- [ ] Adicionar asterisco vermelho se obrigatório: `<span className="text-red-500">*</span>`

#### Input
- [ ] Adicionar classes base
- [ ] Adicionar classes condicionais baseadas em `hasError()`
- [ ] Borda vermelha se erro: `border-red-500`
- [ ] Ring vermelho se erro: `focus:ring-red-400`
- [ ] Adicionar transition: `transition-colors`

```typescript
className={`base-classes ${
  hasError('fieldName')
    ? 'border-red-500 focus:ring-red-400'
    : 'border-stone-200 focus:ring-stone-400'
}`}
```

#### Mensagem de Erro
- [ ] Adicionar componente `<ErrorMessage />` abaixo do input
- [ ] Passar erro obtido com `getError()`

```typescript
<ErrorMessage error={getError('fieldName')} />
```

#### Texto de Ajuda (opcional)
- [ ] Adicionar texto explicativo
- [ ] Classes: `text-xs text-stone-400 mt-1`

### 6. Botão de Submit

- [ ] Adicionar atributo `type="submit"`
- [ ] Adicionar `disabled={hasErrors()}`
- [ ] Classes condicionais baseadas em `hasErrors()`
- [ ] Estado desabilitado: `bg-stone-300 text-stone-500 cursor-not-allowed`
- [ ] Estado ativo: `bg-stone-800 hover:bg-stone-700 text-white`

```typescript
<button
  type="submit"
  disabled={hasErrors()}
  className={hasErrors() ? 'disabled-classes' : 'active-classes'}
>
  Salvar
</button>
```

### 7. Limpeza de Erros

- [ ] Limpar erros ao fechar modal/form
- [ ] Limpar erros ao resetar formulário
- [ ] Adicionar ao `useEffect` se necessário

```typescript
const handleClose = () => {
  clearAllErrors();
  onClose();
};
```

## Testes Manuais

### Para Cada Campo

- [ ] Tentar submeter com campo vazio (se obrigatório)
- [ ] Inserir valor inválido (negativo, muito longo, etc.)
- [ ] Verificar se mensagem de erro aparece
- [ ] Verificar se borda fica vermelha
- [ ] Verificar se botão é desabilitado

### Fluxo Completo

- [ ] Preencher todos os campos corretamente
- [ ] Verificar se botão está habilitado
- [ ] Submeter formulário
- [ ] Verificar se dados são processados
- [ ] Verificar se erros são limpos após submit

### Cenários Edge

- [ ] Colar texto muito longo
- [ ] Inserir caracteres especiais
- [ ] Valores numéricos extremos (0, negativo, muito grande)
- [ ] Strings vazias vs apenas espaços
- [ ] Abrir e fechar modal múltiplas vezes

## Checklist Visual

### Aparência dos Campos

- [ ] Campos normais: borda cinza
- [ ] Campos com erro: borda vermelha
- [ ] Focus ring: azul (normal) ou vermelho (erro)
- [ ] Transições suaves entre estados
- [ ] Mensagens de erro alinhadas e legíveis

### Aparência dos Botões

- [ ] Botão ativo: fundo escuro, hover mais claro
- [ ] Botão desabilitado: fundo cinza claro, sem hover
- [ ] Cursor appropriate (pointer vs not-allowed)
- [ ] Ícones alinhados corretamente

### Responsividade

- [ ] Testar em mobile (< 640px)
- [ ] Testar em tablet (640px - 1024px)
- [ ] Testar em desktop (> 1024px)
- [ ] Layouts de grid responsivos funcionando
- [ ] Textos não quebram inadequadamente

## Acessibilidade

- [ ] Todas as labels estão associadas aos inputs
- [ ] Campos obrigatórios têm indicador visual
- [ ] Mensagens de erro são descritivas
- [ ] Contraste adequado (WCAG AA)
- [ ] Focus visível em todos os campos
- [ ] Navegação por teclado funciona
- [ ] Screen readers podem ler erros

## Performance

- [ ] Validação apenas no submit (não em cada keystroke)
- [ ] Re-renders minimizados
- [ ] Sem memory leaks (cleanup em useEffect)
- [ ] Form responde rapidamente

## Código Limpo

- [ ] Código bem indentado e formatado
- [ ] Nomes de variáveis descritivos
- [ ] Comentários onde necessário
- [ ] Sem console.logs desnecessários
- [ ] Imports organizados
- [ ] TypeScript sem erros

## Documentação

- [ ] Adicionar comentário sobre validações especiais
- [ ] Atualizar README se necessário
- [ ] Documentar edge cases conhecidos

## Validações Comuns por Tipo de Campo

### Texto Curto (nome, título)
```typescript
- [ ] validateRequired
- [ ] validateMaxLength (50-100)
```

### Texto Longo (descrição, notas)
```typescript
- [ ] validateMaxLength (200-500)
- [ ] Opcional: validateRequired
```

### Número Positivo (preço, quantidade)
```typescript
- [ ] validateRequired
- [ ] validatePositiveNumber
```

### Número em Intervalo (dia do mês, idade)
```typescript
- [ ] validateRequired
- [ ] validateRange(min, max)
```

### Select/Dropdown
```typescript
- [ ] validateRequired
- [ ] Verificar se não está no valor placeholder
```

### Email (se implementar futuramente)
```typescript
- [ ] validateRequired
- [ ] validateEmail (a implementar)
```

## Checklist por Componente

### TransactionModal

- [x] Descrição validada (obrigatória, máx 200)
- [x] Valor validado (obrigatório, positivo)
- [x] Categoria validada (obrigatória)
- [x] Mês validado (obrigatório)
- [x] Ano validado (obrigatório)
- [x] Botão desabilitado quando inválido
- [x] Erros limpos ao fechar
- [x] Feedback visual implementado

### Recurring

- [x] Descrição validada (obrigatória, máx 200)
- [x] Valor validado (obrigatório, positivo)
- [x] Dia do mês validado (entre 1 e 31)
- [x] Categoria validada (obrigatória)
- [x] Botão desabilitado quando inválido
- [x] Erros limpos ao esconder form
- [x] Feedback visual implementado

### CompanyList

- [x] Nome validado (obrigatório, máx 100)
- [x] Botão desabilitado quando inválido
- [x] Feedback visual implementado

### CompanyModal

- [x] Nome validado (obrigatório, máx 100)
- [x] Botão desabilitado quando inválido
- [x] Erros limpos ao fechar
- [x] Feedback visual implementado

## Próximo Formulário

Use este template para seu próximo formulário:

```typescript
// ==========================================
// NOVO FORMULÁRIO: [Nome do Formulário]
// ==========================================

// CAMPOS:
// - [ ] Campo1: [tipo] - [validações]
// - [ ] Campo2: [tipo] - [validações]
// - [ ] Campo3: [tipo] - [validações]

// VALIDAÇÕES:
const validateForm = (): boolean => {
  return validateFields({
    // campo1: () => ...,
    // campo2: () => ...,
  });
};

// TESTES:
// - [ ] Submeter vazio
// - [ ] Valores inválidos
// - [ ] Valores válidos
// - [ ] Abrir/fechar modal
```

## Conclusão

Após completar todos os itens deste checklist:

- [ ] Fazer commit das mudanças
- [ ] Testar em diferentes navegadores
- [ ] Solicitar code review (se aplicável)
- [ ] Atualizar documentação do projeto
- [ ] Marcar task como concluída

---

**Lembre-se:** Validação é sobre criar uma boa experiência para o usuário. Mensagens claras, feedback visual imediato e formulários fáceis de usar são fundamentais!
