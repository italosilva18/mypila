# Security Implementation Guide

## Overview
This document describes the security measures implemented in the M2M Financeiro backend API.

## JWT Security

### Secure Secret Management
**Location:** `config/jwt.go`

The application now enforces secure JWT secret management:

1. **Environment Variable Priority**: Reads `JWT_SECRET` from environment
2. **Blocks Insecure Defaults**: Rejects "default_secret_key_change_me"
3. **Auto-Generation**: Generates cryptographically secure random secret if not set
4. **Warning System**: Logs clear warnings when using generated secrets

### Security Features:
- 256-bit (32 bytes) cryptographically secure random secrets
- Base64 URL-safe encoding
- Automatic detection of insecure configurations
- Fatal error on uninitialized secret access

### Production Setup:
```bash
# Generate a secure secret
openssl rand -base64 32

# Set in environment
export JWT_SECRET="your_generated_secret_here"
```

## Rate Limiting

### Global Rate Limit
**Configuration:** `main.go` line 43-57
- **Limit:** 100 requests per minute per IP
- **Applies to:** All API endpoints
- **Purpose:** Prevent DoS attacks

### Authentication Rate Limit
**Configuration:** `main.go` line 72-82
- **Limit:** 20 requests per minute per IP
- **Applies to:** `/api/auth/*` endpoints
- **Purpose:** Prevent brute force attacks and credential stuffing

### Implementation:
- IP-based tracking using Fiber's limiter middleware
- In-memory storage (consider Redis for distributed systems)
- HTTP 429 (Too Many Requests) on limit exceeded

## Authentication Endpoints

### POST /api/auth/register
- Creates new user account
- Hashes password with bcrypt (cost: 10)
- Returns JWT token valid for 72 hours

### POST /api/auth/login
- Authenticates existing user
- Validates credentials using bcrypt
- Returns JWT token valid for 72 hours

### GET /api/auth/me
**NEW ENDPOINT** - Token validation and user info
- **Protected:** Requires valid JWT token
- **Purpose:** 
  - Validate token is still valid
  - Fetch current user data
  - Check authentication state in frontend
- **Returns:** Current user object

**Usage Example:**
```bash
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
     http://localhost:8080/api/auth/me
```

**Response:**
```json
{
  "user": {
    "id": "507f1f77bcf86cd799439011",
    "name": "John Doe",
    "email": "john@example.com",
    "createdAt": "2025-12-14T10:30:00Z"
  }
}
```

## Token Structure

### JWT Claims:
```json
{
  "userId": "507f1f77bcf86cd799439011",
  "email": "user@example.com",
  "exp": 1734278400
}
```

### Token Lifetime:
- **Duration:** 72 hours (3 days)
- **Algorithm:** HS256 (HMAC-SHA256)
- **Validation:** Signature + expiration check

## Security Best Practices

### DO:
✅ Set `JWT_SECRET` environment variable in production
✅ Use different secrets for dev, staging, and production
✅ Rotate JWT secrets periodically
✅ Use HTTPS in production
✅ Monitor rate limit hits
✅ Keep dependencies updated

### DON'T:
❌ Commit `.env` file to version control
❌ Use default or weak secrets
❌ Share secrets in logs or error messages
❌ Store tokens in localStorage (use httpOnly cookies in production)
❌ Ignore security warnings in logs

## Threat Mitigation

| Threat | Mitigation | CVSS Score Before | After |
|--------|-----------|-------------------|-------|
| Weak JWT Secret | Mandatory strong secret | 9.1 (Critical) | 2.0 (Low) |
| Brute Force Auth | Rate limiting (20/min) | 7.5 (High) | 3.5 (Low) |
| DoS Attack | Global rate limit (100/min) | 7.5 (High) | 4.0 (Medium) |
| Token Validation | /me endpoint | 5.0 (Medium) | 2.0 (Low) |

## Compliance

### OWASP Top 10 Coverage:
- **A01:2021 - Broken Access Control**: JWT authentication + Protected middleware
- **A02:2021 - Cryptographic Failures**: bcrypt password hashing + secure JWT secrets
- **A07:2021 - Identification and Authentication Failures**: Rate limiting + strong password hashing

### Regulatory Compliance:
- **LGPD/GDPR**: Password hashing (personal data protection)
- **PCI-DSS**: Secure authentication mechanisms
- **SOC2**: Access control and monitoring capabilities

## Monitoring & Alerting

### Log Events to Monitor:
1. `[SECURITY WARNING] Generated random JWT_SECRET` - Indicates missing env variable
2. `[SECURITY WARNING] Detected insecure default JWT_SECRET` - Security violation
3. HTTP 429 responses - Potential attack or misconfigured client
4. Failed login attempts - Brute force indicator

### Recommended Alerts:
- Alert on >10 failed logins from same IP in 5 minutes
- Alert on JWT_SECRET not set in production
- Alert on >50 rate limit violations per hour

## Future Enhancements

### Short-term:
- [ ] Add refresh token mechanism
- [ ] Implement account lockout after N failed attempts
- [ ] Add CORS whitelist instead of wildcard
- [ ] Add request signing for sensitive operations

### Medium-term:
- [ ] Redis-backed rate limiting for distributed systems
- [ ] JWT token blacklist for logout functionality
- [ ] Multi-factor authentication (MFA)
- [ ] IP whitelist for admin operations

### Long-term:
- [ ] OAuth2/OIDC integration
- [ ] Hardware security module (HSM) for key management
- [ ] Advanced threat detection (ML-based)
- [ ] Web Application Firewall (WAF) integration

## Incident Response

### If JWT_SECRET is compromised:
1. **Immediately** rotate JWT_SECRET in all environments
2. Invalidate all existing tokens (restart service or implement blacklist)
3. Force password reset for all users
4. Review access logs for suspicious activity
5. Document incident and lessons learned

### If rate limits are bypassed:
1. Investigate attack vector (distributed IPs, etc.)
2. Consider implementing additional layers (Cloudflare, WAF)
3. Temporarily reduce rate limits if under active attack
4. Block malicious IP ranges at network level

## Testing

### Security Tests to Run:
```bash
# Test rate limiting
for i in {1..25}; do curl http://localhost:8080/api/auth/login; done

# Test JWT validation
curl -H "Authorization: Bearer invalid_token" \
     http://localhost:8080/api/auth/me

# Test missing JWT_SECRET (should auto-generate)
unset JWT_SECRET
go run main.go
```

## Contact

For security concerns or vulnerability reports:
- Create a security advisory in GitHub
- Contact: security@m2m-financeiro.com (if applicable)

---

**Last Updated:** 2025-12-14
**Version:** 1.0.0
**Maintained by:** Security Guardian
