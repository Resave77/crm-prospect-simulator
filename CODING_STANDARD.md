# Coding and Engineering Standard

Status: `APPROVED`

## 1. General Principles

- Optimize for clarity, correctness, and maintainability over cleverness.
- Keep business rules explicit and testable.
- Make invalid state transitions impossible at the service boundary.
- Preserve approved public API contracts unless a versioned or explicitly approved contract change says otherwise.
- Do not log secrets or sensitive evidence.
- Comments explain intent and constraints, not obvious syntax.

## 2. Git Workflow

- `main` is protected and deployable.
- Work is delivered through small pull requests from short-lived branches.
- Commit messages follow Conventional Commits.
- Database/API contract changes are identified in the PR description.
- Architecture changes require an ADR and documentation update.
- Generated files and compiled binaries are not committed unless explicitly required.
- Existing user changes are never overwritten during migration.

## 3. Go Standard

- Use the Go version pinned in `go.mod` and CI.
- Format with `gofmt`; imports are deterministic.
- Package names are short, lowercase, and responsibility-based.
- Exported identifiers have meaningful documentation.
- Constructors return dependencies explicitly; no mutable global service locators.
- Accept `context.Context` in service/repository/provider operations.
- Use sentinel/typed domain errors and map them to HTTP status centrally.
- Wrap errors with operational context while preserving the cause.
- Never expose raw database or Google provider errors to API clients.
- HTTP clients use explicit timeouts and are injected for testing.
- Goroutines require bounded concurrency, cancellation, and deterministic collection.

### Handler rules

- Parse and validate transport shape.
- Obtain authenticated principal from middleware context.
- Call one application service operation.
- Return the standard response envelope.
- Contain no SQL, provider calls, transactions, or workflow decisions.

### Service rules

- Enforce authorization-sensitive ownership and state transitions.
- Define transaction boundaries.
- Use interfaces owned by the consuming service.
- Remain independent of Fiber types.
- Receive clock/ID generators when deterministic tests require them.

### Repository rules

- Express domain-oriented methods, not generic CRUD abstractions.
- Return domain models or purpose-specific projections.
- Use parameterized SQL only.
- Accept a transaction abstraction when participating in multi-write use cases.
- Distinguish not-found, conflict, and infrastructure failure.

## 4. Vue and TypeScript Standard

- Vue Single File Components use `<script setup lang="ts">`.
- Composition API is mandatory for new code.
- TypeScript strict mode is enabled.
- Avoid `any`; unknown external data is validated before use.
- Route-level code splitting is used for Admin and Sales modules.
- Components use PrimeVue primitives and project design tokens.
- Direct DOM manipulation is limited to external library adapters such as Google Maps.
- API calls go through typed API modules, never directly from arbitrary components.
- Pinia stores do not contain presentation markup or router-instance side effects.
- Forms have explicit loading, validation, success, and failure states.

## 5. Naming

| Area | Convention | Example |
|---|---|---|
| Go package/file | lowercase/snake case file | `pipeline_history.go` |
| Go exported type | PascalCase | `ProspectService` |
| JSON | camelCase | `googlePlaceId` |
| Database | snake_case | `google_place_id` |
| Vue component | PascalCase | `ProspectSearchMap.vue` |
| Composable | `use` prefix | `useGeolocation.ts` |
| Pinia store | domain + `Store` | `useProspectStore` |
| Route name | scoped PascalCase | `AdminProspectDetail` |
| Status constants | uppercase stable codes | `FOLLOW_UP` |

## 6. API and Validation

- Treat all client input as untrusted.
- Normalize strings deliberately; do not silently alter business identifiers.
- Validation errors identify safe field names.
- IDs, status codes, pagination, and timestamps follow `API.md`.
- Breaking API changes require a new version or an approved compatibility plan.
- Clients never construct authorization decisions from hidden UI alone.

## 7. Database Standard

- Prisma migration names describe the business change.
- Applied migrations are never edited.
- Transactions are short and contain no external network calls.
- N+1 query patterns are prohibited in list endpoints.
- Soft deletion is used only when business semantics require it; immutable history is preferred for audit records.
- Query performance is verified for major lists before production.

## 8. Security Standard

- Secrets are environment variables managed by the deployment platform.
- Passwords use bcrypt with a reviewed work factor. Plaintext passwords are never stored or logged.
- JWT signing keys support rotation and are never committed.
- Selfie uploads validate content type, signature, size, and ownership.
- Google server keys never enter frontend bundles.
- Authentication and access denials avoid resource-enumeration leakage.
- Dependency and secret scanning run in CI.

## 9. Testing Standard

Required layers:

- Unit tests for state transitions, validation, distance calculation, and mapping.
- Service tests using repository/provider fakes.
- Repository integration tests against PostgreSQL.
- HTTP contract tests for envelopes, errors, and authorization.
- Frontend unit tests for stores and critical form behavior.
- End-to-end tests for the complete Administrator-to-Sales workflow.
- Deployment routing smoke tests for frontend and `/api/*` ownership.

Tests must not depend on a live Google API by default. Live-provider tests are opt-in and credential-controlled.

## 10. Quality Gates

A pull request cannot merge unless applicable checks pass:

- Go formatting, vet/static analysis, unit, contract, and integration tests.
- TypeScript type-check, lint, unit tests, and production build.
- Migration validation when Prisma files change.
- Secret/dependency scans.
- Routing smoke tests when Vercel configuration changes.
- Documentation consistency for contract or architecture changes.

## 11. Observability

- Logs are structured JSON in hosted environments.
- Every request has a request ID.
- Logs include route template, status, duration, and safe actor ID.
- Provider timing and error category are recorded without credentials/payload leakage.
- Metrics distinguish validation, authorization, conflict, provider, and infrastructure failures.
- Audit logs are business records and are separate from operational logs.
