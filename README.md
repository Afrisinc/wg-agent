# WireGuard Agent

A production-ready HTTP API for managing WireGuard peers. This agent provides secure REST endpoints for adding, removing, and monitoring WireGuard VPN peers.

## Features

- ✅ Secure API with key-based authentication
- ✅ Input validation and error handling
- ✅ Context-aware command execution with timeouts
- ✅ Docker support with multi-stage builds
- ✅ Health check and readiness probes
- ✅ Comprehensive test coverage
- ✅ GitHub Actions CI/CD pipeline
- ✅ Security scanning (Trivy, gosec, staticcheck)
- ✅ Graceful shutdown handling
- ✅ Structured logging

## Quick Start

### Prerequisites

- Go 1.21+
- WireGuard installed and configured
- Docker (optional)

### Running Locally

1. **Clone the repository**
   ```bash
   git clone https://github.com/afrisinc/wg-agent.git
   cd wg-agent
   ```

2. **Set environment variables**
   ```bash
   export API_KEY="your-secure-api-key"
   export HOST="0.0.0.0"
   export PORT="9999"
   export WG_INTERFACE="wg0"
   export PUBLIC_KEY_PATH="/etc/wireguard/publickey"
   ```

3. **Run the agent**
   ```bash
   go run main.go
   ```

   The agent will start on `http://localhost:9999`

### Running with Docker

1. **Build the image**
   ```bash
   docker build -t wg-agent:latest .
   ```

2. **Run the container**
   ```bash
   docker run -d \
     --name wg-agent \
     --cap-add NET_ADMIN \
     --cap-add SYS_MODULE \
     -p 9999:9999 \
     -v /etc/wireguard:/etc/wireguard:ro \
     -e API_KEY="your-secure-api-key" \
     wg-agent:latest
   ```

3. **Using docker-compose**
   ```bash
   docker-compose up -d
   ```

## API Endpoints

### Health Checks

#### GET /health
Returns the health status of the agent.

```bash
curl http://localhost:9999/health
```

Response:
```json
{
  "status": "healthy"
}
```

#### GET /ready
Returns the readiness status (checks if WireGuard interface is available).

```bash
curl http://localhost:9999/ready
```

### Peer Management (Requires Authentication)

All endpoints below require the `X-API-Key` header.

#### POST /add-peer
Add a new peer to the WireGuard interface.

```bash
curl -X POST http://localhost:9999/add-peer \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "public_key": "jI6DsucHvzJzcow3v7CqvJODct9+pWG8V+MlaWL7yGc=",
    "allowed_ip": "192.168.1.100"
  }'
```

Request schema:
```json
{
  "public_key": "string (required)",
  "allowed_ip": "string (required)",
  "endpoint": "string (optional)"
}
```

Response:
```json
{
  "status": "success",
  "message": "Peer jI6DsucH... added successfully"
}
```

#### POST /remove-peer
Remove a peer from the WireGuard interface.

```bash
curl -X POST http://localhost:9999/remove-peer \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "public_key": "jI6DsucHvzJzcow3v7CqvJODct9+pWG8V+MlaWL7yGc="
  }'
```

Response:
```json
{
  "status": "success",
  "message": "Peer jI6DsucH... removed successfully"
}
```

#### GET /status
Get the current status of the WireGuard interface.

```bash
curl -H "X-API-Key: your-api-key" http://localhost:9999/status
```

Response: Plain text WireGuard status output

#### GET /public-key
Get the public key of the WireGuard interface.

```bash
curl -H "X-API-Key: your-api-key" http://localhost:9999/public-key
```

Response:
```json
{
  "public_key": "jI6DsucHvzJzcow3v7CqvJODct9+pWG8V+MlaWL7yGc="
}
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `API_KEY` | (required) | API key for authentication |
| `HOST` | `0.0.0.0` | Server host address |
| `PORT` | `9999` | Server port |
| `WG_INTERFACE` | `wg0` | WireGuard interface name |
| `PUBLIC_KEY_PATH` | `/etc/wireguard/publickey` | Path to public key file |
| `LOG_LEVEL` | `info` | Logging level |

### Security Considerations

1. **API Key**: Use a strong, randomly generated API key. Use a secure method to distribute it (e.g., HashiCorp Vault, AWS Secrets Manager)
2. **HTTPS**: Always use HTTPS in production. Consider using a reverse proxy (nginx, Caddy) for TLS termination
3. **Network Policy**: Restrict network access to the agent's port
4. **Audit Logging**: Monitor API access logs for suspicious activity
5. **Resource Limits**: Configure appropriate CPU and memory limits when running in containers

## Testing

Run the test suite:

```bash
go test -v -race -cover ./...
```

Run tests with coverage report:

```bash
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Deployment

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment instructions.

### Production Checklist

- [ ] Set strong, randomly generated API_KEY
- [ ] Configure HTTPS/TLS
- [ ] Set up proper logging and monitoring
- [ ] Configure resource limits (CPU, memory)
- [ ] Enable health checks in your orchestration system
- [ ] Set up alerting for failed requests
- [ ] Test recovery procedures
- [ ] Document runbooks for common operations

## CI/CD Pipeline

This project includes GitHub Actions workflows for:

- **CI/CD** (`ci.yml`): Tests, linting, building, and security scanning
- **Docker Build** (`docker.yml`): Building and pushing Docker images
- **Release** (`release.yml`): Creating releases with binary artifacts

### Workflow Triggers

- **CI/CD**: Push to main/develop, pull requests
- **Docker Build**: Push to main, tag creation
- **Release**: Tag creation (v*)

## Security

This project includes multiple security scanning tools:

- **Trivy**: Container and filesystem vulnerability scanning
- **gosec**: Go security checker
- **staticcheck**: Static analysis for Go
- **golangci-lint**: Combined Go linter

All security scans run automatically in CI/CD pipelines.

## Contributing

1. Create a feature branch
2. Make your changes
3. Run tests: `go test -v ./...`
4. Run linting: `golangci-lint run`
5. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

- Report issues on [GitHub Issues](https://github.com/afrisinc/wg-agent/issues)
- Check [TROUBLESHOOTING.md](TROUBLESHOOTING.md) for common problems

## Roadmap

- [ ] Multi-interface support
- [ ] Peer management dashboard
- [ ] Peer usage statistics
- [ ] Automatic peer cleanup
- [ ] Integration with LDAP/Active Directory
- [ ] Event webhooks
