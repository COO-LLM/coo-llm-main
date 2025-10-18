---
sidebar_position: 6
tags: [developer-guide, contributing]
---

# Contributing

Thank you for your interest in contributing to COO-LLM! This document provides guidelines and information for contributors.

## Code of Conduct

This project follows a code of conduct to ensure a welcoming environment for all contributors. By participating, you agree to:

- Be respectful and inclusive
- Focus on constructive feedback
- Accept responsibility for mistakes
- Show empathy towards other contributors
- Help create a positive community

## Getting Started

### Development Environment

1. **Prerequisites:**
   - Go 1.21 or later
   - Docker and Docker Compose
   - Git
   - Make

2. **Clone the repository:**
   ```bash
   git clone https://github.com/user/coo-llm.git
   cd coo-llm
   ```

3. **Install dependencies:**
   ```bash
   go mod download
   ```

4. **Run tests:**
   ```bash
   make test
   ```

5. **Build and run:**
   ```bash
   make build
   make run
   ```

### Project Structure

```
coo-llm/
├── cmd/                    # Application entrypoints
│   └── coo-llm/       # Main application
├── internal/              # Private application code
│   ├── api/              # HTTP API handlers
│   ├── balancer/         # Load balancing logic
│   ├── config/           # Configuration management
│   ├── log/              # Logging system
│   ├── provider/         # LLM provider adapters
│   └── store/            # Storage backends
├── pkg/                   # Public packages
├── configs/               # Configuration files
├── docs/                  # Documentation site
├── test/                  # Integration tests
├── docker-compose.yml     # Docker setup
├── Dockerfile            # Container build
├── Makefile              # Build automation
├── go.mod                # Go modules
└── README.md             # Project overview
```

## Development Workflow

### 1. Choose an Issue

