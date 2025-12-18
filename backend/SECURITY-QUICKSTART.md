# Security Quick Start Guide

## 🚀 Setup in 3 Steps

### 1. Generate JWT Secret
```bash
openssl rand -base64 32
```

### 2. Set Environment Variable
```bash
# Linux/Mac
export JWT_SECRET="your_generated_secret_here"

# Windows PowerShell
$env:JWT_SECRET="your_generated_secret_here"

# Or create .env file
echo "JWT_SECRET=your_generated_secret_here" > .env
```

### 3. Run Server
```bash
go run main.go
```

## 🔍 Verify Security

### Check Logs
Look for these messages:
```
✅ GOOD: [SECURITY] JWT_SECRET loaded from environment variable
❌ BAD:  [SECURITY WARNING] JWT_SECRET not set in environment!
```

### Test Rate Limiting
```bash
# Should fail with HTTP 429 after 20 requests
for i in {1..25}; do 
  curl -X POST http://localhost:8080/api/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"test","password":"test"}'
done
```

### Test /me Endpoint
```bash
# Register/Login to get token
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"your@email.com","password":"yourpass"}' | jq -r '.token')

# Test /me endpoint
curl -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/api/auth/me
```

## 📋 Security Checklist

### Development
- [ ] Set JWT_SECRET in .env file
- [ ] Never commit .env to version control
- [ ] Test rate limiting works
- [ ] Test /api/auth/me endpoint
- [ ] Review security warnings in logs

### Production
- [ ] Set strong JWT_SECRET (32+ random characters)
- [ ] Use environment variables or secrets manager
- [ ] Enable HTTPS
- [ ] Configure CORS whitelist (not wildcard)
- [ ] Set up monitoring for rate limit violations
- [ ] Set up alerts for failed authentication attempts
- [ ] Regular security audits
- [ ] Keep dependencies updated

## 🆘 Troubleshooting

### "Invalid or expired token"
- Token expired (72h lifetime)
- JWT_SECRET changed since token was issued
- Token corrupted during transmission

### HTTP 429 - Rate Limited
- Too many requests from your IP
- Wait 1 minute and try again
- Check if you're in a loop

### "JWT secret not initialized"
- JWT_SECRET not set before server start
- Check environment variables
- Check .env file exists and is loaded

## 📞 Need Help?

See [SECURITY.md](./SECURITY.md) for detailed documentation.

## 🎯 New API Endpoints

### GET /api/auth/me
Returns current authenticated user.

**Request:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
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

## 🔧 Rate Limits

| Endpoint | Limit | Window |
|----------|-------|--------|
| Global (all) | 100 req | 1 minute |
| /api/auth/* | 20 req | 1 minute |

## 🛡️ Security Improvements

| Feature | Before | After |
|---------|--------|-------|
| JWT Secret | Insecure default | Mandatory secure secret |
| Brute Force | No protection | 20 req/min limit |
| DoS Attack | No protection | 100 req/min limit |
| Token Validation | No endpoint | /api/auth/me endpoint |

---

**Version:** 1.1.0  
**Last Updated:** 2025-12-14
