# Quick Start Guide

## Prerequisites

- Go 1.21+ installed
- WireGuard installed on your system
- API key generated

## Installation

### 1. Generate API Key

```bash
openssl rand -base64 32
# Example output: jI6DsucHvzJzcow3v7CqvJODct9+pWG8V+MlaWL7yGc=
```

### 2. Create .env File

```bash
cp .env.example .env
```

Edit `.env` and set:
```env
API_KEY=your-generated-api-key-here
HOST=0.0.0.0
PORT=9999
WG_INTERFACE=wg0
```

### 3. Run the Server

#### Option A: Direct (Linux/macOS)

```bash
# Download dependencies
go mod download

# Build the binary
go build -o wg-agent

# Run with environment variables
export $(cat .env | xargs)
./wg-agent
```

#### Option B: Using Go Run

```bash
export $(cat .env | xargs)
go run main.go
```

#### Option C: Docker

```bash
# Build Docker image
docker build -t wg-agent:latest .

# Run container
docker run -d \
  --name wg-agent \
  --cap-add NET_ADMIN \
  --cap-add SYS_MODULE \
  -p 9999:9999 \
  -v /etc/wireguard:/etc/wireguard:ro \
  --env-file .env \
  wg-agent:latest

# View logs
docker logs -f wg-agent
```

#### Option D: Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f wg-agent

# Stop
docker-compose down
```

### 4. Verify it's Running

```bash
# Check health
curl http://localhost:9999/health

# Expected response:
# {"status":"healthy"}
```

## Access Swagger Documentation

Open your browser and navigate to:

```
http://localhost:9999/swagger/
```

You'll see the interactive Swagger UI where you can:
- View all available endpoints
- See request/response schemas
- Try out API calls directly from the browser

## Test API Endpoints

### Get Public Key

```bash
curl -H "X-API-Key: YOUR_API_KEY" \
  http://localhost:9999/public-key
```

### Add a Peer

```bash
curl -X POST http://localhost:9999/add-peer \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "jI6DsucHvzJzcow3v7CqvJODct9+pWG8V+MlaWL7yGc=",
    "allowed_ip": "192.168.1.100"
  }'
```

### Check Status

```bash
curl -H "X-API-Key: YOUR_API_KEY" \
  http://localhost:9999/status
```

### Remove a Peer

```bash
curl -X POST http://localhost:9999/remove-peer \
  -H "X-API-Key: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "public_key": "jI6DsucHvzJzcow3v7CqvJODct9+pWG8V+MlaWL7yGc="
  }'
```

## Troubleshooting

### "Failed to load config: API_KEY environment variable not set"

Make sure you've set the API_KEY environment variable:

```bash
export API_KEY="your-api-key"
# or
export $(cat .env | xargs)
```

### "WireGuard interface not found"

Ensure WireGuard is installed and the interface exists:

```bash
# Check if WireGuard is installed
which wg wg-quick

# List interfaces
ip link show

# Create interface if needed
sudo ip link add dev wg0 type wireguard
sudo ip addr add 10.0.0.1/24 dev wg0
sudo ip link set wg0 up
```

### Port Already in Use

If port 9999 is already in use:

```bash
# Using a different port
export PORT=8080
./wg-agent

# Then access Swagger at: http://localhost:8080/swagger/
```

## Development

### Run Tests

```bash
go test -v ./...
```

### Run Tests with Coverage

```bash
go test -v -cover ./...
```

### Format Code

```bash
go fmt ./...
```

### Run Linters

```bash
golangci-lint run ./...
```

### Build Optimized Binary

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o wg-agent \
  -ldflags="-w -s -X main.Version=v1.0.0"
```

## Useful Make Commands

If you have `make` installed:

```bash
make help           # Show all available commands
make build          # Build the binary
make test           # Run tests
make lint           # Run linters
make docker-build   # Build Docker image
make docker-up      # Start Docker Compose
make docker-down    # Stop Docker Compose
make run            # Run locally
make coverage       # Generate coverage report
```

## Next Steps

1. ✅ Server is running
2. ✅ Swagger UI is accessible
3. Read the [README.md](README.md) for detailed documentation
4. Check [DEPLOYMENT.md](DEPLOYMENT.md) for production deployment
5. Review [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common issues

## API Documentation

The complete API documentation is available at:
- **Swagger UI**: http://localhost:9999/swagger/
- **Swagger JSON**: http://localhost:9999/swagger.json
- **README Endpoints**: See [README.md](README.md#api-endpoints)

## Default Credentials

For testing purposes only, use:

```bash
API_KEY=test-secret-key-123
HOST=0.0.0.0
PORT=9999
```

⚠️ **IMPORTANT**: Change these in production!

## Security Notes

1. Keep `API_KEY` secret and secure
2. Use HTTPS in production (via reverse proxy)
3. Restrict network access to the agent
4. Regularly rotate API keys
5. Monitor logs for suspicious activity

## Getting Help

- Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
- Review logs: `journalctl -u wg-agent -f`
- Check GitHub issues: https://github.com/afrisinc/wg-agent/issues
- Read [CONTRIBUTING.md](CONTRIBUTING.md) for development

---

**Ready to go!** 🚀

Start the server and visit http://localhost:9999/swagger/ to explore the API.
