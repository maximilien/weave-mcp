# HTTPS/TLS Setup Guide

Complete guide for enabling HTTPS/TLS in Weave MCP Server.

**Version:** 0.9.0
**Last Updated:** 2026-01-05

---

## Overview

Weave MCP Server supports both HTTP and HTTPS protocols. By default, the server runs on HTTP for backward compatibility. HTTPS is available as an opt-in feature for secure communications.

## Quick Start

### Development (Self-Signed Certificate)

**1. Generate certificate:**
```bash
./tools/generate-cert.sh
```

**2. Start server with HTTPS:**
```bash
./bin/weave-mcp -tls -tls-cert ./certs/server.crt -tls-key ./certs/server.key
```

**3. Access server:**
```
https://localhost:8030
```

⚠️ **Note:** Browsers will show security warnings for self-signed certificates.

### Production (Trusted Certificate)

**1. Obtain certificate from a trusted CA:**
- Let's Encrypt (free, automated)
- Commercial CA (Digicert, Sectigo, etc.)

**2. Configure in `config.yaml`:**
```yaml
tls:
  enabled: true
  cert_file: /etc/ssl/certs/server.crt
  key_file: /etc/ssl/private/server.key
  auto_redirect: true  # Redirect HTTP to HTTPS
```

**3. Start server:**
```bash
./start.sh
```

---

## Configuration Options

### Option 1: Command-Line Flags

```bash
./bin/weave-mcp \
  -tls \
  -tls-cert /path/to/cert.crt \
  -tls-key /path/to/key.key \
  -tls-redirect
```

**Flags:**
- `-tls`: Enable HTTPS/TLS (default: false)
- `-tls-cert`: Path to TLS certificate file
- `-tls-key`: Path to TLS private key file
- `-tls-redirect`: Auto-redirect HTTP to HTTPS (default: false)

### Option 2: Configuration File

Add to `config.yaml`:

```yaml
tls:
  enabled: true                     # Enable HTTPS
  cert_file: ./certs/server.crt    # Certificate path
  key_file: ./certs/server.key     # Private key path
  auto_redirect: false              # HTTP to HTTPS redirect
```

**Note:** Command-line flags override configuration file settings.

---

## HTTP to HTTPS Redirect

When `auto_redirect` is enabled, the server starts two listeners:

1. **HTTPS server** on configured port (default: 8030)
2. **HTTP redirect server** on port 80

All HTTP requests are automatically redirected to HTTPS with HTTP 301 status.

**Enable redirect:**

```bash
# Command-line
./bin/weave-mcp -tls -tls-cert ./certs/server.crt -tls-key ./certs/server.key -tls-redirect

# Config file
tls:
  enabled: true
  cert_file: ./certs/server.crt
  key_file: ./certs/server.key
  auto_redirect: true
```

**Requirements:**
- Server must have permission to bind to port 80 (may require root/sudo)
- Port 80 must not be in use by another service

---

## Certificate Generation

### Development: Self-Signed Certificate

Use the provided script:

```bash
./tools/generate-cert.sh [cert_dir] [domain]
```

**Examples:**

```bash
# Default: ./certs, localhost
./tools/generate-cert.sh

# Custom directory
./tools/generate-cert.sh /etc/ssl/certs

# Custom domain
./tools/generate-cert.sh ./certs example.com
```

**Manual generation:**

```bash
# Create directory
mkdir -p ./certs

# Generate private key
openssl genrsa -out ./certs/server.key 2048

# Generate certificate signing request
openssl req -new -key ./certs/server.key \
  -out ./certs/server.csr \
  -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"

# Generate self-signed certificate (valid 365 days)
openssl x509 -req -days 365 \
  -in ./certs/server.csr \
  -signkey ./certs/server.key \
  -out ./certs/server.crt

# Set permissions
chmod 600 ./certs/server.key
chmod 644 ./certs/server.crt
```

### Production: Let's Encrypt (Free)

**Using Certbot:**

```bash
# Install certbot
sudo apt-get install certbot  # Ubuntu/Debian
brew install certbot          # macOS

# Generate certificate
sudo certbot certonly --standalone -d your-domain.com

# Certificates will be in:
# /etc/letsencrypt/live/your-domain.com/fullchain.pem
# /etc/letsencrypt/live/your-domain.com/privkey.pem
```

**Configure Weave MCP:**

```yaml
tls:
  enabled: true
  cert_file: /etc/letsencrypt/live/your-domain.com/fullchain.pem
  key_file: /etc/letsencrypt/live/your-domain.com/privkey.pem
  auto_redirect: true
```

**Auto-renewal:**

```bash
# Add to crontab
0 0 * * * certbot renew --quiet && systemctl restart weave-mcp
```

---

## Security Best Practices

### Certificate Management

1. **Never commit certificates to version control**
   - Add `*.crt`, `*.key`, `*.pem` to `.gitignore`
   - Use environment-specific certificate paths

2. **Use strong private keys**
   - Minimum 2048-bit RSA keys
   - Consider 4096-bit for high-security environments
   - Use elliptic curve (ECDSA) for better performance

3. **Restrict key file permissions**
   ```bash
   chmod 600 /path/to/server.key
   chown weave-mcp:weave-mcp /path/to/server.key
   ```

