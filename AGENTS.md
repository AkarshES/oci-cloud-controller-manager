# Repository Guidelines

## Project Structure & Module Organization
Core Go code lives in `cmd/` (entrypoints), `pkg/` (shared libraries), `controllers/` (controller-runtime logic), and `api/` (CRD types). Deployment assets are under `manifests/`, with helper docs in `docs/` and operational scripts in `hack/` and `scripts/`. Built artifacts are written to `dist/` (and `dist/arm/` for ARM builds). End-to-end test assets and images are in `test/` and `images/e2e-tests/`.

## Build, Test, and Development Commands
- `make check`: runs formatting/lint/vet gates (`gofmt`, `golint`, `govet`).
- `make test`: runs unit tests across `cmd`, `controllers`, and `pkg` via `hack/test.sh`.
- `make build`: compiles all components to `dist/`.
- `make build-arm`: cross-builds ARM binaries to `dist/arm/`.
- `make manifests`: copies and version-stamps YAMLs into `dist/`.
- `make run-ccm-dev`: runs CCM locally against a kubeconfig/cloud config.
- `make run-ccm-e2e-tests-local`: executes the local E2E driver script.

Example: `make build COMPONENT="oci-cloud-controller-manager oci-csi-node-driver"`.

## Coding Style & Naming Conventions
Use Go defaults and keep code `gofmt`-clean. `.editorconfig` enforces tabs for `Makefile`/Go, 2-space YAML, and 4-space shell scripts. Follow Go naming (`CamelCase` exported, `camelCase` internal), and keep package names short/lowercase. Run `make check` before opening a PR.

## Testing Guidelines
Unit tests use Go’s `testing` package with files named `*_test.go` adjacent to source files (for example, `pkg/.../*_test.go`). Prefer table-driven tests for controller/client behavior. Run `make test` for standard coverage output (`coverage.out`), and use `make coverage` when you need HTML/text coverage reports.

## Commit & Pull Request Guidelines
Recent commits typically use an issue key prefix (for example, `OKE-41173`, `JIRA OKE-41505`) plus a concise imperative summary. Keep commits focused and reference the tracking ticket. PRs must satisfy Oracle contribution policy: include `Signed-off-by` (use `git commit --signoff`) so OCA verification can pass. In PR descriptions, include scope, risk, and commands run (for example, `make check && make test`).
