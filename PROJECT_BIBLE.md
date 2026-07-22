# Enterprise CRM Field Sales — Project Bible

## 1. Document Status

This document is the architectural source of truth for the Enterprise CRM Field Sales project. It governs design and implementation decisions unless an approved Architecture Decision Record (ADR) explicitly supersedes it.

Status: `APPROVED — P0 FOUNDATION`

No implementation may begin until the architecture documentation set is approved.

## 2. Product Mission

The product manages the complete lifecycle of a business discovered through Google Places, from prospect discovery and field-sales engagement through conversion into an existing customer and recurring customer attendance.

The product is an enterprise field-sales simulator with production-grade boundaries, auditability, security, and deployment design. It is not a throwaway demo.

## 3. Authoritative Business Flow

1. Administrator determines a search area.
2. Administrator searches Google Places using radius, categories, and optional keyword.
3. Administrator selects a business and saves it as a prospect.
4. The system rejects duplicate prospects using Google Place ID.
5. Administrator assigns the prospect to one Sales Executive.
6. The prospect appears in the Sales Executive's mobile `My Prospect` list.
7. Sales Executive performs a prospect visit with GPS, selfie, check-in, and check-out.
8. Pipeline advances through `ASSIGNED`, `VISITED`, and `FOLLOW_UP`.
9. Sales Executive records the decision as `LOST` or `WON`.
10. `LOST` ends the prospect lifecycle.
11. Administrator reviews a `WON` prospect and starts conversion.
12. Administrator completes the existing-customer form and saves the customer.
13. The prospect becomes `CONVERTED` and leaves active prospect work queues.
14. The customer appears in the Sales Executive's mobile `My Customer` list.
15. Future customer visits use Existing Customer Attendance with GPS, selfie, check-in, and check-out.

Detailed transitions and invariants are defined in `FLOW.md`.

## 4. Users and Authorization

Only two roles exist:

- `ADMINISTRATOR`: desktop administration, discovery, assignment, won review, conversion, user administration, and settings.
- `SALES_EXECUTIVE`: mobile prospect/customer work assigned to that user, visits, pipeline updates, attendance, and profile.

No third role, generic permission builder, multi-role user, or public registration is in scope. Authorization must still be enforced server-side for every protected operation.

## 5. Experience Boundaries

The frontend is one Vue application with two deliberately separated experiences:

- Desktop administrator surface: `/admin/*`
- Mobile Sales Executive surface: `/sales/*`
- Authentication surface: `/login`

Route guards, layout components, navigation, viewport policy, and server authorization separate the experiences. Responsive CSS alone is not considered sufficient separation.

## 6. System Architecture

The backend is a modular monolith using Go and Fiber. Each business module owns its models, repository contracts, services, handlers, request/response DTOs, and tests.

Request direction is strictly:

```text
HTTP JSON request
  -> Fiber router/middleware
  -> handler
  -> service/use case
  -> repository or external provider interface
  -> PostgreSQL or Google API
```

Rules:

- Handlers translate HTTP concerns only.
- Services enforce business workflows and transaction boundaries.
- Repositories abstract persistence.
- Domain models do not depend on Fiber, Prisma, PrimeVue, or JSON transport structures.
- Google Places is an external provider, not a repository.
- Cross-module writes are coordinated by an application service and one transaction.
- The backend never renders HTML.

## 7. Technology Decisions

### Frontend

- Vue 3 with Composition API and TypeScript.
- Vite for build and local development.
- PrimeVue as the UI component foundation.
- Pinia for shared application state.
- Vue Router for all frontend navigation.
- Axios for REST communication through one configured client.

### Backend

- Go with Fiber v2.
- REST API returning JSON only.
- PostgreSQL as the system of record.
- Repository pattern and service layer.
- JWT access and refresh-token authentication.
- Runtime PostgreSQL access through a native Go driver/pool.
- Prisma used only to define and execute database migrations; Prisma Client is not used by Go.

### External Platforms

- Google Maps JavaScript API for interactive maps.
- Google Places API for Nearby Search, Text Search, and Place Detail.
- Object storage for selfies; binary images are not stored in PostgreSQL.
- Vercel for the Vue application and Go serverless API.

## 8. Routing Invariant

Routing ownership must never overlap:

| Path | Owner | Response |
|---|---|---|
| `/` | Vue/Vercel static hosting | SPA HTML |
| `/login` | Vue Router | SPA HTML |
| `/admin/*` | Vue Router | SPA HTML |
| `/sales/*` | Vue Router | SPA HTML |
| `/api/*` | Fiber | JSON |

Fiber must not define `/`, `/admin/*`, `/sales/*`, or an SPA fallback. Vercel must not rewrite `/api/*` to `index.html`. These invariants are expanded in `DEPLOYMENT.md`.

## 9. New Project Foundation

This repository is the new CRM Prospect Simulator. It does not retain runtime compatibility with the previous prototype and does not use the prototype's routes, Fiber rendering, templates, or Vercel configuration.

The following approved business capabilities remain required:

- Google Places Nearby Search.
- Google Places Text Search.
- Google Place Detail.
- Google Maps rendering, marker selection, radius circle, draggable center, and browser GPS.
- Category mapping, distance calculation, explicit field masks, and API-key separation.
- A JSON-only Go API under `/api/*` and a Vue-rendered SPA for all user interfaces.
- Contract-tested Vercel routing that prevents API requests from reaching the SPA fallback.

## 10. Domain Modules

- `auth`: identity, sessions, tokens, password operations.
- `dashboard`: role-specific projections and counts.
- `prospect`: discovery snapshots, prospect lifecycle, and prospect queries.
- `assignment`: prospect ownership history.
- `visit`: prospect field visits and evidence.
- `pipeline`: controlled status transitions and history.
- `customer`: won review, conversion, and existing-customer records.
- `attendance`: existing-customer visits and evidence.
- `user`: administrator-managed users.
- `settings`: controlled system configuration.
- `shared/googleplaces`: provider client and provider DTO mapping.

## 11. Non-Functional Requirements

- All timestamps are stored in UTC and rendered in the user's configured timezone.
- Identifiers exposed by the API are UUIDs; database sequence IDs are not exposed.
- Business mutations are auditable.
- Conversion, check-in, check-out, and pipeline decisions are idempotent.
- List APIs are paginated and deterministically sorted.
- API errors follow one documented JSON envelope.
- Logs are structured and must not contain passwords, JWTs, Google keys, or selfie contents.
- GPS coordinates retain enough precision for attendance validation.
- Google API failures must not corrupt CRM state.
- Accessibility target is WCAG 2.1 AA for primary workflows.
- Desktop administrator support begins at 1024 px; Sales Executive workflows are optimized for 360–480 px mobile widths.

## 12. Explicit Non-Goals for Architecture Approval

- No code generation.
- No CRUD implementation.
- No database or Prisma migration execution.
- No authentication implementation.
- No Vue components.
- No Go handlers.
- No deployment or repository creation.
- No unapproved changes to the business flow and public API contracts defined by this documentation.

## 13. Decision Governance

After approval, architectural changes require an ADR containing context, decision, alternatives, consequences, migration impact, and approval date. Documentation and API/database contracts must be updated in the same change as the approved decision.

## 14. Definition of Architecture Approval

Approval confirms agreement on:

- Exact business flow and state machines.
- Two-role authorization model.
- Modular-monolith boundaries.
- Database ownership and conceptual schema.
- REST contracts and error conventions.
- Frontend/backend routing separation.
- Authentication and deployment design.
- UI separation between desktop administrator and mobile sales.
- Preservation plan for existing Google integrations.
