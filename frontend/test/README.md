# Test Suite Documentation

## Visao Geral

Esta suíte de testes foi configurada para o frontend React usando **Vitest**, **React Testing Library** e **Jest DOM**. A estratégia de testes segue a pirâmide de testes, com foco em testes unitários robustos e cobertura significativa.

## Estrutura de Testes

### Configuracao

- **Framework**: Vitest (alternativa moderna ao Jest)
- **Ambiente**: jsdom (simula o DOM do navegador)
- **Testing Library**: @testing-library/react
- **Matchers**: @testing-library/jest-dom
- **User Interactions**: @testing-library/user-event

### Arquivos de Configuracao

1. **vite.config.ts**: Configuracao principal do Vitest integrada ao Vite
2. **vitest.config.ts**: Configuracao dedicada do Vitest (opcional)
3. **test/setup.ts**: Arquivo de setup executado antes de todos os testes

## Scripts Disponíveis

```bash
# Executar todos os testes em modo watch
npm test

# Executar testes com cobertura
npm run test:coverage

# Executar testes com UI interativa
npm run test:ui
```

## Categorias de Testes

### 1. Testes de Utils (validation.test.ts)

**Objetivo**: Testar funcoes utilitarias de validacao de formularios

**Cobertura**:
- validateRequired: 8 casos de teste
- validateMinLength: 7 casos de teste
- validateMaxLength: 6 casos de teste
- validatePositiveNumber: 10 casos de teste
- validateRange: 10 casos de teste
- combineValidations: 8 casos de teste

**Total**: 49 casos de teste para validacoes

**Casos Críticos Testados**:
- Valores válidos e inválidos
- Edge cases (null, undefined, strings vazias)
- Valores nos limites (boundaries)
- Tipos diferentes (string, number)
- Combinacao de validacoes

### 2. Testes de Hooks (useFormValidation.test.ts)

**Objetivo**: Testar hook customizado de validacao de formularios

**Cobertura**:
- Estado inicial
- validateField (validacao de campo único)
- validateFields (validacao de múltiplos campos)
- clearError (limpar erro específico)
- clearAllErrors (limpar todos os erros)
- getError (obter mensagem de erro)
- hasErrors (verificar se há erros)
- hasError (verificar erro de campo específico)
- Cenários complexos (workflow completo)

**Total**: 25+ casos de teste

**Padroes Testados**:
- Isolamento de estado entre testes
- Imutabilidade de estado
- Funcoes callback estáveis
- Workflows realistas de formulários

### 3. Testes de Context (AuthContext.test.tsx)

**Objetivo**: Testar contexto de autenticacao e fluxos de login/logout

**Cobertura**:
- Inicializacao do provider
- Login (sucesso, falha, persistência)
- Register (sucesso, falha, auto-autenticacao)
- Logout (limpeza de estado e localStorage)
- isAuthenticated (computed property)
- Restauracao de sessao (localStorage)
- Uso do hook fora do provider
- Memoizacao de contexto
- Fluxo completo de autenticacao

**Total**: 30+ casos de teste

**Mocks Utilizados**:
- API service (vi.mock)
- localStorage (implementacao customizada no setup)

## Metas de Cobertura

```
Lines:      80%
Functions:  80%
Branches:   80%
Statements: 80%
```

## Boas Práticas Implementadas

### 1. Arrange-Act-Assert (AAA)

Todos os testes seguem o padrão AAA:
```typescript
it('should validate required field', () => {
  // Arrange
  const value = '';
  const fieldName = 'Email';

  // Act
  const result = validateRequired(value, fieldName);

  // Assert
  expect(result.isValid).toBe(false);
  expect(result.error).toBe('Email é obrigatório');
});
```

### 2. Cleanup Automático

```typescript
afterEach(() => {
  cleanup(); // Limpa componentes montados
  localStorage.clear(); // Limpa localStorage
});
```

### 3. Mocks Isolados

```typescript
beforeEach(() => {
  vi.clearAllMocks(); // Limpa histórico de mocks
});
```

### 4. Testes Descritivos

