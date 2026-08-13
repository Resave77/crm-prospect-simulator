import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const profile = () => readFile(new URL('../src/views/Sales/ProfileView.vue', import.meta.url), 'utf8')
const salesLayout = () => readFile(new URL('../src/layouts/SalesLayout.vue', import.meta.url), 'utf8')

test('Sales Profile reuses existing account and performance data', async () => {
  const source = await profile()
  for (const value of ['auth.user?.fullName', 'auth.user?.employeeId', 'auth.user?.email', 'auth.user?.phone', 'roleLabel', 'totalVisits', 'wonProspects', 'completedVisits']) {
    assert.ok(source.includes(value), `${value} must remain rendered`)
  }
  assert.match(source, /await crm\.loadMyProspects\(\)/)
  assert.match(source, /getMyProspect\(p\.id\)/)
  assert.match(source, /@click="logout"/)
})

test('Sales Profile uses a bounded responsive dashboard grid without viewport math', async () => {
  const source = await profile()
  assert.match(source, /grid-template-columns: minmax\(320px, \.9fr\) minmax\(0, 1\.6fr\)/)
  assert.match(source, /@media \(min-width: 768px\) and \(max-width: 1099px\)/)
  assert.match(source, /@media \(max-width: 767px\)/)
  assert.match(source, /\.profile-page > \* \{ min-width: 0; max-width: 100%; box-sizing: border-box; \}/)
  assert.doesNotMatch(source, /width:\s*100vw|calc\(100vw\s*-\s*(?:240|68)px\)/)
  assert.doesNotMatch(source, /overflow-x:\s*hidden/)
})

test('Sales Profile account details and metrics wrap safely', async () => {
  const source = await profile()
  assert.match(source, /profile-card-top/)
  assert.doesNotMatch(source, /profile-card-cover/)
  assert.match(source, /\.strip-value[\s\S]*?white-space: normal; overflow-wrap: anywhere/)
  assert.match(source, /\.perf-stats \{ flex: 1; grid-template-columns: repeat\(3, minmax\(0, 1fr\)\)/)
  assert.match(source, /\.perf-stats \{ grid-template-columns: repeat\(2, minmax\(0, 1fr\)\); \}/)
})

test('Sales Profile redesign leaves mobile Sales navigation intact', async () => {
  const [source, layout] = await Promise.all([profile(), salesLayout()])
  assert.doesNotMatch(source, /sales-nav|mobile-bottom-nav|sidebarCollapsed/)
  assert.match(layout, /class="sales-nav" aria-label="Sales navigation"/)
  assert.match(layout, /--sales-shell-sidebar-w: 240px/)
  assert.match(layout, /--sales-shell-sidebar-w: 68px/)
})
