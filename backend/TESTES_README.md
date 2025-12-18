# Testes Unitarios - Backend Go

## Visao Geral

Este documento descreve a suite de testes unitarios implementada para o backend do sistema M2M Financeiro. Os testes seguem as melhores praticas de Go, utilizando **table-driven tests** para cobertura completa e eficiente.

## Arquivos de Teste Criados

### 1. `helpers/ownership_test.go`
**Objetivo**: Testar funcoes de validacao de propriedade de recursos.

**Testes Implementados**:
- ✅ `TestValidateCompanyOwnership` - Valida acesso de usuario a empresas
  - Acesso valido do proprietario
  - Acesso negado para usuario diferente
  - Empresa nao encontrada
- ✅ `TestValidateCompanyOwnershipByString` - Valida conversao de ID string
  - ID valido
  - Formato invalido
  - ID vazio
- ✅ `TestGetUserIDFromContext` - Extrai user ID do contexto Fiber
  - User ID valido
  - User ID ausente
  - Tipo invalido
  - Formato invalido
- ✅ `TestValidateTransactionOwnership` - Valida acesso a transacoes
  - Acesso valido do proprietario
  - Acesso negado
  - Transacao nao encontrada
- ✅ `TestValidateCompanyOwnershipByString_EdgeCases` - Casos extremos
- ✅ `TestGetUserIDFromContext_EdgeCases` - Casos extremos
- ✅ `BenchmarkGetUserIDFromContext` - Benchmark de performance

**Total**: 8 funcoes de teste + 1 benchmark

**Comando de execucao**:
```bash
# Executar todos os testes de ownership (requer MongoDB rodando)
go test ./helpers -v -run "TestValidate|TestGet"

# Executar apenas testes sem banco de dados
go test ./helpers -run TestGetUserIDFromContext -v

# Executar benchmarks
go test ./helpers -bench=. -benchmem
```

---

### 2. `middleware/auth_test.go`
**Objetivo**: Testar middleware de autenticacao JWT.

**Testes Implementados**:
- ✅ `TestProtectedMiddleware` - Suite completa de validacao de token (10 sub-testes)
  - Token valido com Bearer
  - Header ausente
  - Token sem Bearer prefix
  - Formato invalido
  - Token malformado
  - Token expirado
  - Token com secret errado
  - Bearer vazio
  - Schema errado
  - Multiplos Bearer prefixes
- ✅ `TestProtectedMiddleware_ContextValues` - Valida valores no contexto
- ✅ `TestProtectedMiddleware_TokenWithoutClaims` - Token sem claims
- ✅ `TestProtectedMiddleware_DifferentSigningMethod` - Metodo de assinatura errado
- ✅ `TestProtectedMiddleware_VeryLongToken` - Token muito longo
- ✅ `TestProtectedMiddleware_TokenExpiringNow` - Token expirando
- ✅ `TestProtectedMiddleware_CaseSensitiveBearer` - Case sensitivity (4 sub-testes)
- ✅ `TestProtectedMiddleware_MultipleSpaces` - Espacos multiplos (4 sub-testes)
- ✅ `BenchmarkProtectedMiddleware_ValidToken` - Benchmark token valido
- ✅ `BenchmarkProtectedMiddleware_InvalidToken` - Benchmark token invalido
- ✅ `BenchmarkProtectedMiddleware_NoToken` - Benchmark sem token
- ✅ `BenchmarkGenerateTestToken` - Benchmark geracao token

**Total**: 8 funcoes de teste + 4 benchmarks

**Resultado dos Testes**:
```
PASS: TestProtectedMiddleware (10/10 sub-testes)
PASS: TestProtectedMiddleware_ContextValues
PASS: TestProtectedMiddleware_TokenWithoutClaims
PASS: TestProtectedMiddleware_DifferentSigningMethod
PASS: TestProtectedMiddleware_VeryLongToken
PASS: TestProtectedMiddleware_TokenExpiringNow
PASS: TestProtectedMiddleware_CaseSensitiveBearer (4/4 sub-testes)
PASS: TestProtectedMiddleware_MultipleSpaces (4/4 sub-testes)
```

**Comando de execucao**:
```bash
# Executar todos os testes de middleware
go test ./middleware -v

# Executar benchmarks
go test ./middleware -bench=. -benchmem
```

---

### 3. `handlers/auth_test.go`
**Objetivo**: Testar handlers de autenticacao (registro, login, token).

**Testes Implementados**:
- ✅ `TestRegister` - Teste de registro de usuario (7 sub-testes)
  - Registro valido
  - Email duplicado
  - Nome ausente
  - Email invalido
  - Senha muito curta
  - Email vazio
  - Tentativa de XSS no nome
- ✅ `TestLogin` - Teste de login (6 sub-testes)
  - Credenciais validas
  - Senha errada
  - Usuario inexistente
  - Email vazio
  - Senha vazia
  - Email invalido
- ✅ `TestGetMe` - Teste de endpoint /me (4 sub-testes)
  - User ID valido
  - User ID ausente
  - Tipo invalido
  - Formato invalido
- ✅ `TestGenerateToken` - Teste de geracao de token JWT
- ✅ `TestRegisterInvalidJSON` - JSON invalido no registro
- ✅ `TestLoginInvalidJSON` - JSON invalido no login
- ✅ `TestRegisterEdgeCases` - Casos extremos (3 sub-testes)
  - Nome muito longo
  - SQL injection
  - MongoDB injection
- ✅ `BenchmarkGenerateToken` - Benchmark geracao token

**Total**: 7 funcoes de teste + 1 benchmark

