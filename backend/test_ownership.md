# Guia de Testes - Validação de Ownership

## Pré-requisitos
- Backend rodando em `http://localhost:8080`
- MongoDB conectado
- Duas contas de usuário diferentes

## Setup Inicial

### 1. Registrar Dois Usuários

```bash
# Usuário 1
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice",
    "email": "alice@example.com",
    "password": "senha123"
  }'

# Copie o token retornado
# TOKEN_ALICE="<token_aqui>"

# Usuário 2
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Bob",
    "email": "bob@example.com",
    "password": "senha123"
  }'

# Copie o token retornado
# TOKEN_BOB="<token_aqui>"
```

### 2. Alice Cria uma Empresa

```bash
curl -X POST http://localhost:8080/api/companies \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_ALICE" \
  -d '{
    "name": "Empresa da Alice"
  }'

# Copie o ID da empresa retornado
# COMPANY_ALICE="<id_aqui>"
```

### 3. Bob Cria uma Empresa

```bash
curl -X POST http://localhost:8080/api/companies \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_BOB" \
  -d '{
    "name": "Empresa do Bob"
  }'

# Copie o ID da empresa retornado
# COMPANY_BOB="<id_aqui>"
```

## Testes de Validação

### Teste 1: Listar Empresas (Sucesso)

**Alice vê apenas suas empresas:**
```bash
curl -X GET http://localhost:8080/api/companies \
  -H "Authorization: Bearer $TOKEN_ALICE"

# Esperado: Retorna apenas "Empresa da Alice"
```

**Bob vê apenas suas empresas:**
```bash
curl -X GET http://localhost:8080/api/companies \
  -H "Authorization: Bearer $TOKEN_BOB"

# Esperado: Retorna apenas "Empresa do Bob"
```

### Teste 2: Tentar Acessar Empresa de Outro Usuário (Falha)

**Bob tenta atualizar empresa da Alice:**
```bash
curl -X PUT http://localhost:8080/api/companies/$COMPANY_ALICE \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_BOB" \
  -d '{
    "name": "Tentativa de Hack"
  }'

# Esperado: HTTP 403 Forbidden
# {
#   "error": "Forbidden: you do not have permission to access this company",
#   "details": "This company belongs to another user"
# }
```

**Bob tenta deletar empresa da Alice:**
```bash
curl -X DELETE http://localhost:8080/api/companies/$COMPANY_ALICE \
  -H "Authorization: Bearer $TOKEN_BOB"

# Esperado: HTTP 403 Forbidden
```

### Teste 3: Criar Transação (Sucesso)

**Alice cria transação em sua empresa:**
```bash
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_ALICE" \
  -d '{
    "companyId": "'$COMPANY_ALICE'",
    "month": "Janeiro",
    "year": 2025,
    "amount": 5000,
    "category": "Salário",
    "status": "PAGO",
    "description": "Salário de Janeiro"
  }'

# Esperado: HTTP 201 Created
# Copie o ID da transação
# TRANSACTION_ALICE="<id_aqui>"
```

### Teste 4: Tentar Criar Transação em Empresa de Outro Usuário (Falha)

**Bob tenta criar transação na empresa da Alice:**
```bash
curl -X POST http://localhost:8080/api/transactions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_BOB" \
  -d '{
    "companyId": "'$COMPANY_ALICE'",
    "month": "Janeiro",
    "year": 2025,
    "amount": 1000000,
    "category": "Fraude",
    "status": "PAGO",
    "description": "Tentativa de fraude"
  }'

# Esperado: HTTP 403 Forbidden
```

### Teste 5: Listar Transações (Sucesso)

**Alice lista transações de sua empresa:**
```bash
curl -X GET "http://localhost:8080/api/transactions?companyId=$COMPANY_ALICE" \
  -H "Authorization: Bearer $TOKEN_ALICE"

# Esperado: Retorna transações da empresa da Alice
```

### Teste 6: Tentar Listar Transações de Outro Usuário (Falha)

**Bob tenta listar transações da empresa da Alice:**
```bash
curl -X GET "http://localhost:8080/api/transactions?companyId=$COMPANY_ALICE" \
  -H "Authorization: Bearer $TOKEN_BOB"

# Esperado: HTTP 403 Forbidden
```

### Teste 7: Atualizar Transação (Sucesso)

**Alice atualiza sua própria transação:**
```bash
curl -X PUT http://localhost:8080/api/transactions/$TRANSACTION_ALICE \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_ALICE" \
  -d '{
    "month": "Janeiro",
    "year": 2025,
    "amount": 5500,
    "category": "Salário",
    "status": "PAGO",
    "description": "Salário com aumento"
  }'

# Esperado: HTTP 200 OK
```

### Teste 8: Tentar Atualizar Transação de Outro Usuário (Falha)

**Bob tenta atualizar transação da Alice:**
```bash
curl -X PUT http://localhost:8080/api/transactions/$TRANSACTION_ALICE \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_BOB" \
  -d '{
    "month": "Janeiro",
    "year": 2025,
    "amount": 1,
    "category": "Hack",
    "status": "ABERTO",
    "description": "Tentativa de modificação"
  }'

# Esperado: HTTP 403 Forbidden
```

### Teste 9: Criar Categoria (Sucesso)

**Alice cria categoria em sua empresa:**
```bash
curl -X POST "http://localhost:8080/api/categories?companyId=$COMPANY_ALICE" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_ALICE" \
  -d '{
    "name": "Salário",
    "type": "INCOME",
    "color": "#00FF00",
    "budget": 10000
  }'

# Esperado: HTTP 201 Created
# CATEGORY_ALICE="<id_aqui>"
```

