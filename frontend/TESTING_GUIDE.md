# Guia de Testes - Frontend React

## Instalacao

Para instalar todas as dependencias de teste:

```bash
cd D:\Sexto\frontend
npm install
```

Isso instalara:
- vitest@2.1.8
- @testing-library/react@16.1.0
- @testing-library/jest-dom@6.6.3
- @testing-library/user-event@14.5.2
- jsdom@25.0.1
- @vitest/ui@2.1.8
- @vitest/coverage-v8@2.1.8

## Execucao dos Testes

### Modo Interativo (Watch Mode)
```bash
npm test
```
Os testes serao executados automaticamente quando voce salvar arquivos.

### Interface Visual
```bash
npm run test:ui
```
Abre interface web interativa em http://localhost:51204/

### Cobertura de Codigo
```bash
npm run test:coverage
```
Gera relatorio de cobertura em `coverage/index.html`

## Estrutura de Testes Criados

```
frontend/
├── test/
│   ├── setup.ts                          # Configuracao global dos testes
│   ├── README.md                         # Documentacao detalhada
│   └── examples/
│       └── ComponentTest.example.tsx     # Template de teste de componente
├── utils/
│   └── validation.test.ts                # 49 testes de validacao
├── hooks/
│   └── useFormValidation.test.ts         # 25+ testes do hook
├── contexts/
│   └── AuthContext.test.tsx              # 30+ testes de autenticacao
├── vite.config.ts                        # Config Vitest (integrada)
└── vitest.config.ts                      # Config Vitest (dedicada)
```

## Resumo dos Testes

### 1. Validation Utils (49 testes)
**Arquivo**: `utils/validation.test.ts`

Testa todas as funcoes de validacao:
- `validateRequired`: 8 casos
- `validateMinLength`: 7 casos
- `validateMaxLength`: 6 casos
- `validatePositiveNumber`: 10 casos
- `validateRange`: 10 casos
- `combineValidations`: 8 casos

**Cobertura**: ~100% das funcoes de validacao

### 2. Form Validation Hook (25+ testes)
**Arquivo**: `hooks/useFormValidation.test.ts`

Testa o hook customizado de validacao:
- Estado inicial
- Validacao de campo unico
- Validacao de multiplos campos
- Limpeza de erros
- Workflows completos de formulario

**Cobertura**: ~100% do hook

### 3. Authentication Context (30+ testes)
**Arquivo**: `contexts/AuthContext.test.tsx`

Testa todo o fluxo de autenticacao:
- Login (sucesso, falha, persistencia)
- Register (sucesso, falha)
- Logout
- Restauracao de sessao
- Memoizacao de contexto
- Fluxo completo: register → logout → login

**Cobertura**: ~95% do AuthContext

## Estatisticas Esperadas

Ao executar `npm run test:coverage`, voce devera ver:

```
Test Files  3 passed (3)
     Tests  104+ passed (104+)
  Duration  ~5s

 % Stmts | % Branch | % Funcs | % Lines
---------|----------|---------|----------
   95%+  |   90%+   |   95%+  |   95%+
```

## Configuracao de Cobertura

O projeto esta configurado para exigir:
- **Lines**: 80%
- **Functions**: 80%
- **Branches**: 80%
- **Statements**: 80%

Arquivos excluidos da cobertura:
- node_modules/
- test/
- *.config.*
- *.d.ts
- mockData/
- dist/

## Mocks Globais (test/setup.ts)

O arquivo de setup configura automaticamente:

1. **@testing-library/jest-dom**: Matchers customizados
2. **cleanup**: Limpeza automatica apos cada teste
3. **matchMedia**: Mock do window.matchMedia
4. **IntersectionObserver**: Mock global
5. **localStorage**: Implementacao funcional em memoria
6. **fetch**: Mock do fetch global
7. **console.error**: Supressao de warnings do React

## Comandos Uteis

### Executar teste especifico
```bash
npx vitest validation.test.ts
```

### Executar testes em modo UI filtrado
```bash
npm run test:ui -- validation
```

### Executar com verbosidade
```bash
npx vitest --reporter=verbose
```

### Executar em modo CI (sem watch)
```bash
npx vitest run
```

### Ver apenas testes que falharam
```bash
npx vitest --reporter=verbose --reporter=junit
```

## Debugging

