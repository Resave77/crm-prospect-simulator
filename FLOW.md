# Business Flow and State Machines

Status: `APPROVED — BUSINESS FLOW; IMPLEMENTATION PENDING`

## 1. Global Rules

- Administrator operates the desktop interface.
- Sales Executive operates the mobile interface.
- A prospect is sourced from Google Places and uniquely identified by Google Place ID.
- Saving a search result creates a CRM snapshot; later Google changes must not rewrite historical CRM data automatically.
- A prospect must be assigned before it appears in `My Prospect`.
- A prospect must be `WON` before customer conversion.
- A lost prospect is terminal.
- Conversion creates a customer without deleting the source prospect.
- Prospect visits and customer attendance are different business records.

## 2. End-to-End Flow

| Step | Actor | Action | Result |
|---:|---|---|---|
| 1 | Administrator | Select search center | Latitude and longitude are established |
| 2 | Administrator | Choose radius, categories, keyword | Search criteria are validated |
| 3 | System | Query Google Places | Search results and map markers are displayed |
| 4 | Administrator | Select business | Place Detail is loaded |
| 5 | Administrator | Save as prospect | Duplicate Google Place ID is rejected |
| 6 | Administrator | Assign Sales Executive | Active assignment is created |
| 7 | System | Publish assignment | Prospect appears in assigned user's `My Prospect` |
| 8 | Sales Executive | Check in to prospect | GPS, selfie, time, and device context are recorded |
| 9 | Sales Executive | Check out | Visit is completed; pipeline can become `VISITED` |
| 10 | Sales Executive | Update pipeline | `VISITED` advances to `FOLLOW_UP` |
| 11 | Sales Executive | Record decision | Prospect becomes `LOST` or `WON` |
| 12 | Administrator | Review won prospect | Conversion eligibility is confirmed |
| 13 | Administrator | Complete customer form | Customer code, radius, and required data are validated |
| 14 | System | Convert atomically | Customer is created; prospect becomes `CONVERTED` |
| 15 | System | Publish customer | Customer appears in owner's `My Customer` |
| 16 | Sales Executive | Customer attendance | GPS/selfie check-in and check-out are recorded |

## 3. Prospect State Machine

Canonical statuses:

```text
SAVED -> ASSIGNED -> VISITED -> FOLLOW_UP -> LOST
                                      |
                                      +-> WON -> CONVERTED
```

| Current | Allowed next | Initiator | Preconditions |
|---|---|---|---|
| none | `SAVED` | Administrator | Valid Google Place snapshot; Google Place ID not already used |
| `SAVED` | `ASSIGNED` | Administrator | Active Sales Executive selected |
| `ASSIGNED` | `VISITED` | System/service | A valid prospect visit has completed |
| `VISITED` | `FOLLOW_UP` | Assigned Sales Executive | Visit exists and ownership is active |
| `FOLLOW_UP` | `LOST` | Assigned Sales Executive | Loss reason supplied |
| `FOLLOW_UP` | `WON` | Assigned Sales Executive | Win notes supplied |
| `WON` | `CONVERTED` | Administrator | Conversion form valid; no customer already linked |

Rules:

- Status cannot be skipped through the standard API.
- Replaying the same successful transition returns the existing result where an idempotency key is supplied.
- Every transition creates a pipeline-history entry in the same transaction.
- `LOST` and `CONVERTED` are terminal.
- Reassignment changes ownership but does not move pipeline status backward.
- Status names are stable API/database codes and are not translated.

## 4. Assignment Lifecycle

Only one assignment may be active for a prospect.

```text
ACTIVE -> REASSIGNED
ACTIVE -> COMPLETED
```

- Initial assignment moves `SAVED` to `ASSIGNED`.
- Reassignment closes the prior assignment and creates a new active assignment atomically.
- Conversion closes the active prospect assignment as `COMPLETED`.
- The customer inherits the active Sales Executive by default; Administrator may explicitly select another active Sales Executive during conversion.
- Historical assignments remain immutable for audit.

## 5. Prospect Visit State Machine

```text
CHECKED_IN -> CHECKED_OUT
```

Check-in requires:

- Authenticated assigned Sales Executive.
- Prospect in `ASSIGNED`, `VISITED`, or `FOLLOW_UP`.
- No open visit for the same user.
- Latitude and longitude.
- GPS accuracy value.
- Captured selfie media reference.
- Client-generated idempotency key.

Check-out requires:

- The same Sales Executive who checked in.
- An open visit.
- Checkout latitude, longitude, accuracy, and timestamp.
- Checkout time later than check-in time.

The server records its own receipt timestamps. Client timestamps are retained only as evidence and never treated as the sole authority.

## 6. Decision Flow

From `FOLLOW_UP`, exactly one decision is selected:

### Lost

- A controlled loss reason is required.
- Optional notes may be supplied.
- Prospect is removed from the active mobile queue.
- No conversion is permitted.

### Won

- Win notes are required.
- Prospect enters Administrator's Won Review queue.
- It remains a prospect until conversion completes.
- It does not appear in `My Customer` yet.

## 7. Conversion Transaction

Conversion is one atomic service operation:

1. Lock and reload the prospect.
2. Confirm status is `WON`.
3. Confirm it has no linked customer.
4. Validate unique customer code.
5. Validate attendance radius and required customer data.
6. Create customer from approved form plus immutable prospect/place snapshot.
7. Assign customer owner.
8. Close active prospect assignment.
9. Move prospect to `CONVERTED`.
10. Append pipeline and audit history.
11. Commit all changes together.

Any failure rolls back the entire conversion.

## 8. Existing Customer Attendance

Customer attendance is independent of prospect visits.

```text
CHECKED_IN -> CHECKED_OUT
```

Check-in validation additionally calculates the distance between submitted GPS and the customer's registered coordinates. The record stores:

- Submitted coordinates and accuracy.
- Customer coordinates at that time.
- Configured attendance radius at that time.
- Calculated distance.
- Whether the visit is inside or outside the radius.

Architecture default: outside-radius check-in is rejected unless a future approved business rule introduces an administrator-reviewed exception flow.

## 9. Work-Queue Inclusion

| Queue | Inclusion rule |
|---|---|
| Admin Prospect List | All prospects subject to filters |
| Admin Won Review | Status `WON` and no customer link |
| Sales My Prospect | Active assignment owned by current user; status `ASSIGNED`, `VISITED`, or `FOLLOW_UP` |
| Sales My Customer | Active customer owned by current user |
| Active visits | Current user's records with no checkout |

## 10. Failure and Concurrency Rules

- Duplicate prospect saves return conflict and identify the existing prospect ID when authorized.
- Concurrent assignment updates use transactional locking/version checks.
- A stale pipeline request returns conflict with the current status.
- A second conversion returns the already-created customer for an idempotent replay, otherwise conflict.
- Google Places unavailability blocks discovery/detail refresh but does not make stored CRM records unavailable.
- Media upload must complete before a check-in transaction references the media asset.
