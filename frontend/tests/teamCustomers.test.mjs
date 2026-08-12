import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const VIEW_URL = new URL('../src/views/Sales/Customer/MyCustomersView.vue', import.meta.url)

test('my customers page selects team mode only when backend confirms team scope', async () => {
  const view = await readFile(VIEW_URL, 'utf8')
  assert.match(view, /auth\.hasPermission\('view_team_dashboard'\)/)
  assert.match(view, /crm\.loadTeamCustomers\(\)/)
  assert.match(view, /if \(team\?\.hasTeam\)/)
  assert.match(view, /teamMode\.value = true/)
})

test('individual my customers mode remains the fallback', async () => {
  const view = await readFile(VIEW_URL, 'utf8')
  assert.match(view, /teamMode\.value = false/)
  assert.match(view, /await crm\.loadMyCustomers\(\)/)
  assert.match(view, /teamMode\.value \? crm\.teamCustomers\?\.customers \?\? \[\] : crm\.myCustomers/)
  assert.match(view, /teamMode\.value \? 'Team Customers' : 'My Customers'/)
})

test('team customer cards render assigned sales context and hide field-only actions', async () => {
  const view = await readFile(VIEW_URL, 'utf8')
  assert.match(view, /customer\.assignedSales\?\.fullName/)
  assert.match(view, /customer\.assignedSales\.roleName/)
  assert.match(view, /customer\.assignedSales\.roleLevel/)
  assert.match(view, /v-if="!teamMode" class="mc-action-btn mc-action-detail"/)
  assert.match(view, /v-if="!teamMode" class="mc-action-btn mc-action-checkin"/)
})
