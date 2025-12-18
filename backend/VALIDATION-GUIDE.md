# Guia de Validações - Backend M2M

Sistema completo de validação de inputs implementado para o backend Go/Fiber.

## Estrutura

### Arquivo de Validações
- **Localização**: `helpers/validation.go`
- **Funções reutilizáveis** para validação de diferentes tipos de dados
- **Mensagens em português** para facilitar o uso

## Validações Implementadas por Entidade

### 1. Transaction (handlers/transaction.go)

**Campos validados:**
- `amount`: Deve ser maior que zero
- `description`: Máximo de 200 caracteres
- `category`: Não pode ser vazio
- `month`: Deve ser um mês válido em português
- `year`: Deve estar entre 2000 e 2100
- `status`: Deve ser "PAGO" ou "ABERTO"

**Exemplo de resposta de erro:**
```json
{
  "errors": [
    {
      "field": "amount",
      "message": "amount deve ser maior que zero"
    },
    {
      "field": "month",
      "message": "Mês inválido. Use mês em português (ex: Janeiro, Fevereiro, etc.)"
    }
  ]
}
```

### 2. Category (handlers/category.go)

**Campos validados:**
- `name`: Não pode ser vazio, máximo de 50 caracteres
- `color`: Deve ser formato hex válido (#RRGGBB)

**Exemplo de resposta de erro:**
```json
{
  "errors": [
    {
      "field": "name",
      "message": "name não pode ser vazio"
    },
    {
      "field": "color",
      "message": "Cor deve estar no formato hexadecimal #RRGGBB (ex: #FF5733)"
    }
  ]
}
```

### 3. Company (handlers/company.go)

**Campos validados:**
- `name`: Não pode ser vazio, máximo de 100 caracteres

**Exemplo de resposta de erro:**
```json
{
  "errors": [
    {
      "field": "name",
      "message": "name não pode ser vazio"
    }
  ]
}
```

### 4. RecurringTransaction (handlers/recurring.go)

**Campos validados:**
- `description`: Não pode ser vazio, máximo de 200 caracteres
- `amount`: Deve ser maior que zero
- `dayOfMonth`: Deve estar entre 1 e 31
- `category`: Não pode ser vazio

**Exemplo de resposta de erro:**
```json
{
  "errors": [
    {
      "field": "dayOfMonth",
      "message": "Dia do mês deve estar entre 1 e 31"
    }
  ]
}
```

### 5. Auth (handlers/auth.go)

**Registro - Campos validados:**
- `email`: Deve ser formato email válido
- `password`: Mínimo de 6 caracteres
- `name`: Não pode ser vazio

**Login - Campos validados:**
- `email`: Deve ser formato email válido
- `password`: Não pode ser vazio

**Exemplo de resposta de erro:**
```json
{
  "errors": [
    {
      "field": "email",
      "message": "Formato de email inválido"
    },
    {
      "field": "password",
      "message": "password deve ter no mínimo 6 caracteres"
    }
  ]
}
```

## Funções de Validação Disponíveis

### Funções Básicas

```go
// Valida campo obrigatório
ValidateRequired(value, fieldName string) *ValidationError

// Valida comprimento máximo
ValidateMaxLength(value, fieldName string, maxLength int) *ValidationError

// Valida comprimento mínimo
ValidateMinLength(value, fieldName string, minLength int) *ValidationError

// Valida número positivo
ValidatePositiveNumber(value float64, fieldName string) *ValidationError

// Valida intervalo numérico
ValidateRange(value int, min, max int, fieldName string) *ValidationError
```

### Funções Especializadas

```go
// Valida formato de email
ValidateEmail(email string) *ValidationError

// Valida cor hexadecimal (#RRGGBB)
ValidateHexColor(color string) *ValidationError

// Valida mês em português
ValidateMonth(month string) *ValidationError

// Valida status (PAGO/ABERTO)
ValidateStatus(status string) *ValidationError

// Valida dia do mês (1-31)
ValidateDayOfMonth(day int) *ValidationError
```

### Funções Auxiliares

```go
// Coleta múltiplos erros e retorna apenas os não nulos
CollectErrors(errors ...*ValidationError) []ValidationError

// Verifica se há erros na lista
HasErrors(errors []ValidationError) bool

// Envia erro de validação único
SendValidationError(c *fiber.Ctx, field, message string) error

// Envia múltiplos erros de validação
SendValidationErrors(c *fiber.Ctx, errors []ValidationError) error
```

## Como Usar nas Rotas

### Exemplo Completo

```go
func CreateTransaction(c *fiber.Ctx) error {
    var req models.CreateTransactionRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Corpo da requisição inválido"})
    }

    // Validações
    errors := helpers.CollectErrors(
        helpers.ValidatePositiveNumber(req.Amount, "amount"),
        helpers.ValidateMaxLength(req.Description, "description", 200),
        helpers.ValidateRequired(req.Category, "category"),
        helpers.ValidateMonth(req.Month),
        helpers.ValidateRange(req.Year, 2000, 2100, "year"),
        helpers.ValidateStatus(string(req.Status)),
    )

    if helpers.HasErrors(errors) {
        return helpers.SendValidationErrors(c, errors)
    }

    // Prosseguir com a lógica de negócio...
}
```

## Meses Válidos em Português

O sistema aceita os seguintes meses:
- Janeiro
- Fevereiro
- Março
- Abril
- Maio
- Junho
- Julho
- Agosto
- Setembro
- Outubro
- Novembro
- Dezembro
- Acumulado (para casos especiais como férias/13º)

## Status HTTP de Resposta

- **400 Bad Request**: Erro de validação de input
- **401 Unauthorized**: Credenciais inválidas
- **404 Not Found**: Recurso não encontrado
- **409 Conflict**: Email já cadastrado
- **500 Internal Server Error**: Erro no servidor

## Benefícios

1. **Consistência**: Todas as validações seguem o mesmo padrão
2. **Reutilização**: Funções podem ser usadas em qualquer handler
3. **Mensagens claras**: Erros em português facilitam o debug
4. **Centralização**: Fácil manutenção e atualização
5. **Múltiplos erros**: Retorna todos os erros de uma vez, não apenas o primeiro
6. **Type-safe**: Validações com tipos apropriados (int, float64, string)

## Testando as Validações

### Exemplo com cURL - Transaction Inválida

```bash
curl -X POST http://localhost:5000/api/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "companyId": "123",
    "amount": -100,
    "month": "InvalidMonth",
    "year": 1999,
    "status": "invalid"
  }'
```

**Resposta esperada (400):**
```json
{
  "errors": [
    {
      "field": "amount",
      "message": "amount deve ser maior que zero"
    },
    {
      "field": "month",
      "message": "Mês inválido. Use mês em português (ex: Janeiro, Fevereiro, etc.)"
    },
    {
      "field": "year",
      "message": "year deve estar entre 2000 e 2100"
    },
    {
      "field": "status",
      "message": "Status deve ser 'PAGO' ou 'ABERTO'"
    }
  ]
}
```

### Exemplo com cURL - Registro Inválido

```bash
curl -X POST http://localhost:5000/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "invalid-email",
    "password": "123",
    "name": ""
  }'
```

**Resposta esperada (400):**
```json
{
  "errors": [
    {
      "field": "name",
      "message": "name não pode ser vazio"
    },
    {
      "field": "email",
      "message": "Formato de email inválido"
    },
    {
      "field": "password",
      "message": "password deve ter no mínimo 6 caracteres"
    }
  ]
}
```

## Próximos Passos

Para adicionar novas validações:

1. Adicione a função de validação em `helpers/validation.go`
2. Use a função no handler apropriado
3. Teste com casos válidos e inválidos
4. Documente neste arquivo

## Arquivos Modificados

- `helpers/validation.go` (NOVO)
- `handlers/transaction.go` (MODIFICADO)
- `handlers/category.go` (MODIFICADO)
- `handlers/company.go` (MODIFICADO)
- `handlers/recurring.go` (MODIFICADO)
- `handlers/auth.go` (MODIFICADO)
