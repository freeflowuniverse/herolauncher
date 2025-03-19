#!/bin/bash

# Script to help set up environment variables for the videoconf package

# Check if .env.example exists
if [ ! -f "../.env.example" ]; then
    echo "Error: .env.example file not found!"
    exit 1
fi

# Check if .env already exists
if [ -f "../.env" ]; then
    read -p ".env file already exists. Do you want to overwrite it? (y/n): " overwrite
    if [ "$overwrite" != "y" ]; then
        echo "Setup canceled."
        exit 0
    fi
fi

# Copy the example file
cp "../.env.example" "../.env"
echo "Created .env file from template."

# Prompt for LiveKit credentials
read -p "Enter your LiveKit URL (e.g., wss://your-livekit-instance.livekit.cloud): " livekit_url
read -p "Enter your LiveKit API Key: " livekit_api_key
read -p "Enter your LiveKit API Secret: " livekit_api_secret
read -p "Enter the port for the server (default: 8096): " port
port=${port:-8096}

# Update the .env file with the provided values
sed -i '' "s|LIVEKIT_URL=.*|LIVEKIT_URL=$livekit_url|g" "../.env"
sed -i '' "s|LIVEKIT_API_KEY=.*|LIVEKIT_API_KEY=$livekit_api_key|g" "../.env"
sed -i '' "s|LIVEKIT_API_SECRET=.*|LIVEKIT_API_SECRET=$livekit_api_secret|g" "../.env"
sed -i '' "s|PORT=.*|PORT=$port|g" "../.env"

echo "Environment variables set successfully!"
echo "You can now run the videoconf server."
