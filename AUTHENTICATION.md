# Authentication and Authorization Architecture

Status: `APPROVED — FOUNDATION IMPLEMENTED`

## 1. Scope

The system has exactly two roles:

- `ADMINISTRATOR`
- `SALES_EXECUTIVE`

There is no public registration, social login, guest access, or third role.

## 2. Token Model

- Access token: signed JWT, short lived, sent in the `Authorization` header.
- Refresh token: high-entropy rotating session token stored in a Secure, HttpOnly cookie.
- The database stores a hash of the refresh-token secret, never the raw token.
- The Vue application holds the access token in memory, not local storage.
- Page reload restores a session through the refresh endpoint.

Recommended initial lifetimes, subject to security approval:

- Access JWT: 15 minutes.
- Refresh session: 30 days absolute maximum.
- Inactive/deactivated users cannot refresh or log in.

## 3. Access JWT Claims

Required claims:

| Claim | Purpose |
|---|---|
| `iss` | Trusted issuer |
| `aud` | CRM API audience |
| `sub` | User UUID |
| `role` | One canonical role |
| `sid` | Refresh session UUID |
| `ver` | User token version |
| `jti` | Unique access token ID |
| `iat`, `nbf`, `exp` | Time validation |

The API validates signature, algorithm allowlist, issuer, audience, time claims, session/user state as required, and canonical role. It never trusts a role supplied separately by the client.

## 4. Login Flow

1. Client posts email and password to `/api/v1/auth/login`.
2. Server normalizes email. Login rate limiting is a required production-hardening control and is not part of the current P0 implementation.
3. Server verifies the bcrypt password hash using the approved bcrypt comparison function.
4. Server confirms user is active.
5. Server creates a refresh session with hashed token metadata.
6. Server sets the refresh cookie and returns an access JWT plus safe user profile.
7. Frontend routes Administrator to `/admin/dashboard` or Sales Executive to `/sales/dashboard`.

Authentication failures return one generic message so email existence is not disclosed.

## 5. Refresh Rotation

1. Browser sends the HttpOnly refresh cookie to the refresh endpoint.
2. Server validates session, expiry, user state, token hash, and rotation chain. Explicit `Origin` validation is required before production deployment.
3. Current session is revoked as rotated.
4. A replacement session/token is created atomically.
5. New refresh cookie and access JWT are returned.

Reuse of a revoked rotated token is treated as possible token theft and revokes the affected session family.

## 6. Logout and Revocation

- Logout revokes the current refresh session and clears its cookie.
- The current logout-all operation revokes all refresh sessions for the user. Incrementing token version to invalidate outstanding access tokens is required before production deployment.
- Password reset/change, user deactivation, or suspected compromise performs logout-all.
- Short-lived access tokens expire naturally; critical endpoints may also verify token version/user status against a cache/database policy.

## 7. Cookie and Browser Policy

Production refresh cookie properties:

- `HttpOnly`
- `Secure`
- `SameSite=Strict` where operationally compatible
- Path restricted to authentication endpoints
- No JavaScript access

Production uses same-origin frontend/API hosting, eliminating routine CORS requirements. State-changing cookie-authenticated endpoints must validate `Origin` before production release; bearer-authenticated business endpoints still enforce normal authorization.

## 8. Authorization Layers

Authorization is enforced in two stages:

1. Middleware validates identity and broad role eligibility.
2. Service verifies resource ownership and workflow eligibility.

Examples:

- Sales Executive cannot search Places or save a prospect.
- Sales Executive may view/visit only an actively assigned prospect.
- Sales Executive may attend only a customer currently assigned to that user.
- Administrator cannot use Sales visit/attendance actions.
- Only Administrator can convert a `WON` prospect.

Hiding navigation in Vue is usability, not security.

## 9. Frontend Route Guards

- Unauthenticated access redirects to `/login` while retaining a safe intended route.
- Administrator routes reject Sales Executive sessions.
- Sales routes reject Administrator sessions.
- API `401` triggers one controlled refresh attempt.
- A failed refresh clears in-memory identity and returns to login.
- `403` displays an access-denied state and never loops refresh.

Axios refresh handling must serialize concurrent refresh attempts so multiple failed requests do not rotate the same token simultaneously.

## 10. Password and User Provisioning

- Initial Administrator is provisioned through a controlled deployment/runbook process.
- Administrators create or invite Sales Executive accounts in the later user-management phase.
- Password policy and recovery delivery channel require business/security approval before implementation.
- Password hashes and reset secrets never appear in logs, API responses, or audit diffs.

## 11. Abuse Protection

- Rate-limit login by normalized identifier and network source.
- Add escalating temporary delays/lock policy without creating permanent denial-of-service risk.
- Rate-limit refresh and password-recovery endpoints.
- Record suspicious refresh reuse and repeated failures.
- Do not use CAPTCHA unless monitoring shows it is required and an ADR approves it.

## 12. Key Management

- JWT signing keys are deployment secrets.
- Asymmetric signing is preferred so verification can be separated from signing.
- Tokens include a key ID for rotation.
- Rotation supports an overlap window for outstanding access tokens.
- Development, preview, and production use different keys and issuers.

## 13. Authentication Audit Events

The production security backlog must persist successful login, failed login category, refresh, logout, logout-all, refresh reuse, password change/reset, role/status change, and user deactivation events. Tokens, passwords, and raw cookies must never be recorded.
