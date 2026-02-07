# Contributing Guide

Thank you for your interest in contributing to the WireGuard Agent project! This guide will help you get started.

## Code of Conduct

Be respectful and inclusive. We are committed to providing a welcoming and inspiring community for all.

## How to Contribute

### Reporting Bugs

1. Check if the issue already exists on [GitHub Issues](https://github.com/afrisinc/wg-agent/issues)
2. Create a new issue with:
   - Clear, descriptive title
   - Description of the problem
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant logs or error messages
   - System information (OS, Go version, etc.)

### Suggesting Features

1. Check existing [GitHub Issues](https://github.com/afrisinc/wg-agent/issues) for similar proposals
2. Create an issue describing:
   - The use case and motivation
   - Proposed solution
   - Alternatives considered
   - Any implementation concerns

### Pull Requests

1. **Fork and clone** the repository
   ```bash
   git clone https://github.com/afrisinc/wg-agent.git
   cd wg-agent
   ```

2. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make your changes**
   - Write clean, maintainable code
   - Follow Go style guidelines
   - Add tests for new functionality
   - Update documentation as needed

4. **Run quality checks**
   ```bash
   make fmt      # Format code
   make lint     # Run linters
   make test     # Run tests
   ```

5. **Commit with clear messages**
   ```bash
   git commit -m "feat: add new feature

   Detailed description of the changes made."
   ```

   Use conventional commit format:
   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation changes
   - `test:` - Test additions/modifications
   - `refactor:` - Code refactoring
   - `perf:` - Performance improvements
   - `chore:` - Maintenance tasks

6. **Push and create a Pull Request**
   ```bash
   git push origin feature/your-feature-name
   ```

7. **Fill out the PR template** with:
   - Description of changes
   - Related issues
   - Test plan
   - Screenshots (if applicable)

## Development Setup

### Prerequisites

- Go 1.21+
- Docker and Docker Compose (optional)
- Make (optional)

### Local Development

```bash
# Clone repository
git clone https://github.com/afrisinc/wg-agent.git
cd wg-agent

# Download dependencies
go mod download

# Run tests
go test -v ./...

# Run the agent
API_KEY="test-key" go run main.go

# Format and lint
go fmt ./...
golangci-lint run ./...
```

### Using Make

```bash
# View available commands
make help

# Format, lint, test, and build
make all

# Run tests with coverage
make test-coverage

# Build Docker image
make docker-build

# Start Docker Compose
make docker-up
```

## Testing

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run with coverage
go test -v -cover ./...

# Run specific test
go test -v -run TestAddPeerHandler ./...

# Run with race detector
go test -race ./...
```

### Writing Tests

- Place test files alongside source code with `_test.go` suffix
- Use table-driven tests for multiple test cases
- Test both happy paths and error cases
- Mock external dependencies
- Aim for >80% code coverage

Example:
```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {
            name:    "valid input",
            input:   "test",
            want:    "expected",
            wantErr: false,
        },
        {
            name:    "invalid input",
            input:   "bad",
            want:    "",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("wantErr %v, got error %v", tt.wantErr, err)
            }
            if got != tt.want {
                t.Errorf("got %s, want %s", got, tt.want)
            }
        })
    }
}
```

## Code Style

### Go Style Guide

Follow the [Effective Go](https://golang.org/doc/effective_go) guidelines:

- Use camelCase for variables and functions
- Use PascalCase for exported identifiers
- Keep lines under 100 characters (soft limit)
- Use meaningful variable names
- Write comments for exported functions
- Avoid naked returns
- Handle errors explicitly

### Code Organization

```
project/
├── main.go              # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   └── wireguard/      # WireGuard operations
├── .github/
│   └── workflows/      # CI/CD pipelines
├── tests/              # Integration tests
└── docs/               # Documentation
```

### Comments

- Exported functions should have a comment explaining their purpose
- Complex logic should have explanatory comments
- Avoid obvious comments (e.g., `i++  // increment i`)
- Keep comments updated with code changes

## Documentation

### Code Documentation

- Document exported packages with a comment
- Document exported functions with their purpose
- Include examples in comments when helpful
- Keep README.md updated with new features

### Commit Messages

Good commit message example:
```
fix: handle empty public key validation

- Add check for empty string before regex validation
- Add test case for empty key
- Fixes #42
```

## Performance Considerations

- Profile code before optimizing
- Use benchmarks for performance-critical code
- Avoid allocations in hot paths
- Document performance implications of changes

## Security

### Security Checklist

Before submitting code:
- [ ] No hardcoded secrets or credentials
- [ ] Input validation on all external data
- [ ] Proper error handling (no info leaks)
- [ ] No use of deprecated crypto functions
- [ ] Consider OWASP top 10
- [ ] Reviewed for SQL injection, XSS, command injection risks

### Reporting Security Issues

Do NOT open a public issue for security vulnerabilities. Please email security concerns to your security contact.

## Release Process

### Version Numbering

Follow [Semantic Versioning](https://semver.org/):
- MAJOR.MINOR.PATCH
- Increment MAJOR for incompatible API changes
- Increment MINOR for new features (backwards compatible)
- Increment PATCH for bug fixes

### Creating a Release

1. Update CHANGELOG.md
2. Update version in files (if needed)
3. Create and push tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
4. GitHub Actions will automatically:
   - Build binaries
   - Create release
   - Build and push Docker image

## Getting Help

- **Documentation**: Check [README.md](README.md), [DEPLOYMENT.md](DEPLOYMENT.md), [TROUBLESHOOTING.md](TROUBLESHOOTING.md)
- **Issues**: Search existing [GitHub Issues](https://github.com/afrisinc/wg-agent/issues)
- **Discussions**: Start a discussion for design feedback
- **Email**: Contact maintainers for sensitive topics

## Thank You!

We appreciate all contributions, whether they are:
- Code changes
- Bug reports
- Feature requests
- Documentation improvements
- Testing feedback
- Security reports

Thank you for making WireGuard Agent better!
