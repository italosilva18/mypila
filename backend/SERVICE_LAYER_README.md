# Service Layer Pattern - Guia de Uso

## Visao Rapida

Este backend implementa o **Service Layer Pattern** com as seguintes camadas:

- **Handler** → Gerencia HTTP (requests/responses)
- **Service** → Logica de negocio (validacao, regras)
- **Repository** → Acesso a dados (MongoDB)

## Estrutura de Arquivos

```
backend/
├── handlers/company.go              # HTTP handlers
├── services/company_service.go      # Business logic
├── repositories/company_repository.go  # Data access
├── models/company.go                # Domain models
└── main.go                          # Server + Graceful shutdown
```

## Como Usar

### 1. Iniciar o Servidor

```bash
cd D:\Sexto\backend
go run main.go
```

### 2. Compilar

```bash
go build -o bin/server.exe .
./bin/server.exe
```

### 3. Testar

```bash
go test -v ./services/...
```

### 4. Graceful Shutdown

Pressione `Ctrl+C` ou envie `SIGTERM`:

```bash
# Windows
taskkill /F /IM server.exe

# Linux/Mac
kill -SIGTERM <pid>
```

O servidor ira:
1. Parar de aceitar novas requisicoes
2. Completar requisicoes ativas
3. Fechar conexao MongoDB
4. Encerrar graciosamente

## Exemplos de Uso

### Criar uma Nova Company

```http
POST /api/companies
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Minha Empresa"
}
```

**Fluxo Interno:**
```
Handler → Service → Repository → MongoDB
   ↓         ↓          ↓           ↓
Parse    Validate   Insert      Save
Request  Sanitize   Document    Data
```

### Atualizar Company

```http
PUT /api/companies/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Empresa Atualizada"
}
```

**Fluxo Interno:**
```
Handler → Service → Repository → MongoDB
   ↓         ↓          ↓           ↓
Parse    Validate   Update      Modify
Request  Ownership  Document    Data
         Sanitize
```

### Deletar Company (Cascade)

```http
DELETE /api/companies/:id
Authorization: Bearer <token>
```

**Fluxo Interno:**
```
Handler → Service → Repository → MongoDB
   ↓         ↓          ↓           ↓
Parse    Validate   Delete      Remove:
ID       Ownership  Related     - Company
                    Data        - Transactions
                                - Categories
                                - Recurring
```

## Adicionar um Novo Dominio

### Exemplo: Product Service

#### 1. Criar Model

```go
// models/product.go
package models

type Product struct {
    ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    CompanyID primitive.ObjectID `json:"companyId" bson:"companyId"`
    Name      string             `json:"name" bson:"name"`
    Price     float64            `json:"price" bson:"price"`
    CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}
```

#### 2. Criar Repository

```go
// repositories/product_repository.go
package repositories

type ProductRepository struct {
    collection *mongo.Collection
}

func NewProductRepository() *ProductRepository {
    return &ProductRepository{
        collection: database.GetCollection("products"),
    }
}

func (r *ProductRepository) Create(ctx context.Context, product *models.Product) error {
    // Implementation
}
```

#### 3. Criar Service

```go
// services/product_service.go
package services

type ProductService struct {
    repo *repositories.ProductRepository
}

func NewProductService() *ProductService {
    return &ProductService{
        repo: repositories.NewProductRepository(),
    }
}

func (s *ProductService) CreateProduct(companyID primitive.ObjectID, req CreateProductRequest) (*models.Product, []helpers.ValidationError, error) {
    // Validate
    validationErrors := helpers.CollectErrors(
        helpers.ValidateRequired(req.Name, "name"),
        helpers.ValidatePositiveNumber(req.Price, "price"),
    )

    if helpers.HasErrors(validationErrors) {
        return nil, validationErrors, ErrInvalidInput
    }

    // Sanitize
    sanitizedName := helpers.SanitizeString(req.Name)

    // Create
    product := &models.Product{
        ID:        primitive.NewObjectID(),
        CompanyID: companyID,
        Name:      sanitizedName,
        Price:     req.Price,
        CreatedAt: time.Now(),
    }

    // Save
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    err := s.repo.Create(ctx, product)
    if err != nil {
        return nil, nil, err
    }

    return product, nil, nil
}
```

#### 4. Criar Handler

```go
// handlers/product.go
package handlers

var productService = services.NewProductService()

func CreateProduct(c *fiber.Ctx) error {
    // Get user ID
    userID, err := helpers.GetUserIDFromContext(c)
    if err != nil {
        return err
    }

    // Parse company ID
    companyID, err := primitive.ObjectIDFromHex(c.Params("companyId"))
    if err != nil {
        return helpers.SendValidationError(c, "companyId", "Invalid company ID")
    }

    // Parse request
    var req CreateProductRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
    }

    // Call service
    product, validationErrors, err := productService.CreateProduct(companyID, req)
    if err != nil {
        if err == services.ErrInvalidInput && validationErrors != nil {
            return helpers.SendValidationErrors(c, validationErrors)
        }
        return c.Status(500).JSON(fiber.Map{"error": "Failed to create product"})
    }

    return c.Status(201).JSON(product)
}
```

#### 5. Registrar Rotas

```go
// main.go
products := api.Group("/products")
products.Post("/", handlers.CreateProduct)
products.Get("/", handlers.GetProducts)
products.Put("/:id", handlers.UpdateProduct)
products.Delete("/:id", handlers.DeleteProduct)
```

## Validacoes Disponiveis

### helpers/validation.go

