#!/bin/bash
# ==============================================================
# Generate self-signed SSL certificates for local development
# ==============================================================
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CERT_DIR="$SCRIPT_DIR"

echo "Generating self-signed SSL certificates..."

openssl req -x509 -nodes \
  -days 365 \
  -newkey rsa:2048 \
  -keyout "$CERT_DIR/server.key" \
  -out "$CERT_DIR/server.crt" \
  -subj "/C=ID/ST=Local/L=Local/O=Singkatin/OU=Dev/CN=localhost" \
  -addext "subjectAltName=DNS:localhost,DNS:*.localhost,IP:127.0.0.1"

echo "Certificates generated:"
echo "    📄 $CERT_DIR/server.crt"
echo "    🔑 $CERT_DIR/server.key"
echo ""
echo "These are self-signed certs for development only."
echo "For production, use proper certificates (e.g., Let's Encrypt)."