### 1. Vitest UI
```bash
npm run test:ui
```
Melhor opcao para debugging visual.

### 2. Console.log em Testes
```typescript
it('should debug', () => {
  const result = validateRequired('', 'Nome');
  console.log('Result:', result);
  expect(result.isValid).toBe(false);
});
```

### 3. Screen.debug()
```typescript
import { screen } from '@testing-library/react';

it('should debug component', () => {
  render(<MyComponent />);
  screen.debug(); // Imprime o DOM
});
```

### 4. Coverage HTML
```bash
npm run test:coverage
# Abrir coverage/index.html no navegador
```

## Boas Praticas

### 1. Nomenclatura de Testes
```typescript
// ✅ BOM: Descreve o comportamento esperado
it('should return error when email is empty', () => {});

// ❌ RUIM: Vago
it('should work', () => {});
```

### 2. Arrange-Act-Assert
```typescript
it('should validate email', () => {
  // Arrange
  const email = 'test@example.com';

  // Act
  const result = validateEmail(email);

  // Assert
  expect(result.isValid).toBe(true);
});
```

### 3. Um Assert por Conceito
```typescript
// ✅ BOM: Testa um comportamento
it('should return error for empty field', () => {
  const result = validateRequired('', 'Email');
  expect(result.isValid).toBe(false);
  expect(result.error).toBe('Email é obrigatório');
});

// ❌ RUIM: Testa multiplos comportamentos nao relacionados
it('should validate everything', () => {
  // Testa 10 coisas diferentes
});
```

### 4. Isolamento de Testes
```typescript
beforeEach(() => {
  vi.clearAllMocks();
  localStorage.clear();
});
```

### 5. Testes Assincronos
```typescript
// ✅ BOM: Usa async/await
it('should login user', async () => {
  await act(async () => {
    await result.current.login(credentials);
  });
  expect(result.current.user).toBeDefined();
});

// ❌ RUIM: Nao aguarda promises
it('should login user', () => {
  result.current.login(credentials); // ❌ Missing await
});
```

## Proximos Passos

### Testes a Criar

1. **Components**
   - [ ] LoginForm.test.tsx
   - [ ] RegisterForm.test.tsx
   - [ ] TransactionList.test.tsx
   - [ ] Dashboard.test.tsx
   - [ ] Header.test.tsx
   - [ ] Sidebar.test.tsx

2. **Pages**
   - [ ] LoginPage.test.tsx
   - [ ] DashboardPage.test.tsx
   - [ ] TransactionsPage.test.tsx

3. **Integration Tests**
   - [ ] Complete login flow
   - [ ] Transaction CRUD flow
   - [ ] Category management flow

4. **E2E Tests** (Playwright/Cypress)
   - [ ] User registration
   - [ ] Complete transaction lifecycle
   - [ ] Multi-company management

## Troubleshooting

### Erro: "Cannot find module 'vitest'"
```bash
npm install -D vitest
```

### Erro: "jsdom is not defined"
```bash
npm install -D jsdom
```

### Erro: "localStorage is not defined"
Verifique se `test/setup.ts` esta sendo carregado no `vite.config.ts`:
```typescript
test: {
  setupFiles: './test/setup.ts'
}
```

### Testes passando localmente mas falhando no CI
- Verifique timezone
- Verifique variaveis de ambiente
- Use `npx vitest run` para simular CI

### Coverage muito baixa
```bash
npm run test:coverage
# Abrir coverage/index.html
# Linhas vermelhas = nao cobertas
# Adicionar testes para essas linhas
```

## Recursos Adicionais

- [Vitest Docs](https://vitest.dev/)
- [React Testing Library](https://testing-library.com/react)
- [Jest DOM Matchers](https://github.com/testing-library/jest-dom)
- [User Event](https://testing-library.com/docs/user-event/intro)
- [Testing Best Practices](https://kentcdodds.com/blog/common-mistakes-with-react-testing-library)

## Suporte

Para duvidas ou problemas:
1. Consulte `test/README.md` para documentacao detalhada
2. Veja `test/examples/ComponentTest.example.tsx` para templates
3. Execute `npm run test:ui` para debugging visual
4. Verifique coverage com `npm run test:coverage`

---

**Testing Virtuoso** - Guardiao da qualidade atraves de testes rigorosos.
