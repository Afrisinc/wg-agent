# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Initial project setup
- Production-ready agent with security features
- Docker support with multi-stage builds
- GitHub Actions CI/CD pipeline
- Comprehensive documentation
- Unit tests with high coverage
- Security scanning (Trivy, gosec, staticcheck)
- Health check and readiness probes
- Graceful shutdown handling
- Input validation for public keys and IP addresses

### Changed
- Refactored original code for production readiness
- Improved error handling and logging
- Added context support for command execution

### Fixed
- Security vulnerabilities in command execution
- Missing input validation

## [1.0.0] - 2024-02-07

### Added
- Initial release
- `/add-peer` endpoint for adding WireGuard peers
- `/remove-peer` endpoint for removing peers
- `/status` endpoint for viewing WireGuard status
- `/public-key` endpoint for retrieving the server's public key
- `/health` endpoint for health checks
- `/ready` endpoint for readiness checks
- API key authentication
- Configuration management via environment variables
- Docker support
- Docker Compose for local development
- Kubernetes deployment examples
- Systemd service file for standalone deployments
- Comprehensive documentation
- Unit tests
- GitHub Actions workflows

### Security
- Input validation for public keys and IP addresses
- API key authentication on protected endpoints
- Non-root container user
- Read-only filesystem support
- Security capability restrictions
- HTTPS/TLS support via reverse proxy documentation
- Rate limiting recommendations

## Future Releases

### Planned for 2.0.0
- Multi-interface support
- Peer management dashboard
- Peer usage statistics and monitoring
- Automatic peer cleanup based on inactivity
- LDAP/Active Directory integration
- Event webhooks for peer changes
- Metrics endpoint for Prometheus
- gRPC API support

### Planned for 2.1.0
- Batch operations for adding/removing peers
- Peer metadata storage
- Backup and restore functionality
- Peer bandwidth limiting
- Advanced access control lists

### Under consideration
- Web UI for peer management
- Mobile app companion
- Terraform provider
- Ansible playbook
- Helm chart