```go
ValidateRequired(value, field)        // Campo obrigatorio
ValidateMaxLength(value, field, max)  // Tamanho maximo
ValidateMinLength(value, field, min)  // Tamanho minimo
ValidatePositiveNumber(value, field)  // Numero positivo
ValidateEmail(email)                  // Formato de email
ValidateHexColor(color)               // Cor hexadecimal
ValidateMonth(month)                  // Mes em portugues
ValidateStatus(status)                // PAGO ou ABERTO
```

### helpers/sanitization.go

```go
SanitizeString(input)        // Remove HTML/scripts
SanitizeHTML(input)          // Permite apenas tags seguras
EscapeHTML(input)            // Escapa caracteres especiais
SanitizeAlphanumeric(input)  // Apenas alfanumericos

ValidateNoScriptTags(value, field)     // Previne XSS
ValidateSQLInjection(value, field)     // Previne SQL injection
ValidateMongoInjection(value, field)   // Previne NoSQL injection
ValidatePathTraversal(value, field)    // Previne path traversal
```

## Tratamento de Erros

### Erros de Service

```go
services.ErrCompanyNotFound   // 404 Not Found
services.ErrUnauthorized      // 403 Forbidden
services.ErrInvalidInput      // 400 Bad Request
```

### Exemplo no Handler

```go
company, validationErrors, err := companyService.CreateCompany(userID, req)

if err != nil {
    switch err {
    case services.ErrInvalidInput:
        if validationErrors != nil {
            return helpers.SendValidationErrors(c, validationErrors)
        }
        return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})

    case services.ErrCompanyNotFound:
        return c.Status(404).JSON(fiber.Map{"error": "Company not found"})

    case services.ErrUnauthorized:
        return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})

    default:
        return c.Status(500).JSON(fiber.Map{"error": "Internal server error"})
    }
}

return c.Status(201).JSON(company)
```

## Boas Praticas

### 1. Handler Layer
- ✅ Apenas logica HTTP
- ✅ Parse de requests
- ✅ Validacao de parametros de URL
- ✅ Retorno de status codes
- ❌ NAO fazer validacao de negocio
- ❌ NAO acessar database diretamente

### 2. Service Layer
- ✅ Validacao de regras de negocio
- ✅ Sanitizacao de inputs
- ✅ Orquestracao de operacoes
- ✅ Validacao de ownership
- ❌ NAO fazer parse de HTTP
- ❌ NAO retornar status codes HTTP

### 3. Repository Layer
- ✅ Queries ao banco de dados
- ✅ Operacoes CRUD
- ✅ Tratamento de erros do DB
- ❌ NAO fazer validacao de negocio
- ❌ NAO fazer sanitizacao

## Performance

### Timeouts Configurados

```go
// Handler: 10 segundos
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

// Repository: 5 segundos (ownership validation)
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

// Delete cascade: 15 segundos
ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
```

### Otimizacoes Futuras

1. **Connection Pooling**
   - MongoDB driver ja faz pooling automatico
   - Configure maxPoolSize no connection string

2. **Indices**
   - migrations/indexes.go (ja implementado)
   - Indices em userId, companyId

3. **Caching**
   - Adicionar Redis no Repository Layer
   - Cache de queries frequentes

4. **Rate Limiting**
   - Ja implementado no main.go
   - 100 req/min global
   - 20 req/min para auth

## Seguranca

### HTTPS (Producao)

Descomente no main.go:
```go
c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
```

### CORS

Configure origins permitidas:
```go
app.Use(cors.New(cors.Config{
    AllowOrigins: "https://yourdomain.com",  // Em producao
    AllowHeaders: "Origin, Content-Type, Accept, Authorization",
    AllowMethods: "GET, POST, PUT, PATCH, DELETE, OPTIONS",
}))
```

### Environment Variables

```bash
# .env
PORT=8080
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=m2m_financeiro
JWT_SECRET=your-secret-key-here
```

## Troubleshooting

### Erro: "invalid memory address or nil pointer dereference"
**Causa:** MongoDB nao conectado
**Solucao:** Certifique-se que MongoDB esta rodando

### Erro: "company not found" ao deletar
**Causa:** Usuario nao e dono da company
**Solucao:** Service valida ownership antes de deletar

### Erro: "validation errors"
**Causa:** Input invalido
**Solucao:** Verifique os campos do request

## Monitoramento

### Logs Estruturados

```go
log.Printf("[INFO] Company created: %s", company.ID.Hex())
log.Printf("[ERROR] Failed to create company: %v", err)
log.Printf("[WARN] Invalid ownership attempt: user=%s company=%s", userID, companyID)
```

### Metricas (Futuro)

```go
// services/company_service.go
func (s *CompanyService) CreateCompany(...) {
    metrics.IncrementCounter("company.create.attempts")
    defer metrics.RecordDuration("company.create.duration", time.Now())

    company, err := s.repo.Create(...)
    if err != nil {
        metrics.IncrementCounter("company.create.errors")
        return nil, err
    }

    metrics.IncrementCounter("company.create.success")
    return company, nil
}
```

## Documentacao Adicional

- `ARCHITECTURE.md` - Arquitetura detalhada
- `SERVICE_LAYER_DIAGRAM.md` - Diagramas visuais
- `IMPLEMENTATION_SUMMARY.md` - Resumo da implementacao

## Suporte

Para duvidas ou problemas:
1. Consulte a documentacao
2. Verifique os testes em `services/*_test.go`
3. Analise os logs do servidor
