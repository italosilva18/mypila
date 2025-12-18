# Exemplos Práticos de Validação

## Casos de Teste Reais

### 1. Transaction - Múltiplos Erros

**Request:**
```bash
POST /api/transactions
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "companyId": "67677d6e7e1f8e54e8d7ee3b",
  "amount": -500,
  "description": "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.",
  "category": "",
  "month": "January",
  "year": 1999,
  "status": "pending"
}
```

**Response (400):**
```json
{
  "errors": [
    {
      "field": "amount",
      "message": "amount deve ser maior que zero"
    },
    {
      "field": "description",
      "message": "description deve ter no máximo 200 caracteres"
    },
    {
      "field": "category",
      "message": "category não pode ser vazio"
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

---

### 2. Transaction - Sucesso

**Request:**
```bash
POST /api/transactions
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "companyId": "67677d6e7e1f8e54e8d7ee3b",
  "amount": 3500.50,
  "description": "Salário do mês",
  "category": "Salário",
  "month": "Dezembro",
  "year": 2024,
  "status": "PAGO"
}
```

**Response (201):**
```json
{
  "id": "67677d6e7e1f8e54e8d7ee3c",
  "companyId": "67677d6e7e1f8e54e8d7ee3b",
  "amount": 3500.50,
  "description": "Salário do mês",
  "category": "Salário",
  "month": "Dezembro",
  "year": 2024,
  "status": "PAGO"
}
```

---

### 3. Category - Cor Inválida

**Request:**
```bash
POST /api/categories?companyId=67677d6e7e1f8e54e8d7ee3b
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "name": "Transporte",
  "color": "red",
  "type": "EXPENSE",
  "budget": 500
}
```

**Response (400):**
```json
{
  "errors": [
    {
      "field": "color",
      "message": "Cor deve estar no formato hexadecimal #RRGGBB (ex: #FF5733)"
    }
  ]
}
```

---

### 4. Category - Sucesso

**Request:**
```bash
POST /api/categories?companyId=67677d6e7e1f8e54e8d7ee3b
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "name": "Transporte",
  "color": "#FF5733",
  "type": "EXPENSE",
  "budget": 500
}
```

**Response (201):**
```json
{
  "id": "67677d6e7e1f8e54e8d7ee3d",
  "companyId": "67677d6e7e1f8e54e8d7ee3b",
  "name": "Transporte",
  "color": "#FF5733",
  "type": "EXPENSE",
  "budget": 500,
  "createdAt": "2024-12-15T10:30:00Z"
}
```

---

### 5. Company - Nome Muito Longo

**Request:**
```bash
POST /api/companies
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "name": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
}
```

**Response (400):**
```json
{
  "errors": [
    {
      "field": "name",
      "message": "name deve ter no máximo 100 caracteres"
    }
  ]
}
```

---

### 6. RecurringTransaction - Dia Inválido

**Request:**
```bash
POST /api/recurring
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "companyId": "67677d6e7e1f8e54e8d7ee3b",
  "description": "Aluguel mensal",
  "amount": 1200,
  "category": "Moradia",
  "dayOfMonth": 35
}
```

**Response (400):**
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

---

### 7. RecurringTransaction - Sucesso

**Request:**
```bash
POST /api/recurring
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

{
  "companyId": "67677d6e7e1f8e54e8d7ee3b",
  "description": "Aluguel mensal",
  "amount": 1200,
  "category": "Moradia",
  "dayOfMonth": 5
}
```

**Response (200):**
```json
{
  "id": "67677d6e7e1f8e54e8d7ee3e",
  "companyId": "67677d6e7e1f8e54e8d7ee3b",
  "description": "Aluguel mensal",
  "amount": 1200,
  "category": "Moradia",
  "dayOfMonth": 5,
  "createdAt": "2024-12-15T10:35:00Z"
}
```

---

### 8. Auth - Registro com Email Inválido

**Request:**
```bash
POST /api/auth/register
Content-Type: application/json

{
  "name": "João Silva",
  "email": "joao@invalid",
  "password": "123"
}
```

**Response (400):**
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

---

### 9. Auth - Registro Bem-Sucedido

**Request:**
```bash
POST /api/auth/register
Content-Type: application/json