- Nomes claros e específicos
- Estrutura Given-When-Then implícita
- Agrupamento lógico com describe()

### 5. Testing Library Best Practices

```typescript
// Usar renderHook para hooks
const { result } = renderHook(() => useFormValidation());

// Usar act() para atualizacoes de estado
act(() => {
  result.current.validateField('email', validationFn);
});

// Usar waitFor() para operacoes assíncronas
await waitFor(() => {
  expect(result.current.loading).toBe(false);
});
```

## Estratégia de Testes por Camada

### Utils (Funcoes Puras)
- **Foco**: Entradas e saídas
- **Tipo**: Testes unitários simples
- **Mock**: Nenhum necessário

### Hooks (Lógica de Estado)
- **Foco**: Estado e side effects
- **Tipo**: Testes de integração leve
- **Mock**: Dependencies externas

### Contexts (Estado Global)
- **Foco**: Provider, consumidores, persistência
- **Tipo**: Testes de integração
- **Mock**: API, localStorage, external services

### Components (UI)
- **Foco**: Renderização, interações, acessibilidade
- **Tipo**: Testes de componente
- **Mock**: Contexts, hooks, APIs

## Patterns de Teste

### 1. Test Data Builders

```typescript
const mockUser: User = {
  id: '1',
  name: 'Test User',
  email: 'test@example.com',
};
```

### 2. Custom Render com Providers

```typescript
const { result } = renderHook(() => useAuth(), {
  wrapper: AuthProvider,
});
```

### 3. Async Testing

```typescript
await act(async () => {
  await result.current.login(credentials);
});
```

### 4. Error Boundary Testing

```typescript
const originalError = console.error;
console.error = vi.fn();
// ... test code
console.error = originalError;
```

## Debugging Tips

### 1. Usar screen.debug()

```typescript
import { screen } from '@testing-library/react';
screen.debug(); // Mostra o DOM atual
```

### 2. Usar logRoles()

```typescript
import { logRoles } from '@testing-library/react';
logRoles(container); // Mostra roles ARIA
```

### 3. Vitest UI

```bash
npm run test:ui
```

Abre interface visual para debugging de testes.

### 4. Coverage Report

```bash
npm run test:coverage
```

Gera relatório HTML em `coverage/index.html`

## Próximos Passos

### Testes a Adicionar

1. **Component Tests**
   - LoginForm.test.tsx
   - RegisterForm.test.tsx
   - TransactionList.test.tsx
   - Dashboard.test.tsx

2. **Integration Tests**
   - Complete user flows
   - Form submission flows
   - Navigation flows

3. **E2E Tests** (com Playwright/Cypress)
   - Critical user paths
   - Payment flows
   - Multi-page workflows

### Melhorias Futuras

1. **Visual Regression Testing**
   - Configurar Storybook
   - Adicionar chromatic ou percy

2. **Performance Testing**
   - React DevTools Profiler
   - Lighthouse CI

3. **Accessibility Testing**
   - jest-axe
   - Automated a11y checks

## Recursos

- [Vitest Documentation](https://vitest.dev/)
- [React Testing Library](https://testing-library.com/react)
- [Testing Library Best Practices](https://kentcdodds.com/blog/common-mistakes-with-react-testing-library)
- [Vitest UI](https://vitest.dev/guide/ui.html)

## Contribuindo

Ao adicionar novos testes:

1. Siga o padrão AAA
2. Use nomes descritivos
3. Teste comportamento, não implementação
4. Mantenha testes isolados e independentes
5. Use mocks apenas quando necessário
6. Documente casos de teste complexos
7. Mantenha cobertura acima de 80%

## Troubleshooting

### Erro: Cannot find module 'vitest/config'

```bash
npm install -D vitest
```

### Erro: jsdom not found

```bash
npm install -D jsdom
```

### Testes falhando com localStorage

Verifique se `test/setup.ts` está configurado corretamente com o mock de localStorage.

### Hooks falhando

Certifique-se de usar `renderHook` do `@testing-library/react` e não do `@testing-library/react-hooks` (deprecated).
