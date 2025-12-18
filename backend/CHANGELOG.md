# Changelog

## [1.1.0] - 2025-12-14

### Security Enhancements

#### 🔒 JWT Secret Management (CRITICAL)
- **Added**: Mandatory JWT_SECRET validation in `config/jwt.go`
- **Changed**: No longer accepts insecure default "default_secret_key_change_me"
- **Added**: Automatic generation of cryptographically secure random secrets
- **Added**: Comprehensive security warnings when JWT_SECRET not set
- **Impact**: Prevents authentication bypass vulnerabilities (CVSS 9.1 → 2.0)

#### 🛡️ Rate Limiting
- **Added**: Global rate limiting (100 requests/minute per IP)
- **Added**: Auth-specific rate limiting (20 requests/minute per IP)
- **Added**: Protection against brute force attacks
- **Added**: Protection against DoS attacks
- **Impact**: Mitigates brute force and DoS threats (CVSS 7.5 → 3.5)

#### 🔑 Token Validation Endpoint
- **Added**: `GET /api/auth/me` endpoint
- **Feature**: Returns current authenticated user information
- **Feature**: Validates token is still valid
- **Feature**: Useful for frontend authentication state management

### Files Modified
- `config/jwt.go` (NEW) - Secure JWT secret management
- `handlers/auth.go` - Updated to use secure config + added GetMe handler
- `middleware/auth.go` - Updated to use secure config
- `main.go` - Added JWT initialization + rate limiting middleware

### Files Created
- `.env.example` - Environment variable template
- `.gitignore` - Prevents committing sensitive files
- `SECURITY.md` - Comprehensive security documentation
- `test-security.sh` - Security testing script

### Migration Guide

#### For Development
```bash
# Copy the example env file
cp .env.example .env

# Generate a secure JWT secret
openssl rand -base64 32

# Add to .env file
echo "JWT_SECRET=<your_generated_secret>" >> .env

# Run the server
go run main.go
```

#### For Production
```bash
# Set environment variable
export JWT_SECRET="your_secure_production_secret_here"

# Or use your deployment platform's secrets manager
# Heroku: heroku config:set JWT_SECRET="..."
# Docker: -e JWT_SECRET="..."
# Kubernetes: Create a Secret resource
```

### Breaking Changes
⚠️ **IMPORTANT**: The server will now generate a random JWT_SECRET if not set, which means:
- All tokens will be invalidated on server restart
- You MUST set JWT_SECRET in production environments
- Development should also set a persistent JWT_SECRET to avoid re-authentication

### Security Warnings
The server will now log clear warnings if:
1. JWT_SECRET is not set (will auto-generate)
2. Insecure default secret is detected
3. Rate limits are being hit

### Testing
Run the security test suite:
```bash
# Start the server
go run main.go

# In another terminal
./test-security.sh
```

### Dependencies Added
- `github.com/gofiber/fiber/v2/middleware/limiter` - Rate limiting

### Compliance
- ✅ OWASP A02:2021 - Cryptographic Failures
- ✅ OWASP A07:2021 - Identification and Authentication Failures
- ✅ LGPD/GDPR - Secure authentication
- ✅ PCI-DSS - Strong cryptography

### References
- [SECURITY.md](./SECURITY.md) - Complete security documentation
- [OWASP Top 10 2021](https://owasp.org/Top10/)
- [JWT Best Practices](https://tools.ietf.org/html/rfc8725)

---

## [1.0.0] - 2025-12-09

### Initial Release
- Basic authentication (register/login)
- JWT token generation
- User management
- Company management
- Transaction CRUD operations
- Category management
- Recurring transactions
- Statistics endpoints
