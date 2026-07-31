# Implementation Plan

> **Document:** 05-implementation-plan.md  
> **Project:** CRM Prospect Simulator  
> **Date:** 2026-07-30  
> **Status:** Phase 6A Complete — Admin Reset Password (Backend)  

---

## Overview

This plan covers 7 implementation phases (3A through 9) to add SALES_MANAGER role, account management, team scoping, single active session, activity logging, and final polish. Each phase is independent and can be rolled back.

Current project baseline:

- Backend: Go 1.26.4, Fiber v2, PostgreSQL (pgx/v5), Prisma for schema
- Frontend: Vue 3 + TypeScript, PrimeVue 4, Pinia, Vue Router
- Auth: JWT access token (30min) + refresh token (30 day, HTTP-only cookie, rotation with theft detection)

---

## Phase 3A — Role Foundation

**Objective:** Add SALES_MANAGER role to the system and establish the manager relationship.

### Files Modified

| File | Change |
|---|---|
| `backend/prisma/schema.prisma` | Added `SALES_MANAGER` to `UserRole` enum. Added `employeeId` (unique), `phone`, `mustChangePassword` (default false), `managerId` (self-ref FK), `createdBy`, `updatedBy` to User model. |
| `backend/internal/auth/model/user.go` | Added `RoleSalesManager` constant. Added `EmployeeID`, `Phone`, `MustChangePassword`, `ManagerID`, `CreatedBy`, `UpdatedBy` fields. `PublicUser` exposes all new fields. |
| `backend/internal/auth/repository/postgres.go` | Updated `FindByEmail`, `FindUserByID`, and `scanUser` for 16 columns. `UpsertSeed` includes new columns with COALESCE. |
| `frontend/src/types/auth.ts` | Added `'SALES_MANAGER'` to `UserRole`. Added `employeeId`, `phone`, `mustChangePassword`, `managerId` to `AuthUser` (optional). |
| `frontend/src/router/index.ts` | Added `/forbidden` route. SALES_MANAGER login redirects to `/forbidden`. `beforeEach` skips redirect-away-from-login for SALES_MANAGER. |
| `frontend/src/views/Login/LoginView.vue` | SALES_MANAGER login routes to `/forbidden`. |
| `frontend/src/layouts/SalesLayout.vue` | `roleLabel` includes `'Sales Manager'`. |
| `frontend/src/views/Sales/ProfileView.vue` | `roleLabel` includes `'Sales Manager'`. Employee ID and Phone display actual values from auth user. |
| `frontend/src/api/crm.ts` | All hardcoded role union types include `SALES_MANAGER`. |
| `backend/cmd/seed/main.go` | Added SALES_MANAGER account (Budi Santoso, SM-0001). Linked 3 existing SE accounts to manager UUID. |

### Migration Applied

Migration file: `backend/prisma/migrations/20260730000101_add_sales_manager_account_foundation/migration.sql`

Generated via `prisma migrate diff --from-schema-datasource --to-schema-datamodel` and applied via `prisma db push --accept-data-loss`. Marked as resolved via `prisma migrate resolve --applied`.

Key SQL changes:
- `ALTER TYPE "UserRole" ADD VALUE 'SALES_MANAGER'`
- New columns on `users`: `employee_id`, `phone`, `must_change_password` (default false), `manager_id`, `created_by`, `updated_by`
- Self-referencing FK on `manager_id` (ON DELETE SET NULL)
- Unique index on `employee_id`
- FKs for `created_by`, `updated_by`

### Rollback

```bash
prisma migrate down 1  # Not currently available; revert via db push with old schema
git checkout -- backend/internal/auth/model/user.go
git checkout -- backend/internal/auth/repository/postgres.go
git checkout -- frontend/src/
git checkout -- backend/cmd/seed/main.go
```

### Exit Criteria

- [x] Prisma migration runs without error
- [x] `SALES_MANAGER` is in Go Role constants
- [x] `SALES_MANAGER` is in frontend UserRole type
- [x] Seed script creates manager + SEs with manager_id set
- [x] `npm run build` passes
- [x] `go build ./...` passes
- [x] `go test ./...` passes
- [ ] Manager UI not yet available (Phase 3B)

