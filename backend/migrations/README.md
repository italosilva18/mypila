# Database Migrations and Indexes

Este diretório contém scripts de migração e configurações de índices para o MongoDB.

---

## Índices MongoDB

### Índices Criados Automaticamente

Os índices são criados automaticamente ao iniciar o servidor através do arquivo `indexes.go`.

#### Collection: users
- **email_unique_idx**: Índice único em `email`
  - Garante unicidade de emails
  - Otimiza queries de autenticação
  - Performance: O(1) para lookup por email

#### Collection: companies
- **userId_idx**: Índice em `userId`
  - Otimiza queries de empresas por usuário
  - Performance: O(log n) para lookup por userId

#### Collection: transactions
- **companyId_idx**: Índice em `companyId`
- **companyId_year_month_idx**: Índice composto (companyId + year DESC + month)
- **status_idx**: Índice em `status`
- **companyId_status_idx**: Índice composto (companyId + status)

#### Collection: categories
- **companyId_idx**: Índice em `companyId`
- **companyId_type_idx**: Índice composto (companyId + type)

#### Collection: recurring
- **companyId_idx**: Índice em `companyId`
- **companyId_dayOfMonth_idx**: Índice composto (companyId + dayOfMonth)

### Paginação de Transações

**Endpoint:** `GET /api/transactions`

**Query Parameters:**
- `page` (default: 1): Número da página
- `limit` (default: 50, max: 100): Itens por página
- `companyId` (opcional): Filtrar por empresa

**Exemplo de Response:**
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 50,
    "total": 150,
    "totalPages": 3
  }
}
```

### Verificar Índices

```javascript
// Ver índices criados
db.transactions.getIndexes()

// Verificar uso de índices
db.transactions.find({companyId: ObjectId("...")}).explain("executionStats")
```

---

## Migrações

## Scripts Disponíveis

### add_userid_to_companies.js

Adiciona o campo `userId` a todas as empresas existentes no banco de dados.

**Quando usar:**
- Ao atualizar de uma versão anterior sem ownership validation
- Quando empresas existentes não têm o campo `userId`

**Como executar:**

#### Opção 1: Via mongosh (recomendado)

```bash
mongosh mongodb://localhost:27017/m2m --file migrations/add_userid_to_companies.js
```

#### Opção 2: Dentro do mongosh

```javascript
use m2m
load("migrations/add_userid_to_companies.js")
```

#### Opção 3: Copiar e colar no mongosh

Abra o arquivo, copie o conteúdo e cole no shell do MongoDB.

**O que o script faz:**

1. Verifica se existem usuários no banco
2. Conta quantas empresas não têm `userId`
3. Atribui o primeiro usuário encontrado como dono de todas as empresas sem dono
4. Exibe estatísticas da migração
5. Sugere próximos passos

**Importante:**

Se você precisa atribuir empresas específicas a usuários específicos, edite o script ou execute manualmente:

```javascript
// Atribuir empresa específica a usuário específico
db.companies.updateOne(
  { name: "Nome da Empresa" },
  { $set: { userId: ObjectId("id_do_usuario") } }
)
```

## Criar Índices

Após a migração, é importante criar índices para performance:

```javascript
// Conecte ao MongoDB
mongosh mongodb://localhost:27017/m2m

// Execute os comandos
use m2m

// Índice para companies.userId
db.companies.createIndex({ userId: 1 })

// Verificar índices criados
db.companies.getIndexes()
```

## Verificação Pós-Migração

### Verificar se todas as empresas têm userId

```javascript
db.companies.countDocuments({ userId: { $exists: false } })
// Deve retornar 0
```

### Listar empresas por usuário

```javascript
db.companies.aggregate([
  { $group: { _id: "$userId", companies: { $push: "$name" } } },
  { $lookup: { from: "users", localField: "_id", foreignField: "_id", as: "user" } },
  { $unwind: "$user" },
  { $project: { userName: "$user.name", companies: 1 } }
])
```

### Verificar integridade dos dados

```javascript
// Verificar se todos os userIds são válidos
db.companies.aggregate([
  {
    $lookup: {
      from: "users",
      localField: "userId",
      foreignField: "_id",
      as: "user"
    }
  },
  { $match: { user: { $size: 0 } } },
  { $project: { name: 1, userId: 1 } }
])
// Não deve retornar resultados (empresas com userId inválido)
```

## Rollback

Se precisar reverter a migração:

```javascript
// ATENÇÃO: Isso remove o campo userId de TODAS as empresas
db.companies.updateMany(
  {},
  { $unset: { userId: "" } }
)
```

## Backup Recomendado

Antes de executar qualquer migração, faça backup do banco de dados:

```bash
# Backup
mongodump --db=m2m --out=backup_$(date +%Y%m%d_%H%M%S)

# Restore (se necessário)
mongorestore --db=m2m backup_YYYYMMDD_HHMMSS/m2m
```

## Troubleshooting

### Erro: "No users found in database"

**Solução:** Crie pelo menos um usuário antes de executar a migração:

```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Admin",
    "email": "admin@example.com",
    "password": "senha123"
  }'
```

### Empresas ainda sem userId após migração

**Diagnóstico:**
```javascript
db.companies.find({ userId: { $exists: false } })
```

**Solução:**
Execute a migração novamente ou atribua manualmente.

### Performance lenta após migração

**Diagnóstico:**
```javascript
db.companies.getIndexes()
```

**Solução:**
Certifique-se de que o índice foi criado:
```javascript
db.companies.createIndex({ userId: 1 })
```

## Referências

- [MongoDB Migration Best Practices](https://www.mongodb.com/blog/post/6-rules-of-thumb-for-mongodb-schema-design)
- [mongosh Documentation](https://www.mongodb.com/docs/mongodb-shell/)