4. **Rotate certificates regularly**
   - Set expiration dates (Let's Encrypt: 90 days)
   - Automate renewal process
   - Monitor expiration dates

### TLS Configuration

The server uses secure defaults:

- **TLS versions:** TLS 1.2+ only (Go stdlib default)
- **Cipher suites:** Strong ciphers only (Go stdlib default)
- **Timeouts:**
  - Read: 30 seconds
  - Write: 30 seconds
  - Idle: 120 seconds

### Firewall Configuration

```bash
# Allow HTTPS
sudo ufw allow 8030/tcp

# Allow HTTP redirect (if auto_redirect enabled)
sudo ufw allow 80/tcp

# Block direct HTTP if using HTTPS only
sudo ufw deny 8030/tcp  # Optional: if only using HTTPS
```

---

## Troubleshooting

### Common Issues

**1. Certificate file not found**

```
FATAL: TLS certificate file not found cert_file=/path/to/cert.crt
```

**Solution:**
- Verify certificate path is correct
- Check file permissions (must be readable)
- Use absolute paths instead of relative paths

**2. Permission denied (port 80)**

```
ERROR: HTTP redirect server error error="listen tcp :80: bind: permission denied"
```

**Solution:**
```bash
# Option 1: Run as root/sudo (not recommended)
sudo ./bin/weave-mcp -tls -tls-redirect ...

# Option 2: Use setcap (recommended)
sudo setcap 'cap_net_bind_service=+ep' ./bin/weave-mcp

# Option 3: Use different port for redirect
# (not standard, but works)
```

**3. Browser security warnings**

```
NET::ERR_CERT_AUTHORITY_INVALID
```

**Solution:**
- **Development:** This is expected for self-signed certificates. Click "Advanced" → "Proceed anyway"
- **Production:** Use a certificate from a trusted CA (Let's Encrypt, commercial CA)

**4. Certificate/key mismatch**

```
FATAL: Failed to start HTTPS server error="tls: private key does not match public key"
```

**Solution:**
- Ensure certificate and key are from the same generation
- Regenerate both certificate and key
- Verify you're not mixing certificates from different domains

### Testing HTTPS

**1. Test with curl:**

```bash
# Self-signed (skip verification)
curl -k https://localhost:8030/health

# Trusted certificate
curl https://localhost:8030/health

# Check certificate details
openssl s_client -connect localhost:8030 -showcerts
```

**2. Test with browser:**

Navigate to `https://localhost:8030`

**3. Test redirect:**

```bash
# Should redirect to HTTPS
curl -L http://localhost/health
```

---

## Migration from HTTP to HTTPS

### Step-by-Step Migration

**1. Generate/obtain certificates:**
```bash
./tools/generate-cert.sh
```

**2. Test HTTPS alongside HTTP:**
```bash
# Start with HTTPS enabled, no redirect
./bin/weave-mcp -tls -tls-cert ./certs/server.crt -tls-key ./certs/server.key -port 8031
```

**3. Verify HTTPS works:**
```bash
curl -k https://localhost:8031/health
```

**4. Update clients to use HTTPS:**
- Update URLs in MCP clients
- Test all integrations

**5. Enable redirect (optional):**
```bash
./bin/weave-mcp -tls -tls-cert ./certs/server.crt -tls-key ./certs/server.key -tls-redirect
```

**6. Update configuration:**
```yaml
tls:
  enabled: true
  cert_file: ./certs/server.crt
  key_file: ./certs/server.key
  auto_redirect: true
```

---

## Examples

### Example 1: Development with Self-Signed Cert

```bash
# Generate certificate
./tools/generate-cert.sh

# Start server
./bin/weave-mcp -tls -tls-cert ./certs/server.crt -tls-key ./certs/server.key

# Test
curl -k https://localhost:8030/health
```

### Example 2: Production with Let's Encrypt

```bash
# Generate certificate with certbot
sudo certbot certonly --standalone -d mcp.example.com

# Configure config.yaml
cat >> config.yaml <<EOF
tls:
  enabled: true
  cert_file: /etc/letsencrypt/live/mcp.example.com/fullchain.pem
  key_file: /etc/letsencrypt/live/mcp.example.com/privkey.pem
  auto_redirect: true
EOF

# Start server
./start.sh
```

### Example 3: Docker with HTTPS

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o weave-mcp src/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/weave-mcp .
COPY --from=builder /app/config.yaml .
COPY certs/ ./certs/

EXPOSE 8030 80
CMD ["./weave-mcp", "-tls", "-tls-cert", "./certs/server.crt", "-tls-key", "./certs/server.key", "-tls-redirect"]
```

---

## FAQ

**Q: Does HTTPS affect performance?**
A: Minimal impact. TLS encryption adds ~1-2ms latency. Benefits of security outweigh small performance cost.

**Q: Can I use both HTTP and HTTPS simultaneously?**
A: Not on the same port. You can run HTTPS on 8030 and optionally redirect HTTP (port 80) to HTTPS.

**Q: Do I need to restart the server when renewing certificates?**
A: Yes, certificates are loaded at startup. Restart the server after renewal.

**Q: What certificate formats are supported?**
A: PEM format (`.crt`, `.pem` for certificate, `.key`, `.pem` for private key).

**Q: Can I use wildcard certificates?**
A: Yes, wildcard certificates (*.example.com) work perfectly.

**Q: Is HTTP/2 supported?**
A: Yes, Go's http.Server automatically supports HTTP/2 when using HTTPS.

---

## Additional Resources

- [Let's Encrypt Documentation](https://letsencrypt.org/docs/)
- [OpenSSL Documentation](https://www.openssl.org/docs/)
- [Mozilla SSL Configuration Generator](https://ssl-config.mozilla.org/)
- [SSL Labs Server Test](https://www.ssllabs.com/ssltest/)

---

## Support

For issues or questions:
- GitHub Issues: https://github.com/maximilien/weave-mcp/issues
- MCP Tools Reference: [MCP_TOOLS.md](./MCP_TOOLS.md)
- Usage Examples: [EXAMPLES.md](./EXAMPLES.md)
