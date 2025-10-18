# Contributing to COO-LLM

We welcome contributions from the community! Whether you're fixing bugs, adding features, improving documentation, or helping with testing, your help is appreciated.

## Ways to Contribute

- **Report Bugs**: Use the [Issue Tracker](https://github.com/user/coo-llm/issues) to report bugs.
- **Suggest Features**: Open an issue for feature requests.
- **Code Contributions**: Submit pull requests for code changes.
- **Documentation**: Help improve docs in the `docs/` directory.
- **Testing**: Write or run tests.

## Development Workflow

### Branch Naming Convention

All branches should follow this naming convention:

```
<type>/<version>/<description>
```

**Types:**
- `feat`: New features
- `fix`: Bug fixes
- `docs`: Documentation updates
- `refactor`: Code refactoring
- `test`: Testing improvements
- `chore`: Maintenance tasks

**Version:** Target version (e.g., `v1.2.x`, `v1.3.0`)

**Description:** Brief, kebab-case description

**Examples:**
- `feat/v1.2.x/update-web-ui`
- `fix/v1.1.x/rate-limit-bug`
- `docs/v1.2.x/api-reference`
- `refactor/v1.3.0/cleanup-balancer`

### Getting Started

1. Fork the repository.
2. Clone your fork: `git clone https://github.com/coo-llm/coo-llm-main.git`
3. Create a feature branch: `git checkout -b feat/v1.2.x/your-feature`
4. Set up the development environment (see [Guidelines](Contributing/Guidelines.md)).
5. Make your changes.
6. Run tests: `make test`
7. Commit with clear messages.
8. Push and submit a pull request.

## Guidelines

- Follow the [Development Guidelines](Contributing/Guidelines.md).
- Ensure code passes linting and tests.
- Write clear commit messages following conventional commits.
- Keep PRs focused on single features/bugs.

## Commit Message Format

Use conventional commits:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

**Examples:**
- `feat(balancer): add hybrid scoring algorithm`
- `fix(api): resolve rate limiting issue`
- `docs(api): update stats endpoint reference`

## Resources

- [GitHub Repository](https://github.com/user/coo-llm)
- [Issue Tracker](https://github.com/user/coo-llm/issues)
- [Discussions](https://github.com/user/coo-llm/discussions)
- [Changelog](Contributing/Changelog.md)