### Teste 10: Tentar Criar Categoria em Empresa de Outro Usuário (Falha)

**Bob tenta criar categoria na empresa da Alice:**
```bash
curl -X POST "http://localhost:8080/api/categories?companyId=$COMPANY_ALICE" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_BOB" \
  -d '{
    "name": "Fraude",
    "type": "EXPENSE",
    "color": "#FF0000",
    "budget": 0
  }'

# Esperado: HTTP 403 Forbidden
```

### Teste 11: Criar Regra Recorrente (Sucesso)

**Alice cria regra recorrente:**
```bash
curl -X POST http://localhost:8080/api/recurring \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN_ALICE" \
  -d '{
    "companyId": "'$COMPANY_ALICE'",
    "description": "Aluguel mensal",
    "amount": 1500,
    "category": "Despesas",
    "dayOfMonth": 5
  }'

# Esperado: HTTP 200 OK
# RECURRING_ALICE="<id_aqui>"
```

### Teste 12: Tentar Deletar Regra de Outro Usuário (Falha)

**Bob tenta deletar regra da Alice:**
```bash
curl -X DELETE http://localhost:8080/api/recurring/$RECURRING_ALICE \
  -H "Authorization: Bearer $TOKEN_BOB"

# Esperado: HTTP 403 Forbidden
```

### Teste 13: Stats com Ownership (Sucesso)

**Alice vê estatísticas de sua empresa:**
```bash
curl -X GET "http://localhost:8080/api/stats?companyId=$COMPANY_ALICE" \
  -H "Authorization: Bearer $TOKEN_ALICE"

# Esperado: Retorna estatísticas corretas
```

**Bob tenta ver estatísticas da empresa da Alice:**
```bash
curl -X GET "http://localhost:8080/api/stats?companyId=$COMPANY_ALICE" \
  -H "Authorization: Bearer $TOKEN_BOB"

# Esperado: HTTP 403 Forbidden
```

### Teste 14: Toggle Status (Sucesso e Falha)

**Alice alterna status de sua transação:**
```bash
curl -X PATCH http://localhost:8080/api/transactions/$TRANSACTION_ALICE/toggle-status \
  -H "Authorization: Bearer $TOKEN_ALICE"

# Esperado: HTTP 200 OK
```

**Bob tenta alterar transação da Alice:**
```bash
curl -X PATCH http://localhost:8080/api/transactions/$TRANSACTION_ALICE/toggle-status \
  -H "Authorization: Bearer $TOKEN_BOB"

# Esperado: HTTP 403 Forbidden
```

## Testes de Edge Cases

### Teste 15: ID Inválido

```bash
curl -X GET http://localhost:8080/api/companies/invalid-id \
  -H "Authorization: Bearer $TOKEN_ALICE"

# Esperado: HTTP 400 Bad Request
# { "error": "Invalid ID format" }
```

### Teste 16: Recurso Inexistente

```bash
curl -X GET http://localhost:8080/api/companies/507f1f77bcf86cd799439011 \
  -H "Authorization: Bearer $TOKEN_ALICE"

# Esperado: HTTP 404 Not Found
# { "error": "Company not found" }
```

### Teste 17: Sem Token de Autenticação

```bash
curl -X GET http://localhost:8080/api/companies

# Esperado: HTTP 401 Unauthorized
# { "error": "Missing authorization token" }
```

### Teste 18: Token Inválido

```bash
curl -X GET http://localhost:8080/api/companies \
  -H "Authorization: Bearer token_invalido"

# Esperado: HTTP 401 Unauthorized
# { "error": "Invalid or expired token" }
```

## Resumo dos Códigos Esperados

| Cenário | Código HTTP | Descrição |
|---------|-------------|-----------|
| Sucesso em operação | 200 OK | Operação bem-sucedida |
| Criação bem-sucedida | 201 Created | Recurso criado |
| Deleção bem-sucedida | 204 No Content | Recurso deletado |
| ID inválido | 400 Bad Request | Formato de ID incorreto |
| Não autenticado | 401 Unauthorized | Token ausente ou inválido |
| Ownership violado | 403 Forbidden | Tentativa de acessar dados de outro usuário |
| Recurso não encontrado | 404 Not Found | Recurso não existe |
| Erro de servidor | 500 Internal Server Error | Erro inesperado |

## Verificação de Logs

Durante os testes, monitore os logs do servidor para ver as validações acontecendo:

```bash
# No terminal onde o servidor está rodando, você verá:
# Validações bem-sucedidas passam silenciosamente
# Tentativas de acesso não autorizado são logadas
```

## Checklist de Validação

- [ ] Alice pode criar suas próprias empresas
- [ ] Alice pode ver apenas suas empresas
- [ ] Bob não pode ver empresas da Alice
- [ ] Bob não pode modificar empresas da Alice
- [ ] Bob não pode deletar empresas da Alice
- [ ] Alice pode criar transações em suas empresas
- [ ] Bob não pode criar transações nas empresas da Alice
- [ ] Bob não pode modificar transações da Alice
- [ ] Bob não pode deletar transações da Alice
- [ ] Alice pode criar categorias em suas empresas
- [ ] Bob não pode criar categorias nas empresas da Alice
- [ ] Alice pode criar regras recorrentes em suas empresas
- [ ] Bob não pode modificar regras da Alice
- [ ] Estatísticas respeitam ownership
- [ ] IDs inválidos retornam erro apropriado
- [ ] Recursos inexistentes retornam 404
- [ ] Requests sem token retornam 401
- [ ] Tokens inválidos retornam 401
