---
sidebar_position: 5
tags: [developer-guide, testing]
---

# Testing

Testing guide for COO-LLM development.

## Running Tests

### Unit Tests

```bash
# Run all tests
go test ./...

# Run specific package
go test ./internal/config

# Run with verbose output
go test -v ./internal/balancer

# Run with race detection
go test -race ./...
```

### Integration Tests

```bash
# Run e2e tests (requires running server)
go test -tags=e2e ./test

# With environment setup
export COO_TEST_MODE=true
go test -tags=e2e ./test
```

### Web UI Tests

```bash
cd webui
npm test

# With coverage
npm test -- --coverage
```

## Test Structure

### Unit Tests

- **Config**: `internal/config/config_test.go`
- **Balancer**: `internal/balancer/selector_test.go`
- **API**: `internal/api/*_test.go`
- **Providers**: `internal/provider/*_test.go`

### Test Categories

- **Unit tests**: Test individual functions/components
- **Integration tests**: Test component interactions
- **E2E tests**: Full request flow testing

## Writing Tests

### Basic Unit Test

```go
func TestConfigValidation(t *testing.T) {
    cfg := &config.Config{
        Version: "1.0",
        Server: config.Server{
            Listen: ":2906",
            AdminAPIKey: "test-key",
        },
    }

    err := config.ValidateConfig(cfg)
    assert.NoError(t, err)
}
```

### Table-Driven Tests

```go
func TestSelectorAlgorithm(t *testing.T) {
    tests := []struct {
        name     string
        cfg      *config.Config
        expected string
    }{
        {
            name: "round_robin",
            cfg: &config.Config{
                Policy: config.Policy{Algorithm: "round_robin"},
            },
            expected: "key1",
        },
        {
            name: "least_loaded",
            cfg: &config.Config{
                Policy: config.Policy{Algorithm: "least_loaded"},
            },
            expected: "key2",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Mocking Dependencies

```go
type mockStore struct {
    data map[string]float64
}

func (m *mockStore) GetUsage(provider, key, metric string) (float64, error) {
    return m.data[fmt.Sprintf("%s:%s:%s", provider, key, metric)], nil
}

func TestCalculateScore(t *testing.T) {
    store := &mockStore{
        data: map[string]float64{
            "openai:key1:req": 10,
        },
    }

    selector := balancer.NewSelector(&config.Config{}, store)
    // Test scoring logic
}
```

## Test Coverage

### Generate Coverage Report

```bash
# HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# Text summary
go test -cover ./...
```

### Coverage Goals

- **Unit tests**: >80% coverage
- **Critical paths**: 100% coverage (auth, balancing, rate limiting)
- **Integration**: Key user flows covered

## CI/CD Testing

### GitHub Actions Workflow

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test -race -cover ./...
      - run: go build ./cmd/coo-llm
```

### Pre-commit Hooks

```bash
# Install pre-commit
pip install pre-commit

# Run hooks
pre-commit run --all-files
```

## Performance Testing

### Load Testing with Vegeta

```bash
echo "POST http://localhost:2906/api/v1/chat/completions
Authorization: Bearer test-key
Content-Type: application/json

{
  \"model\": \"openai:gpt-4o\",
  \"messages\": [{\"role\": \"user\", \"content\": \"Hello\"}]
}" | vegeta attack -rate=10 -duration=30s | vegeta report
```

### Benchmark Tests

```go
func BenchmarkChatCompletion(b *testing.B) {
    // Setup
    cfg := &config.Config{ /* ... */ }
    selector := balancer.NewSelector(cfg, store)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // Benchmark code
    }
}
```

## Debugging Tests

### Verbose Output

```bash
go test -v -run TestSpecificFunction
```

### Debug Logging

```go
// In test
t.Logf("Debug info: %+v", variable)

// Enable debug logs
os.Setenv("LOG_LEVEL", "debug")
```

### Test Timeouts

```go
func TestSlowOperation(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping slow test in short mode")
    }

    done := make(chan bool)
    go func() {
        // Slow operation
        done <- true
    }()

    select {
    case <-done:
        // Success
    case <-time.After(30 * time.Second):
        t.Fatal("Test timed out")
    }
}
```

## Best Practices

### Test Organization

1. **One assertion per test**: Keep tests focused
2. **Descriptive names**: `TestCalculateScore_EmptyWeights`
3. **Cleanup**: Use `t.Cleanup()` for resource cleanup
4. **Parallel tests**: Use `t.Parallel()` when safe

### Mocking

1. **Interface-based**: Mock via interfaces, not concrete types
2. **Minimal mocks**: Only mock what's necessary
3. **Real dependencies**: Use real DB for integration tests

### CI Best Practices

1. **Fast feedback**: Run unit tests first
2. **Fail fast**: Stop on first failure in CI
3. **Cache dependencies**: Speed up builds
4. **Matrix testing**: Test multiple Go versions

## Contributing Tests

When adding new features:

1. **Write tests first**: TDD approach
2. **Cover edge cases**: Error conditions, boundary values
3. **Update existing tests**: Don't break backward compatibility
4. **Document test scenarios**: Comments explaining test purpose

## Troubleshooting Test Issues

### Flaky Tests

- **Race conditions**: Use `-race` flag
- **Timing issues**: Avoid `time.Sleep()`, use channels
- **External dependencies**: Mock external APIs

### Common Errors

- **Import cycles**: Refactor to break cycles
- **Missing dependencies**: Add to `go.mod`
- **Test data**: Use consistent test fixtures