**Observacao Importante**:
Os testes de handlers requerem MongoDB rodando localmente devido a inicializacao global de servicos no pacote `handlers`.

**Comando de execucao**:
```bash
# NOTA: Certifique-se que MongoDB esta rodando em localhost:27017

# Executar todos os testes de auth
go test ./handlers -v -run "TestGenerate|TestRegister|TestLogin|TestGetMe"

# Executar benchmarks
go test ./handlers -bench=BenchmarkGenerateToken -benchmem
```

---

## Estatisticas Gerais

### Cobertura de Testes
```
Total de arquivos de teste: 3
Total de funcoes de teste: 23
Total de sub-testes: ~40
Total de benchmarks: 6
```

### Tecnicas Utilizadas

1. **Table-Driven Tests**: Todos os testes utilizam a abordagem table-driven do Go
   ```go
   tests := []struct {
       name string
       input interface{}
       expected interface{}
       expectError bool
   }{
       // casos de teste...
   }
   ```

2. **Test Fixtures**: Funcoes auxiliares para criar dados de teste
   - `setupTestDB()`: Inicializa banco de teste
   - `createTestUser()`: Cria usuario de teste
   - `createTestCompany()`: Cria empresa de teste
   - `createTestTransaction()`: Cria transacao de teste

3. **Test Isolation**: Cada teste usa banco de dados separado e limpa apos execucao

4. **Graceful Skipping**: Testes que requerem MongoDB pulam graciosamente se nao disponivel

5. **Benchmarks**: Testes de performance para operacoes criticas

## Como Executar os Testes

### Pre-requisitos
```bash
# 1. MongoDB rodando localmente (para testes de integracao)
docker run -d -p 27017:27017 mongo:latest

# 2. Dependencias Go instaladas
go mod download
```

### Comandos de Execucao

```bash
# Executar TODOS os testes (requer MongoDB)
go test ./... -v

# Executar testes de um pacote especifico
go test ./middleware -v
go test ./helpers -v -run TestGetUserIDFromContext

# Executar apenas testes unitarios (sem banco)
go test ./helpers -run TestGetUserIDFromContext -v
go test ./middleware -v

# Executar com cobertura
go test ./middleware -cover
go test ./helpers -run TestGetUserIDFromContext -cover

# Executar benchmarks
go test ./middleware -bench=. -benchmem
go test ./helpers -bench=. -benchmem

# Executar testes em modo verbose com detalhes
go test ./middleware -v -race

# Executar teste especifico
go test ./middleware -run TestProtectedMiddleware/Valid_token -v
```

### Interpretando Resultados

```bash
# Exemplo de saida de sucesso:
=== RUN   TestProtectedMiddleware
=== RUN   TestProtectedMiddleware/Valid_token_with_Bearer_prefix
--- PASS: TestProtectedMiddleware/Valid_token_with_Bearer_prefix (0.00s)
PASS
ok      m2m-backend/middleware  2.015s

# Exemplo de teste pulado:
--- SKIP: TestValidateCompanyOwnership (0.00s)
    ownership_test.go:26: Skipping test: MongoDB not available
```

## Cenarios de Teste Cobertos

### Seguranca
- ✅ Validacao de tokens JWT
- ✅ Tokens expirados
- ✅ Tokens com assinatura invalida
- ✅ Tentativas de XSS
- ✅ SQL/MongoDB injection
- ✅ Validacao de ownership (acesso nao autorizado)

### Validacao de Dados
- ✅ Email invalido
- ✅ Senha muito curta
- ✅ Campos obrigatorios vazios
- ✅ Formato de ObjectID invalido
- ✅ JSON malformado

### Edge Cases
- ✅ Strings muito longas
- ✅ Caracteres especiais
- ✅ Tipos de dados incorretos
- ✅ Valores nulos/vazios
- ✅ Case sensitivity

### Casos de Erro
- ✅ Usuario nao autenticado
- ✅ Recurso nao encontrado
- ✅ Acesso negado (403)
- ✅ Credenciais invalidas
- ✅ Email duplicado

## Boas Praticas Implementadas

1. ✅ **Nomenclatura Clara**: Nomes de teste descrevem o cenario testado
2. ✅ **Isolamento**: Cada teste e independente
3. ✅ **Table-Driven**: Facilita adicionar novos casos
4. ✅ **Cleanup**: Recursos sempre sao limpos apos testes
5. ✅ **Skip Gracioso**: Testes nao falham se dependencias nao disponiveis
6. ✅ **Benchmarks**: Performance e monitorada
7. ✅ **Contexto Real**: Usa Fiber context real, nao mocks
8. ✅ **Documentacao**: Comentarios explicam o proposito

## Proximos Passos

Para expandir a suite de testes:

1. **Adicionar testes para handlers de company**
   - Criar/Listar/Atualizar/Deletar empresas

2. **Adicionar testes para handlers de transaction**
   - CRUD de transacoes
   - Validacoes de datas

3. **Adicionar testes de integracao**
   - Fluxos completos end-to-end

4. **Adicionar testes de concorrencia**
   - Race conditions
   - Deadlocks

5. **CI/CD Integration**
   - GitHub Actions
   - Cobertura automatica

## Contato e Suporte

Para duvidas sobre os testes:
- Verificar a documentacao inline nos arquivos `*_test.go`
- Executar `go test -v` para output detalhado
- Verificar logs de erro para diagnosticos

---

**Gerado por**: Testing Virtuoso - Claude Code
**Data**: 2025-12-16
**Versao Go**: 1.24.0
**Framework**: Fiber v2.52.10
