#!/bin/bash

# This script helps connect to the WebDAV server from macOS Finder

# WebDAV server details
SERVER_HOST="localhost"
SERVER_PORT="9999"
USERNAME="admin"
PASSWORD="1234"
PROTOCOL="http"  # Change to https if your server uses HTTPS

# Create the WebDAV URL
WEBDAV_URL="${PROTOCOL}://${SERVER_HOST}:${SERVER_PORT}/"

echo "===== WebDAV Connection Helper ====="
echo "WebDAV URL: ${WEBDAV_URL}"
echo "Username: ${USERNAME}"
echo "Password: ${PASSWORD}"
echo ""

echo "To connect from macOS Finder:"
echo "1. In Finder, press Cmd+K or select 'Go > Connect to Server...'"
echo "2. Enter the WebDAV URL: ${WEBDAV_URL}"
echo "3. Click 'Connect'"
echo "4. When prompted, enter the username and password"
echo ""

echo "Testing WebDAV connection..."
curl -u "${USERNAME}:${PASSWORD}" -X PROPFIND "${WEBDAV_URL}" -H "Depth: 1" -v

echo ""
echo "If you see a 207 Multi-Status response above, the WebDAV server is accessible."
echo "If you're still having trouble connecting from Finder, try these troubleshooting steps:"
echo ""
echo "1. Make sure the server is running with authentication enabled:"
echo "   ./bin/webdavserver -auth -debug"
echo ""
echo "2. For HTTPS connections, macOS requires trusted certificates. You may need to:"
echo "   a. Import the self-signed certificate into your macOS Keychain"
echo "   b. Or run the server with HTTP instead for testing purposes"
echo ""
echo "3. Try connecting with a different WebDAV client like Cyberduck or Transmit"
echo ""
echo "4. Check if your firewall is blocking the connection"
