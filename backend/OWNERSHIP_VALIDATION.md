# Implementação de Validação de Ownership

## Resumo
Sistema de validação de ownership implementado para garantir que usuários só possam acessar e manipular dados de empresas que pertencem a eles.

## Arquitetura

### 1. Modelo de Dados Atualizado
**Arquivo:** `models/company.go`
- Adicionado campo `UserID` ao modelo `Company`
- Tipo: `primitive.ObjectID`
- Liga cada empresa a um usuário específico

### 2. Helper de Validação
**Arquivo:** `helpers/ownership.go`

Funções principais:

#### `GetUserIDFromContext(c *fiber.Ctx)`
Extrai e valida o userID do contexto do Fiber (inserido pelo middleware de autenticação).

#### `ValidateCompanyOwnership(c *fiber.Ctx, companyID primitive.ObjectID)`
Valida que o usuário autenticado é o dono da empresa:
1. Busca a empresa no banco de dados
2. Compara `company.UserID` com o `userID` do contexto
3. Retorna erro 403 Forbidden se não for o owner
4. Retorna erro 404 se a empresa não existir

#### `ValidateCompanyOwnershipByString(c *fiber.Ctx, companyIDStr string)`
Wrapper conveniente que aceita companyID como string.

#### `ValidateTransactionOwnership(c *fiber.Ctx, transactionID primitive.ObjectID)`
Valida ownership de uma transação através da validação da empresa associada.

#### `ValidateCategoryOwnership(c *fiber.Ctx, categoryID primitive.ObjectID)`
Valida ownership de uma categoria através da validação da empresa associada.

#### `ValidateRecurringOwnership(c *fiber.Ctx, recurringID primitive.ObjectID)`
Valida ownership de uma regra recorrente através da validação da empresa associada.

## Handlers Atualizados

### Company Handler (`handlers/company.go`)

#### `GetCompanies()`
- Filtra empresas por `userID`
- Retorna apenas empresas do usuário autenticado

#### `CreateCompany()`
- Atribui automaticamente o `userID` do usuário autenticado à nova empresa

#### `UpdateCompany()`
- Valida ownership antes de permitir atualização

#### `DeleteCompany()`
- Valida ownership antes de permitir exclusão
- Mantém cascade delete de dados relacionados

### Transaction Handler (`handlers/transaction.go`)

#### `GetAllTransactions()`
- Se `companyId` fornecido: valida ownership da empresa
- Se não fornecido: retorna transações de todas as empresas do usuário

#### `GetTransaction()`
- Valida ownership antes de retornar transação individual

#### `CreateTransaction()`
- Valida ownership da empresa antes de criar transação

#### `UpdateTransaction()`
- Valida ownership antes de permitir atualização

#### `DeleteTransaction()`
- Valida ownership antes de permitir exclusão

#### `GetStats()`
- Se `companyId` fornecido: valida ownership
- Se não fornecido: retorna estatísticas de todas as empresas do usuário

#### `ToggleStatus()`
- Valida ownership antes de alterar status

#### `SeedTransactions()`
- Cria empresa com `userID` do usuário autenticado ao fazer seed

### Category Handler (`handlers/category.go`)

#### `GetCategories()`
- Valida ownership da empresa antes de listar categorias

#### `CreateCategory()`
- Valida ownership da empresa antes de criar categoria

#### `UpdateCategory()`
- Valida ownership da categoria antes de atualizar

#### `DeleteCategory()`
- Valida ownership da categoria antes de deletar

### Recurring Handler (`handlers/recurring.go`)

#### `GetRecurring()`
- Valida ownership da empresa antes de listar regras

#### `CreateRecurring()`
- Valida ownership da empresa antes de criar regra

#### `DeleteRecurring()`
- Valida ownership da regra antes de deletar

#### `ProcessRecurring()`
- Valida ownership da empresa antes de processar regras

## Códigos de Status HTTP

### 400 Bad Request
- Formato de ID inválido
- Parâmetros obrigatórios faltando

### 401 Unauthorized
- Usuário não autenticado
- Token JWT inválido ou expirado

### 403 Forbidden
- Usuário autenticado tenta acessar dados de empresa de outro usuário
- Mensagem: "Forbidden: you do not have permission to access this company"

### 404 Not Found
- Recurso não encontrado (empresa, transação, categoria, etc.)

### 500 Internal Server Error
- Erros de banco de dados
- Erros inesperados do servidor

## Fluxo de Validação

```
1. Request chega ao handler
   ↓
2. Middleware de autenticação extrai userID do JWT
   ↓
3. Handler chama helper de validação
   ↓
4. Helper busca recurso no banco de dados
   ↓
5. Helper valida se recurso.companyID pertence ao userID
   ↓
6. Se válido: continua processamento
   Se inválido: retorna 403 Forbidden
```

## Rotas Protegidas

Todas as rotas abaixo do middleware `api.Use(middleware.Protected())` em `main.go`:

- `GET /api/companies`
- `POST /api/companies`
- `PUT /api/companies/:id`
- `DELETE /api/companies/:id`
- `GET /api/transactions`
- `GET /api/transactions/:id`
- `POST /api/transactions`
- `PUT /api/transactions/:id`
- `DELETE /api/transactions/:id`
- `PATCH /api/transactions/:id/toggle-status`
- `GET /api/stats`
- `GET /api/categories`
- `POST /api/categories`
- `PUT /api/categories/:id`
- `DELETE /api/categories/:id`
- `GET /api/recurring`
- `POST /api/recurring`
- `DELETE /api/recurring/:id`
- `POST /api/recurring/process`
- `POST /api/seed`

## Segurança

### Princípios Implementados

1. **Least Privilege**: Usuários só acessam seus próprios dados
2. **Defense in Depth**: Validação em múltiplas camadas
3. **Fail Secure**: Em caso de erro, nega acesso por padrão
4. **Clear Error Messages**: Mensagens de erro claras mas sem expor informações sensíveis

### Prevenção de Ataques

- **IDOR (Insecure Direct Object Reference)**: Impedido pela validação de ownership
- **Horizontal Privilege Escalation**: Impedido - usuários não podem acessar dados de outros usuários
- **Data Leakage**: Queries filtradas por userID previnem vazamento de dados

## Testes Recomendados

1. **Cenário Positivo**: Usuário acessa seus próprios dados
2. **Cenário Negativo**: Usuário tenta acessar dados de outro usuário
3. **Edge Case**: IDs inválidos ou malformados
4. **Edge Case**: Recursos inexistentes
5. **Edge Case**: Usuário não autenticado

## Performance

- Validações usam índices de banco de dados em `_id` e `userId`
- Timeout de 5 segundos para operações de validação
- Cache de contexto do Fiber evita múltiplas extrações de userID

## Manutenção

Para adicionar validação a novos recursos:

1. Adicionar campo `userID` ao modelo se for um recurso raiz
2. Ou adicionar `companyID` se for um recurso filho de Company
3. Criar função de validação em `helpers/ownership.go`
4. Adicionar chamada de validação no handler
5. Atualizar documentação

## Migração de Dados

**IMPORTANTE**: Empresas existentes no banco de dados não terão `userID`. É necessário:

1. Criar script de migração para popular `userID` em empresas existentes
2. Ou limpar banco de dados de desenvolvimento
3. Adicionar índice no campo `userId` em `companies` collection para performance

```javascript
// Script MongoDB para criar índice
db.companies.createIndex({ "userId": 1 })
```