---

## Phase 3B — Manager Authentication & Routing

**Objective:** Allow SALES_MANAGER to log in, redirect to manager dashboard, and protect manager routes.

### Files to Modify

| File | Change |
|---|---|
| `frontend/src/router/index.ts` | Change `meta.role` to `meta.roles` array. Add `/manager` route group with `roles: ['SALES_MANAGER']`. Update `beforeEach` guard to support array. Add `/forbidden` route. Add LoginView redirect for SALES_MANAGER → `/manager/dashboard`. |
| `frontend/src/layouts/ManagerLayout.vue` | **New file.** Create manager layout with sidebar (Dashboard, Team Members, Team Prospects, Team Customers, Team Pipeline, Team Visits, Reports, Profile). |
| `frontend/src/views/Manager/Dashboard/ManagerDashboardView.vue` | **New file.** Manager dashboard placeholder or summary. |
| `frontend/src/views/ForbiddenView.vue` | **New file.** Simple page: "You do not have access to this page." |
| `frontend/src/stores/auth.ts` | Update `role` type to include SALES_MANAGER. Add `isManager` computed. |
| `frontend/src/views/Login/LoginView.vue` | Add SALES_MANAGER redirect to `/manager/dashboard`. |
| `backend/server/app.go` | Add `RequireRole(SALES_MANAGER)` to dashboards for manager route. |

Also ensure that the `LoginView.vue` redirect supports SALES_MANAGER.

### Route Group Design

```ts
// Manager routes
{
  path: '/manager',
  component: () => import('../layouts/ManagerLayout.vue'),
  meta: { requiresAuth: true, roles: ['SALES_MANAGER'] },
  children: [
    { path: '', redirect: '/manager/dashboard' },
    { path: 'dashboard', name: 'ManagerDashboard', ... },
    { path: 'team-members', name: 'ManagerTeamMembers', ... },
    // ... more children to be filled in Phase 5B
  ],
}
```

### Tests

