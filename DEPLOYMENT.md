# Deployment and Routing Architecture

Status: `APPROVED — CONFIGURED; PREVIEW VERIFICATION PENDING`

## 1. Delivery Topology

- One new GitHub repository containing frontend and backend.
- One new Vercel project connected to that repository.
- One public origin for the SPA and API.
- Vercel serves the Vue build for frontend paths.
- Vercel invokes a Go Function for `/api/*`.
- PostgreSQL and object storage are managed external services.

## 2. Non-Negotiable Routing Ownership

```text
/                     -> Vue SPA
/login                -> Vue SPA -> Vue Router
/admin/*              -> Vue SPA -> Vue Router
/sales/*              -> Vue SPA -> Vue Router
/api/*                -> Go Function -> Fiber -> JSON
```

Rules:

- API routing rule has precedence over SPA fallback.
- SPA fallback applies only to paths not beginning with `/api/`.
- Fiber registers only API and operational JSON routes.
- Fiber has no HTML template engine and no static-file fallback.
- Vue never handles `/api/*` as a client route.
- Unknown `/api/*` paths return JSON 404.
- Direct navigation and refresh of `/admin/...` or `/sales/...` returns `index.html`, after which Vue Router resolves the page.

## 3. Repository Establishment

1. Create a dedicated repository for the new CRM Prospect Simulator.
2. Configure the real remote before adopting a hosted Go module path.
3. Protect `main` and require the documented quality gates.
4. Connect only the new repository to the new Vercel project.
5. Keep unrelated repositories and deployment settings outside this project.

## 4. Vercel Project

The Vercel project uses only the version-controlled routing configuration in this repository.

Expected build responsibilities:

- Install/build frontend from `frontend`.
- Publish the Vite output directory.
- Discover root `api/index.go` as the Go Function adapter.
- Route `/api/:path*` to the adapter before the SPA fallback and forward the captured path through the private `__api_path` rewrite parameter.
- Route remaining non-file frontend paths to the SPA entrypoint.

The adapter restores `/api/<captured-path>`, removes the private rewrite parameter, preserves public query parameters, and then invokes Fiber. Adapter contract tests are mandatory; a Vercel preview smoke test remains required before production cutover.

## 5. Environment Separation

| Environment | Purpose | Data policy |
|---|---|---|
| Local | Developer workflow | Local/test accounts and services |
| Preview | Pull-request verification | Isolated non-production services/credentials |
| Production | Approved release | Production services and restricted secrets |

Preview deployments must never connect to the production database, production object bucket, or unrestricted Google keys.

## 6. Environment Variables

Categories include:

- Application environment and public base URL.
- PostgreSQL connection and pool limits.
- JWT issuer, audience, signing/verification keys.
- Google Places server API key.
- Vite Google Maps browser API key.
- Object-storage endpoint, bucket, and credentials.
- Allowed origin and security settings.
- Observability configuration.

Only explicitly prefixed public frontend variables may enter the Vite bundle. Server secrets must never use the frontend public prefix.

## 7. Google Key Restrictions

- Browser Maps key: HTTP referrer restrictions and only required Maps APIs.
- Server Places key: server-side restrictions supported by the hosting architecture and only required Places APIs.
- Separate keys for preview and production.
- Quota alerts and billing alerts enabled.
- Keys are rotated without source-code changes.

## 8. Database Deployment

- Prisma migration validation runs in CI.
- Production migrations run as an explicit release step, not on every serverless cold start.
- Application deployment occurs only after compatible migrations succeed.
- Expand/migrate/contract sequencing maintains compatibility during rolling/serverless deployments.
- Connection pooling and limits are sized for serverless concurrency.

## 9. CI/CD Gates

Pull request:

- Backend formatting/static analysis/tests.
- Frontend type-check/lint/tests/build.
- Prisma validation when relevant.
- Security scanning.
- Preview deployment.
- Routing smoke test.

Production:

- Merge to protected `main`.
- Approved migration step when required.
- Vercel production deployment.
- Post-deployment smoke tests.
- Rollback/forward-fix decision if checks fail.

## 10. Mandatory Routing Test Matrix

| Request | Expected |
|---|---|
| `GET /` | Vue HTML, 200 |
| `GET /login` | Vue HTML, 200 |
| `GET /admin/dashboard` | Vue HTML, then frontend auth guard |
| `GET /sales/dashboard` | Vue HTML, then frontend auth guard |
| `GET /api/health` | JSON, 200 |
| `GET /api/v1/health` | JSON, 200 compatibility alias |
| `GET /api/v1/unknown` | JSON, 404 |
| `GET /api/v1/prospects` unauthenticated | JSON, 401 |
| `GET /not-a-real-page` | Vue HTML, then Vue 404 view |

Tests also verify that an API response never contains SPA HTML and a direct frontend route never receives Fiber JSON.

## 11. Operational Endpoints

- `/api/health`: process health, no dependency secrets.
- `/api/v1/health`: versioned compatibility alias for API clients.
- `/api/v1/ready`: dependency readiness where appropriate.

Operational responses are JSON and expose no configuration values.

## 12. Rollback and Recovery

- Vercel deployment rollback handles application artifacts.
- Database rollback favors forward-fix migrations; destructive down migrations are not assumed safe.
- Object-storage changes are versioned/retained according to approved policy.
- The prior stable deployment and database compatibility window remain available during each future rollout.
- Routing configuration changes require immediate smoke tests because they can make the entire application unreachable.

## 13. Cutover Criteria

- All required Google Places/Maps capabilities pass their acceptance tests.
- All routing tests pass on the production candidate domain.
- No API path returns frontend HTML.
- Direct Vue route refresh works.
- Production secrets and key restrictions are verified.
- Database backup and migration plan are approved.
- End-to-end Admin-to-Sales flow passes.
- The current release has an application rollback and database forward-fix plan.
