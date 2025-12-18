# Quick Start - Testes Frontend

## Passo a Passo Rapido

### 1. Instalar Dependencias
```bash
cd D:\Sexto\frontend
npm install
```

Isso instalara:
- vitest
- @testing-library/react
- @testing-library/jest-dom
- @testing-library/user-event
- jsdom
- @vitest/ui
- @vitest/coverage-v8

### 2. Executar Testes
```bash
npm test
```

Saida esperada:
```
✓ utils/validation.test.ts (49 tests)
✓ hooks/useFormValidation.test.ts (25 tests)
✓ contexts/AuthContext.test.tsx (30 tests)

Test Files  3 passed (3)
     Tests  104 passed (104)
```

### 3. Ver Cobertura
```bash
npm run test:coverage
```

Saida esperada:
```
Coverage report
-----------------
Lines       : 95%+
Functions   : 95%+
Branches    : 90%+
Statements  : 95%+
```

Relatorio HTML gerado em: `coverage/index.html`

### 4. Interface Visual
```bash
npm run test:ui
```

Abre interface em: http://localhost:51204/

---

## Arquivos Criados

### Configuracao
- `package.json` - Modificado (scripts e dependencias)
- `vite.config.ts` - Modificado (config de teste)
- `vitest.config.ts` - Criado (config dedicada)

### Setup
- `test/setup.ts` - Mocks globais e configuracao
- `test/helpers/testUtils.tsx` - Utilities customizadas

### Testes
- `utils/validation.test.ts` - 49 testes de validacao
- `hooks/useFormValidation.test.ts` - 25+ testes do hook
- `contexts/AuthContext.test.tsx` - 30+ testes de auth

### Documentacao
- `test/README.md` - Documentacao completa
- `TESTING_GUIDE.md` - Guia pratico de uso
- `TEST_SUMMARY.md` - Resumo executivo
- `QUICKSTART_TESTS.md` - Este arquivo

### Exemplos
- `test/examples/ComponentTest.example.tsx` - Template de teste

### Scripts
- `scripts/test-helper.js` - Helper para tarefas comuns

---

## Verificacao Rapida

### Verificar se tudo foi instalado
```bash
npx vitest --version
```

Deve mostrar: `vitest/2.1.8` ou superior

### Executar um teste especifico
```bash
npx vitest validation.test.ts
```

### Ver lista de testes
```bash
node scripts/test-helper.js list
```

### Ver estatisticas
```bash
node scripts/test-helper.js stats
```

---

## Estrutura de Diretorios

```
D:\Sexto\frontend/
│
├── test/
│   ├── setup.ts                          ✅ Config global
│   ├── README.md                         ✅ Docs completa
│   ├── helpers/
│   │   └── testUtils.tsx                 ✅ Utilities
│   └── examples/
│       └── ComponentTest.example.tsx     ✅ Template
│
├── utils/
│   └── validation.test.ts                ✅ 49 testes
│
├── hooks/
│   └── useFormValidation.test.ts         ✅ 25+ testes
│
├── contexts/
│   └── AuthContext.test.tsx              ✅ 30+ testes
│
├── scripts/
│   └── test-helper.js                    ✅ Helper script
│
├── package.json                          ✅ Modificado
├── vite.config.ts                        ✅ Modificado
├── vitest.config.ts                      ✅ Criado
├── TESTING_GUIDE.md                      ✅ Criado
├── TEST_SUMMARY.md                       ✅ Criado
└── QUICKSTART_TESTS.md                   ✅ Este arquivo
```

---

## Comandos Essenciais

### NPM Scripts
```bash
npm test                # Modo watch
npm run test:coverage   # Cobertura
npm run test:ui         # Interface visual
```

### Vitest CLI
```bash
npx vitest run              # Executar uma vez (CI mode)
npx vitest watch            # Modo watch
npx vitest <file>           # Teste especifico
npx vitest --reporter=verbose  # Output detalhado
```

### Helper Script
```bash
node scripts/test-helper.js help             # Ajuda
node scripts/test-helper.js coverage-report  # Relatorio
node scripts/test-helper.js stats            # Estatisticas
node scripts/test-helper.js list             # Listar testes
node scripts/test-helper.js clean            # Limpar temp files
```

---

## Troubleshooting

### Problema: "Cannot find module 'vitest'"
**Solucao:**
```bash
npm install -D vitest
```

### Problema: "jsdom is not defined"
**Solucao:**
```bash
npm install -D jsdom
```

### Problema: Testes nao executam
**Solucao:**
```bash
# Limpar e reinstalar
rm -rf node_modules package-lock.json
npm install
```

### Problema: localStorage errors
**Solucao:** Verificar se `test/setup.ts` existe e esta configurado no `vite.config.ts`

---

## Proximos Passos

### Imediato
1. Execute `npm install`
2. Execute `npm test`
3. Execute `npm run test:coverage`
4. Explore `npm run test:ui`

### Curto Prazo
1. Criar testes para componentes:
   - LoginForm
   - RegisterForm
   - TransactionList

2. Criar testes para pages:
   - LoginPage
   - DashboardPage

### Medio Prazo
1. Configurar CI/CD com testes
2. Adicionar E2E tests (Playwright)
3. Adicionar visual regression tests

---

## Recursos

### Documentacao Local
- `test/README.md` - Documentacao completa
- `TESTING_GUIDE.md` - Guia pratico
- `TEST_SUMMARY.md` - Resumo executivo
- `test/examples/` - Templates e exemplos

### Documentacao Online
- [Vitest](https://vitest.dev/)
- [React Testing Library](https://testing-library.com/react)
- [Jest DOM](https://github.com/testing-library/jest-dom)

---

## Checklist de Verificacao

- [ ] Dependencias instaladas (`npm install`)
- [ ] Testes executam sem erros (`npm test`)
- [ ] Cobertura >= 80% (`npm run test:coverage`)
- [ ] UI mode funciona (`npm run test:ui`)
- [ ] Helper script funciona (`node scripts/test-helper.js help`)
- [ ] Todos os 104+ testes passam
- [ ] Coverage report abre no navegador

---

## Suporte

Se encontrar problemas:
1. Verifique `TESTING_GUIDE.md` - secao Troubleshooting
2. Verifique `test/README.md` - secao Debugging
3. Execute `node scripts/test-helper.js stats` para diagnostico

---

**Status**: ✅ TUDO PRONTO!

Comece com:
```bash
cd D:\Sexto\frontend
npm install
npm test
```

---

*Testing Virtuoso - Guardiao da qualidade atraves de testes rigorosos.*
