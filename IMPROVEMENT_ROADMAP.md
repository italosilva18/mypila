# ROADMAP DE MELHORIAS - M2M SALARY MANAGER

**Criado em:** 2025-12-16
**Baseado em:** Auditoria completa de 5 agentes especializados

---

## PRIORIDADE CRITICA (Fazer Imediatamente)

### SEC-001: Implementar CSRF Protection
**Arquivo:** `backend/main.go`
```go
// Adicionar apos imports
import "github.com/gofiber/fiber/v2/middleware/csrf"

// Adicionar middleware
app.Use(csrf.New(csrf.Config{
    KeyLookup:      "header:X-CSRF-Token",
    CookieName:     "csrf_",
    CookieSameSite: "Strict",
    Expiration:     1 * time.Hour,
}))
```

### SEC-002: Restringir CORS
**Arquivo:** `backend/main.go` linha 65
```go
// DE:
AllowOrigins: "*"

// PARA:
AllowOrigins: os.Getenv("ALLOWED_ORIGINS") // "http://localhost:3333"
```

### SEC-003: Migrar JWT para httpOnly Cookies
**Arquivos:** `backend/handlers/auth.go`, `frontend/services/api.ts`
- Backend: Retornar token em cookie httpOnly
- Frontend: Remover localStorage, usar credentials: 'include'

### DB-001: Criar Indices MongoDB
**Arquivo:** Criar `backend/migrations/indexes.go`
```go
func CreateIndexes(db *mongo.Database) error {
    // users: email (unique)
    // companies: userId
    // transactions: companyId, (companyId + year + month)
    // categories: companyId
    // recurring: companyId
}
```

### DB-002: Implementar Paginacao
**Arquivos:** Todos os handlers de listagem
```go
page := c.QueryInt("page", 1)
limit := c.QueryInt("limit", 50)
// Aplicar SetSkip e SetLimit nas queries
```

---

## PRIORIDADE ALTA (Proximas 2 Semanas)

### ARCH-001: Padronizar CompanyID para ObjectID
**Arquivos:**
- `backend/models/category.go` - Mudar `CompanyID string` para `primitive.ObjectID`
- `backend/models/recurring.go` - Mudar `CompanyID string` para `primitive.ObjectID`

### ARCH-002: Criar Service Layer
**Criar:** `backend/services/`
```
services/
├── auth_service.go
├── company_service.go
├── transaction_service.go
└── category_service.go
```

### ARCH-003: Implementar Repository Pattern
**Criar:** `backend/repositories/`
```
repositories/
├── base_repository.go
├── user_repository.go
├── company_repository.go
└── transaction_repository.go
```

### TEST-001: Testes de Ownership
**Criar:** `backend/helpers/ownership_test.go`
- Testar ValidateCompanyOwnership
- Testar ValidateTransactionOwnership
- Testar acesso nao autorizado

### TEST-002: Testes de Auth
**Criar:** `backend/handlers/auth_test.go`
- Testar registro
- Testar login
- Testar token invalido

### PERF-001: Implementar Cache
**Criar:** `backend/cache/redis.go`
- Cache de companies (5 min)
- Cache de categories (5 min)
- Cache de stats (1 min)

---

## PRIORIDADE MEDIA (Proximo Mes)

### FE-001: Setup de Testes Frontend
**Arquivo:** `frontend/package.json`
```json
{
  "devDependencies": {
    "vitest": "^1.0.4",
    "@testing-library/react": "^14.1.2",
    "@testing-library/jest-dom": "^6.1.5"
  },
  "scripts": {
    "test": "vitest",
    "test:coverage": "vitest --coverage"
  }
}
```

### FE-002: Memoizar Componentes
**Arquivos:** Dashboard.tsx, FinancialChart.tsx
```typescript
// Usar useMemo para calculos
const filteredTransactions = useMemo(() => {...}, [deps]);

// Usar useCallback para funcoes
const handleClick = useCallback(() => {...}, [deps]);

// Usar React.memo para componentes
export const Dashboard = memo(() => {...});
```

### FE-003: Habilitar Strict Mode TypeScript
**Arquivo:** `frontend/tsconfig.json`
```json
{
  "compilerOptions": {
    "strict": true,
    "noImplicitAny": true,
    "strictNullChecks": true,
    "noUnusedLocals": true
  }
}
```

