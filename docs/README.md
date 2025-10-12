# TruckLLM Documentation

This directory contains the documentation for TruckLLM, built with [Docusaurus](https://docusaurus.io/).

## Structure

```
docs/
├── intro/              # Introduction and overview
│   ├── intro.md       # Main intro page
│   ├── overview.md    # Detailed overview
│   └── architecture.md # System architecture
├── guides/            # User guides
│   ├── configuration.md
│   ├── deployment.md
│   └── providers.md
├── reference/         # Technical reference
│   ├── api.md
│   ├── balancer.md
│   ├── storage.md
│   └── logging.md
├── contributing/      # Contributor docs
│   ├── guidelines.md
│   └── changelog.md
├── assets/            # Images and media
├── .gitignore
└── README.md
```

## Local Development

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Start development server:**
   ```bash
   npm start
   ```

3. **Build for production:**
   ```bash
   npm run build
   ```

4. **Serve production build:**
   ```bash
   npm run serve
   ```

## Writing Documentation

### Frontmatter

Each markdown file should include frontmatter for Docusaurus:

```yaml
---
sidebar_position: 1
---

# Page Title
```

### Links

Use relative links for internal documentation:

```markdown
[Configuration](../guides/configuration.md)
```

### Code Blocks

Use appropriate language for syntax highlighting:

```go
func example() {
    // Go code
}
```

### Images

Place images in the `assets/` directory:

```markdown
![Architecture](assets/architecture.png)
```

## Deployment

The documentation is automatically deployed via GitHub Actions when changes are pushed to the main branch.

For manual deployment:

```bash
npm run build
npm run deploy
```

## Contributing

When adding new documentation:

1. Follow the existing structure
2. Add appropriate frontmatter
3. Update navigation if needed
4. Test locally before committing
5. Update this README if adding new sections