#!/bin/bash

# Default values
HOST="localhost"
PORT="9999"
PATH_PREFIX=""
USE_HTTPS=false
USERNAME=""
PASSWORD=""

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    -h|--host)
      HOST="$2"
      shift 2
      ;;
    -p|--port)
      PORT="$2"
      shift 2
      ;;
    -path|--path-prefix)
      PATH_PREFIX="$2"
      shift 2
      ;;
    -s|--https)
      USE_HTTPS=true
      shift
      ;;
    -u|--username)
      USERNAME="$2"
      shift 2
      ;;
    -pw|--password)
      PASSWORD="$2"
      shift 2
      ;;
    -h|--help)
      echo "Usage: $0 [options]"
      echo "Options:"
      echo "  -h, --host HOSTNAME      WebDAV server hostname (default: localhost)"
      echo "  -p, --port PORT          WebDAV server port (default: 9999)"
      echo "  -path, --path-prefix PATH WebDAV path prefix (default: none)"
      echo "  -s, --https              Use HTTPS instead of HTTP"
      echo "  -u, --username USERNAME  Username for authentication"
      echo "  -pw, --password PASSWORD Password for authentication"
      echo "  --help                   Show this help message"
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      echo "Use --help for usage information"
      exit 1
      ;;
  esac
done

# Construct the WebDAV URL
PROTOCOL="http"
if [ "$USE_HTTPS" = true ]; then
  PROTOCOL="https"
fi

# Build the URL with authentication if provided
if [ -n "$USERNAME" ] && [ -n "$PASSWORD" ]; then
  # URL encode the username and password
  USERNAME_ENCODED=$(echo -n "$USERNAME" | xxd -plain | tr -d '\n' | sed 's/\(.\{2\}\)/%\1/g')
  PASSWORD_ENCODED=$(echo -n "$PASSWORD" | xxd -plain | tr -d '\n' | sed 's/\(.\{2\}\)/%\1/g')
  
  WEBDAV_URL="${PROTOCOL}://${USERNAME_ENCODED}:${PASSWORD_ENCODED}@${HOST}:${PORT}${PATH_PREFIX}"
  echo "Opening WebDAV connection to ${PROTOCOL}://${USERNAME}:****@${HOST}:${PORT}${PATH_PREFIX} in Finder..."
else
  WEBDAV_URL="${PROTOCOL}://${HOST}:${PORT}${PATH_PREFIX}"
  echo "Opening WebDAV connection to $WEBDAV_URL in Finder..."
fi

# Open the WebDAV URL in Finder
open "$WEBDAV_URL"

echo "WebDAV connection opened in Finder."

if [ -z "$USERNAME" ] || [ -z "$PASSWORD" ]; then
  echo "If prompted, enter your credentials to connect."
fi

echo "Note: macOS may require you to reconnect after system restarts."
echo "Debug tip: If connection fails, try the following:"
echo "  1. Check that the WebDAV server is running"
echo "  2. Verify the hostname and port are correct"
echo "  3. If using HTTPS, ensure the certificate is trusted"
echo "  4. Try connecting manually in Finder with Go > Connect to Server"