{
  "name": "João Silva",
  "email": "joao.silva@example.com",
  "password": "senha123"
}
```

**Response (201):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiI2NzY3N2Q2ZTdlMWY4ZTU0ZThkN2VlM2YiLCJlbWFpbCI6ImpvYW8uc2lsdmFAZXhhbXBsZS5jb20iLCJleHAiOjE3MzQ2MTQ0MDB9.xyz...",
  "user": {
    "id": "67677d6e7e1f8e54e8d7ee3f",
    "name": "João Silva",
    "email": "joao.silva@example.com",
    "createdAt": "2024-12-15T10:40:00Z"
  }
}
```

---

### 10. Auth - Login com Credenciais Inválidas

**Request:**
```bash
POST /api/auth/login
Content-Type: application/json

{
  "email": "joao.silva@example.com",
  "password": "senhaerrada"
}
```

**Response (401):**
```json
{
  "error": "Credenciais inválidas"
}
```

---

## Testando com cURL

### Script Completo de Teste

```bash
#!/bin/bash

# Configuração
BASE_URL="http://localhost:5000/api"

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "==================================="
echo "   TESTES DE VALIDAÇÃO - M2M"
echo "==================================="

# 1. Registrar usuário
echo -e "\n${YELLOW}[1] Registrando novo usuário...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "test123"
  }')

echo "$REGISTER_RESPONSE" | json_pp

# Extrair token
TOKEN=$(echo "$REGISTER_RESPONSE" | grep -o '"token":"[^"]*' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo -e "${RED}Erro: Não foi possível obter o token${NC}"
  exit 1
fi

echo -e "${GREEN}Token obtido com sucesso!${NC}"

# 2. Criar empresa
echo -e "\n${YELLOW}[2] Criando empresa...${NC}"
COMPANY_RESPONSE=$(curl -s -X POST "$BASE_URL/companies" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Minha Empresa Teste"
  }')

echo "$COMPANY_RESPONSE" | json_pp

# Extrair company ID
COMPANY_ID=$(echo "$COMPANY_RESPONSE" | grep -o '"id":"[^"]*' | cut -d'"' -f4)

if [ -z "$COMPANY_ID" ]; then
  echo -e "${RED}Erro: Não foi possível obter o ID da empresa${NC}"
  exit 1
fi

echo -e "${GREEN}Empresa criada: $COMPANY_ID${NC}"

# 3. Testar transaction inválida
echo -e "\n${YELLOW}[3] Testando transaction com erros...${NC}"
curl -s -X POST "$BASE_URL/transactions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"companyId\": \"$COMPANY_ID\",
    \"amount\": -100,
    \"category\": \"\",
    \"month\": \"January\",
    \"year\": 1999,
    \"status\": \"invalid\"
  }" | json_pp

# 4. Criar transaction válida
echo -e "\n${YELLOW}[4] Criando transaction válida...${NC}"
curl -s -X POST "$BASE_URL/transactions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"companyId\": \"$COMPANY_ID\",
    \"amount\": 1500.50,
    \"description\": \"Pagamento teste\",
    \"category\": \"Salário\",
    \"month\": \"Dezembro\",
    \"year\": 2024,
    \"status\": \"PAGO\"
  }" | json_pp

# 5. Testar category com cor inválida
echo -e "\n${YELLOW}[5] Testando category com cor inválida...${NC}"
curl -s -X POST "$BASE_URL/categories?companyId=$COMPANY_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Teste",
    "color": "red",
    "type": "EXPENSE"
  }' | json_pp

# 6. Criar category válida
echo -e "\n${YELLOW}[6] Criando category válida...${NC}"
curl -s -X POST "$BASE_URL/categories?companyId=$COMPANY_ID" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Transporte",
    "color": "#FF5733",
    "type": "EXPENSE",
    "budget": 500
  }' | json_pp

# 7. Testar recurring com dia inválido
echo -e "\n${YELLOW}[7] Testando recurring com dia inválido...${NC}"
curl -s -X POST "$BASE_URL/recurring" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"companyId\": \"$COMPANY_ID\",
    \"description\": \"Teste\",
    \"amount\": 100,
    \"category\": \"Teste\",
    \"dayOfMonth\": 50
  }" | json_pp

# 8. Criar recurring válido
echo -e "\n${YELLOW}[8] Criando recurring válido...${NC}"
curl -s -X POST "$BASE_URL/recurring" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"companyId\": \"$COMPANY_ID\",
    \"description\": \"Aluguel mensal\",
    \"amount\": 1200,
    \"category\": \"Moradia\",
    \"dayOfMonth\": 5
  }" | json_pp

echo -e "\n${GREEN}==================================="
echo "   TESTES CONCLUÍDOS COM SUCESSO"
echo "===================================${NC}"
```

