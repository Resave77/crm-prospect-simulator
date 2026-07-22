# REST API Architecture

Status: `APPROVED — AUTHENTICATION FOUNDATION IMPLEMENTED; CRM CONTRACTS PLANNED`

## 1. Boundary

All backend endpoints live under `/api/*` and return JSON. Versioned business APIs use `/api/v1/*`.

`GET /api/health` is the unversioned operational health endpoint. `/api/v1/health` is retained as a versioned alias; neither endpoint renders HTML.

Fiber never returns application HTML. Unknown API routes return a JSON `404`; they never fall through to Vue Router.

## 2. Conventions

- JSON fields use `camelCase`.
- Resource IDs are UUID strings.
- Timestamps use RFC 3339 UTC.
- List endpoints use cursor pagination where records can grow significantly.
- Mutating requests validate `Content-Type: application/json`, except media upload endpoints.
- `Idempotency-Key` is required for conversion, check-in, check-out, and other retry-sensitive actions.
- `X-Request-ID` is accepted or generated and returned.
- Authentication uses `Authorization: Bearer <access-token>`.

## 3. Response Envelopes

Single resource:

```json
{
  "data": {},
  "meta": { "requestId": "uuid" }
}
```

Collection:

```json
{
  "data": [],
  "meta": {
    "requestId": "uuid",
    "nextCursor": null
  }
}
```

Error:

```json
{
  "error": {
    "code": "PROSPECT_INVALID_TRANSITION",
    "message": "Prospect cannot transition to the requested status.",
    "fields": {},
    "requestId": "uuid"
  }
}
```

Error messages are safe for clients; internal errors and provider payloads remain in protected logs.

## 4. HTTP Status Policy

| Status | Use |
|---:|---|
| 200 | Successful read/action replay |
| 201 | Resource created |
| 204 | Successful action with no body |
| 400 | Malformed request |
| 401 | Missing/invalid authentication |
| 403 | Authenticated but unauthorized |
| 404 | Resource unavailable to caller |
| 409 | Duplicate, stale state, invalid transition, or idempotency conflict |
| 422 | Field validation failure |
| 429 | Rate limit exceeded |
| 502 | Required upstream provider failed |
| 503 | Service dependency unavailable |

## 5. Authentication API

| Method | Path | Access | Purpose |
|---|---|---|---|
| POST | `/api/v1/auth/login` | Public | Authenticate user and establish token session |
| POST | `/api/v1/auth/refresh` | Refresh session | Rotate refresh session and issue access token |
| POST | `/api/v1/auth/logout` | Authenticated | Revoke current refresh session |
| POST | `/api/v1/auth/logout-all` | Authenticated | Revoke all user's sessions |
| GET | `/api/v1/auth/me` | Authenticated | Return identity, role, and session-safe profile |

No public registration endpoint exists.

## 6. Google Places API

| Method | Path | Role | Purpose |
|---|---|---|---|
| GET | `/api/v1/places/nearby` | Administrator | Nearby Search by coordinates/radius/categories |
| POST | `/api/v1/places/search` | Administrator | Text Search with location context |
| GET | `/api/v1/places/:placeId` | Administrator | Curated Place Detail |

The API returns curated CRM DTOs, not raw Google payloads. Google API keys, quota headers, and debug traces are never exposed to normal clients.

Nearby parameters include `latitude`, `longitude`, `radiusMeters`, and repeated/category-list filters. Text Search includes `query` plus the same location context.

## 7. Prospect API

| Method | Path | Role | Purpose |
|---|---|---|---|
| POST | `/api/v1/prospects` | Administrator | Save selected Google business as prospect |
| GET | `/api/v1/prospects` | Administrator | Filtered/paginated prospect list |
| GET | `/api/v1/prospects/:id` | Administrator or assigned Sales | Prospect detail projection |
| GET | `/api/v1/me/prospects` | Sales Executive | Current user's active assigned prospects |

Saving a prospect submits the selected Google Place ID and approved snapshot. The service may revalidate essential place data before saving. Duplicate Place ID returns `409 PROSPECT_ALREADY_EXISTS`.

## 8. Assignment API

| Method | Path | Role | Purpose |
|---|---|---|---|
| POST | `/api/v1/prospects/:id/assignments` | Administrator | Initial assignment or reassignment |
| GET | `/api/v1/prospects/:id/assignments` | Administrator | Assignment history |