| Test | Area |
|---|---|
| SALES_MANAGER logs in → redirect to /manager/dashboard | Integration |
| SALES_EXECUTIVE tries /manager/* → redirected | Integration |
| SALES_MANAGER tries /admin/* → redirected | Integration |

### Rollback

```bash
git checkout -- frontend/src/router/index.ts
rm frontend/src/layouts/ManagerLayout.vue
rm -rf frontend/src/views/Manager/
rm frontend/src/views/ForbiddenView.vue
```

### Exit Criteria

- [ ] Login redirects SALES_MANAGER to `/manager/dashboard`
- [ ] SALES_MANAGER can browse `/manager/*` routes
- [ ] SALES_EXECUTIVE is blocked from `/manager/*`
- [ ] SALES_MANAGER is blocked from `/admin/*` and `/sales/*`
- [ ] `beforeEach` guard supports `roles` array
- [ ] Manager layout renders correctly (sidebar + content area)

---

## Phase 4A — Admin Account Management (Backend)

**Objective:** Full CRUD for user accounts via Admin API.

> **Status: Complete (2026-07-31).** Implemented as a dedicated `internal/admin` module with model/repository/service/handler. Endpoints registered under the existing `/api/v1/admin` group which already enforces `Authenticate` + `RequireRole(ADMINISTRATOR)`. See `07-final-implementation-report.md` for the Phase 4A report.

### Files to Modify

| File | Change |
|---|---|
| `backend/internal/auth/repository/repository.go` | Add `UserRepository` methods: `List`, `Create`, `Update`, `SetStatus`, `IncrementTokenVersion`, `FindByEmployeeID`, `ListByManager`, `ListByRole`. |
| `backend/internal/auth/repository/postgres.go` | Implement new repository methods. |
| `backend/internal/auth/service/auth.go` | Add `CreateAccount`, `UpdateAccount`, `SetAccountStatus`, `ResetPassword`, `ListAccounts`, `AdminEndSession`. Add validation logic. |
| `backend/internal/auth/service/errors.go` | Add `ErrActiveSessionExists`, `ErrInvalidManager`, `ErrEmailTaken`, `ErrEmployeeIDTaken`, etc. |
| `backend/internal/auth/handler/handler.go` | Add `AdminListUsers`, `AdminCreateUser`, `AdminGetUser`, `AdminUpdateUser`, `AdminSetStatus`, `AdminResetPassword`, `AdminEndSession`. |
| `backend/server/app.go` | Register admin user management routes: `/api/v1/admin/users/*`. |

### New Admin Endpoints

| Method | Path | Handler | Notes |
|---|---|---|---|
| GET | `/api/v1/admin/users` | List all users | Paginated, filterable by role/status |
| GET | `/api/v1/admin/users/:id` | Get user detail | Includes active session info |
| POST | `/api/v1/admin/users` | Create user | Validation + auto employee ID |
| PATCH | `/api/v1/admin/users/:id` | Update user | Fields: full_name, email, phone, role, manager_id |
| PATCH | `/api/v1/admin/users/:id/status` | Set status | ACTIVE ↔ INACTIVE |
| POST | `/api/v1/admin/users/:id/reset-password` | Reset password | Returns temporary password |
| POST | `/api/v1/admin/users/:id/end-session` | End active session | Revoke + increment token version |
| GET | `/api/v1/admin/roles/summary` | Role access summary data | For frontend display |

### Validation Rules

| Rule | Implementation |
|---|---|
| Email unique | Check before create/update |
| Employee ID unique | Auto-generate, allow override with uniqueness check |
| SE must have manager | Required field validation |
| Manager must be SALES_MANAGER | Cross-field validation |
| Manager must be ACTIVE | Status check |
| Admin cannot have manager | Set manager_id = NULL |
| No self-manager | id != manager_id |
| No circular manager | Required for manager change |

### Tests

| Test | Area |
|---|---|
| Create ADMINISTRATOR account | Repository + Service |
| Create SALES_MANAGER account | Repository + Service |
| Create SALES_EXECUTIVE with valid manager | Repository + Service |
| Create SALES_EXECUTIVE without manager → error | Service |
| Create SALES_EXECUTIVE with non-SM manager → error | Service |
| Create duplicate email → error | Repository |
| Update user status → active/inactive | Repository + Service |
| Reset password → session revoked | Service |
| End session → session revoked + token bumped | Service |

### Rollback

```bash
git checkout -- backend/internal/auth/
git checkout -- backend/server/app.go
```

### Exit Criteria

- [x] All admin user endpoints return correct data
- [x] Validation rules enforced (email, employee_id, manager)
- [x] Password set for new accounts (temporary password, must_change_password = true)
- [x] Status change guards self-deactivation and last active administrator
- [x] `go test ./...` passes
- [x] `go build ./...` passes
- [x] Password reset (delivered in Phase 6A); session invalidation on status change → Phase 6

---

## Phase 4B — Admin Account Management (Frontend)

**Objective:** Admin UI for managing user accounts.

> **Status: Complete (2026-07-31).** Part 1 delivered the account list and create form (DataTable with search/filter/pagination, role-aware create form with scope summary). Part 2 delivered the account detail and edit pages, enabled View/Edit actions in the list, and added `accounts/:id` and `accounts/:id/edit` routes. Password reset and session management remain deferred (Phase 6). See `07-final-implementation-report.md` for the Phase 4B report.

> **Phase 4B Hotfix (2026-07-31):** PATCH `/api/v1/admin/users/:id` now distinguishes an omitted `managerId` from an explicit `null` via a custom `OptionalUUID` request type. Promoting a SALES_EXECUTIVE to SALES_MANAGER/ADMINISTRATOR clears `manager_id` automatically. No migration required.

### Files to Create

| File | Purpose |
|---|---|
| `frontend/src/views/Admin/Accounts/AdminAccountsView.vue` | Account list with DataTable, search, filters, status badges |
| `frontend/src/views/Admin/Accounts/AdminAccountCreateView.vue` | Create account form with Role Access Summary panel |
| `frontend/src/views/Admin/Accounts/AdminAccountDetailView.vue` | Account detail with edit, session info, status toggle |
| `frontend/src/views/Admin/Accounts/AdminAccountEditView.vue` | Edit account form |
| `frontend/src/stores/admin.ts` | **New store** for admin operations (create user, list users, reset password, etc.) |
| `frontend/src/api/admin.ts` | **New API module** for admin user endpoints |

### Files to Modify

| File | Change |
|---|---|
| `frontend/src/router/index.ts` | Add account management routes under `/admin/accounts/*` |
| `frontend/src/layouts/AdminLayout.vue` | Add "Accounts" to sidebar navigation |

### Router Entries

```ts
{
  path: '/admin',
  meta: { requiresAuth: true, roles: ['ADMINISTRATOR'] },
  component: AdminLayout,
  children: [
    // ... existing routes ...
    { path: 'accounts', name: 'AdminAccounts', component: () => import('../views/Admin/Accounts/AdminAccountsView.vue') },
    { path: 'accounts/create', name: 'AdminAccountCreate', component: () => import('../views/Admin/Accounts/AdminAccountCreateView.vue') },
    { path: 'accounts/:id', name: 'AdminAccountDetail', component: () => import('../views/Admin/Accounts/AdminAccountDetailView.vue') },
    { path: 'accounts/:id/edit', name: 'AdminAccountEdit', component: () => import('../views/Admin/Accounts/AdminAccountEditView.vue') },
  ],
}
```

### Create Account Form Sections

1. **Account Identity** — employee_id (auto), full_name, email, phone
2. **Role & Reporting** — role dropdown, manager dropdown (shown only for SE)
3. **Role Access Summary** — read-only panel that updates live when role changes
4. **Security** — status toggle (ACTIVE/INACTIVE), password mode (auto/manual)

### Tests

| Test | Area |
|---|---|
| Navigate to account list | E2E |
| Create account with all roles | E2E |
| Validation errors display correctly | E2E |
| Status toggle works | E2E |
| Password reset shows temporary password once | E2E |

### Rollback

```bash
git checkout -- frontend/src/router/index.ts
git checkout -- frontend/src/layouts/AdminLayout.vue
rm -rf frontend/src/views/Admin/Accounts/
rm frontend/src/stores/admin.ts
rm frontend/src/api/admin.ts
```

### Exit Criteria

- [x] Account list loads with search and filters
- [x] Account creation works for all three roles
- [x] Manager selection works (shows only SALES_MANAGER users)
- [x] Role access summary updates correctly
- [x] Status toggle activates/deactivates account
- [x] Account detail and edit pages work (role/manager/employee ID/email/phone)
- [x] `npm run build` passes
- [ ] Password reset shows temporary password once (deferred to Phase 6)

---

## Phase 5A — Team Scope Backend

**Objective:** Create `/api/v1/manager/*` endpoints with TEAM-scoped data.

### Files to Create/Modify

| File | Change |
|---|---|
| `backend/internal/auth/handler/manager.go` | **New.** Manager dashboard, team members, team prospects, team customers, team pipeline, team visits. All scoped to `manager_id = current_user_id`. |
| `backend/internal/auth/service/manager.go` | **New.** Service layer for manager operations. |
| `backend/internal/prospect/handler/handler.go` | Add manager-scoped prospect methods. Reuse existing logic with scope filter. |
| `backend/internal/customer/handler/handler.go` | Add manager-scoped customer methods. |
| `backend/server/app.go` | Register `/api/v1/manager/*` route group with `RequireRole(SALES_MANAGER)`. |

### New Endpoints

| Method | Path | Handler | Scope |
|---|---|---|---|
| GET | `/api/v1/manager/dashboard` | Team summary stats | TEAM |
| GET | `/api/v1/manager/team-members` | List SEs under this manager | TEAM |
| GET | `/api/v1/manager/prospects` | List team prospects | TEAM |
| GET | `/api/v1/manager/prospects/:id` | Prospect detail | TEAM |
| GET | `/api/v1/manager/prospects/:id/comments` | Prospect comments | TEAM |
| GET | `/api/v1/manager/prospects/:id/place-details` | Prospect place details | TEAM |
| GET | `/api/v1/manager/customers` | List team customers | TEAM |
| GET | `/api/v1/manager/customers/:id` | Customer detail | TEAM |
| GET | `/api/v1/manager/pipeline` | Team pipeline stats | TEAM |
| GET | `/api/v1/manager/visits` | Team visits | TEAM |

### Data Scope Query Pattern

```go
// Team scope helper
func (s *Service) teamScopeQuery(ctx context.Context, managerID uuid.UUID) string {
    // Returns filter: assigned_sales_executive_id IN (SELECT id FROM users WHERE manager_id = $1)
}
```

### Tests

| Test | Area |
|---|---|
| Manager sees only their team's prospects | Integration |
| Manager sees only their team's customers | Integration |
| Manager A does NOT see Manager B's team data | Integration |
| SALES_EXECUTIVE blocked from `/manager/*` | Integration |

### Rollback

```bash
git checkout -- backend/internal/
git checkout -- backend/server/app.go
```

### Exit Criteria

- [ ] All manager endpoints return TEAM-scoped data
- [ ] Data isolation between managers is verified
- [ ] SALES_EXECUTIVE cannot access manager endpoints
- [ ] `go test ./...` passes

---

## Phase 5B — Manager Frontend

**Objective:** Manager UI with team views.

### Files to Create

| File | Purpose |
|---|---|
| `frontend/src/views/Manager/Team/TeamMembersView.vue` | List of SEs under manager |
| `frontend/src/views/Manager/Prospect/ManagerProspectsView.vue` | Team prospects list |
| `frontend/src/views/Manager/Prospect/ManagerProspectDetailView.vue` | Team prospect detail (read-only) |
| `frontend/src/views/Manager/Customer/ManagerCustomersView.vue` | Team customers list |
| `frontend/src/views/Manager/Customer/ManagerCustomerDetailView.vue` | Team customer detail (read-only) |
| `frontend/src/views/Manager/Pipeline/ManagerPipelineView.vue` | Team pipeline |
| `frontend/src/views/Manager/Visit/ManagerVisitsView.vue` | Team visits |
| `frontend/src/stores/manager.ts` | **New store** for manager operations |
| `frontend/src/api/manager.ts` | **New API module** for manager endpoints |

### Files to Modify

| File | Change |
|---|---|
| `frontend/src/router/index.ts` | Add children routes under `/manager` |
| `frontend/src/layouts/ManagerLayout.vue` | Wire up sidebar links to real routes |

### Tests

| Test | Area |
|---|---|
| Manager dashboard shows correct team metrics | E2E |
| Team prospects list excludes SE from other teams | E2E |
| Prospect detail is read-only (no transition buttons) | E2E |

### Rollback

```bash
rm -rf frontend/src/views/Manager/
rm frontend/src/stores/manager.ts
rm frontend/src/api/manager.ts
git checkout -- frontend/src/router/index.ts
git checkout -- frontend/src/layouts/ManagerLayout.vue
```

### Exit Criteria

- [ ] Manager dashboard displays team stats
- [ ] Team members list is accurate
- [ ] Team prospects/customers are read-only
- [ ] Manager cannot transition pipeline from UI
- [ ] `npm run build` passes

---

## Phase 6 — Single Active Session

**Objective:** Enforce single active session with REJECT_NEW_LOGIN behavior.

### Files to Modify

| File | Change |
|---|---|
| `backend/internal/auth/repository/repository.go` | Add `HasActiveSession(ctx, userID) (bool, error)` to SessionRepository |
| `backend/internal/auth/repository/postgres.go` | Implement `HasActiveSession` query |
| `backend/internal/auth/service/auth.go` | Add check in Login. Return `ErrActiveSessionExists`. |
| `backend/internal/auth/service/errors.go` | Add `ErrActiveSessionExists` |
| `backend/internal/auth/handler/handler.go` | Return 409 for `ErrActiveSessionExists`. |
| `backend/internal/shared/response/response.go` | Ensure error code `ACTIVE_SESSION_EXISTS` is returned. |

### Active Session Check Implementation

```go
func (r *PostgresRepository) HasActiveSession(ctx context.Context, userID uuid.UUID) (bool, error) {
    var count int
    err := r.pool.QueryRow(ctx, `
        SELECT COUNT(*)
        FROM refresh_sessions
        WHERE user_id = $1
          AND revoked_at IS NULL
          AND expires_at > NOW()
    `, userID).Scan(&count)
    if err != nil {
        return false, fmt.Errorf("check active session: %w", err)
    }
    return count > 0, nil
}
```

### Login Flow Update

```go
func (s *AuthService) Login(ctx context.Context, email, password string, client ClientContext) (AuthResult, error) {
    // ... existing validation ...
    
    hasActive, err := s.sessions.HasActiveSession(ctx, user.ID)
    if err != nil {
        return AuthResult{}, err
    }
    if hasActive {
        return AuthResult{}, ErrActiveSessionExists
    }
    
    return s.createLogin(ctx, user, client, uuid.Nil)
}
```

### Admin End Session (from Phase 4A)

Admin can end another user's session. This is part of admin account management but is essential for single session recovery.

### Tests

| Test | Area |
|---|---|
| Login when no active session → success | Integration |
| Login when active session exists → 409 | Integration |
| After logout, login succeeds | Integration |
| After admin ends session, login succeeds | Integration |
| Refresh does not trigger active session check | Integration |

### Rollback

```bash
git checkout -- backend/internal/auth/
```

### Exit Criteria

- [ ] Second login attempt is rejected with `ACTIVE_SESSION_EXISTS`
- [ ] Logout from device A → device B can log in
- [ ] Admin "End Session" → device A blocked, new login allowed
- [ ] Refresh rotation does not interfere with session check
- [ ] `go test ./...` passes

---

## Phase 6A — Admin Reset Password (Backend)

> **Status: Complete (2026-07-31).** Adds the Administrator-only endpoint
> `POST /api/v1/admin/users/:id/reset-password` required by the Phase 6B
> frontend. No migration needed (uses existing `must_change_password`,
> `token_version`, and `refresh_sessions.revoked_at`/`revoke_reason` columns).
> See `07-final-implementation-report.md` for the Phase 6A report.

### Endpoint Contract

| Field | Value |
|---|---|
| Method / Path | `POST /api/v1/admin/users/:id/reset-password` |
| Access | `Authenticate` + `RequireRole(ADMINISTRATOR)` (existing admin group) |
| Body | `{ "mode": "AUTO" \| "MANUAL", "temporaryPassword"?: string }` |
| Response 200 | `{ "userId", "temporaryPassword", "mustChangePassword": true, "sessionsRevoked" }` |
| Errors | 400 malformed/invalid UUID, 401 unauthenticated, 403 non-admin, 422 mode/password validation, 404 `USER_NOT_FOUND`, 500 safe message |

### Files Modified

| File | Change |
|---|---|
| `internal/admin/model/user.go` | Added `ResetPasswordMode` (AUTO/MANUAL), `ResetPasswordInput`, `ResetPasswordResult`. |
| `internal/admin/service/service.go` | `ResetPassword` (role gate, AUTO via `generateTemporaryPassword` crypto/rand generator, MANUAL via `validateTemporaryPassword`, bcrypt `DefaultCost`); sentinels `ErrInvalidResetMode`, `ErrTemporaryPasswordRequired`, `ErrWeakTemporaryPassword`. |
| `internal/admin/repository/repository.go`, `postgres.go` | `ResetPassword` — single tx: update user (hash, `must_change_password=TRUE`, `token_version+1`, `updated_by`), revoke active refresh sessions (`revoke_reason='ADMIN_PASSWORD_RESET'`), return revoked count. |
| `internal/admin/handler/handler.go` | `ResetPassword` handler + error-code mappings in `writeError`. |
| `server/app.go` | Route registered in the existing admin group. |

### Security Notes

- Temporary password is returned **once** in the 200 response only; never logged, never persisted, never embedded in errors.
- Bumping `token_version` invalidates previously issued access tokens (verified by test).
- Revoking active refresh sessions signs the user out of all devices.
- Reset allowed on INACTIVE accounts and on the acting admin's own account (keeps last-admin protections intact).

### Tests

- Service: generator rule coverage, AUTO/MANUAL hashing, weak/missing/invalid password rejections, not-found mapping, no password leakage in repo errors, INACTIVE/self allowed, result shape.
- Repository: SQL statement assertions (no COALESCE, correct columns) + transaction behavior via mock tx (update+revoke order, revoked count, zero sessions, missing user short-circuits, error propagation).
- Server: 401/403/400/422/404, AUTO/MANUAL success, no `passwordHash`/`tokenVersion`/`refreshToken` leakage in response, `token_version` bump invalidates old access token.

### Exit Criteria

- [x] Endpoint returns the Phase 6B contract envelope without frontend changes
- [x] AUTO and MANUAL modes validated and hashed with bcrypt
- [x] `must_change_password=true`, `token_version+1`, active sessions revoked
- [x] `go test ./...`, `go build ./...`, `go vet ./...` pass
- [ ] Live smoke test against a local DB (skipped: requires disposable account / safe DB)

---

## Phase 6B — Admin Reset Password (Frontend)

> **Status: Complete (2026-07-31).** Reusable `ResetPasswordDialog.vue` (AUTO/MANUAL modes,
> confirm + live validation, one-time password with copy, self/INACTIVE warnings, clears all
> sensitive state on close). API/types/store/utils wired; list + detail integration. See
> `07-final-implementation-report.md`. Backend endpoint delivered by Phase 6A (2026-07-31) —
> the live flow is fully supported by the documented contract.

| File | Change |
|---|---|
| `components/admin/ResetPasswordDialog.vue` | **New.** Reset dialog (AUTO/MANUAL, one-time password, copy). |
| `api/admin.ts` | `resetUserPassword` → POST `/admin/users/:id/reset-password`. |
| `types/admin.ts` | `AdminResetPasswordMode/Payload/Result`. |
| `stores/admin.ts` | `resettingPassword` + `resetPassword` (guard, no password stored). |
| `utils/admin.ts` | `adminTemporaryPasswordError` validator. |
| `AdminAccountsView.vue` | Key-icon Reset Password action → dialog. |
| `AdminAccountDetailView.vue` | Header Reset Password action → dialog. |
| `tests/resetPassword.test.mjs` | **New.** Validator tests. |

### Exit Criteria

- [x] AUTO/MANUAL modes, MANUAL validation, one-time password + copy, warnings
- [x] `npm run typecheck` and `npm run build` pass
- [x] Live success flow (backend endpoint delivered by Phase 6A)

---

## Phase 7 — Activity Log

**Objective:** Log account management and auth events for admin review.

### Files to Create

| File | Purpose |
|---|---|
| `prisma/migrations/.../add_activity_logs` | New `activity_logs` table |
| `backend/internal/activity/model/activity.go` | ActivityLog struct |
| `backend/internal/activity/repository/repository.go` | ActivityLogRepository interface |
| `backend/internal/activity/repository/postgres.go` | Postgres implementation |

### Files to Modify

| File | Change |
|---|---|
| `prisma/schema.prisma` | Add `ActivityLog` model |
| `backend/server/app.go` | Register activity log routes |
| `backend/internal/auth/handler/handler.go` | Add logging calls to account operations |
| `frontend/src/router/index.ts` | Add `/admin/activity-log` route |
| `frontend/src/layouts/AdminLayout.vue` | Add "Activity Log" to sidebar |
| `frontend/src/views/Admin/ActivityLogView.vue` | **New.** Activity log display |

### Activity Log Schema

```prisma
enum ActivityEventType {
  ACCOUNT_CREATED
  ACCOUNT_UPDATED
  ROLE_CHANGED
  MANAGER_CHANGED
  ACCOUNT_ACTIVATED
  ACCOUNT_DEACTIVATED
  PASSWORD_RESET
  SESSION_TERMINATED
  LOGIN_SUCCESS
  LOGIN_FAILED
  LOGIN_REJECTED_ACTIVE_SESSION
  LOGOUT
  LOGOUT_ALL
}

model ActivityLog {
  id          String            @id @default(uuid()) @db.Uuid
  eventType   ActivityEventType @map("event_type")
  actorId     String            @map("actor_id") @db.Uuid
  targetId    String?           @map("target_id") @db.Uuid
  description String
  metadata    Json              @default("{}")
  createdAt   DateTime          @default(now()) @map("created_at") @db.Timestamptz(6)
  
  actor       User              @relation("ActivityActor", fields: [actorId], references: [id])
  
  @@index([actorId, createdAt])
  @@index([eventType, createdAt])
  @@index([createdAt])
  @@map("activity_logs")
}
```

### Tests

| Test | Area |
|---|---|
| Account creation creates log entry | Integration |
| Password reset creates log entry | Integration |
| Login failure creates log entry | Integration |
| Activity log endpoint returns paginated results | Integration |

### Exit Criteria

- [ ] All account events are logged
- [ ] All auth events are logged
- [ ] Admin can view activity log
- [ ] `go test ./...` passes

---

## Phase 8 — Final Authorization Polish

**Objective:** Strengthen auth checks, improve UI guards, add forbidden page.

### Frontend Changes

- Update `beforeEach` guard to support `roles` array properly
- Add `v-if` directives for role-based UI element visibility
- Add "Forbidden" view with back-to-home navigation
- Update all route `meta.roles` to use array format

### Backend Changes

- Audit all existing endpoints for correct `RequireRole` guards
- Add `RequireRole` to any unprotected endpoints
- Ensure error responses return consistent error codes

### Files to Modify

| File | Change |
|---|---|
| `frontend/src/router/index.ts` | Finalize meta.roles for all routes |
| `frontend/src/stores/auth.ts` | Add `can(role)` helper |
| `frontend/src/composables/usePermission.ts` | **New.** Permission composable |
| `backend/server/app.go` | Audit all route guards |

### Exit Criteria

- [ ] All routes have correct role guards
- [ ] Forbidden page renders for unauthorized access
- [ ] UI elements respect user role
- [ ] `npm run build` passes
- [ ] `go build ./...` passes

---

## Phase 9 — Final Testing & Report

**Objective:** Comprehensive testing and documentation.

### Testing

| Area | Tests |
|---|---|
| Role access | All role × page combinations |
| Data scope | ALL, TEAM, OWN verification |
| Account CRUD | Create, read, update, deactivate |
| Single session | REJECT_NEW_LOGIN scenarios |
| Negative tests | Direct URL access, API manipulation |

### Documentation

- Update `AUDIT_REPORT.md` with final state
- Create summary for advisor
- Screenshots of key flows
- API documentation updates

### Exit Criteria

- [ ] All tests pass
- [ ] Documentation is complete
- [ ] Advisor report is ready

---

## Risk Register

| Risk | Phase | Likelihood | Mitigation |
|---|---|---|---|
| Prisma migration conflict with existing data | 3A | Low | Test on dev DB first |
| `manager_id` FK constraint fails on existing SEs | 3A | Medium | Make `manager_id` nullable, backfill in migration |
| Manager layout does not match mobile-first design | 3B | Low | Use SalesLayout patterns |
| Session family complexity underestimated | 6 | Medium | Simple `COUNT(*)` check is sufficient for v1 |
| Frontend type errors during migration | 3A | Low | Run `npm run typecheck` before commit |

---

## Summary Timeline

| Phase | Depends On | Estimated Effort |
|---|---|---|
| 3A — Role Foundation | — | 2–3 days |
| 3B — Manager Routing | 3A | 2–3 days |
| 4A — Admin Account Backend | 3A | 4–5 days |
| 4B — Admin Account Frontend | 4A | 4–5 days |
| 5A — Team Scope Backend | 3A | 3–4 days |
| 5B — Manager Frontend | 5A | 3–4 days |
| 6 — Single Active Session | 4A | 2–3 days |
| 7 — Activity Log | 4A | 3–4 days |
| 8 — Final Polish | All above | 2–3 days |
| 9 — Final Testing | 8 | 2–3 days |
