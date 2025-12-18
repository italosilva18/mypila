# Service Layer Pattern - Diagrama Visual

## Estrutura de Pastas

```
D:\Sexto\backend\
├── handlers/              # HTTP Layer - Gerencia requisicoes
│   └── company.go         # Company HTTP handlers
│
├── services/              # Service Layer - Logica de negocio
│   ├── company_service.go        # Company business logic
│   └── company_service_test.go   # Service tests
│
├── repositories/          # Data Layer - Acesso ao banco
│   ├── base.go                   # Base repository interface
│   └── company_repository.go     # Company data access
│
├── models/                # Domain Models
│   └── company.go         # Company struct and DTOs
│
├── helpers/               # Shared utilities
│   ├── validation.go      # Validation helpers
│   ├── sanitization.go    # Input sanitization
│   └── ownership.go       # Ownership validation
│
├── database/              # Database connection
│   └── mongodb.go         # MongoDB connection + Disconnect()
│
└── main.go                # Entry point + Graceful shutdown
```

## Fluxo de Requisicao: POST /api/companies

```
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT (HTTP Request)                     │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ POST /api/companies
                             │ Body: {"name": "Empresa XYZ"}
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    HANDLER LAYER (handlers/)                     │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  handlers/company.go                                    │    │
│  │  ------------------------------------------------       │    │
│  │  func CreateCompany(c *fiber.Ctx) error {             │    │
│  │    1. Extrai userID do contexto (JWT)                 │    │
│  │    2. Parse do request body                           │    │
│  │    3. Chama service.CreateCompany()                   │    │
│  │    4. Trata erros e retorna JSON                      │    │
│  │  }                                                     │    │
│  └────────────────────────────────────────────────────────┘    │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ service.CreateCompany(userID, request)
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                   SERVICE LAYER (services/)                      │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  services/company_service.go                           │    │
│  │  ------------------------------------------------       │    │
│  │  func (s *CompanyService) CreateCompany() {           │    │
│  │    1. Valida input:                                   │    │
│  │       - ValidateRequired                              │    │
│  │       - ValidateMaxLength                             │    │
│  │       - ValidateNoScriptTags (XSS)                    │    │
│  │       - ValidateMongoInjection                        │    │
│  │       - ValidateSQLInjection                          │    │
│  │                                                        │    │
│  │    2. Sanitiza dados:                                 │    │
│  │       - SanitizeString (remove HTML/scripts)          │    │
│  │                                                        │    │
│  │    3. Cria model Company:                             │    │
│  │       - Gera ID                                       │    │
│  │       - Define userID                                 │    │
│  │       - Define timestamp                              │    │
│  │                                                        │    │
│  │    4. Chama repository.Create()                       │    │
│  │  }                                                     │    │
│  └────────────────────────────────────────────────────────┘    │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ repo.Create(ctx, company)
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                REPOSITORY LAYER (repositories/)                  │
│  ┌────────────────────────────────────────────────────────┐    │
│  │  repositories/company_repository.go                    │    │
│  │  ------------------------------------------------       │    │
│  │  func (r *CompanyRepository) Create() {               │    │
│  │    1. Garante ID e timestamp setados                  │    │
│  │    2. Executa collection.InsertOne()                  │    │
│  │    3. Retorna erro se falhar                          │    │
│  │  }                                                     │    │
│  └────────────────────────────────────────────────────────┘    │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ InsertOne(document)
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                         MONGODB DATABASE                         │
│  Collection: companies                                           │
│  {                                                               │
│    "_id": ObjectId("..."),                                       │
│    "userId": ObjectId("..."),                                    │
│    "name": "Empresa XYZ",                                        │
│    "createdAt": ISODate("...")                                   │
│  }                                                               │
└─────────────────────────────────────────────────────────────────┘
```

## Fluxo de Resposta

```
┌─────────────────────────────────────────────────────────────────┐
│                         MONGODB DATABASE                         │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ Success (nil error)
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                REPOSITORY LAYER (repositories/)                  │
│  return nil (sucesso)                                            │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ return company, nil, nil
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                   SERVICE LAYER (services/)                      │
│  return company (com ID, timestamps, dados sanitizados)          │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ return c.Status(201).JSON(company)
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                    HANDLER LAYER (handlers/)                     │
│  HTTP 201 Created                                                │
│  {                                                               │
│    "id": "...",                                                  │
│    "userId": "...",                                              │
│    "name": "Empresa XYZ",                                        │
│    "createdAt": "2025-12-16T..."                                 │
│  }                                                               │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ JSON Response
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                        CLIENT (HTTP Response)                    │
└─────────────────────────────────────────────────────────────────┘
```

## Tratamento de Erros