Salve como `full-validation-test.sh` e execute:

```bash
chmod +x full-validation-test.sh
./full-validation-test.sh
```

---

## Validações por Mês

### Meses Válidos (em Português)

```json
{
  "valid_months": [
    "Janeiro",
    "Fevereiro",
    "Março",
    "Abril",
    "Maio",
    "Junho",
    "Julho",
    "Agosto",
    "Setembro",
    "Outubro",
    "Novembro",
    "Dezembro",
    "Acumulado"
  ]
}
```

### Exemplo de Erro

```bash
# Tentando usar mês em inglês
curl -X POST http://localhost:5000/api/transactions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "companyId": "...",
    "amount": 100,
    "category": "Test",
    "month": "December",
    "year": 2024,
    "status": "PAGO"
  }'
```

**Resposta:**
```json
{
  "errors": [
    {
      "field": "month",
      "message": "Mês inválido. Use mês em português (ex: Janeiro, Fevereiro, etc.)"
    }
  ]
}
```

---

## Validações de Cor Hexadecimal

### Cores Válidas

```
#000000  ✓ (Preto)
#FFFFFF  ✓ (Branco)
#FF5733  ✓ (Laranja)
#3498DB  ✓ (Azul)
#2ECC71  ✓ (Verde)
```

### Cores Inválidas

```
red      ✗ (Nome de cor)
#FFF     ✗ (Formato curto)
FF5733   ✗ (Sem #)
#GG5733  ✗ (Caracteres inválidos)
#FF57331 ✗ (Muitos dígitos)
```

---

## Integração com Frontend

### Exemplo React/TypeScript

```typescript
interface ValidationError {
  field: string;
  message: string;
}

interface ValidationErrorResponse {
  errors?: ValidationError[];
  error?: string;
  field?: string;
}

async function createTransaction(data: TransactionData) {
  try {
    const response = await fetch('/api/transactions', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(data)
    });

    if (!response.ok) {
      const errorData: ValidationErrorResponse = await response.json();

      // Múltiplos erros
      if (errorData.errors) {
        errorData.errors.forEach(err => {
          console.error(`${err.field}: ${err.message}`);
          // Mostrar erro no campo específico do formulário
          setFieldError(err.field, err.message);
        });
      }

      // Erro único
      if (errorData.error && errorData.field) {
        setFieldError(errorData.field, errorData.error);
      }

      return;
    }

    const transaction = await response.json();
    console.log('Transaction criada:', transaction);
  } catch (error) {
    console.error('Erro na requisição:', error);
  }
}
```

---

## Dicas de Uso

1. **Sempre valide no frontend também** - Melhora UX
2. **Mostre erros específicos** - Use o campo `field` para destacar inputs
3. **Agrupe erros** - Mostre todos os erros de uma vez
4. **Use português** - Mensagens já vêm traduzidas
5. **Teste edge cases** - Valores extremos, strings vazias, etc.

---

## Resumo de Status HTTP

| Status | Significado | Quando Ocorre |
|--------|-------------|---------------|
| 200 | OK | Operação bem-sucedida |
| 201 | Created | Recurso criado |
| 400 | Bad Request | Validação falhou |
| 401 | Unauthorized | Token inválido/ausente |
| 404 | Not Found | Recurso não existe |
| 409 | Conflict | Email já existe |
| 500 | Server Error | Erro interno |
