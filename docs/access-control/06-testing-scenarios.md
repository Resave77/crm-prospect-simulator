# Testing Scenarios

> **Document:** 06-testing-scenarios.md  
> **Project:** CRM Prospect Simulator  
> **Date:** 2026-07-30  
> **Status:** Draft — Design Phase  

---

## 1. Role Access Tests

### 1.1 Page Access by Role

| # | Test Scenario | Given | When | Expected Result | Type |
|---|---|---|---|---|---|
| 1.1 | Admin accesses admin pages | Logged in as ADMINISTRATOR | Navigate to `/admin/dashboard`, `/admin/customers`, `/admin/prospects/list`, etc. | All admin pages load successfully | E2E |
| 1.2 | Admin blocked from Check In | Logged in as ADMINISTRATOR | Navigate to `/sales/my-prospects/:id/check-in` | Redirected to Forbidden or admin dashboard | Integration |
| 1.3 | SM accesses manager pages | Logged in as SALES_MANAGER | Navigate to `/manager/dashboard`, `/manager/team-members`, etc. | All manager pages load successfully | E2E |
| 1.4 | SM blocked from Account Management | Logged in as SALES_MANAGER | Navigate to `/admin/accounts` | Redirected to Forbidden or manager dashboard | Integration |
| 1.5 | SM blocked from Check In | Logged in as SALES_MANAGER | Navigate to `/sales/my-prospects/:id/check-in` | Redirected to Forbidden | Integration |
| 1.6 | SE accesses sales pages | Logged in as SALES_EXECUTIVE | Navigate to `/sales/dashboard`, `/sales/my-prospects`, etc. | All sales pages load successfully | E2E |
| 1.7 | SE blocked from Admin pages | Logged in as SALES_EXECUTIVE | Navigate to `/admin/accounts` | Redirected to Forbidden or sales dashboard | Integration |
| 1.8 | SE blocked from Manager pages | Logged in as SALES_EXECUTIVE | Navigate to `/manager/dashboard` | Redirected to Forbidden or sales dashboard | Integration |
| 1.9 | Direct URL bypass attempt | Logged in as SALES_EXECUTIVE | Manually enter `/admin/accounts` in browser URL bar | Redirected to Forbidden | E2E |
| 1.10 | API bypass attempt | Use SALES_EXECUTIVE token | Send GET `/api/v1/admin/users` | HTTP 403 Forbidden | API |
| 1.11 | Unauthenticated access | No auth token | Navigate to any protected route | Redirected to `/login` | E2E |
| 1.12 | Unauthenticated API access | No auth header | Send GET `/api/v1/sales/prospects` | HTTP 401 Unauthorized | API |

### 1.2 Negative Tests

| # | Test Scenario | Given | When | Expected Result | Type |
|---|---|---|---|---|---|
| 1.13 | Expired token | Expired JWT | Send any authenticated API request | HTTP 401 → refresh called | Integration |
| 1.14 | Revoked token | Token for user whose session was revoked | Send API request | HTTP 401 | Integration |
| 1.15 | Token after password reset | Old JWT (before password reset) | Send API request | HTTP 401 (token version mismatch) | Integration |
| 1.16 | Invalid role in token | Token with unknown role | Send API request | HTTP 401 | Integration |
| 1.17 | Admin tries Check In API | ADMINISTRATOR token | Send POST `/api/v1/sales/prospects/:id/visits/check-in` | HTTP 403 Forbidden | API |

---

## 2. Data Scope Tests

### 2.1 Administrator (ALL)

| # | Test Scenario | Given | When | Expected Result | Type |
|---|---|---|---|---|---|
| 2.1 | Admin sees all prospects | Admin with 3 SE teams, each with 5 prospects | GET `/api/v1/admin/prospects/pipeline` | Response includes all 15 prospects | API |
| 2.2 | Admin sees all customers | Multiple customers across all SEs | GET `/api/v1/admin/customers` | Response includes all customers | API |
| 2.3 | Admin sees all visits | Visits by all SEs | GET `/api/v1/admin/visits` | Response includes all visits | API |
| 2.4 | Admin sees all users | Accounts of all roles | GET `/api/v1/admin/users` | Response includes Admin, SM, SE accounts | API |

### 2.2 Sales Manager (TEAM)

