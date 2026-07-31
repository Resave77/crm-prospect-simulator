# Authentication & Session Flow

> **Document:** 04-authentication-session-flow.md  
> **Project:** CRM Prospect Simulator  
> **Date:** 2026-07-30  
> **Status:** Draft — Design Phase  

---

## 1. Existing Authentication Summary

The current auth system is documented in detail in `AUDIT_REPORT.md` (sections 5–6). This section summarises what exists and what changes are required.

### 1.1 What Already Works

| Feature | Status | Notes |
|---|---|---|
| JWT access token (30min TTL) | ✅ | Standard JWT with `sub`, `role`, `session`, `tokenVersion` |
| Refresh token (30-day TTL) | ✅ | HTTP-only cookie, path `/api/v1/auth` |
| Refresh token rotation | ✅ | Old session revoked, new session created on each refresh |
| Theft detection | ✅ | Reuse of rotated session revokes ALL user sessions |
| Token version (`token_version`) | ✅ | Incremented to invalidate all sessions globally |
| bcrypt password hashing | ✅ | |
| Login with email/password | ✅ | |
| Logout (single session) | ✅ | Revokes current refresh session |
| Logout all | ✅ | Revokes all refresh sessions |
| Role-based route guard (backend) | ✅ | `RequireRole(...)` middleware |
| Role-based route guard (frontend) | ✅ | `meta.role` in router |

### 1.2 What Needs to Change

| Feature | Status | Phase |
|---|---|---|
| Single active session | ❌ Not implemented | Phase 6 |
| Active session rejection | ❌ Not implemented | Phase 6 |
| Admin session management (list/end) | ❌ Not implemented | Phase 6 |
| `SALES_MANAGER` support in guards | ❌ Not implemented | Phase 3A/B |
| `roles` array (frontend router meta) | ❌ Single role only | Phase 3B |
| Rate limiting on auth endpoints | ❌ Not implemented | Future |
| Password reset | ❌ Not implemented | Phase 4A |

---

## 2. Single Active Session Decision

### 2.1 Decision: REJECT_NEW_LOGIN

**Chosen behavior: Reject new login when an active session exists.**

### 2.2 Rationale

| Factor | REJECT_NEW_LOGIN | Auto-logout-old |
|---|---|---|
| User predictability | ✅ Expected — user knows they must log out first | ❌ Surprising — old device loses access silently |
| Audit clarity | ✅ Single session always active | ❌ Hard to explain which device was logged out |
| Implementation simplicity | ✅ Simple active session check | ✅ Simple, but UX is worse |
| Security | ✅ Good — prevents concurrent use | ✅ Good — forces single device |
| Help desk burden | Lower — user controls logout | Higher — user confused why old device stopped |

### 2.3 User Experience

```
User on Device A: logged in, working.
User on Device B: tries to log in with same account.

Device B receives:
  HTTP 409 Conflict
  {
    "error": {
      "code": "ACTIVE_SESSION_EXISTS",
      "message": "Account already signed in on another device."
    }
  }

Frontend displays:
  "Account already signed in on another device.
   Sign out from the other device first, or contact your administrator
   to end the active session."

User must:
  1. Log out from Device A, OR
  2. Ask Admin to end the session from Account Management
```

---

## 3. Active Session Definition

### 3.1 Current Session Model

A `refresh_sessions` row is considered **active** when:

```sql
SELECT *
FROM refresh_sessions
WHERE user_id = $1
  AND revoked_at IS NULL
  AND expires_at > CURRENT_TIMESTAMP
ORDER BY created_at DESC;
```

### 3.2 Single Active Session Query

The query for single session enforcement must account for rotation:

```sql
-- Find the most recently created active session
SELECT *
FROM refresh_sessions
WHERE user_id = $1
  AND revoked_at IS NULL
  AND expires_at > CURRENT_TIMESTAMP
ORDER BY created_at DESC
LIMIT 1;
```

If this query returns a row, and the row's `id` does not match the current login attempt's session family, block the login.

### 3.3 Important: Rotation ≠ New Login

Refresh token rotation creates a **new session row** each time a token is refreshed. This means:

| Event | Old Row | New Row |
|---|---|---|
| Login | — | Created (active) |
| Refresh #1 | Revoked (ROTATED) | Created (active) |
| Refresh #2 | Revoked (ROTATED) | Created (active) |

Rotation is part of normal single-device usage and must **not** be treated as a concurrent login.

### 3.4 Session Family Concept

To distinguish between:
- **New login** (should be rejected if active session exists)
- **Refresh rotation** (should be allowed)

The system should track a **session family** — a chain of sessions linked by `replaced_by_session_id`.

**Implementation approach:** During login, check if there is **any** active session. If the current login is not a rotation continuation of that session, block it.

**Simpler alternative for initial implementation:**

Add a `login_id` or `device_id` field to `refresh_sessions`. All sessions in the same login chain share the same `login_id`. New login → new `login_id`. Refresh → same `login_id`.

Or, even simpler:

- On login (not refresh), check if any active session exists.
- If yes, reject.
- On refresh, do not check (rotation is allowed).
- This works because `createLogin()` is called both on initial login and on refresh. The handler must distinguish which path called it.

**Proposed implementation:**

```go
func (s *AuthService) Login(ctx, email, password, client) (AuthResult, error) {
  // ... existing validation ...

  // NEW: Check for existing active session
  hasActive, err := s.sessions.HasActiveSession(ctx, user.ID)
  if err != nil { return ..., err }
  if hasActive {
    return ..., ErrActiveSessionExists
  }

  return s.createLogin(ctx, user, client, uuid.Nil, true) // isLogin = true
}

func (s *AuthService) createLogin(ctx, user, client, replaceSessionID, isLogin) (AuthResult, error) {
  // ... existing logic ...
  // Pass isLogin flag to differentiate new login from refresh
}
```

