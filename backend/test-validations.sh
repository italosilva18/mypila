#!/bin/bash

# Script para testar as validações do backend
# Certifique-se de que o servidor está rodando em localhost:5000

BASE_URL="http://localhost:5000/api"
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "  Teste de Validações - Backend M2M"
echo "========================================="
echo ""

# 1. Teste de Registro com dados inválidos
echo -e "${YELLOW}1. Testando Registro com dados inválidos${NC}"
echo "POST /api/auth/register"
curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "invalid-email",
    "password": "123",
    "name": ""
  }' | json_pp
echo ""
echo ""

# 2. Teste de Login com email inválido
echo -e "${YELLOW}2. Testando Login com email inválido${NC}"
echo "POST /api/auth/login"
curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "not-an-email",
    "password": "test123"
  }' | json_pp
echo ""
echo ""

# 3. Teste de criação de Company com nome vazio
echo -e "${YELLOW}3. Testando Company com nome vazio${NC}"
echo "POST /api/companies"
curl -s -X POST "$BASE_URL/companies" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "name": ""
  }' | json_pp
echo ""
echo ""

# 4. Teste de criação de Company com nome muito longo
echo -e "${YELLOW}4. Testando Company com nome muito longo (>100 chars)${NC}"
echo "POST /api/companies"
curl -s -X POST "$BASE_URL/companies" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "name": "'"$(printf 'A%.0s' {1..150})"'"
  }' | json_pp
echo ""
echo ""

# 5. Teste de criação de Category com dados inválidos
echo -e "${YELLOW}5. Testando Category com cor inválida${NC}"
echo "POST /api/categories?companyId=YOUR_COMPANY_ID"
curl -s -X POST "$BASE_URL/categories?companyId=123456789012345678901234" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "name": "",
    "color": "invalid-color",
    "type": "EXPENSE"
  }' | json_pp
echo ""
echo ""

# 6. Teste de criação de Transaction com múltiplos erros
echo -e "${YELLOW}6. Testando Transaction com múltiplos erros${NC}"
echo "POST /api/transactions"
curl -s -X POST "$BASE_URL/transactions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "companyId": "123",
    "amount": -100,
    "description": "'"$(printf 'X%.0s' {1..250})"'",
    "category": "",
    "month": "InvalidMonth",
    "year": 1999,
    "status": "invalid"
  }' | json_pp
echo ""
echo ""

# 7. Teste de criação de Transaction com mês inválido
echo -e "${YELLOW}7. Testando Transaction com mês em inglês (inválido)${NC}"
echo "POST /api/transactions"
curl -s -X POST "$BASE_URL/transactions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "companyId": "123456789012345678901234",
    "amount": 100,
    "category": "Teste",
    "month": "January",
    "year": 2024,
    "status": "PAGO"
  }' | json_pp
echo ""
echo ""

# 8. Teste de RecurringTransaction com dados inválidos
echo -e "${YELLOW}8. Testando RecurringTransaction com dia inválido${NC}"
echo "POST /api/recurring"
curl -s -X POST "$BASE_URL/recurring" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "companyId": "123",
    "description": "",
    "amount": -50,
    "category": "",
    "dayOfMonth": 35
  }' | json_pp
echo ""
echo ""

# 9. Teste com dados VÁLIDOS - Transaction
echo -e "${GREEN}9. Testando Transaction com dados VÁLIDOS${NC}"
echo "POST /api/transactions"
curl -s -X POST "$BASE_URL/transactions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "companyId": "123456789012345678901234",
    "amount": 100.50,
    "description": "Descrição válida",
    "category": "Salário",
    "month": "Janeiro",
    "year": 2024,
    "status": "PAGO"
  }' | json_pp
echo ""
echo ""

# 10. Teste com dados VÁLIDOS - Category
echo -e "${GREEN}10. Testando Category com dados VÁLIDOS${NC}"
echo "POST /api/categories?companyId=YOUR_COMPANY_ID"
curl -s -X POST "$BASE_URL/categories?companyId=123456789012345678901234" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{
    "name": "Nova Categoria",
    "color": "#FF5733",
    "type": "EXPENSE",
    "budget": 1000
  }' | json_pp
echo ""
echo ""

echo "========================================="
echo "  Testes Concluídos"
echo "========================================="
echo ""
echo -e "${YELLOW}NOTAS:${NC}"
echo "- Substitua YOUR_TOKEN_HERE pelo token JWT válido"
echo "- Substitua YOUR_COMPANY_ID por um ID de empresa válido"
echo "- Os testes com dados válidos (#9 e #10) requerem autenticação"
echo "- Erros 401/403 indicam problemas de autenticação, não de validação"
echo ""