### FE-004: Adicionar ESLint
```bash
npm install -D eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin
```

### FE-005: Melhorar Acessibilidade
- Adicionar aria-labels em botoes de icone
- Implementar focus trap em modais
- Adicionar role="dialog" em modais
- Suportar navegacao por teclado (ESC fecha modal)

### BE-001: Graceful Shutdown
**Arquivo:** `backend/main.go`
```go
c := make(chan os.Signal, 1)
signal.Notify(c, os.Interrupt, syscall.SIGTERM)
go func() {
    <-c
    database.DB.Client().Disconnect(ctx)
    os.Exit(0)
}()
```

### BE-002: Health Check Detalhado
**Arquivo:** `backend/main.go`
```go
app.Get("/health", func(c *fiber.Ctx) error {
    if err := database.DB.Client().Ping(ctx, nil); err != nil {
        return c.Status(503).JSON(...)
    }
    return c.JSON(fiber.Map{"status": "healthy"})
})
```

### BE-003: Connection Pool MongoDB
**Arquivo:** `backend/database/mongodb.go`
```go
clientOptions := options.Client().
    ApplyURI(mongoURI).
    SetMaxPoolSize(100).
    SetMinPoolSize(10).
    SetMaxConnIdleTime(30 * time.Second)
```

---

## PRIORIDADE BAIXA (Backlog)

### FEAT-001: Refresh Token System
- Criar model RefreshToken
- Endpoint POST /auth/refresh
- Automatizar renovacao no frontend

### FEAT-002: Logout com Blacklist
- Implementar blacklist de tokens (Redis)
- Endpoint POST /auth/logout
- Verificar blacklist no middleware

### FEAT-003: Password Policy
```go
func ValidatePasswordStrength(password string) *ValidationError {
    // Minimo 8 caracteres
    // Maiusculas, minusculas, numeros, especiais
}
```

### FEAT-004: Rate Limiting Distribuido
- Usar Redis para contadores compartilhados
- Permitir multiplas instancias do backend

### FEAT-005: API Versioning
```go
v1 := app.Group("/api/v1")
v1.Get("/transactions", handlers.GetAllTransactions)
```

### FEAT-006: Code Splitting Frontend
```typescript
const Dashboard = lazy(() => import('./components/Dashboard'));
const Categories = lazy(() => import('./pages/Categories'));
```

### FEAT-007: React Query para Cache
```typescript
const { data, isLoading } = useQuery({
    queryKey: ['transactions', companyId],
    queryFn: () => api.getTransactions(companyId),
    staleTime: 5000,
});
```

### INFRA-001: CI/CD Pipeline
```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: go test -v ./...
      - run: npm test
```

### DOCS-001: OpenAPI/Swagger Completo
- Documentar todos os endpoints
- Exemplos de request/response
- Codigos de erro

### MONITOR-001: Observabilidade
- Logging estruturado (zerolog)
- Metricas (Prometheus)
- Tracing (OpenTelemetry)

---

## CHECKLIST DE CONCLUSAO

### Seguranca
- [ ] CSRF protection implementado
- [ ] CORS restrito
- [ ] JWT em httpOnly cookies
- [ ] Rate limiting por email
- [ ] Password policy forte

### Performance
- [ ] Indices MongoDB criados
- [ ] Paginacao implementada
- [ ] Cache Redis configurado
- [ ] Componentes memoizados
- [ ] Code splitting

### Qualidade
- [ ] Testes backend > 80%
- [ ] Testes frontend > 70%
- [ ] ESLint configurado
- [ ] TypeScript strict mode
- [ ] Error boundaries

### Arquitetura
- [ ] Service Layer criado
- [ ] Repository Pattern implementado
- [ ] Tipos padronizados
- [ ] DTOs separados

---

## ESTIMATIVAS

| Fase | Duracao | Esforco |
|------|---------|---------|
| Critica | 1 semana | Alto |
| Alta | 2 semanas | Alto |
| Media | 4 semanas | Medio |
| Baixa | Continuo | Baixo |

**Total estimado para producao-ready:** 8-10 semanas

---

*Roadmap gerado em 2025-12-16*
