#!/bin/bash

# Script to generate a self-signed certificate for WebDAV HTTPS testing

# Default values
CERT_DIR="./certs"
CERT_FILE="$CERT_DIR/webdav.crt"
KEY_FILE="$CERT_DIR/webdav.key"
DAYS=365
COMMON_NAME="localhost"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -d|--dir)
      CERT_DIR="$2"
      CERT_FILE="$CERT_DIR/webdav.crt"
      KEY_FILE="$CERT_DIR/webdav.key"
      shift 2
      ;;
    -cn|--common-name)
      COMMON_NAME="$2"
      shift 2
      ;;
    -days)
      DAYS="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: $0 [options]"
      echo "Options:"
      echo "  -d, --dir DIR          Directory to store certificates (default: ./certs)"
      echo "  -cn, --common-name CN  Common name for certificate (default: localhost)"
      echo "  -days DAYS             Validity period in days (default: 365)"
      echo "  -h, --help             Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

# Create certificate directory if it doesn't exist
mkdir -p "$CERT_DIR"

echo "Generating self-signed certificate for $COMMON_NAME"
echo "Certificate will be valid for $DAYS days"
echo "Files will be stored in $CERT_DIR"

# Generate private key and certificate
openssl req -x509 -newkey rsa:4096 -keyout "$KEY_FILE" -out "$CERT_FILE" -days "$DAYS" -nodes -subj "/CN=$COMMON_NAME"

# Check if generation was successful
if [ $? -eq 0 ]; then
  echo "Certificate generation successful!"
  echo "Certificate: $CERT_FILE"
  echo "Private key: $KEY_FILE"
  
  # Make the files readable
  chmod 644 "$CERT_FILE"
  chmod 600 "$KEY_FILE"
  
  echo ""
  echo "To use these certificates with the WebDAV server, run:"
  echo "go run cmd/webdavserver/main.go -https -cert $CERT_FILE -key $KEY_FILE [other options]"
else
  echo "Certificate generation failed!"
fi
