#!/bin/bash
# SPDX-License-Identifier: MIT
# Generate self-signed TLS certificate for development/testing
# DO NOT use self-signed certificates in production!

set -e

CERT_DIR="${1:-./certs}"
DOMAIN="${2:-localhost}"
DAYS=365

echo "🔐 Generating self-signed TLS certificate for development"
echo "📁 Certificate directory: $CERT_DIR"
echo "🌐 Domain: $DOMAIN"
echo "📅 Valid for: $DAYS days"
echo ""

# Create certs directory if it doesn't exist
mkdir -p "$CERT_DIR"

# Generate private key
echo "🔑 Generating private key..."
openssl genrsa -out "$CERT_DIR/server.key" 2048

# Generate certificate signing request
echo "📝 Generating certificate signing request..."
openssl req -new -key "$CERT_DIR/server.key" \
    -out "$CERT_DIR/server.csr" \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=$DOMAIN"

# Generate self-signed certificate
echo "📜 Generating self-signed certificate..."
openssl x509 -req -days $DAYS \
    -in "$CERT_DIR/server.csr" \
    -signkey "$CERT_DIR/server.key" \
    -out "$CERT_DIR/server.crt"

# Cleanup CSR
rm "$CERT_DIR/server.csr"

# Set appropriate permissions
chmod 600 "$CERT_DIR/server.key"
chmod 644 "$CERT_DIR/server.crt"

echo ""
echo "✅ Certificate generated successfully!"
echo ""
echo "📋 Generated files:"
echo "  Certificate: $CERT_DIR/server.crt"
echo "  Private Key: $CERT_DIR/server.key"
echo ""
echo "🚀 To start the server with HTTPS:"
echo "  ./bin/weave-mcp -tls -tls-cert $CERT_DIR/server.crt -tls-key $CERT_DIR/server.key"
echo ""
echo "⚠️  WARNING: This is a self-signed certificate for development only!"
echo "   Browsers will show security warnings. For production, use a"
echo "   certificate from a trusted Certificate Authority (CA)."