| # | Test Scenario | Given | When | Expected Result | Type |
|---|---|---|---|---|---|
| 2.5 | Manager A sees team A prospects | Manager A manages SE1, SE2. SE1 has 3 prospects, SE2 has 2 | GET `/api/v1/manager/prospects` | Returns 5 prospects (SE1 + SE2) | API |
| 2.6 | Manager A sees team A customers | SE1 has 2 customers, SE2 has 1 | GET `/api/v1/manager/customers` | Returns 3 customers | API |
| 2.7 | Manager A sees team A visits | SE1 has 4 visits, SE2 has 1 | GET `/api/v1/manager/visits` | Returns 5 visits | API |
| 2.8 | Manager A does NOT see Manager B's data | Manager B manages SE3 with 10 prospects | GET `/api/v1/manager/prospects` | Results contain NO SE3 prospects | API |
| 2.9 | Manager A sees team members correctly | Manager A has SE1, SE2 under her | GET `/api/v1/manager/team-members` | Returns SE1 and SE2 only | API |
| 2.10 | Manager A sees prospect detail for team | Manager A, prospect owned by SE1 | GET `/api/v1/manager/prospects/:id` | Returns prospect detail | API |
| 2.11 | Manager A cannot see SE3's prospect | Manager A tries SE3's prospect ID | GET `/api/v1/manager/prospects/:id` | HTTP 404 or 403 | API |

### 2.3 Sales Executive (OWN)