- Check [GitHub Issues](https://github.com/user/coo-llm/issues) for open tasks
- Look for issues labeled `good first issue` or `help wanted`
- Comment on the issue to indicate you're working on it

### 2. Create a Branch

Follow the branch naming convention:

```bash
git checkout -b feat/v1.2.x/your-feature-name
# or
git checkout -b fix/v1.1.x/issue-description
```

See [CONTRIBUTING.md](https://github.com/your-org/coo-llm/blob/main/CONTRIBUTING.md) for full branch naming guidelines.

### 3. Make Changes

- Write clear, concise commit messages
- Follow the existing code style
- Add tests for new functionality
- Update documentation as needed

### 4. Test Your Changes

```bash
# Run unit tests
go test ./...

# Run integration tests
go test ./test/...

# Run linting
make lint

# Build the project
make build
```

### 5. Submit a Pull Request

- Push your branch to GitHub
- Create a pull request with a clear description
- Reference any related issues
- Wait for review and address feedback

## Coding Standards

### Go Code Style

- Follow standard Go formatting (`go fmt`)
- Use `gofmt -s` for additional simplifications
- Run `go vet` to check for common mistakes
- Use `golint` for style issues

### Naming Conventions

- Use descriptive names for variables and functions
- Follow Go naming conventions (camelCase for private, PascalCase for public)
- Use consistent naming patterns

### Code Organization

- Keep functions small and focused
- Use interfaces for abstraction
- Separate concerns properly
- Add comments for complex logic

### Error Handling

- Return errors instead of panicking
- Use error wrapping for context
- Handle errors at appropriate levels
- Log errors with sufficient context

## Testing

### Unit Tests

- Place test files alongside source code (`*_test.go`)
- Use table-driven tests for multiple test cases
- Mock external dependencies
- Aim for high coverage (>80%)

```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {"valid input", "hello", "HELLO", false},
        {"empty input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := Function(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.expected, result)
            }
        })
    }
}
```

### Integration Tests

- Place integration tests in `test/` directory
- Test complete workflows
- Use real dependencies where possible
- Clean up resources after tests

### Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# Coverage summary
go tool cover -func=coverage.out
```

## Documentation

### Code Documentation

- Add package comments for exported packages
- Document exported functions, types, and methods
- Use examples in documentation

```go
// Package provider contains LLM provider implementations.
//
// Example:
//
//	p, err := provider.NewOpenAIProvider(config)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	resp, err := p.Generate(ctx, req)
package provider

// Generate sends a request to the LLM provider and returns the response.
// It handles authentication, request formatting, and response parsing.
func (p *OpenAIProvider) Generate(ctx context.Context, req *Request) (*Response, error) {
    // implementation
}
```

### API Documentation

- Document REST API endpoints
- Include request/response examples
- Specify error conditions

### README Updates

- Update README.md for significant changes
- Add examples for new features
- Update installation instructions

## Adding New Providers

1. **Implement the Provider interface:**

```go
type CustomProvider struct {
    cfg *config.Provider
}

func NewCustomProvider(cfg *config.Provider) *CustomProvider {
    return &CustomProvider{cfg: cfg}
}

func (p *CustomProvider) Name() string {
    return p.cfg.ID
}

func (p *CustomProvider) Generate(ctx context.Context, req *Request) (*Response, error) {
    // Implementation
}

func (p *CustomProvider) ListModels(ctx context.Context) ([]string, error) {
    // Implementation
}
```

2. **Register in the provider registry:**

```go
func (r *Registry) LoadFromConfig(cfg *config.Config) error {
    // ... existing providers ...
    case "custom":
        p = NewCustomProvider(&pCfg)
    // ...
}
```

3. **Add configuration example:**

```yaml
providers:
  - id: custom
    name: "Custom Provider"
    base_url: "https://api.custom.com/v1"
    keys:
      - secret: "${CUSTOM_API_KEY}"
        pricing:
          input_token_cost: 0.001
          output_token_cost: 0.002
```

4. **Add tests and documentation**

## Adding Storage Backends

1. **Implement the storage interface:**

```go
type CustomStore struct {
    // fields
}

func (s *CustomStore) GetUsage(provider, keyID, metric string) (float64, error) {
    // Implementation
}

func (s *CustomStore) SetUsage(provider, keyID, metric string, value float64) error {
    // Implementation
}

func (s *CustomStore) IncrementUsage(provider, keyID, metric string, delta float64) error {
    // Implementation
}
```

2. **Add to main application:**

```go
func initStorage(cfg *config.Config) store.RuntimeStore {
    switch cfg.Storage.Runtime.Type {
    case "custom":
        return NewCustomStore(cfg.Storage.Runtime)
    // ... other cases ...
    }
}
```

## Pull Request Process

### Before Submitting

- [ ] Tests pass (`make test`)
- [ ] Code is formatted (`go fmt`)
- [ ] No linting errors (`make lint`)
- [ ] Documentation updated
- [ ] Commit messages are clear

### PR Template

```markdown
## Description
Brief description of the changes.

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing performed

## Checklist
- [ ] Code follows project style
- [ ] Documentation updated
- [ ] Tests pass
- [ ] No breaking changes
```

### Review Process

1. **Automated Checks:** CI runs tests and linting
2. **Code Review:** Maintainers review code changes
3. **Testing:** Additional testing may be requested
4. **Approval:** PR approved and merged

## Release Process

### Versioning

COO-LLM follows [Semantic Versioning](https://semver.org/):

- **MAJOR:** Breaking changes
- **MINOR:** New features
- **PATCH:** Bug fixes

### Release Checklist

- [ ] Update version in code
- [ ] Update CHANGELOG.md
- [ ] Create git tag
- [ ] Build and test release artifacts
- [ ] Publish to package registry
- [ ] Update documentation

## Community

### Communication

- **GitHub Issues:** Bug reports and feature requests
- **Discussions:** General questions and ideas
- **Pull Requests:** Code contributions

### Getting Help

- Check existing issues and documentation
- Ask questions in GitHub Discussions
- Join our community chat (if available)

### Recognition

Contributors are recognized in:
- CHANGELOG.md for releases
- GitHub contributor statistics
- Release notes

## License

By contributing to COO-LLM, you agree that your contributions will be licensed under the same license as the project (MIT License).