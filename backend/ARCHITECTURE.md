# Backend Architecture - Service Layer Pattern

## Visao Geral

Este backend implementa o **Service Layer Pattern**, separando claramente as responsabilidades entre camadas:

```
HTTP Request -> Handler -> Service -> Repository -> MongoDB
HTTP Response <- Handler <- Service <- Repository <- MongoDB
```

## Estrutura de Camadas

### 1. Handler Layer (handlers/)
**Responsabilidade:** Gerenciar requisicoes HTTP e respostas

- Valida parametros de entrada (ID, body parsing)
- Extrai dados do contexto (user ID, params)
- Chama a camada de Service
- Formata e retorna respostas HTTP
- Trata erros especificos do HTTP (400, 404, 500)

**Exemplo:** `handlers/company.go`

### 2. Service Layer (services/)
**Responsabilidade:** Logica de negocio e orquestracao

- Validacao de regras de negocio
- Sanitizacao de dados
- Orquestracao de multiplas operacoes
- Gestao de transacoes complexas
- Validacao de propriedade (ownership)
- Erros de dominio (ErrCompanyNotFound, ErrUnauthorized)

**Exemplo:** `services/company_service.go`

### 3. Repository Layer (repositories/)
**Responsabilidade:** Acesso a dados e persistencia

- Operacoes CRUD no MongoDB
- Queries especificas
- Gestao de indices
- Erros de repositorio (RepositoryError)
- Isolamento da tecnologia de persistencia

**Exemplo:** `repositories/company_repository.go`

## Fluxo de Dados - Exemplo: Criar Company

```go
1. HTTP POST /api/companies
   Body: {"name": "Empresa XYZ"}

2. Handler (company.go)
   - Extrai userID do contexto
   - Parse do body
   - Chama service.CreateCompany(userID, request)

3. Service (company_service.go)
   - Valida input (required, max length, XSS, SQL injection)
   - Sanitiza dados (remove HTML, script tags)
   - Cria model Company com ID e timestamp
   - Chama repository.Create(company)

4. Repository (company_repository.go)
   - Insere no MongoDB
   - Retorna erro se falhar

5. Resposta
   <- Repository retorna nil (sucesso)
   <- Service retorna company criada
   <- Handler retorna 201 com JSON da company
```

## Vantagens da Arquitetura

### Separacao de Responsabilidades
- **Handler:** So lida com HTTP
- **Service:** So lida com logica de negocio
- **Repository:** So lida com banco de dados

### Testabilidade
- Cada camada pode ser testada independentemente
- Facil criar mocks para testes unitarios
- Services podem ser testados sem HTTP
- Repositories podem ser testados sem logica de negocio

### Manutenibilidade
- Codigo organizado e facil de navegar
- Mudancas em uma camada nao afetam outras
- Facil adicionar novos recursos

### Reusabilidade
- Services podem ser chamados de multiplos handlers
- Repositories podem ser reutilizados por multiplos services
- Logica de negocio centralizada

### Escalabilidade
- Facil adicionar cache na camada de repository
- Facil implementar circuit breakers
- Facil migrar para microservicos

## Graceful Shutdown

O servidor implementa graceful shutdown para garantir que:

1. Requisicoes em andamento sejam concluidas
2. Conexao MongoDB seja fechada corretamente
3. Recursos sejam liberados adequadamente

```go
// Captura sinais SIGINT e SIGTERM
quit := make(chan os.Signal, 1)
signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

// Aguarda sinal
<-quit

// Shutdown gracioso do Fiber
app.Shutdown()

// Fecha conexao MongoDB
database.Disconnect()
```

## Tratamento de Erros

### Erros de Dominio (Service Layer)
```go
var (
    ErrCompanyNotFound = errors.New("company not found")
    ErrUnauthorized = errors.New("unauthorized access")
    ErrInvalidInput = errors.New("invalid input")
)
```

### Erros de Repositorio (Repository Layer)
```go
type RepositoryError struct {
    Operation string
    Err       error
}
```

### Erros de Validacao (Helpers)
```go
type ValidationError struct {
    Field   string
    Message string
}
```

## Seguranca

### Validacoes Implementadas
- XSS Prevention (ValidateNoScriptTags)
- SQL Injection Prevention (ValidateSQLInjection)
- NoSQL Injection Prevention (ValidateMongoInjection)
- Input Sanitization (SanitizeString)

### Ownership Validation
- Toda operacao valida que o usuario e dono do recurso
- Implementado na camada de Service
- Retorna ErrUnauthorized se falhar

## Proximos Passos

1. **Implementar testes unitarios**
   - Testar services com mocks de repositories
   - Testar repositories com banco in-memory
   - Testar handlers com Fiber Test

2. **Adicionar mais services**
   - TransactionService
   - CategoryService
   - RecurringService

3. **Implementar cache**
   - Redis na camada de repository
   - Cache de queries frequentes

4. **Adicionar observabilidade**
   - Logging estruturado
   - Metricas (Prometheus)
   - Tracing (OpenTelemetry)

5. **Melhorias de performance**
   - Connection pooling
   - Batch operations
   - Query optimization
