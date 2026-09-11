# Master Data UI Update Audit

## Scope

This isolated change is for the Master Data view only:

- `frontend/src/components/admin/MasterDataPanel.vue`
- `frontend/src/components/admin/MasterDataTrashPanel.vue`
- `frontend/src/views/Admin/Customer/CustomerListView.vue`

The existing Segment and Category records are still read from the existing API/fallback data. No table seed/content migration is included in this UI commit.

## UI and behavior checked

- Segment and Category cards use the same responsive card layout and stretch to the same height.
- The card background remains white and aligned with the surrounding page surface.
- The card does not add an internal scrolling area.
- Category Google Place Types wrap into pills inside the available cell width.
- Category `Show All` / `Show Less` remains available for the list view.
- Normal row actions remain available for View, Edit, and Delete.
- Edit Order is inline: the Action column is hidden while ordering, Cancel restores the draft, and Save Order applies the reordered rows in the current view.
- Trash uses a white surface and supports loading, empty state, and restore actions.
- Existing form/detail status handling remains available; the visible table Status column is intentionally omitted according to the latest UI request.

## Validation results

- `frontend: npm.cmd run typecheck` — passed.
- `frontend: npm.cmd test -- --run` — passed, 96/96 tests.
- `frontend: npm.cmd run build` — passed, 576 modules transformed.
- `backend: go test ./...` — passed.
- `git diff --check` — passed; only normal Git line-ending warnings were reported.

Interactive browser click-through was not available in this environment, so the final QA should still open `/admin/customers?tab=master` and verify View, Edit, Delete, Trash/Restore, Show All/Show Less, and Edit Order manually.

## Intentionally excluded from this UI commit

These existing staged/untracked changes are kept outside the UI commit so they cannot alter the Master Data table contents or introduce unrelated work:

- Shortcut AI backend/model/API changes.
- `backend/prisma/migrations/202608280001_master_data_shortcut_ai/`.
- `backend/prisma/migrations/202609010001_master_data_category_place_types/` (updates stored Google Place Types values).

## Conflict-safe handoff prompt

> Pull only the Master Data UI change from branch `feature/master-data-ui-sync`. Fetch the branch first, create a local feature branch from the team's current base, and apply the single isolated commit. Do not reset, force-push, or overwrite existing files. Before applying it, run `git status`; if the same three files have local edits, preserve them and resolve only those files manually. Then run the frontend typecheck/tests/build and open `/admin/customers?tab=master` for the manual QA checklist above.

The new remote cannot be configured until its repository URL is supplied. The existing `origin` is intentionally left unchanged.
