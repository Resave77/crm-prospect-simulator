# Repository and Folder Structure

Status: `APPROVED — IMPLEMENTED INCREMENTALLY`

## 1. Repository Strategy

The new CRM Prospect Simulator uses one repository containing frontend, backend, deployment configuration, and architecture documentation. No repository hosting path is assumed until an actual remote is explicitly configured.

## 2. Target Structure

```text
enterprise-crm-field-sales/
├── .github/
│   ├── workflows/
│   │   ├── pull-request.yml
│   │   └── main.yml
│   ├── CODEOWNERS
│   └── pull_request_template.md
│
├── api/
│   └── index.go
│
├── go.mod                         # Required at repository root by Vercel Go runtime
├── go.sum
│
├── backend/
│   ├── cmd/
│   │   └── server/
│   │       └── main.go
│   ├── config/
│   ├── internal/
│   │   ├── auth/
│   │   │   ├── handler/
│   │   │   ├── service/
│   │   │   ├── repository/
│   │   │   ├── model/
│   │   │   └── dto/
│   │   ├── dashboard/
│   │   ├── prospect/
│   │   ├── assignment/
│   │   ├── visit/
│   │   ├── pipeline/
│   │   ├── customer/
│   │   ├── attendance/
│   │   ├── user/
│   │   ├── settings/
│   │   └── shared/
│   │       ├── googleplaces/
│   │       ├── middleware/
│   │       ├── response/
│   │       ├── validation/
│   │       ├── pagination/
│   │       ├── clock/
│   │       └── errors/
│   ├── platform/
│   │   ├── database/
│   │   ├── storage/
│   │   ├── httpclient/
│   │   ├── security/
│   │   └── observability/
│   ├── server/
│   │   ├── app.go
│   │   ├── routes.go
│   │   └── dependencies.go
│   ├── prisma/
│   │   ├── schema.prisma
│   │   └── migrations/
│   ├── test/
│   │   ├── contract/
│   │   └── integration/
│   └── package.json               # Prisma CLI only
│
├── frontend/
│   ├── public/
│   ├── src/
│   │   ├── api/
│   │   ├── assets/
│   │   ├── components/
│   │   │   ├── common/
│   │   │   ├── feedback/
│   │   │   ├── maps/
│   │   │   ├── prospect/
│   │   │   ├── customer/
│   │   │   └── attendance/
│   │   ├── composables/
│   │   ├── layouts/
│   │   ├── router/
│   │   ├── stores/
│   │   ├── types/
│   │   ├── utils/
│   │   └── views/
│   │       ├── Login/
│   │       ├── Admin/
│   │       │   ├── Dashboard/
│   │       │   ├── Prospect/
│   │       │   ├── Assignment/
│   │       │   ├── Customer/
│   │       │   ├── User/
│   │       │   └── Settings/
│   │       └── Sales/
│   │           ├── Dashboard/
│   │           ├── Prospect/
│   │           ├── Visit/
│   │           ├── Pipeline/
│   │           ├── Customer/
│   │           ├── Attendance/
│   │           └── Profile/
│   ├── test/
│   │   ├── unit/
│   │   └── e2e/
│   ├── index.html
│   ├── package.json
│   ├── tsconfig.json
│   └── vite.config.ts
│
├── docs/
│   ├── adr/
│   └── runbooks/
│
├── PROJECT_BIBLE.md
├── FLOW.md
├── DATABASE.md
├── API.md
├── FOLDER_STRUCTURE.md
├── CODING_STANDARD.md
├── AUTHENTICATION.md
├── DEPLOYMENT.md
├── UI_GUIDELINES.md
├── README.md
└── vercel.json
```

The displayed `.go`, `.vue`, and configuration filenames describe intended ownership only; this document does not authorize their creation.

## 3. Backend Module Template

Each business module may contain:

```text
module/
├── handler/       # Fiber request/response adapters
├── service/       # Use cases and business orchestration
├── repository/    # Interfaces and PostgreSQL implementations
├── model/         # Domain entities/value objects
├── dto/           # Transport input/output contracts
└── errors.go      # Stable domain error definitions
```

Rules:

- A folder is created only when it has a real responsibility.
- Cross-module imports target a public service/contract, not another module's PostgreSQL implementation.
- `shared` contains genuinely cross-cutting primitives, not miscellaneous business logic.
- Provider-specific Google DTOs remain under `shared/googleplaces` and do not leak into prospect domain models.

## 4. Frontend Ownership

- `views` are route-level orchestration screens.
- `components` are reusable presentation and interaction units.
- `stores` hold cross-route state only; local form state remains local.
- `api` contains Axios client configuration and typed resource clients.
- `types` contains frontend contracts; generated API types may later replace duplicated types.
- Admin and Sales view trees may reuse primitives but never share a route layout.

## 5. Vercel Entrypoint

Root `api/index.go` exists only because Vercel discovers Go Functions under the root `api` directory. The root `go.mod` is also required by the official Vercel Go runtime. The adapter delegates to the backend application bootstrap and contains no business rules.

## 6. Project-Owned Source

Only packages belonging to the approved new architecture are kept in the source tree. Generated dependencies, build output, structure snapshots, server-rendered templates, and obsolete compatibility folders are not project source.

## 7. Dependency Direction

Allowed:

```text
api entrypoint -> backend server
handler -> service -> repository interface
repository implementation -> platform/database
service -> external provider interface
frontend view -> store/api/component
```

Forbidden:

```text
model -> handler
repository -> handler
service -> Fiber context
domain module -> Vercel adapter
backend -> frontend source
frontend -> database
```

## 8. Test Placement

- Go unit tests live beside the package under test.
- Cross-package API contract tests live in `backend/test/contract`.
- PostgreSQL integration tests live in `backend/test/integration`.
- Vue unit tests live in `frontend/test/unit` or beside complex modules by agreed convention.
- Browser flow tests live in `frontend/test/e2e`.
- Every implemented module adds its own unit, contract, and integration coverage as applicable.
