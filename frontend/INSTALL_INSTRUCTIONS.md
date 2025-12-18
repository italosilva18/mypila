# Instrucoes de Instalacao - ESLint e TypeScript Strict

## Passo 1: Instalar Dependencias

Execute o seguinte comando no diretorio `D:\Sexto\frontend`:

```bash
npm install
```

Este comando instalara automaticamente todas as novas dependencias adicionadas ao `package.json`:

- @types/react@^19.0.0
- @types/react-dom@^19.0.0
- @typescript-eslint/eslint-plugin@^8.0.0
- @typescript-eslint/parser@^8.0.0
- eslint@^9.0.0
- eslint-plugin-jsx-a11y@^6.10.0
- eslint-plugin-react@^7.37.0
- eslint-plugin-react-hooks@^5.0.0
- prettier@^3.4.0

---

## Passo 2: Verificar Configuracoes

Apos a instalacao, verifique se os seguintes arquivos foram criados/atualizados:

### Arquivos de Configuracao:
- [x] `.eslintrc.json` - Configuracao do ESLint
- [x] `.prettierrc` - Configuracao do Prettier
- [x] `.eslintignore` - Arquivos ignorados pelo ESLint
- [x] `.prettierignore` - Arquivos ignorados pelo Prettier
- [x] `tsconfig.json` - TypeScript strict mode ativado

---

## Passo 3: Executar Verificacoes

### Verificar Lint:
```bash
npm run lint
```

### Verificar Formatacao:
```bash
npm run format:check
```

### Compilar TypeScript:
```bash
npm run build
```

---

## Passo 4: Corrigir Problemas Automaticamente

### Corrigir Lint:
```bash
npm run lint:fix
```

### Formatar Codigo:
```bash
npm run format
```

---

## Passo 5: Integrar no Workflow

### VS Code Settings (recomendado)

Crie/atualize `.vscode/settings.json`:

```json
{
  "editor.formatOnSave": true,
  "editor.defaultFormatter": "esbenp.prettier-vscode",
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "eslint.validate": [
    "javascript",
    "javascriptreact",
    "typescript",
    "typescriptreact"
  ]
}
```

### Extensoes VS Code Recomendadas:
- ESLint (dbaeumer.vscode-eslint)
- Prettier - Code formatter (esbenp.prettier-vscode)
- TypeScript and JavaScript Language Features (embutida)

---

## Passo 6: Testar Aplicacao

Execute a aplicacao para garantir que tudo funciona:

```bash
npm run dev
```

Abra o navegador e acesse: http://localhost:5173

---

## Resolucao de Problemas

### Erro: "Module not found"
```bash
rm -rf node_modules package-lock.json
npm install
```

### Erro: "Cannot find module '@typescript-eslint/parser'"
```bash
npm install --save-dev @typescript-eslint/parser @typescript-eslint/eslint-plugin
```

### ESLint nao funciona no VS Code:
1. Reinicie o VS Code
2. Execute: "Developer: Reload Window" (Ctrl+Shift+P)
3. Verifique se a extensao ESLint esta instalada e ativada

---

## Comandos Rapidos

```bash
# Instalacao completa
npm install

# Verificacoes
npm run lint
npm run format:check
npm run build

# Correcoes automaticas
npm run lint:fix
npm run format

# Executar aplicacao
npm run dev
```

---

## Checklist Final

- [ ] `npm install` executado com sucesso
- [ ] `npm run lint` sem erros criticos
- [ ] `npm run format:check` sem problemas
- [ ] `npm run build` compila sem erros
- [ ] `npm run dev` inicia a aplicacao
- [ ] Navegador abre sem erros no console
- [ ] Extensoes VS Code instaladas
- [ ] Settings do VS Code configurados

---

## Suporte

Se encontrar problemas, verifique:

1. **Versao do Node.js:** >= 18.x
2. **Versao do NPM:** >= 9.x
3. **Permissoes de escrita** no diretorio
4. **Espaco em disco** suficiente

---

## Proximos Passos Apos Instalacao

1. Revisar e corrigir warnings do ESLint
2. Configurar pre-commit hooks (opcional)
3. Integrar no CI/CD pipeline
4. Documentar padroes de codigo da equipe

---

Configuracao implementada por: Frontend Artisan
Data: 2025-12-16
