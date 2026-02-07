#!/bin/bash

# WireGuard Agent Setup Script
# This script sets up the WireGuard Agent for local development or production

set -e

echo "🚀 WireGuard Agent Setup"
echo "======================="
echo ""

# Check Go installation
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ from https://golang.org/"
    exit 1
fi

echo "✅ Go is installed: $(go version)"
echo ""

# Check if .env exists
if [ -f ".env" ]; then
    echo "⚠️  .env file already exists"
    read -p "Do you want to overwrite it? (y/N): " -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cp .env.example .env
        echo "✅ .env file created from example"
    fi
else
    cp .env.example .env
    echo "✅ .env file created from example"
fi

echo ""
echo "📝 Generating API Key..."

# Generate API key if not already set
if ! grep -q "API_KEY=.*[a-zA-Z0-9]" .env 2>/dev/null; then
    API_KEY=$(openssl rand -base64 32)
    if [ -f ".env" ]; then
        # Update .env with generated key
        if [[ "$OSTYPE" == "darwin"* ]]; then
            sed -i '' "s/API_KEY=.*/API_KEY=$API_KEY/" .env
        else
            sed -i "s/API_KEY=.*/API_KEY=$API_KEY/" .env
        fi
        echo "✅ API Key generated and saved to .env"
        echo "   Key: $API_KEY"
    fi
else
    echo "✅ API_KEY already set in .env"
fi

echo ""
echo "📦 Downloading dependencies..."
go mod download
echo "✅ Dependencies downloaded"

echo ""
echo "🔨 Building the application..."
go build -o wg-agent
echo "✅ Build complete: ./wg-agent"

echo ""
echo "✨ Setup complete!"
echo ""
echo "📚 Next steps:"
echo ""
echo "1. Start the server:"
echo "   export \$(cat .env | xargs)"
echo "   ./wg-agent"
echo ""
echo "2. Access Swagger UI:"
echo "   http://localhost:9999/swagger/"
echo ""
echo "3. Test the API:"
echo "   curl -H 'X-API-Key: YOUR_API_KEY' http://localhost:9999/health"
echo ""
echo "4. Read the documentation:"
echo "   cat QUICK_START.md"
echo ""