Assignment input identifies one active Sales Executive plus optional effective dates. Reassignment never overwrites history.

## 9. Prospect Visit API

| Method | Path | Role | Purpose |
|---|---|---|---|
| POST | `/api/v1/media/selfies` | Sales Executive | Upload and validate selfie evidence |
| POST | `/api/v1/prospects/:id/visits/check-in` | Assigned Sales | Start prospect visit |
| POST | `/api/v1/prospects/:id/visits/:visitId/check-out` | Visit owner | Complete prospect visit |
| GET | `/api/v1/prospects/:id/visits` | Administrator or assigned Sales | Visit history projection |

Check-in references a `READY` media asset and includes coordinates, GPS accuracy, and client capture time.

## 10. Pipeline API

| Method | Path | Role | Purpose |
|---|---|---|---|
| POST | `/api/v1/prospects/:id/pipeline/transitions` | Assigned Sales | Request permitted transition |
| GET | `/api/v1/prospects/:id/pipeline/history` | Administrator or assigned Sales | Read transition history |

The request specifies `targetStatus`, `expectedCurrentStatus`, reason code where applicable, and notes. The service—not the client—decides whether the transition is legal.

## 11. Conversion and Customer API

| Method | Path | Role | Purpose |
|---|---|---|---|
| GET | `/api/v1/prospects/won` | Administrator | Won Review queue |
| GET | `/api/v1/prospects/:id/conversion-preview` | Administrator | Pre-fill conversion form |
| POST | `/api/v1/prospects/:id/conversion` | Administrator | Atomically create customer |
| GET | `/api/v1/customers` | Administrator | Existing Customer list |
| GET | `/api/v1/customers/:id` | Administrator or assigned Sales | Customer detail projection |
| GET | `/api/v1/me/customers` | Sales Executive | Current user's active customers |

Conversion requires `WON`, unique customer code, location, attendance radius, customer data, and owner.

## 12. Existing Customer Attendance API

| Method | Path | Role | Purpose |
|---|---|---|---|
| POST | `/api/v1/customers/:id/attendances/check-in` | Assigned Sales | Start attendance |
| POST | `/api/v1/customers/:id/attendances/:attendanceId/check-out` | Attendance owner | Complete attendance |
| GET | `/api/v1/customers/:id/attendances` | Administrator or assigned Sales | Attendance history |

The server calculates distance and determines radius compliance. Client-calculated distance is never trusted.

## 13. Administration API

| Method | Path | Role | Purpose |
|---|---|---|---|
| GET | `/api/v1/users` | Administrator | User list |
| GET | `/api/v1/users/:id` | Administrator | User detail |
| POST/PATCH | `/api/v1/users...` | Administrator | Future approved user management |
| GET/PATCH | `/api/v1/settings...` | Administrator | Future approved settings management |
| GET | `/api/v1/dashboard/admin` | Administrator | Admin dashboard projection |
| GET | `/api/v1/dashboard/sales` | Sales Executive | Sales dashboard projection |

CRUD details remain intentionally unspecified until the relevant implementation phase is approved.

## 14. Authorization Matrix

| Capability | Administrator | Sales Executive |
|---|:---:|:---:|
| Search Google Places | Yes | No |
| Save/assign prospect | Yes | No |
| View all prospects | Yes | No |
| View assigned prospect | Yes | Yes |
| Prospect visit | No | Own assignment only |
| Pipeline decision | No | Own assignment only |
| Won review/conversion | Yes | No |
| View all customers | Yes | No |
| View assigned customers | Yes | Yes |
| Customer attendance | No | Own customer only |
| Manage users/settings | Yes | No |

## 15. Versioning and Deprecation

Only endpoints defined for this new project are supported. Future breaking changes require a new API version or an explicitly approved deprecation window; routes from unrelated applications are not compatibility contracts.

## 16. API Documentation and Testing Policy

- OpenAPI becomes the machine-readable companion to this document during implementation.
- Contract tests cover status, envelope, authorization, and validation behavior.
- Repository/provider fakes support deterministic service tests.
- Google sandbox/live calls are separated from the default test suite.
- Every endpoint requires positive, unauthorized, forbidden, validation, conflict, and dependency-failure coverage as applicable.