| # | Test Scenario | Given | When | Expected Result | Type |
|---|---|---|---|---|---|
| 2.12 | SE1 sees own prospects only | SE1 has 3 prospects, SE2 has 5 | GET `/api/v1/sales/prospects` | Returns 3 prospects | API |
| 2.13 | SE1 sees own customers only | SE1 has 2 customers, SE2 has 4 | GET `/api/v1/sales/customers` | Returns 2 customers | API |
| 2.14 | SE1 sees own visits only | SE1 has 6 visits, SE2 has 2 | GET `/api/v1/sales/visits` | Returns 6 visits | API |
| 2.15 | SE cannot access SE2's prospect detail | SE1 logged in, tries SE2's prospect ID | GET `/api/v1/sales/prospects/:id` (SE2's ID) | HTTP 404 or 403 | API |
| 2.16 | SE cannot access SE2's customer detail | SE1 logged in, tries SE2's customer ID | GET `/api/v1/sales/customers/:id` (SE2's ID) | HTTP 404 or 403 | API |

### 2.4 Cross-Scope Verification

| # | Test Scenario | Given | Expected Result |
|---|---|---|---|
| 2.17 | Admin count > SM count > SE count | Same data set | Admin returns most records, SM returns subset, SE returns own only |
| 2.18 | SM with no SEs assigned | SM has 0 SEs under them | `/manager/*` endpoints return empty arrays |
| 2.19 | SE with no prospects | SE has 0 assigned prospects | `/sales/prospects` returns empty array |
| 2.20 | SE with no customers | SE has 0 customers | `/sales/customers` returns empty array |

---

## 3. Account Tests

### 3.1 Account Creation

| # | Test Scenario | Given | When | Expected Result | Type |
|---|---|---|---|---|---|
| 3.1 | Create ADMINISTRATOR | Admin logged in | Fill form with ADMIN role, submit | Account created, no manager_id | E2E |
| 3.2 | Create SALES_MANAGER | Admin logged in | Fill form with SM role, submit | Account created, manager_id = null | E2E |
| 3.3 | Create SALES_EXECUTIVE with manager | Admin logged in | Fill form with SE role, select SM | Account created, manager_id = selected SM | E2E |
| 3.4 | Create SE without manager → error | Admin logged in | Fill form with SE role, no manager selected | Validation error: manager required | E2E |
| 3.5 | Create SE with invalid manager role → error | Admin logged in | Select ADMIN as manager for SE | Validation error: manager must be SM | E2E |
| 3.6 | Create with duplicate email → error | Admin logged in | Use existing email | Validation error: email taken | E2E |
| 3.7 | Create ADMIN with manager → error | Admin logged in | Try to set manager for ADMIN role | Field hidden or validation error | E2E |
| 3.8 | Employee ID auto-generation | Admin logged in | Create SM account | Employee ID = SM-000X (incrementing) | E2E |
| 3.9 | Employee ID uniqueness | Admin logged in | Manual entry of existing employee ID | Validation error: ID taken | E2E |

### 3.2 Account Status

| # | Test Scenario | Given | When | Expected Result | Type |
|---|---|---|---|---|---|
| 3.10 | Create account as ACTIVE | Admin | Status = ACTIVE | User can login immediately | E2E |
| 3.11 | Create account as INACTIVE | Admin | Status = INACTIVE | User cannot login | E2E |
| 3.12 | Deactivate active account | Admin, user is ACTIVE | Set status to INACTIVE | User cannot login, session revoked | E2E |
| 3.13 | Activate inactive account | Admin, user is INACTIVE | Set status to ACTIVE | User can login again | E2E |
| 3.14 | Inactive user login attempt | INACTIVE user | Try to login | HTTP 401 ACCOUNT_INACTIVE | API |
| 3.15 | Inactive user refresh attempt | INACTIVE user has existing refresh cookie | POST /auth/refresh | HTTP 401 | API |
| 3.16 | Deactivation does not delete records | SE deactivated | Check prospects, customers, visits | All records remain, SE name still shown | E2E |

### 3.3 Password Reset

> Backend support delivered by Phase 6A (2026-07-31): AUTO/MANUAL, temp password returned once,
> `token_version` bump, active sessions revoked. Scenarios 3.17, 3.19, 3.20 covered by
> service/repository/server tests; 3.18 depends on the login flow and is not yet automated.

| # | Test Scenario | Given | When | Expected Result | Type |
|---|---|---|---|---|---|
| 3.17 | Admin resets password | Admin, target user exists | Click "Reset Password" | Temporary password shown once | E2E |
| 3.18 | Login with new temporary password | User whose password was reset | Login with temporary password | Login successful | E2E |
| 3.19 | Old password rejected after reset | User whose password was reset | Login with old password | HTTP 401 INVALID_CREDENTIALS | API |
| 3.20 | Password reset revokes sessions | User was logged in on Device A | Admin resets password | Device A gets 401 on next request | Integration |

### 3.4 Account Edit

| # | Test Scenario | Given | When | Expected Result | Type |
|---|---|---|---|---|---|
| 3.21 | Update full_name | Admin | Edit full_name | Updated successfully | E2E |
| 3.22 | Update email | Admin | Edit email | Updated, must remain unique | E2E |
| 3.23 | Update phone | Admin | Edit phone | Updated successfully | E2E |
| 3.24 | Change SE manager | Admin, SE currently under Manager A | Change to Manager B | SE now under Manager B | E2E |
| 3.25 | Change SE role to SM | Admin, SE has no prospects | Change role to SALES_MANAGER | Updated, manager_id cleared | E2E |

---

## 4. Single Session Tests

### 4.1 REJECT_NEW_LOGIN Behavior

| # | Test Scenario | Given | When | Expected Result | Type |
|---|---|---|---|---|---|
| 4.1 | Login on Device A succeeds | No active session | POST /auth/login | HTTP 200, session created | API |
| 4.2 | Login same account on Device B rejected | Device A has active session | POST /auth/login (same credentials) | HTTP 409 ACTIVE_SESSION_EXISTS | API |
| 4.3 | Different account on Device B succeeds | Device A has active session (user1) | POST /auth/login (user2 credentials) | HTTP 200, session created for user2 | API |
| 4.4 | Login after logout from Device A | Device A logs out | POST /auth/login from Device B | HTTP 200, session created | API |
| 4.5 | Login after Admin ends session | Admin ends Device A session | POST /auth/login from Device B | HTTP 200, session created | API |
| 4.6 | Login after session expires | Wait for refresh token to expire | POST /auth/login | HTTP 200, session created | Integration |
| 4.7 | Login after session is revoked (manual DB) | Revoke all sessions for user | POST /auth/login | HTTP 200, session created | Integration |

### 4.2 Concurrent Login Race Condition

| # | Test Scenario | When | Expected Result | Type |
|---|---|---|---|---|
| 4.8 | Two simultaneous login requests | Two parallel POST /auth/login | First succeeds, second gets 409. Exactly one active session. | Integration |

### 4.3 Session Lifecycle

| # | Test Scenario | Given | Expected Result |
|---|---|---|---|
| 4.9 | Active session after login | 1 login | 1 row in refresh_sessions with revoked_at = NULL |
| 4.10 | Active session after multiple refreshes | 1 login + multiple refreshes | 1 active row (others are ROTATED) |
| 4.11 | No active session after logout | 1 login + logout | 0 active rows |
| 4.12 | No active session after admin ends | 1 login + admin ends session | 0 active rows |
| 4.13 | Expired session is not active | 1 login + wait for expiry | 0 active rows |

### 4.4 Rotation Does Not Block New Login (Negative)

| # | Test Scenario | Given | Expected Result |
|---|---|---|---|
| 4.14 | After refresh, login from another device still blocked | Login → Refresh → New device login attempt | 409 ACTIVE_SESSION_EXISTS (refresh is not a new login) |

---

## 5. Manager Tests

### 5.1 Manager Relationship

| # | Test Scenario | Given | When | Expected Result | Type |
|---|---|---|---|---|---|
| 5.1 | SM sees correct team members | Manager A has SE1, SE2 | GET /manager/team-members | Returns SE1, SE2 | API |
| 5.2 | SM with no SEs sees empty team | New SM with no SEs | GET /manager/team-members | Returns [] | API |
| 5.3 | SE reassigned to new SM | SE moved from SM-A to SM-B | SM-B GET /manager/team-members | SE now appears in SM-B's list | E2E |
| 5.4 | Former SM loses access to SE data | SE moved from SM-A to SM-B | SM-A GET /manager/team-members | SE no longer in SM-A's list | API |

### 5.2 Read-Only for Manager

| # | Test Scenario | Given | When | Expected Result | Type |
|---|---|---|---|---|---|
| 5.5 | SM cannot transition SE's prospect | SM logged in | PATCH /sales/prospects/:id/transition | HTTP 403 | API |
| 5.6 | SM cannot check in for SE | SM logged in | POST /sales/prospects/:id/visits/check-in | HTTP 403 | API |
| 5.7 | SM cannot convert prospect | SM logged in | POST /admin/prospects/:id/convert | HTTP 403 | API |
| 5.8 | SM cannot manage accounts | SM logged in | POST /api/v1/admin/users | HTTP 403 | API |
| 5.9 | SM prospect detail has no action buttons | SM viewing team prospect | UI renders | No Edit, Transition, Convert buttons | E2E |
| 5.10 | SM can add comments | SM viewing team prospect | POST comment | Comment added (read-write for comments) | API |

---

## 6. Frontend UI Tests

### 6.1 Navigation Visibility

| # | Test Scenario | Logged In As | Expected Menu Items |
|---|---|---|---|
| 6.1 | Admin navigation | ADMINISTRATOR | Dashboard, Accounts, Sales Executives, Customers, Visit Monitoring, Prospect Finder, Prospect List, Reports, Activity Log |
| 6.2 | SM navigation | SALES_MANAGER | Dashboard, Team Members, Team Prospects, Team Customers, Team Pipeline, Team Visits, Reports, Profile |
| 6.3 | SE navigation | SALES_EXECUTIVE | Home, Customer, Prospect, Pipeline, History, Profile |

### 6.2 Redirect Tests

| # | Test Scenario | Given | Action | Expected Redirect |
|---|---|---|---|---|
| 6.4 | Unauthenticated → protected page | Not logged in | Navigate to `/admin/dashboard` | `/login?redirect=/admin/dashboard` |
| 6.5 | Unauthenticated → public page | Not logged in | Navigate to `/login` | Stay on `/login` |
| 6.6 | Authenticated → login page | Logged in as ADMIN | Navigate to `/login` | Redirect to `/admin/dashboard` |
| 6.7 | Authenticated → wrong role page | Logged in as SE | Navigate to `/admin/dashboard` | Redirect to `/forbidden` |

---

## 7. Admin Session Management Tests

| # | Test Scenario | Given | When | Expected Result | Type |
|---|---|---|---|---|---|
| 7.1 | Admin views active session info | User has active session | Admin opens account detail page | Shows device, IP, login time | E2E |
| 7.2 | Admin ends active session | User has active session | Admin clicks "End Session" | Session revoked, user blocked | E2E |
| 7.3 | End session confirmation | — | Admin clicks "End Session" | Confirmation dialog appears | E2E |
| 7.4 | End session for user with no active session | User has no active session | Admin clicks "End Session" | No-op, show "No active session" | E2E |
| 7.5 | User re-logs in after admin ended session | Session was ended by admin | User logs in again | Success (no active session to block) | API |

---

## 8. Negative Security Tests

| # | Test Scenario | Given | When | Expected Result | Type |
|---|---|---|---|---|---|
| 8.1 | Manipulate frontend role in localStorage | SE manually edits stored user object | Navigate to protected admin page | Backend still enforces 403 | E2E |
| 8.2 | Send API request with manipulated JWT role | Tamper JWT payload in transit | POST to admin endpoint | JWT signature verification fails → 401 | Security |
| 8.3 | Replay old refresh token | Use old crm_refresh cookie after rotation | POST /auth/refresh | Theft detected, all sessions revoked | Security |
| 8.4 | Brute force login (future) | Multiple rapid login attempts | POST /auth/login repeatedly | Rate limited (future phase) | Security |
| 8.5 | SQL injection in login | Email = `' OR 1=1--` | POST /auth/login | Parameterized query prevents injection | Security |
| 8.6 | XSS via profile fields | Email contains `<script>` | Account created with XSS payload | Stored XSS prevented (backend encodes responses) | Security |

---

## 9. Test Data Setup

For manual testing, the following accounts are required:

| # | Email | Password | Role | Manager |
|---|---|---|---|---|
| 1 | `admin@yummy.test` | `password123` | ADMINISTRATOR | — |
| 2 | `manager1@yummy.test` | `password123` | SALES_MANAGER | — |
| 3 | `manager2@yummy.test` | `password123` | SALES_MANAGER | — |
| 4 | `se1@yummy.test` | `password123` | SALES_EXECUTIVE | manager1 |
| 5 | `se2@yummy.test` | `password123` | SALES_EXECUTIVE | manager1 |
| 6 | `se3@yummy.test` | `password123` | SALES_EXECUTIVE | manager2 |

Each SE should have at least 2–3 prospects and 1–2 customers assigned.
