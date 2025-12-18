# M2M Salary Manager

Sistema de gestao financeira pessoal desenvolvido com Go (Fiber) e React (TypeScript).

## Stack Tecnologica

### Backend
- **Go 1.24** - Linguagem de programacao
- **Fiber v2** - Framework web de alta performance
- **MongoDB** - Banco de dados NoSQL
- **JWT** - Autenticacao
- **bluemonday** - Sanitizacao HTML/XSS

### Frontend
- **React 19.2** - Biblioteca UI
- **TypeScript 5.8** - Tipagem estatica
- **Vite 6.2** - Build tool
- **Tailwind CSS 3.4** - Framework CSS
- **Recharts** - Graficos
- **Lucide React** - Icones

### Infraestrutura
- **Docker** - Containerizacao
- **Docker Compose** - Orquestracao
- **GitHub Actions** - CI/CD
- **NGINX** - Reverse proxy (production)

## Inicio Rapido

### Pre-requisitos
- Docker e Docker Compose instalados
- Git
- Make (opcional, mas recomendado)

### Instalacao

```bash
# Clonar repositorio
git clone <repo-url>
cd Sexto

# Usando Makefile (recomendado)
make run

# Ou usando Docker Compose diretamente
docker-compose up -d

# Verificar status
make docker-ps
# ou
docker-compose ps

# Verificar saude dos servicos
make check-health
```

### Acessos

| Servico | URL |
|---------|-----|
| Frontend | http://localhost:3333 |
| Backend API | http://localhost:8081/api |
| MongoDB | localhost:27018 |

## Funcionalidades

- Autenticacao (registro/login)
- Gerenciamento de empresas (multi-tenancy)
- Controle de transacoes financeiras
- Categorias personalizadas com cores
- Transacoes recorrentes
- Dashboard com graficos
- Filtros por mes/ano
- Exportacao de dados
- Interface responsiva (mobile-first)

## Estrutura de Pastas

```
.
├── backend/
│   ├── config/          # Configuracoes (JWT)
│   ├── database/        # Conexao MongoDB
│   ├── handlers/        # Controllers
│   ├── helpers/         # Validacao, sanitizacao
│   ├── middleware/      # Auth middleware
│   ├── models/          # Entidades
│   └── main.go
│
├── frontend/
│   ├── components/      # Componentes React
│   ├── contexts/        # State management
│   ├── hooks/           # Custom hooks
│   ├── pages/           # Paginas
│   ├── services/        # API client
│   ├── utils/           # Utilitarios
│   └── types.ts         # Tipos TypeScript
│
└── docker-compose.yml
```

## API Endpoints

### Autenticacao
- `POST /api/auth/register` - Registro
- `POST /api/auth/login` - Login
- `GET /api/auth/me` - Usuario atual

### Empresas
- `GET /api/companies` - Listar
- `POST /api/companies` - Criar
- `PUT /api/companies/:id` - Atualizar
- `DELETE /api/companies/:id` - Deletar

### Transacoes
- `GET /api/transactions?companyId=` - Listar
- `POST /api/transactions` - Criar
- `PUT /api/transactions/:id` - Atualizar
- `DELETE /api/transactions/:id` - Deletar
- `PATCH /api/transactions/:id/toggle-status` - Alternar status

### Categorias
- `GET /api/categories?companyId=` - Listar
- `POST /api/categories?companyId=` - Criar
- `PUT /api/categories/:id` - Atualizar
- `DELETE /api/categories/:id` - Deletar

### Transacoes Recorrentes
- `GET /api/recurring?companyId=` - Listar
- `POST /api/recurring` - Criar
- `DELETE /api/recurring/:id` - Deletar
- `POST /api/recurring/process` - Processar mes

### Outros
- `GET /api/stats?companyId=` - Estatisticas
- `GET /health` - Health check

## Desenvolvimento

### Comandos Rapidos (Makefile)

```bash
make help              # Listar todos os comandos
make install           # Instalar dependencias
make build             # Compilar backend e frontend
make test              # Executar testes
make lint              # Executar linters
make format            # Formatar codigo

# Docker
make docker-build      # Build das imagens
make docker-up         # Subir servicos
make docker-down       # Parar servicos
make docker-logs       # Ver logs
make docker-restart    # Reiniciar servicos

# Desenvolvimento
make dev-backend       # Executar backend em modo dev
make dev-frontend      # Executar frontend em modo dev

# Operacoes
make backup            # Backup do MongoDB
make restore           # Restaurar backup
make clean             # Limpar artifacts
make ci                # Executar CI localmente
```

### Backend
```bash
cd backend
go run main.go
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

### Rebuild Docker
```bash
# Usando Makefile
make docker-build

# Ou manualmente
docker-compose build --no-cache
docker-compose up -d
```

## Testes

### Backend
```bash
cd backend
go test ./... -v
go test -cover ./...
```

## Variaveis de Ambiente

### Backend (.env)
```env
PORT=8080
MONGODB_URI=mongodb://admin:admin123@mongodb:27017
MONGODB_DATABASE=m2m_financeiro
JWT_SECRET=your-secret-key
ENVIRONMENT=development
```

### Frontend (.env)
```env
VITE_API_URL=http://localhost:8081/api
```

## Documentacao

- [Status do Projeto](PROJECT_STATUS.md) - Auditoria completa
- [Roadmap de Melhorias](IMPROVEMENT_ROADMAP.md) - Proximos passos
- [DevOps Guide](DEVOPS.md) - CI/CD, Docker, deployment
- [Seguranca](backend/SECURITY.md) - Detalhes de seguranca
- [API OpenAPI](backend/docs/openapi.yaml) - Especificacao da API

## Seguranca Implementada

- Validacao de inputs (XSS, SQL/NoSQL injection)
- Sanitizacao com bluemonday
- Ownership validation (IDOR prevention)
- Rate limiting (100 req/min global, 20 req/min auth)
- Security headers (X-Frame-Options, CSP, etc.)
- Bcrypt para senhas

## CI/CD Pipeline

O projeto possui pipelines automatizados com GitHub Actions:

### CI Pipeline
- Testes automatizados (backend e frontend)
- Analise estatica de codigo (go vet, staticcheck)
- Verificacao de formatacao
- Build de imagens Docker
- Scan de vulnerabilidades (Trivy)
- Lint de Dockerfiles (hadolint)

### Deploy Pipeline
- Build multi-arch (amd64/arm64)
- Push para GitHub Container Registry
- Deploy automatico para staging
- Deploy manual para producao
- Rollback automatico em falhas

Ver [DEVOPS.md](DEVOPS.md) para detalhes completos.

## Proximos Passos

Ver [IMPROVEMENT_ROADMAP.md](IMPROVEMENT_ROADMAP.md) para lista completa.

Prioridades imediatas:
1. Implementar CSRF protection
2. Criar indices MongoDB
3. Implementar paginacao
4. Adicionar testes de integracao

## Licenca

MIT

---

*Ultima atualizacao: 2025-12-16*
