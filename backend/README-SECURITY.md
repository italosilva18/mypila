# Security Improvements - M2M Financeiro Backend

## Quick Overview

This backend now includes **enterprise-grade security** for JWT authentication and API protection.

## What Changed?

### 1. JWT Secret is Now Mandatory and Secure

**Before:**
```go
// INSECURE - Used hardcoded default
secret := os.Getenv("JWT_SECRET")
if secret == "" {
    secret = "default_secret_key_change_me"  // BAD!
}
```

**After:**
```go
// SECURE - Enforces strong secrets
config.InitializeJWTSecret()  // Fails or generates secure random if not set
secret := config.GetJWTSecret()
```

**Impact:** Prevents authentication bypass vulnerabilities (CVSS 9.1 Critical)

---

### 2. New Token Validation Endpoint

**Endpoint:** `GET /api/auth/me`

**Purpose:** Validate tokens and fetch current user data

**Example:**
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

**Use in Frontend:**
```javascript
// Check if user is still authenticated on app load
async function checkAuth() {
  const token = localStorage.getItem('token');
  if (!token) return null;
  
  try {
    const response = await fetch('/api/auth/me', {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    
    if (response.ok) {
      return await response.json();
    }
    
    // Token invalid, clear it
    localStorage.removeItem('token');
    return null;
  } catch (error) {
    return null;
  }
}
```

---

### 3. Rate Limiting Protection

**Global Limit:** 100 requests/minute per IP
**Auth Limit:** 20 requests/minute per IP (login/register)

**Example Response when limited:**
```json
{
  "error": "Too many authentication attempts. Please try again later."
}
```

**HTTP Status:** 429 (Too Many Requests)

**Impact:** Protects against brute force and DoS attacks

---

## Setup Instructions

### Option 1: Use Environment Variable

```bash
# Generate a secure secret
openssl rand -base64 32

# Set it
export JWT_SECRET="YourGeneratedSecretHere"

# Run
go run main.go
```

### Option 2: Use .env File

```bash
# Copy example
cp .env.example .env

# Generate secret
openssl rand -base64 32

# Edit .env and add:
JWT_SECRET=YourGeneratedSecretHere

# Run
go run main.go
```

### Option 3: Let It Auto-Generate (Development Only)

```bash
# Just run without setting JWT_SECRET
go run main.go
```

**Warning:** Auto-generated secrets change on every restart, invalidating all tokens!

---

## Security Logs

### Good (Secure Setup):
```
[SECURITY] JWT_SECRET loaded from environment variable
```

### Warning (Development Mode):
```
[SECURITY WARNING] JWT_SECRET not set in environment!
[SECURITY WARNING] Generated random JWT_SECRET for this session
[SECURITY WARNING] All tokens will be invalidated on server restart
[SECURITY WARNING] Set JWT_SECRET environment variable for production!
[SECURITY INFO] Generated secret: AbCdEf123456...
[SECURITY INFO] Add this to your .env file or environment variables
```

---

## Testing

### Test JWT Secret Security
```bash
# Without JWT_SECRET (should generate random)
unset JWT_SECRET
go run main.go

# With JWT_SECRET (should use it)
export JWT_SECRET="test_secret_for_development"
go run main.go
```

### Test /me Endpoint
```bash
# Register a user
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@test.com","password":"test123"}' \
  | jq -r '.token')

# Test /me
curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/api/auth/me
```

### Test Rate Limiting
```bash
# Should fail after 20 requests
for i in {1..25}; do 
  echo "Request $i"
  curl -X POST http://localhost:8080/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"test","password":"test"}'
  echo ""
done
```

---

## File Structure

```
backend/
├── config/
│   └── jwt.go              # NEW: Secure JWT secret management
├── handlers/
│   └── auth.go             # MODIFIED: Added GetMe() + secure JWT
├── middleware/
│   └── auth.go             # MODIFIED: Uses secure JWT config
├── main.go                 # MODIFIED: JWT init + rate limiting
├── .env.example            # NEW: Environment template
├── .gitignore              # NEW: Protects secrets
├── SECURITY.md             # NEW: Comprehensive security docs
├── SECURITY-QUICKSTART.md  # NEW: Quick setup guide
├── CHANGELOG.md            # NEW: Version history
└── test-security.sh        # NEW: Security test suite
```

---

## API Reference

### Authentication Endpoints

#### POST /api/auth/register
Creates a new user account.

**Request:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "SecurePassword123!"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "507f1f77bcf86cd799439011",
    "name": "John Doe",
    "email": "john@example.com",
    "createdAt": "2025-12-14T10:30:00Z"
  }
}
```

#### POST /api/auth/login
Authenticates an existing user.

**Request:**
```json
{
  "email": "john@example.com",
  "password": "SecurePassword123!"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "507f1f77bcf86cd799439011",
    "name": "John Doe",
    "email": "john@example.com",
    "createdAt": "2025-12-14T10:30:00Z"
  }
}
```

#### GET /api/auth/me (NEW)
Returns the current authenticated user.

**Headers:**
```
Authorization: Bearer <your_jwt_token>
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

**Error Responses:**
```json
// Missing token
{
  "error": "Missing authorization token"
}

// Invalid token
{
  "error": "Invalid or expired token"
}

// User not found (deleted after token was issued)
{
  "error": "User not found"
}
```

---

## Security Best Practices

### DO:
- Generate a strong random JWT_SECRET for production
- Use different secrets for dev, staging, and production
- Set JWT_SECRET via environment variables or secrets manager
- Enable HTTPS in production
- Monitor rate limit violations
- Regularly update dependencies
- Review authentication logs

### DON'T:
- Commit .env files to version control
- Use weak or default secrets
- Share secrets in logs or error messages
- Store tokens in localStorage (consider httpOnly cookies)
- Ignore security warnings in logs
- Use the same JWT_SECRET across environments

---

## Production Deployment

### Heroku
```bash
heroku config:set JWT_SECRET="$(openssl rand -base64 32)"
```

### Docker
```bash
docker run -e JWT_SECRET="your_secret_here" your-image
```

### Kubernetes
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: jwt-secret
type: Opaque
data:
  JWT_SECRET: <base64_encoded_secret>
```

### Docker Compose
```yaml
services:
  backend:
    environment:
      - JWT_SECRET=${JWT_SECRET}
```

---

## Troubleshooting

### Tokens Keep Getting Invalidated
**Cause:** JWT_SECRET is changing between restarts  
**Solution:** Set a persistent JWT_SECRET in environment variables

### Rate Limited on Local Development
**Cause:** Making too many requests (>100/min or >20/min on auth)  
**Solution:** Wait 1 minute or adjust limits in `main.go` for development

### /me Endpoint Returns 401
**Cause:** Token expired (72h lifetime) or invalid  
**Solution:** Login again to get a fresh token

---

## Support

For questions or security concerns:
- Read: `SECURITY.md` (comprehensive guide)
- Read: `SECURITY-QUICKSTART.md` (quick setup)
- Read: `CHANGELOG.md` (what changed)

---

**Version:** 1.1.0  
**Last Updated:** 2025-12-14  
**Security Level:** ENHANCED