---

## 4. Login Flow (Updated for Single Session)

```
Client (Device B)                          Server
      │                                       │
      │  POST /api/v1/auth/login              │
      │  {email, password}                    │
      │──────────────────────────────────────>│
      │                                       │
      │                                       │── FindByEmail()
      │                                       │── bcrypt compare password
      │                                       │── Check status == ACTIVE
      │                                       │
      │                                       │── HasActiveSession(userId)?
      │                                       │    │
      │                                       │    ├── Yes → 409 ACTIVE_SESSION_EXISTS
      │                                       │    │
      │                                       │    └── No  → Continue
      │                                       │
      │                                       │── Create RefreshSession
      │                                       │── Issue JWT
      │                                       │
      │  <── 200 {accessToken, user}           │
      │  Set-Cookie: crm_refresh (new)         │
      │                                       │
```

### 4.1 Active Session Check on Login (Detail)

```go
// Repository
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

**Risk:** After rotation, the old session is revoked (ROTATED), so only the latest session is active. This means:
- 1 login → 1 active session → subsequent logins blocked ✅
- Refresh → old revoked → 1 active session → subsequent logins still blocked ✅
- After logout → 0 active sessions → login allowed ✅
- After expiry → 0 active sessions → login allowed ✅
- After admin end-session → 0 active sessions → login allowed ✅

---

## 5. Refresh Flow

No changes to the existing refresh flow for single session. Refresh is authentication maintenance, not a new login.

```
Client                                    Server
  │                                           │
  │  POST /api/v1/auth/refresh                │
  │  Cookie: crm_refresh                      │
  │──────────────────────────────────────────>│
  │                                           │
  │                                           │── Parse sessionID + secret
  │                                           │── FindSessionByID()
  │                                           │── Verify token hash
  │                                           │── Check not revoked, not expired
  │                                           │── Check user is ACTIVE
  │                                           │
  │                                           │── Rotate: revoke old, insert new
  │                                           │── Issue new JWT
  │                                           │
  │  <── 200 {accessToken, user}               │
  │  Set-Cookie: crm_refresh (new)             │
  │                                           │
```

---

## 6. Logout Flow

No changes. Existing flow is correct.

```
Client                                    Server
  │                                           │
  │  POST /api/v1/auth/logout                  │
  │  Cookie: crm_refresh                      │
  │──────────────────────────────────────────>│
  │                                           │
  │                                           │── Parse sessionID
  │                                           │── Revoke session (reason: LOGOUT)
  │                                           │
  │  <── 204 No Content                        │
  │  Set-Cookie: crm_refresh (expired)         │
  │                                           │
```

---

## 7. Admin End Session Flow

### 7.1 Endpoint Design

```
Admin clicks "End Session" on /admin/accounts/:id
         │
         ▼
    Confirmation dialog: "End active session for [full_name]?"
         │
         ▼
    POST /api/v1/admin/users/:id/end-session
         │
         ▼
    Backend:
      1. Find active session for user
      2. Revoke session (reason: ADMIN_ENDED)
      3. Increment token_version (optional, immediate effect)
         │
         ▼
    Response: 200 { message: "Session ended." }
         │
         ▼
    Device A (previous session holder):
      Next API call → 401 → refreshSession() fails → redirect to login
```

### 7.2 Implementation Options

| Option | Effect | Complexity |
|---|---|---|
| Revoke only active refresh session | Device A stays logged in until token expires (30min max) | Low |
| Revoke + increment tokenVersion | Device A immediately rejected on next API call | Medium |

**Recommendation:** Revoke + increment `token_version` for immediate effect.

```go
func (s *AuthService) AdminEndSession(ctx context.Context, targetUserID uuid.UUID) error {
    now := s.now().UTC()
    // Revoke all active sessions
    if err := s.sessions.RevokeAllForUser(ctx, targetUserID, "ADMIN_ENDED", now); err != nil {
        return err
    }
    // Increment token version to invalidate all issued JWTs
    if err := s.users.IncrementTokenVersion(ctx, targetUserID); err != nil {
        return err
    }
    return nil
}
```

---

## 8. Password Reset Impact on Sessions

When Admin resets a user's password:

1. New password hash is stored
2. `must_change_password` is set to `true`
3. `token_version` is incremented
4. All active sessions are revoked

This ensures:
- Old access tokens are invalidated (token version mismatch)
- Old refresh sessions are revoked
- User must log in with new password
- User will be prompted to change password on next login

---

## 9. Session Display (Admin View)

### 9.1 Active Session Info

On admin account detail page:

```
Active Session

  Status:       Active
  Device:       Chrome 126 / Windows 10    (from user-agent)
  IP Address:   192.168.1.100
  Login Time:   2026-07-30 14:23:00 WIB
  Expires At:   2026-08-29 14:23:00 WIB

  [End Session Button]
```

### 9.2 What NOT to Display

- Raw refresh token
- Token hash
- JWT contents
- Full token secrets
- Password hash

---

## 10. Session Cleanup

Not a priority for initial implementation. Expired and revoked sessions can accumulate. A periodic cleanup job can be added in a future phase:

```sql
DELETE FROM refresh_sessions
WHERE expires_at < NOW() - INTERVAL '90 days'
   OR (revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '90 days');
```
