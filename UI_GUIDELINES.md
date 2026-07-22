# UI and Interaction Guidelines

Status: `APPROVED — AUTH SHELLS IMPLEMENTED; CRM SCREENS PENDING`

## 1. Experience Principle

Administrator and Sales Executive experiences are functionally and visually separated while sharing a coherent design system.

- Administrator: information-dense desktop workspace optimized for search, review, assignment, and management.
- Sales Executive: touch-first mobile workflow optimized for current tasks, location evidence, camera use, and quick status updates.

The application is not designed as one desktop screen compressed onto mobile.

## 2. Route and Layout Separation

### Authentication

- `/login`
- Neutral authentication layout.
- Role-based redirect after successful login.

### Administrator desktop

- `/admin/dashboard`
- `/admin/prospects/finder`
- `/admin/prospects`
- `/admin/prospects/:id`
- `/admin/assignments`
- `/admin/prospects/won`
- `/admin/customers`
- `/admin/customers/:id`
- `/admin/users`
- `/admin/settings`

### Sales Executive mobile

- `/sales/dashboard`
- `/sales/prospects`
- `/sales/prospects/:id`
- `/sales/prospects/:id/visit`
- `/sales/prospects/:id/pipeline`
- `/sales/customers`
- `/sales/customers/:id`
- `/sales/customers/:id/attendance`
- `/sales/profile`

## 3. PrimeVue Policy

- PrimeVue supplies accessible base controls, overlays, data tables, menus, messages, and form primitives.
- Project design tokens define colors, spacing, radii, elevation, and typography.
- Components are wrapped only when consistent business behavior or styling requires it.
- PrimeVue internals are not deeply overridden with fragile selectors.
- Icons have accessible labels when meaning is not conveyed by adjacent text.

## 4. Administrator Navigation

Desktop uses a persistent side navigation and top application bar. Primary navigation:

- Dashboard
- Prospect Finder
- Prospect List
- Assignments
- Won Review
- Existing Customers
- User Management
- Settings

Prospect Finder prioritizes a map/list split view. Search filters remain visible, selected-place detail is explicit, and saving a prospect is distinct from assignment and conversion.

## 5. Sales Navigation

Mobile uses a compact header and bottom navigation for frequent destinations:

- Dashboard
- My Prospect
- My Customer
- Profile

Visit, pipeline, and attendance are contextual flows reached from the selected record. They are not crowded into global navigation.

## 6. Workflow UX Rules

- Primary action labels describe the exact next step.
- `Save as Prospect`, `Assign Sales Executive`, `Mark as Won`, and `Convert to Customer` are separate actions.
- Conversion action is absent/disabled unless status is `WON` and user is Administrator.
- Lost decisions require confirmation and reason because they are terminal.
- Pipeline cannot be changed through a generic unrestricted status dropdown.
- Success screens identify the resulting status and next destination.
- Destructive/terminal actions are visually distinct but not dependent on color alone.

## 7. Google Maps and Places

Protected interactions:

- Search center from map or device location.
- Draggable center marker.
- Radius circle.
- Radius, category, and keyword filters.
- Nearby and Text Search result modes.
- Category-specific markers.
- Synchronized list/marker selection.
- Place Detail panel.
- Phone, website, address, Maps link, rating, status, and opening hours when available.

The map must have a list-based alternative. Empty, quota, network, permission-denied, and unavailable-key states require useful messages without leaking Google errors or credentials.

## 8. Mobile Visit and Attendance

The workflow is sequential:

1. Confirm business/customer identity.
2. Request location permission in context.
3. Show current accuracy and distance/radius information.
4. Capture a live selfie or approved camera input.
5. Preview evidence.
6. Submit check-in once.
7. Show active visit state and elapsed time.
8. Submit check-out once.

Requirements:

- Minimum touch target: 44 by 44 CSS pixels.
- Primary check-in/out action remains easy to reach one-handed.
- Loading state prevents double submission.
- GPS accuracy warnings are explicit.
- Permission errors explain how to recover.
- Uploaded selfie is not permanently displayed through a public URL.
- Offline behavior initially reports that connectivity is required; offline mutation queues are out of scope unless separately approved.

## 9. Forms

- Labels remain visible; placeholders are examples, not labels.
- Required fields are identified textually.
- Validation appears beside fields and in an accessible summary for long forms.
- Server validation remains authoritative.
- Unsaved-form navigation warns the user.
- Customer conversion groups identity, contact, location/radius, ownership, and additional data.
- Coordinates may be selected on a map but are also presented textually for verification.

## 10. Tables and Lists

Administrator tables support server-driven pagination, filtering, sorting, and clear empty states. Mobile uses cards/compact lists rather than desktop tables.

Every prospect/customer status uses:

- Stable text label.
- Consistent badge treatment.
- Optional icon.
- Never color alone.

## 11. Feedback States

Every asynchronous screen defines:

- Initial loading.
- Refresh loading.
- Empty result.
- Validation failure.
- Authorization failure.
- Network/dependency failure.
- Success confirmation.
- Retry where safe.

Skeletons are preferred for predictable content; blocking spinners are reserved for atomic actions. Toast messages do not replace persistent error details for failed workflows.

## 12. Accessibility

- Target WCAG 2.1 AA for primary workflows.
- Full keyboard access for Administrator workflows.
- Visible focus indicators.
- Semantic headings and landmarks.
- Accessible names for icon buttons.
- Dialog focus trapping and restoration.
- Sufficient contrast in both normal and status states.
- Maps never provide the only representation of business results.
- Motion respects reduced-motion preferences.

## 13. Responsive Baselines

- Sales baseline: 360 px width, optimized through 480 px.
- Administrator baseline: 1024 px and above.
- Tablet widths may show an intentional unsupported/redirect message for role-inappropriate surfaces rather than a broken hybrid layout.
- The role's primary workflow must remain usable under browser zoom and dynamic text sizing.

## 14. Localization and Formatting

- Architecture supports centralized labels and formatting even if the initial UI has one language.
- API status codes remain English uppercase constants; UI supplies localized labels.
- Dates, times, numbers, and distances use locale-aware formatting.
- Stored timestamps remain UTC; the UI shows configured local timezone.

## 15. UI Acceptance Criteria

- Admin and Sales routes use different layouts and navigation.
- Direct frontend route refresh works on Vercel.
- All business actions follow `FLOW.md` without shortcuts.
- Google Maps/Places behavior satisfies the approved discovery and map acceptance criteria.
- Mobile check-in/out is touch-first and prevents accidental duplication.
- Terminal decisions require confirmation.
- All primary states are keyboard/screen-reader understandable where applicable.
- No secret, raw provider payload, or internal stack error is visible.
