# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [1.0.0] - 2026-06-08

### Added
- Created complete, clean open-source repository templates: `LICENSE` (MIT), `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, and `SECURITY.md`.
- Added multi-stage `Dockerfile` and `docker-compose.yml` for containerized deployments.
- Added GitHub Actions continuous integration workflow (`.github/workflows/ci.yml`) to automatically build and run tests on pull requests.
- Integrated a comprehensive, modern root `README.md` with visual architecture diagrams and usage guides.

### Changed
- Rebranded the entire project from **Signal/Zignal Proxy** to **Multi-Protocol Proxy**.
- Renamed the default public proxy mode from `signal` to `sni` (representing SNI-based TLS routing).
- Rebranded the startup ASCII banner, log taglines, and error prefix markers to use `Multi-Protocol Proxy`.
- Updated Go module import paths across the entire codebase to `multi-protocol-proxy`.
- Renamed the Prometheus metrics namespace to `multiproxy` (from `signalproxy`).

### Fixed
- Fixed in-memory self-signed TLS certificate generation in unit tests to prevent file-system dependencies and panics on fresh environments.
- Corrected Go printf style warnings inside CLI theme printer functions.

---

## [0.4.0] - 2026-02-20

### Added
- Integrated user-specific bandwidth speed limits (throttling in Mbps) using token bucket rate limiting on read/write connections.
- Added monthly bandwidth usage caps (GB) with automatic calendar month usage resets.
- Added `/api/usage` endpoint to query live bandwidth usage statistics.

### Fixed
- Fixed IP-based bypass authentication logic for `super_admin` role users.

---

## [0.3.0] - 2026-02-15

### Added
- Added dynamic Proxy Auto-Config (PAC) endpoint at `/proxy.pac` with token authentication and dynamic credential injection.
- Integrated `bcrypt` password hashing for user accounts configuration database.
- Added interactive user management script `scripts/manage-users.go` to add, remove, and configure users.

---

## [0.2.0] - 2026-02-07

### Added
- Added TLS-in-TLS SNI passthrough capability for clients.
- Integrated landing page stats telemetry endpoints (`/api/stats` and `/api/history`) for active connection counting and success rate calculations.

---

## [0.1.0] - 2026-01-11

### Added
- Initial release containing the core TCP connection redirector and CLI theme output parser.
