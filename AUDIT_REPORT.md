# Auth / Role / Session — Comprehensive Audit Report

**Project:** CRM Prospect Simulator (Yummy CRM)  
 **Date:** 2026-07-30  
 **Scope:** Read-only audit of authentication, authorization, role management, and session architecture across backend (Go/Fiber) and frontend (Vue 3/PrimeVue).

---

## Table of Contents

1. [Git Baseline](#1-git-baseline)
2. [Build / Test Baseline](#2-build--test-baseline)
3. [User Model](#3-user-model)
4. [Role Definitions](#4-role-definitions)
5. [Authentication Flow](#5-authentication-flow)
6. [Session Architecture](#6-session-architecture)
7. [Admin Account Management](#7-admin-account-management)
8. [Route Matrix — Backend](#8-route-matrix--backend)
9. [Endpoint Matrix — Backend](#9-endpoint-matrix--backend)
10. [Data Scope](#10-data-scope)
11. [Menu / Navigation Audit](#11-menu--navigation-audit)
12. [Database Audit](#12-database-audit)
13. [Gap Matrix](#13-gap-matrix)
14. [Implementation Phases](#14-implementation-phases)
15. [Documentation Plan](#15-documentation-plan)
16. [Files Needed Later](#16-files-needed-later)
17. [Uncertainties & Risks](#17-uncertainties--risks)

---

## 1. Git Baseline

| Property      | Value                                                                                                                   |
| ------------- | ----------------------------------------------------------------------------------------------------------------------- |
| Branch        | `main`                                                                                                                  |
| Origin        | `https://github.com/Resave77/crm-prospect-simulator.git`                                                                |
| Latest Commit | `251d09e` — feat: overhaul pipeline UX with stage direction panel, clear action hierarchy, PrimeVue toast notifications |
| Working Tree  | Clean (3 untracked: prisma migration dir, uploads selfie, frontend/public/)                                             |
| Unstaged      | None                                                                                                                    |
| Staged        | None                                                                                                                    |

---

## 2. Build / Test Baseline

| Area                         | Status                                                                                                                                                                                   |
| ---------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Frontend `npm run build`     | ✅ Passes                                                                                                                                                                                |
| Frontend `npm run typecheck` | ✅ Passes                                                                                                                                                                                |
| Frontend tests               | `tests/pipeline.test.mjs`, `tests/navigation.test.mjs` (vanilla Node `--test`)                                                                                                           |
| Backend tests                | `backend/server/app_test.go`, `backend/internal/auth/service/auth_test.go`, `backend/internal/prospect/service/service_test.go`, `backend/internal/customer/repository/postgres_test.go` |
| Go build                     | `go build ./cmd/server/` — expected to pass                                                                                                                                              |
| Go test                      | `go test ./...` — expected to pass                                                                                                                                                       |
| No lint config found         | No `.golangci.yml`, no ESLint config, no Prettier config                                                                                                                                 |

**Note:** Backend tests exist but auth test coverage is limited to `auth/service/auth_test.go`. No test for middleware, handlers, or repository layer.

---

## 3. User Model

### Backend (`prisma/schema.prisma` + `internal/auth/model/user.go`)

| Field                     | Type            | Notes                                    |
| ------------------------- | --------------- | ---------------------------------------- |
| `id`                      | UUID (PK)       | Auto-generated                           |
| `email`                   | String (unique) | Lowercased at login                      |
| `passwordHash`            | String          | bcrypt, mapped to `password_hash`        |
| `fullName`                | String          | Mapped to `full_name`                    |
| `role`                    | UserRole enum   | `ADMINISTRATOR` or `SALES_EXECUTIVE`     |
| `status`                  | UserStatus enum | `ACTIVE` or `INACTIVE`, default `ACTIVE` |
| `tokenVersion`            | Int             | Default 1, used for session invalidation |
| `lastLoginAt`             | DateTime?       | Nullable                                 |
| `createdAt` / `updatedAt` | Timestamps      | Managed by Prisma                        |

**Relationships:**

- `refreshSessions` (1:N)
- `assignedProspects` (1:N, relation "ProspectSalesExecutive")
- `prospectStatusChanges` (1:N, relation "ProspectStatusActor")
- `assignedCustomerSites` (1:N, relation "CustomerSalesExecutive")
- `convertedCustomerSites` (1:N, relation "CustomerConverter")
- `prospectVisits` (1:N, relation "ProspectVisitSalesExecutive")
- `prospectComments` (1:N, relation "ProspectCommentUser")

### Frontend (`types/auth.ts`)

```ts
export type UserRole = "ADMINISTRATOR" | "SALES_EXECUTIVE";

export interface AuthUser {
  id: string;
  email: string;
  fullName: string;
  role: UserRole;
}
```

### Public User (API response)

Only `id`, `email`, `fullName`, `role` are exposed. `status`, `tokenVersion`, timestamps are never sent to the client.

---

## 4. Role Definitions

### Current State

Two roles exist — **there is no `SALES_MANAGER` role anywhere** (not in Prisma enum, not in Go model, not in frontend types):

| Role              | Backend Constant     | Frontend Type       | Menu/Views                                   |
| ----------------- | -------------------- | ------------------- | -------------------------------------------- |
| `ADMINISTRATOR`   | `RoleAdministrator`  | `'ADMINISTRATOR'`   | Full admin sidebar (11 items)                |
| `SALES_EXECUTIVE` | `RoleSalesExecutive` | `'SALES_EXECUTIVE'` | Sales bottom nav (5 items) + desktop sidebar |

### Role Validation

- Go `UserRole.Valid()` checks `r == "ADMINISTRATOR" || r == "SALES_EXECUTIVE"`
- No DB-level check constraint beyond the Prisma enum
- No frontend enum validation; role is a string union type

### Role-Based Frontend Guards

The router (`router/index.ts`) uses `meta.role` on route groups:

- `/admin/*` routes have `meta: { role: 'ADMINISTRATOR' }`
- `/sales/*` routes have `meta: { role: 'SALES_EXECUTIVE' }`

The `beforeEach` guard checks:

```ts
if (to.meta.role && to.meta.role !== auth.role) return homeFor(auth.role!);
```

**Gap:** This only checks the group-level role. Individual routes within a group do not enforce roles independently. If a future role like `SALES_MANAGER` needs access to a subset of admin routes, the current guard cannot express that.

---

## 5. Authentication Flow

### Login Flow

```
Client                    Server
  |                         |
  |-- POST /api/v1/auth/login {email, password} -->|
  |                         |-- FindByEmail()
  |                         |-- bcrypt.CompareHashAndPassword()
  |                         |-- Check User.Status == ACTIVE
  |                         |-- Create RefreshSession (new UUID token)
  |                         |-- Issue JWT (accessToken, 30min TTL)
  |<-- 200 {accessToken, accessExpiresAt, user} -->|
  |    Set-Cookie: crm_refresh (HTTPOnly, Secure, SameSite=Strict)
```

### Access Token (JWT)

- **Algorithm:** Standard JWT (library: `github.com/golang-jwt/jwt/v5`)
- **TTL:** 30 minutes (configurable via `ACCESS_TOKEN_TTL` env var, default 15min)
- **Claims:**
  - `sub`: User UUID
  - `role`: User role string
  - `session`: Session UUID
  - `tokenVersion`: Current `token_version` from DB
  - `iss`, `aud`, `exp`, `iat`, `jti`

### Refresh Token

- **Format:** `sessionID.base64(secret)` — not a JWT, a custom opaque token
- **TTL:** 30 days (configurable via `REFRESH_TOKEN_TTL` env var)
- **Storage:** HTTP-only cookie `crm_refresh`, path `/api/v1/auth`
- **Rotation:** Every refresh revokes the old session and creates a new one (full rotation)
- **Reuse Detection:** If a revoked (ROTATED) session is used again, all sessions for that user are revoked (token theft protection)

### Token Refresh Flow

```
Client                    Server
  |                         |
  |-- POST /api/v1/auth/refresh (Cookie: crm_refresh) -->|
  |                         |-- Parse sessionID + secret from cookie
  |                         |-- FindSessionByID()
  |                         |-- Verify tokenHash matches
  |                         |-- Check not revoked, not expired
  |                         |-- If revoked with reason "ROTATED": revoke ALL user sessions (theft detected)
  |                         |-- Check user is ACTIVE
  |                         |-- Rotate: revoke old session, create new
  |                         |-- Issue new JWT + new refresh cookie
  |<-- 200 {accessToken, accessExpiresAt, user} -->|
  |    Set-Cookie: crm_refresh (new value)
```

### Logout Flow

```
Client                    Server
  |                         |
  |-- POST /api/v1/auth/logout (Cookie: crm_refresh) -->|
  |                         |-- Revoke session with reason "LOGOUT"
  |<-- 204 (No Content)    |
  |    Set-Cookie: crm_refresh (expired)
```

### Logout All Flow (requires authentication)

```
Client                    Server
  |                         |
  |-- POST /api/v1/auth/logout-all (Bearer token) -->|
  |                         |-- Revoke ALL sessions for user with reason "LOGOUT_ALL"
  |<-- 204 (No Content)    |
```

### Frontend Token Management (`api/client.ts`)

- `accessToken` stored in-memory (module-level variable), not in localStorage
- `refreshSession()` called on:
  1. Page load (`auth.bootstrap()`)
  2. 401 interceptor (automatic retry with new token)
- Axios request interceptor attaches `Authorization: Bearer` header
- Axios response interceptor catches 401, calls `refreshSession()`, retries the failed request once
- If refresh also fails, session observer is called with `null`

### Frontend Auth Store (`stores/auth.ts`)

- `auth.bootstrap()` is called in router `beforeEach` to restore session on page load
- `auth.login()` calls backend, stores payload, sets `bootstrapped = true`
- `auth.logout()` calls backend, clears all state
- `auth.authenticated` is a computed ref: `user.value !== null`
- `auth.role` is a computed ref: `user.value?.role ?? null`

### Cookie Configuration

| Property | Value                                                     |
| -------- | --------------------------------------------------------- |
| Name     | `crm_refresh`                                             |
| Path     | `/api/v1/auth` (only sent to auth endpoints)              |
| HTTPOnly | `true`                                                    |
| Secure   | Configurable via `COOKIE_SECURE` env var (default `true`) |
| SameSite | `Strict`                                                  |
| Expires  | Refresh token expiry                                      |

---

## 6. Session Architecture

### Database Model (`RefreshSession`)

| Field                     | Type                 | Notes                                                                                     |
| ------------------------- | -------------------- | ----------------------------------------------------------------------------------------- |
| `id`                      | UUID (PK)            | Generated on creation                                                                     |
| `userId`                  | UUID (FK → users.id) | CASCADE on restrict                                                                       |
| `tokenHash`               | String               | bcrypt hash of refresh token secret                                                       |
| `userAgent`               | String               | Stored for audit                                                                          |
| `ipAddress`               | String               | Stored for audit                                                                          |
| `expiresAt`               | Timestamptz          | 30 days from creation                                                                     |
| `revokedAt`               | Timestamptz?         | Nullable, set on logout/rotation/theft                                                    |
| `revokeReason`            | String?              | `"LOGOUT"`, `"LOGGED_OUT_ALL"`, `"ROTATED"`, `"ROTATION_REUSE"`, `"ACCESS_TOKEN_FAILURE"` |
| `replacedBySessionId`     | UUID?                | Points to the new session after rotation                                                  |
| `createdAt` / `updatedAt` | Timestamps           |                                                                                           |

### Session Lifecycle

```
Created ──→ Active ──→ Revoked (LOGOUT)
  │                    └── Revoked (LOGGED_OUT_ALL)
  │                    └── Revoked (ROTATED) ──→ New session replaces it
  │                    └── Revoked (ACCESS_TOKEN_FAILURE)
  │                    └── Revoked (ROTATION_REUSE → all user sessions revoked)
  └──→ Expired (auto, no cleanup)
```

### Session Repository Interface (`repository.go`)

```go
type SessionRepository interface {
  Create(ctx, session) error
  FindSessionByID(ctx, id) (RefreshSession, error)
  Rotate(ctx, oldID, newSession, at) error    // transactional: revoke old + insert new
  Revoke(ctx, id, reason, at) error
  RevokeAllForUser(ctx, userID, reason, at) error
}
```

### Key Security Properties

1. **Refresh token rotation** — every refresh invalidates the previous session
2. **Reuse detection** — if a rotated session is presented, all user sessions are revoked (implies token theft)
3. **Token version** — `User.tokenVersion` is embedded in the JWT and validated on every authenticated request; incrementing it invalidates all existing access tokens
4. **No session table** — sessions are only persisted server-side via `refresh_sessions` table; there is no concept of a "current session" list exposed to the user

### Missing Session Features

- **No session listing UI** — admin cannot see active sessions
- **No session revocation UI** — admin cannot selectively revoke sessions
- **No session expiry cleanup** — no periodic job to clean up expired/revoked sessions
- **No max sessions per user** — no limit on concurrent refresh sessions

---

## 7. Admin Account Management

### Current State: ❌ Not Implemented

There is **no admin user management functionality** anywhere:

- No API endpoint to list users
- No API endpoint to create users
- No API endpoint to update users (role, status, name, email)
- No API endpoint to reset passwords
- No API endpoint to activate/deactivate users
- No frontend view for user management
- No seed command supports dynamic user creation (only 4 hardcoded accounts)
- The `AdminLayout` sidebar has no "User Management" or "Account" link

The only way to create/manage users is via the seed script (`cmd/seed/main.go`) which has 4 hardcoded accounts:

| Email               | Name                | Role            |
| ------------------- | ------------------- | --------------- |
| `admin@yummy.test`  | Yummy Administrator | ADMINISTRATOR   |
| `sales@yummy.test`  | Nurdin Pratama      | SALES_EXECUTIVE |
| `sales2@yummy.test` | Alicia Ramadhan     | SALES_EXECUTIVE |
| `sales3@yummy.test` | Rizky Ananda        | SALES_EXECUTIVE |

All share the same password: `password123`.

---

## 8. Route Matrix — Backend

### Public Routes (no auth)

| Method | Path                   | Handler       |
| ------ | ---------------------- | ------------- |
| GET    | `/api/health`          | Health check  |
| GET    | `/api/v1/health`       | Health check  |
| POST   | `/api/v1/auth/login`   | Login         |
| POST   | `/api/v1/auth/refresh` | Refresh token |

### Authenticated Routes (any active user)

| Method | Path                      | Handler              |
| ------ | ------------------------- | -------------------- |
| GET    | `/api/v1/auth/me`         | Current user profile |
| POST   | `/api/v1/auth/logout-all` | Revoke all sessions  |

### Sales Executive Routes (auth + RequireRole SALES_EXECUTIVE)

| Method | Path                                                    | Handler                |
| ------ | ------------------------------------------------------- | ---------------------- |
| GET    | `/api/v1/sales/prospects`                               | My assigned prospects  |
| GET    | `/api/v1/sales/prospects/:id`                           | Single prospect detail |
| PATCH  | `/api/v1/sales/prospects/:id/transition`                | Transition status      |
| PATCH  | `/api/v1/sales/prospects/:id/decision`                  | Alias for transition   |
| POST   | `/api/v1/sales/prospects/:id/visits/check-in`           | Check in               |
| PATCH  | `/api/v1/sales/prospects/:id/visits/:visitId/check-out` | Check out              |
| GET    | `/api/v1/sales/prospects/:id/comments`                  | List comments          |
| POST   | `/api/v1/sales/prospects/:id/comments`                  | Add comment            |
| GET    | `/api/v1/sales/prospects/:id/place-details`             | Place details          |
| POST   | `/api/v1/sales/prospects/:id/request-deletion`          | Request deletion       |
| GET    | `/api/v1/sales/visits`                                  | My visit history       |
| GET    | `/api/v1/sales/customers`                               | My customers           |
| GET    | `/api/v1/sales/customers/:id`                           | Customer detail        |
| GET    | `/api/v1/sales/customers/:id/place-details`             | Customer place details |
| GET    | `/api/v1/dashboard/sales`                               | Dashboard data         |

### Administrator Routes (auth + RequireRole ADMINISTRATOR)

| Method | Path                                                         | Handler                   |
| ------ | ------------------------------------------------------------ | ------------------------- |
| GET    | `/api/v1/admin/prospects/won`                                | Won prospects             |
| GET    | `/api/v1/admin/prospects/pipeline`                           | Pipeline                  |
| GET    | `/api/v1/admin/sales-executives`                             | List sales execs          |
| GET    | `/api/v1/admin/prospect-finder/search`                       | Search places             |
| GET    | `/api/v1/admin/prospect-finder/places/:placeId`              | Place detail              |
| GET    | `/api/v1/admin/prospect-finder/place-details/:googlePlaceId` | Place details             |
| POST   | `/api/v1/admin/prospects`                                    | Create prospect           |
| DELETE | `/api/v1/admin/prospects/:id`                                | Delete prospect           |
| GET    | `/api/v1/admin/prospects/:id`                                | Review prospect           |
| GET    | `/api/v1/admin/prospects/:id/comments`                       | List comments             |
| POST   | `/api/v1/admin/prospects/:id/comments`                       | Add comment               |
| GET    | `/api/v1/admin/prospects/:id/place-details`                  | Place details             |
| GET    | `/api/v1/admin/visits`                                       | Visit monitoring          |
| GET    | `/api/v1/admin/prospects/:prospectId/visits`                 | Prospect visits           |
| POST   | `/api/v1/admin/visits/:visitId/delete`                       | Delete visit              |
| POST   | `/api/v1/admin/prospects/:id/approve-deletion`               | Approve deletion          |
| POST   | `/api/v1/admin/prospects/:id/reject-deletion`                | Reject deletion           |
| GET    | `/api/v1/admin/prospects/:id/conversion-form`                | Conversion form data      |
| POST   | `/api/v1/admin/prospects/:id/convert`                        | Convert to customer       |
| GET    | `/api/v1/admin/parent-companies`                             | Search parent companies   |
| GET    | `/api/v1/admin/customers`                                    | Customer list             |
| GET    | `/api/v1/admin/customers/list`                               | Customer list (paginated) |
| GET    | `/api/v1/admin/customers/filter-options`                     | Filter options            |
| GET    | `/api/v1/admin/customers/:id`                                | Customer detail           |
| GET    | `/api/v1/admin/customers/:id/place-details`                  | Place details             |
| DELETE | `/api/v1/admin/customers/:id`                                | Delete customer           |
| GET    | `/api/v1/admin/companies/:id`                                | Parent company            |
| PATCH  | `/api/v1/admin/companies/:id`                                | Update parent company     |
| GET    | `/api/v1/dashboard/admin`                                    | Dashboard data            |

### Auth Endpoints Inside Authenticated Groups

- `POST /api/v1/auth/logout` — Note: this is **not** in the authenticated group; it reads the refresh cookie (no token required)
- `GET /api/v1/auth/me` — **is** behind `authMiddleware.Authenticate`
- `POST /api/v1/auth/logout-all` — **is** behind `authMiddleware.Authenticate`

---

## 9. Endpoint Matrix — Backend

### Data Access Patterns

| Entity           | Sales (role: SE)                                | Admin (role: ADMIN) |
| ---------------- | ----------------------------------------------- | ------------------- |
| Prospects        | Own only (`assignedSalesExecutiveId = user.id`) | All prospects       |
| Customers        | Own only (`salesExecutiveId = user.id`)         | All customers       |
| Visits           | Own only                                        | All visits          |
| Comments         | On own prospects only                           | On any prospect     |
| Parent Companies | ❌ No access                                    | All (via search)    |
| Sales Executives | ❌ No access                                    | List all            |

### Key Observations

- **No "manager" scoping** — there is no way for a SALES_MANAGER (if implemented) to see their team's data
- **Prospect assignment** is set at creation time; no reassignment endpoint exists
- **Customer assignment** is set at conversion time; no reassignment endpoint exists
- **Deletion flow** has a request/approve/reject cycle between SE and Admin

---

## 10. Data Scope

### Prospect Scoping

```sql
-- Sales exec sees only their assigned prospects
WHERE assigned_sales_executive_id = $currentUserId

-- Admin sees all prospects (no WHERE clause on user)
```

### Customer Scoping

```sql
-- Sales exec sees only their customers
WHERE sales_executive_id = $currentUserId

-- Admin sees all customers (no WHERE clause on user)
```

### Visit Scoping

```sql
-- Sales exec sees only visits they created
WHERE sales_executive_id = $currentUserId

-- Admin sees all visits
```

### No Team/Manager Scoping

The data model has **no concept of teams, managers, or hierarchies**. There is no `manager_id` on the User model, no team membership table, no organizational unit.

---

## 11. Menu / Navigation Audit

### Admin Sidebar (AdminLayout.vue)

| #   | Menu Item           | Route                        | Works? | Notes            |
| --- | ------------------- | ---------------------------- | ------ | ---------------- |
| 1   | Dashboard           | `/admin/dashboard`           | ✅     |                  |
| 2   | Sales Executive     | `/admin/sales-executives`    | ⚠️     | Placeholder view |
| 3   | Customer            | `/admin/customers`           | ✅     |                  |
| 4   | Customer Assignment | `/admin/customer-assignment` | ⚠️     | Placeholder view |
| 5   | Visit Monitoring    | `/admin/visit-monitoring`    | ✅     |                  |
| 6   | Prospect Finder     | `/admin/prospect-finder`     | ✅     |                  |
| 7   | Prospect List       | `/admin/prospects/list`      | ✅     |                  |
| 8   | Prospect Assignment | `/admin/prospect-assignment` | ⚠️     | Placeholder view |
| 9   | Reports             | `/admin/reports`             | ⚠️     | Placeholder view |

**Missing Admin Menu Items:**

- User / Account Management
- System Settings

### Sales Navigation (SalesLayout.vue — Desktop Sidebar + Mobile Bottom Nav)

| #   | Menu Item | Route                 | Works? |
| --- | --------- | --------------------- | ------ |
| 1   | Home      | `/sales/dashboard`    | ✅     |
| 2   | Customer  | `/sales/my-customers` | ✅     |
| 3   | Prospect  | `/sales/my-prospects` | ✅     |
| 4   | History   | `/sales/history`      | ✅     |
| 5   | Profile   | `/sales/profile`      | ✅     |

**Missing Sales Menu Items:**

- Pipeline (accessible via `/sales/pipeline` but not in navigation)

### Login View

`LoginView.vue` has a role-based redirect:

- ADMINISTRATOR → `/admin/dashboard`
- SALES_EXECUTIVE → `/sales/dashboard`

---

## 12. Database Audit

### Tables

| Table                     | Purpose              | Key Columns                                                                                                     |
| ------------------------- | -------------------- | --------------------------------------------------------------------------------------------------------------- |
| `users`                   | User accounts        | `id`, `email`, `password_hash`, `full_name`, `role`, `status`, `token_version`, `last_login_at`                 |
| `refresh_sessions`        | Auth sessions        | `id`, `user_id`, `token_hash`, `expires_at`, `revoked_at`, `revoke_reason`, `replaced_by_session_id`            |
| `prospects`               | Prospect records     | `id`, `google_place_id`, `assigned_sales_executive_id`, `status`, `deletion_requested`                          |
| `prospect_visits`         | Visit check-ins/outs | `id`, `prospect_id`, `sales_executive_id`, `check_in_at`, `check_out_at`                                        |
| `prospect_status_history` | Status change audit  | `id`, `prospect_id`, `from_status`, `to_status`, `changed_by_user_id`                                           |
| `parent_companies`        | Company entities     | `id`, `parent_code`, `name`, `kam_assignments` (JSON), `company_contacts` (JSON)                                |
| `customer_sites`          | Converted customers  | `id`, `customer_code`, `parent_company_id`, `source_prospect_id`, `sales_executive_id`, `converted_by_admin_id` |
| `prospect_comments`       | Comments             | `id`, `prospect_id`, `user_id`, `content`                                                                       |

### Indexes

| Table                     | Index                                                             |
| ------------------------- | ----------------------------------------------------------------- |
| `refresh_sessions`        | `(user_id, expires_at)`, `(expires_at)`                           |
| `prospects`               | `(assigned_sales_executive_id, status)`, `(status, createdAt)`    |
| `prospect_visits`         | `(prospect_id, check_in_at)`, `(sales_executive_id, check_in_at)` |
| `prospect_status_history` | `(prospect_id, created_at)`                                       |
| `parent_companies`        | `(name)`                                                          |
| `customer_sites`          | `(sales_executive_id, converted_at)`, `(parent_company_id)`       |
| `prospect_comments`       | `(prospect_id, created_at)`, `(user_id)`                          |

### Foreign Keys (all `ON DELETE Restrict`)

| Source                                       | Target                | Notes             |
| -------------------------------------------- | --------------------- | ----------------- |
| `refresh_sessions.user_id`                   | `users.id`            |                   |
| `prospects.assigned_sales_executive_id`      | `users.id`            |                   |
| `prospect_visits.sales_executive_id`         | `users.id`            |                   |
| `prospect_visits.prospect_id`                | `prospects.id`        |                   |
| `prospect_status_history.changed_by_user_id` | `users.id`            |                   |
| `prospect_status_history.prospect_id`        | `prospects.id`        |                   |
| `customer_sites.sales_executive_id`          | `users.id`            |                   |
| `customer_sites.converted_by_admin_id`       | `users.id`            |                   |
| `customer_sites.parent_company_id`           | `parent_companies.id` |                   |
| `customer_sites.source_prospect_id`          | `prospects.id`        |                   |
| `prospect_comments.prospect_id`              | `prospects.id`        | ON DELETE Cascade |
| `prospect_comments.user_id`                  | `users.id`            |                   |

### Database Gaps

| Gap                                                       | Impact                                       |
| --------------------------------------------------------- | -------------------------------------------- |
| No `manager_id` on `users`                                | Cannot implement hierarchy                   |
| No `teams` or `team_memberships` table                    | Cannot group sales execs                     |
| No `audit_log` table                                      | No generic audit trail                       |
| No `permissions` table                                    | Role-based access is hardcoded               |
| `kam_assignments` and `sales_assignments` are JSON fields | Cannot query or join on them                 |
| No unique constraint on `full_name`                       | No display name uniqueness                   |
| No `deleted_at` soft-delete on users                      | Deleting a user would violate FK constraints |

---

## 13. Gap Matrix

| #   | Gap                                                | Severity | Location           | Impact                                                       |
| --- | -------------------------------------------------- | -------- | ------------------ | ------------------------------------------------------------ |
| 1   | **No SALES_MANAGER role**                          | High     | Everywhere         | Cannot implement tiered access; admin must manage everything |
| 2   | **No admin account management UI/API**             | High     | Backend + Frontend | Users are only seedable; no dynamic management possible      |
| 3   | **No password reset flow**                         | High     | Backend + Frontend | Cannot recover accounts without DB manipulation              |
| 4   | **No user activation/deactivation API**            | Medium   | Backend            | Cannot disable accounts without DB manipulation              |
| 5   | **No team/manager hierarchy**                      | Medium   | Prisma schema      | Cannot scope data by org unit                                |
| 6   | **No session listing/revocation UI**               | Medium   | Backend + Frontend | Admin cannot see or manage active sessions                   |
| 7   | **No rate limiting on auth endpoints**             | Medium   | Backend            | Brute force protection missing                               |
| 8   | **No soft-delete for users**                       | Low      | Prisma schema      | FK constraints prevent user deletion                         |
| 9   | **No audit log table**                             | Low      | Prisma schema      | No generic audit trail                                       |
| 10  | **No Permissions model**                           | Low      | Prisma schema      | Role-based checks are hardcoded strings                      |
| 11  | **Frontend has no role-based component guards**    | Low      | Frontend           | UI elements not hidden by role                               |
| 12  | **`/me` endpoint doesn't return full permissions** | Low      | Backend            | Frontend must hardcode role-to-permission mapping            |
| 13  | **No session cleanup job**                         | Low      | Backend            | Expired/revoked sessions accumulate                          |
| 14  | **Hardcoded seed password**                        | Low      | `cmd/seed/main.go` | All accounts use `password123`                               |
| 15  | **No `updatedAt` on PublicUser**                   | Low      | Backend            | Frontend can't detect stale user info                        |

---

## 14. Implementation Phases

### Phase 1 — SALES_MANAGER Role & Backend Foundation

**Effort:** Medium | **Risk:** Medium

1. Add `SALES_MANAGER` to `UserRole` enum in Prisma
2. Add `SALES_MANAGER` to Go `model.Role` constants
3. Add `SALES_MANAGER` to frontend `UserRole` type
4. Add `manager_id` (self-referencing FK to users) to Prisma User model
5. Create migration
6. Update `RequireRole()` calls to include `SALES_MANAGER` where appropriate
7. Create new route group `/api/v1/manager` with auth + `RequireRole(SALES_MANAGER)`
8. Implement team-scoped data access (prospects/customers/visits filtered by team membership)

### Phase 2 — Admin Account Management

**Effort:** High | **Risk:** Medium

1. Create `UserRepository` methods: `List`, `Create`, `Update`, `Activate`, `Deactivate`, `ResetPassword`
2. Create `AuthHandler` endpoints: `ListUsers`, `CreateUser`, `UpdateUser`, `ActivateUser`, `DeactivateUser`, `ResetPassword`
3. Register routes under `/api/v1/admin/users` (authenticated + RequireRole ADMINISTRATOR)
4. Create frontend `UserManagementView.vue` with PrimeVue DataTable
5. Create user creation/edit dialog/form
6. Add "Users" or "Account Management" to AdminLayout sidebar

### Phase 3 — Session Management

**Effort:** Medium | **Risk:** Low

1. Add `ListSessions` method to SessionRepository (find by user ID)
2. Add `RevokeSession` endpoint to AuthHandler
3. Create frontend session management view (list active sessions, revoke individual sessions)
4. Add session cleanup goroutine (cron-like ticker to delete expired sessions)

### Phase 4 — Password Reset & Security

**Effort:** Medium | **Risk:** Low

1. Add forgot password endpoint (generate reset token, store in DB, email)
2. Add reset password endpoint (validate token, update password)
3. Add rate limiting middleware for `/api/v1/auth/login` and `/api/v1/auth/refresh`
4. Add brute force protection (lockout after N failed attempts)
5. Update seed script to accept password via env var or generate random password

### Phase 5 — Frontend Authorization Polish

**Effort:** Low | **Risk:** Low

1. Create a composable or directive for role-based UI visibility (e.g., `v-if="can('manage_users')"`)
2. Add permission helper to auth store
3. Add role-based component guards to sensitive UI elements
4. Ensure pipeline, menu items, and action buttons respect the user's role

### Phase 6 — Permissions System (Future)

**Effort:** High | **Risk:** High (architectural)

1. Create `permissions` table
2. Create `role_permissions` join table
3. Replace hardcoded `RequireRole()` calls with permission-based checks
4. Add permission endpoint to return user's effective permissions
5. Frontend permission-based rendering

---

## 15. Documentation Plan

The following docs need updating/creation:

| Doc                                     | Content                                                             | Priority |
| --------------------------------------- | ------------------------------------------------------------------- | -------- |
| `backend/README.md` (if exists)         | API auth flow, JWT details, cookie config                           | Medium   |
| `backend/internal/auth/README.md` (new) | Package overview, adding new roles, session lifecycle               | High     |
| `frontend/README.md` (if exists)        | Auth store usage, route guard patterns                              | Medium   |
| Seeding docs                            | How to add users, reset password, generate seed data                | Medium   |
| Deployment notes                        | Required env vars (`JWT_SECRET` min 32 chars, `DATABASE_URL`, etc.) | Low      |

---

## 16. Files Needed Later

### Backend New Files

| File                                              | Purpose                              | Phase |
| ------------------------------------------------- | ------------------------------------ | ----- |
| `backend/internal/auth/handler/admin.go`          | Admin account management endpoints   | 2     |
| `backend/internal/auth/service/admin.go`          | Admin account business logic         | 2     |
| `backend/internal/auth/handler/session.go`        | Session listing/revocation endpoints | 3     |
| `backend/internal/auth/service/ratelimit.go`      | Rate limiting logic                  | 4     |
| `backend/internal/shared/middleware/ratelimit.go` | Rate limit middleware                | 4     |
| `backend/internal/auth/service/permissions.go`    | Permission checking logic (future)   | 6     |
| `backend/internal/auth/model/permission.go`       | Permission data model (future)       | 6     |

### Backend Modified Files

| File                                             | Changes                                                    | Phase |
| ------------------------------------------------ | ---------------------------------------------------------- | ----- |
| `prisma/schema.prisma`                           | Add SALES_MANAGER, manager_id, teams table                 | 1     |
| `backend/internal/auth/model/user.go`            | Add SALES_MANAGER constant                                 | 1     |
| `backend/server/app.go`                          | Add manager route group, admin user routes, session routes | 1-3   |
| `backend/internal/auth/repository/repository.go` | Add User CRUD methods, session listing                     | 2-3   |
| `backend/internal/auth/repository/postgres.go`   | Implement new repository methods                           | 2-3   |

### Frontend New Files

| File                                                   | Purpose                           | Phase |
| ------------------------------------------------------ | --------------------------------- | ----- |
| `frontend/src/views/Admin/Users/AdminUsersView.vue`    | User management list              | 2     |
| `frontend/src/views/Admin/Users/AdminUserForm.vue`     | User create/edit form             | 2     |
| `frontend/src/views/Admin/Users/AdminUserSessions.vue` | Session management                | 3     |
| `frontend/src/composables/usePermission.ts`            | Role/permission helper composable | 5     |

### Frontend Modified Files

| File                                   | Changes                                | Phase |
| -------------------------------------- | -------------------------------------- | ----- |
| `frontend/src/types/auth.ts`           | Add SALES_MANAGER to UserRole          | 1     |
| `frontend/src/router/index.ts`         | Add manager routes, meta for new roles | 1     |
| `frontend/src/stores/auth.ts`          | Add permission helpers                 | 5     |
| `frontend/src/layouts/AdminLayout.vue` | Add User Management link               | 2     |

---

## 17. Uncertainties & Risks

### Technical Risks

| Risk                                                                                                   | Likelihood | Mitigation                                                               |
| ------------------------------------------------------------------------------------------------------ | ---------- | ------------------------------------------------------------------------ |
| JWT secret management: `JWT_SECRET` env var in .env file (not in CI/CD secrets)                        | Medium     | Document requirement; add to README                                      |
| `COOKIE_SECURE=true` in dev causes cookie issues over HTTP                                             | Medium     | Default is `true`; dev setup must use HTTPS or set `COOKIE_SECURE=false` |
| No database migration tooling visible (Prisma migrations directory exists but no CLI usage in scripts) | Low        | Confirm migration workflow before schema changes                         |
| `onDelete: Restrict` on all user FKs prevents user deletion                                            | Medium     | Must implement soft-delete or reassign records before deactivating       |
| Tests use hardcoded test data that may break with schema changes                                       | Low        | Update test fixtures alongside schema changes                            |

### Business Uncertainties

| Question                                                                                         | Impact                          |
| ------------------------------------------------------------------------------------------------ | ------------------------------- |
| Should SALES_MANAGER be able to act on prospects (transition status, check in/out) or only view? | Affects route guard design      |
| What is the SLA for session cleanup? Weeks? Months? Forever?                                     | Affects DB cleanup strategy     |
| Should password reset go through email or just admin-reset?                                      | Affects Phase 4 implementation  |
| Should SALES_MANAGER have their own route prefix or share `/sales`?                              | Affects route design in Phase 1 |
| What's the max team size?                                                                        | Affects data scoping approach   |

### Architectural Notes

1. **The current refresh token rotation with reuse detection is well-implemented** and follows security best practices (NIST SP 800-63B).
2. **The JWT + refresh token split** (short-lived access token + long-lived refresh via HTTP-only cookie) is good practice.
3. **The TokenVersion mechanism** provides a simple but effective way to invalidate all sessions globally.
4. **The Prisma schema is clean** but lacks audit trail tables and proper soft-delete support.
5. **The frontend auth store is minimal** but correct; the refresh interceptor pattern is solid.
6. **Route metadata for roles exists** in the router but is not exhaustive (no manager routes yet).
7. **The Go architecture follows clean-ish layered pattern** (handler → service → repository), but there's some leakage (auth handler has cookie logic, middleware has `Principal()` helper).
8. **No dependency injection framework** — services are manually wired in `bootstrap.Build()`. This is fine for the current scale but may need revisiting as the system grows.
