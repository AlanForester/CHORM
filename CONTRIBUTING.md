# Contributing to CHORM

Thank you for your interest in contributing to CHORM! This document provides guidelines and information for contributors.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Development Setup](#development-setup)
3. [Code Style](#code-style)
4. [Testing](#testing)
5. [Pull Request Process](#pull-request-process)
6. [Feature Requests](#feature-requests)
7. [Bug Reports](#bug-reports)
8. [Documentation](#documentation)
9. [Release Process](#release-process)

## Getting Started

### Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (for testing with ClickHouse)
- Git

### Fork and Clone

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/chorm.git
   cd chorm
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/forester/chorm.git
   ```

## Development Setup

### Quick Start

1. Install development tools:
   ```bash
   make install-tools
   ```

2. Start ClickHouse with Docker:
   ```bash
   make docker-run
   ```

3. Run tests:
   ```bash
   make test
   ```

### Manual Setup

If you prefer to run ClickHouse manually:

1. Install ClickHouse following the [official documentation](https://clickhouse.com/docs/en/install)
2. Start ClickHouse server
3. Create a test database:
   ```sql
   CREATE DATABASE test;
   ```

### Environment Variables

Set these environment variables for development:

```bash
export CLICKHOUSE_HOST=localhost
export CLICKHOUSE_PORT=9000
export CLICKHOUSE_DB=test
export CLICKHOUSE_USER=default
export CLICKHOUSE_PASSWORD=""
```

## Code Style

### Go Code Style

- Follow the [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for code formatting
- Use `golint` for linting
- Follow Go naming conventions

### Struct Tags

When defining structs for ClickHouse mapping:

```go
type User struct {
    ID       uint32    `ch:"id" ch_type:"UInt32" ch_pk:"true"`
    Name     string    `ch:"name" ch_type:"String"`
    Email    string    `ch:"email" ch_type:"String"`
    Age      uint8     `ch:"age" ch_type:"UInt8"`
    Created  time.Time `ch:"created" ch_type:"DateTime"`
    IsActive bool      `ch:"is_active" ch_type:"Boolean"`
    Score    float64   `ch:"score" ch_type:"Float64"`
}
```

### Error Handling

- Always check and handle errors
- Use `fmt.Errorf` with `%w` verb for error wrapping
- Provide meaningful error messages

```go
if err != nil {
    return fmt.Errorf("failed to connect to ClickHouse: %w", err)
}
```

### Documentation

- Add comments for exported functions and types
- Use [godoc](https://golang.org/pkg/go/doc/) style comments
- Include examples in documentation

```go
// Connect creates a new connection to ClickHouse
// Example:
//   db, err := chorm.Connect(ctx, config)
//   if err != nil {
//       log.Fatal(err)
//   }
func Connect(ctx context.Context, config Config) (*DB, error) {
    // implementation
}
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with race detection
make test-race

# Run tests with coverage
make test-coverage

# Run benchmarks
make bench
```

### Writing Tests

- Write tests for all new functionality
- Use descriptive test names
- Test both success and failure cases
- Use table-driven tests for multiple scenarios

```go
func TestConnect(t *testing.T) {
    tests := []struct {
        name    string
        config  Config
        wantErr bool
    }{
        {
            name: "valid connection",
            config: Config{
                Host:     "localhost",
                Port:     9000,
                Database: "test",
            },
            wantErr: false,
        },
        {
            name: "invalid host",
            config: Config{
                Host:     "invalid-host",
                Port:     9000,
                Database: "test",
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            db, err := Connect(context.Background(), tt.config)
            if (err != nil) != tt.wantErr {
                t.Errorf("Connect() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && db == nil {
                t.Error("Connect() returned nil database")
            }
        })
    }
}
```

### Integration Tests

- Use Docker Compose for integration tests
- Test with real ClickHouse instances
- Test cluster functionality

```go
func TestIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }
    
    // Test with real ClickHouse
    db, err := Connect(context.Background(), Config{
        Host:     "localhost",
        Port:     9000,
        Database: "test",
    })
    if err != nil {
        t.Skipf("ClickHouse not available: %v", err)
    }
    defer db.Close()
    
    // Run integration tests
}
```

## Pull Request Process

### Before Submitting

1. Ensure your code follows the style guidelines
2. Run all tests and ensure they pass
3. Add tests for new functionality
4. Update documentation if needed
5. Update examples if needed

### Creating a Pull Request

1. Create a feature branch:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and commit them:
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```

3. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

4. Create a pull request on GitHub

### Commit Message Format

Use conventional commit format:

- `feat:` for new features
- `fix:` for bug fixes
- `docs:` for documentation changes
- `style:` for formatting changes
- `refactor:` for code refactoring
- `test:` for adding tests
- `chore:` for maintenance tasks

Example:
```
feat: add support for ClickHouse Array types

- Add Array type mapping in mapper
- Support Array in struct tags
- Add tests for Array functionality
- Update documentation with Array examples
```

### Pull Request Template

Use this template for pull requests:

```markdown
## Description
Brief description of the changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Examples updated (if applicable)
```

## Feature Requests

### Before Submitting

1. Check existing issues to avoid duplicates
2. Search the documentation for existing functionality
3. Consider if the feature aligns with the project goals

### Feature Request Template

```markdown
## Feature Description
Clear description of the feature

## Use Case
Why is this feature needed? What problem does it solve?

## Proposed Implementation
How would you implement this feature?

## Alternatives Considered
What other approaches were considered?

## Additional Context
Any other relevant information
```

## Bug Reports

### Before Submitting

1. Check existing issues for similar bugs
2. Try to reproduce the issue
3. Check if it's a configuration issue

### Bug Report Template

```markdown
## Bug Description
Clear description of the bug

## Steps to Reproduce
1. Step 1
2. Step 2
3. Step 3

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- Go version:
- ClickHouse version:
- Operating system:
- CHORM version:

## Additional Context
Error messages, logs, etc.
```

## Documentation

### Contributing to Documentation

1. Update API documentation for new features
2. Add examples for new functionality
3. Update README if needed
4. Add inline documentation for complex code

### Documentation Standards

- Use clear, concise language
- Include code examples
- Keep documentation up to date
- Use consistent formatting

## Release Process

### Versioning

We use [Semantic Versioning](https://semver.org/):

- MAJOR version for incompatible API changes
- MINOR version for backwards-compatible functionality
- PATCH version for backwards-compatible bug fixes

### Release Checklist

Before creating a release:

- [ ] All tests pass
- [ ] Documentation is up to date
- [ ] Examples work correctly
- [ ] CHANGELOG is updated
- [ ] Version is updated in code
- [ ] Release notes are prepared

### Creating a Release

1. Update version in `go.mod`
2. Update CHANGELOG.md
3. Create a release tag:
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```
4. Create a GitHub release with release notes

## Getting Help

If you need help with contributing:

1. Check the documentation
2. Search existing issues
3. Create a new issue for questions
4. Join our community discussions

## Code of Conduct

Please be respectful and inclusive in all interactions. We follow the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/0/code_of_conduct/).

## License

By contributing to CHORM, you agree that your contributions will be licensed under the MIT License. 