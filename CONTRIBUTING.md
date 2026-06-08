# Contributing to Multi-Protocol Proxy

First off, thank you for taking the time to contribute! 🎉

This project is an open-source, highly modular, multi-protocol proxy server. We welcome contributions of all kinds: bug fixes, new features, documentation improvements, and feedback.

Here are the guidelines to help you get started.

---

## Code of Conduct

By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). Please report any unacceptable behavior to the project maintainers.

## How Can I Contribute?

### Reporting Bugs
If you find a bug, please open an issue on GitHub. Include:
- A clear, descriptive title.
- Steps to reproduce the issue.
- Expected vs. actual behavior.
- System information (OS, Go version).
- Relevant logs or command output.

### Suggesting Enhancements
We are always looking to improve! If you have ideas for new features or optimizations:
- Open a GitHub issue and explain the use case.
- Discuss your design before starting code implementation.

### Submitting Pull Requests
1. **Fork the repository** and create your branch from `main`.
2. **Make your changes** in a new branch (e.g., `feature/cool-new-protocol` or `fix/connection-leak`).
3. **Write/update unit tests** for your changes.
4. **Ensure the code builds** and all tests pass locally.
5. **Run Go formatting and linting**:
   ```bash
   gofmt -s -w .
   go vet ./...
   ```
6. **Submit a Pull Request** to the main repository, referencing any related issues.

---

## Local Development Setup

To get set up for local development:

1. **Prerequisites**:
   - Go 1.25 or higher installed.
   - Git.

2. **Clone the Repository**:
   ```bash
   git clone https://github.com/tamecalm/multi-protocol-proxy.git
   cd multi-protocol-proxy
   ```

3. **Install Dependencies**:
   ```bash
   ./scripts/start.sh install
   ```

4. **Run Tests**:
   ```bash
   go test -v ./...
   ```

5. **Build and Run (Development Mode)**:
   ```bash
   # Copies the environment template
   cp env.development.example .env
   
   # Builds and launches the proxy
   ./scripts/start.sh dev
   ```

---

## Code Style Guidelines

- **Standard Go Formatting**: We strictly enforce Go formatting. Make sure you run `gofmt` before committing.
- **Explicit Context**: Always pass `context.Context` to network connections and long-running routines to allow graceful shutdown.
- **Metrics & Instrumentation**: If you add new components, register appropriate Prometheus metrics inside your package or `internal/proxy/metrics.go`.
- **Graceful Error Handling**: Do not panic in production code. Always return error bubbles and log failures using the `internal/ui` logger.