```
┌──────────────────────────┐
│  Validation Error        │
│  (Empty name)            │
└────────┬─────────────────┘
         │
         ▼
┌──────────────────────────┐      ┌──────────────────────────┐
│  SERVICE LAYER           │      │  HTTP Response           │
│  Retorna:                │─────▶│  400 Bad Request         │
│  - nil                   │      │  {                       │
│  - validationErrors      │      │    "errors": [           │
│  - ErrInvalidInput       │      │      {                   │
└──────────────────────────┘      │        "field": "name",  │
                                  │        "message": "..."  │
                                  │      }                   │
                                  │    ]                     │
                                  │  }                       │
                                  └──────────────────────────┘

┌──────────────────────────┐
│  Company Not Found       │
└────────┬─────────────────┘
         │
         ▼
┌──────────────────────────┐      ┌──────────────────────────┐
│  SERVICE LAYER           │      │  HTTP Response           │
│  Retorna:                │─────▶│  404 Not Found           │
│  - nil                   │      │  {                       │
│  - nil                   │      │    "error": "Company     │
│  - ErrCompanyNotFound    │      │     not found"           │
└──────────────────────────┘      │  }                       │
                                  └──────────────────────────┘

┌──────────────────────────┐
│  Unauthorized Access     │
│  (Not owner)             │
└────────┬─────────────────┘
         │
         ▼
┌──────────────────────────┐      ┌──────────────────────────┐
│  SERVICE LAYER           │      │  HTTP Response           │
│  ValidateOwnership()     │─────▶│  403 Forbidden           │
│  Retorna nil (not owned) │      │  {                       │
└──────────────────────────┘      │    "error": "Forbidden", │
                                  │    "details": "..."      │
                                  │  }                       │
                                  └──────────────────────────┘
```

## Graceful Shutdown

```
┌─────────────────────────────────────────────────────────────────┐
│                         MAIN.GO                                  │
│                                                                  │
│  1. Setup signal handler:                                       │
│     quit := make(chan os.Signal, 1)                             │
│     signal.Notify(quit, os.Interrupt, syscall.SIGTERM)          │
│                                                                  │
│  2. Start server em goroutine:                                  │
│     go func() { app.Listen(":8080") }()                         │
│                                                                  │
│  3. Wait for signal:                                            │
│     <-quit                                                       │
│                                                                  │
│  4. Graceful shutdown:                                          │
│     a) app.Shutdown()        // Fecha servidor HTTP             │
│     b) database.Disconnect()  // Fecha conexao MongoDB          │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘

Sequencia de Shutdown:

SIGTERM/SIGINT received
         │
         ▼
┌────────────────────┐
│  Stop accepting    │
│  new requests      │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│  Wait for active   │
│  requests to finish│
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│  Shutdown Fiber    │
│  app.Shutdown()    │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│  Close MongoDB     │
│  connection        │
│  database.         │
│  Disconnect()      │
└─────────┬──────────┘
          │
          ▼
┌────────────────────┐
│  Log shutdown      │
│  complete          │
└────────────────────┘
```

## Validacoes de Seguranca

```
┌─────────────────────────────────────────────────────────────────┐
│                    INPUT: "Company XYZ"                          │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      VALIDATION CHAIN                            │
│                                                                  │
│  1. ValidateRequired         ✓ Not empty                        │
│  2. ValidateMaxLength        ✓ <= 100 chars                     │
│  3. ValidateNoScriptTags     ✓ No <script> tags                 │
│  4. ValidateMongoInjection   ✓ No $ne, $gt operators            │
│  5. ValidateSQLInjection     ✓ No SELECT, DROP, etc             │
│                                                                  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                        SANITIZATION                              │
│                                                                  │
│  SanitizeString():                                              │
│  - Remove HTML tags                                             │
│  - Escape dangerous characters                                  │
│  - Trim whitespace                                              │
│  - Normalize whitespace                                         │
│                                                                  │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                   SAFE TO STORE: "Company XYZ"                   │
└─────────────────────────────────────────────────────────────────┘
```

## Beneficios da Arquitetura

```
┌─────────────────────────────────────────────────────────────────┐
│                      SEPARATION OF CONCERNS                      │
│                                                                  │
│  Handler  ──▶  HTTP only     (routing, status codes)            │
│  Service  ──▶  Business only (validation, logic)                │
│  Repository ─▶ Data only     (queries, persistence)             │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                         TESTABILITY                              │
│                                                                  │
│  Unit Tests:     Service logic (with mock repositories)         │
│  Integration:    Repository queries (with test database)        │
│  E2E Tests:      Handlers (with test server)                    │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                       MAINTAINABILITY                            │
│                                                                  │
│  Change DB?       ──▶  Only Repository layer                    │
│  Change Logic?    ──▶  Only Service layer                       │
│  Change API?      ──▶  Only Handler layer                       │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                        SCALABILITY                               │
│                                                                  │
│  Add Cache?       ──▶  Repository layer (transparent)           │
│  Add Queue?       ──▶  Service layer (async operations)         │
│  Add Microservice? ─▶ Extract Service + Repository              │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```
