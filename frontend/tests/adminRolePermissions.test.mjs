import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import test from 'node:test'

const SFC_URL = new URL('../src/views/Admin/Roles/AdminRoleFormView.vue', import.meta.url)

test('menu checkbox state reflects all descendants plus the menu key itself', async () => {
  const sfc = await readFile(SFC_URL, 'utf8')
  const menuState = sfc.match(/function menuState[\s\S]*?\n}/)?.[0]
  assert.ok(menuState, 'menuState must be defined')
  assert.match(menuState, /descendantsOf\(menuKey\)/)
  assert.match(menuState, /selectedKeys\.value\.has\(menuKey\)/)
  assert.match(menuState, /descendants\.length === 0/)
  assert.match(menuState, /selectedDescendants === descendants\.length/)
})

test('toggling a menu affects every descendant regardless of search visibility', async () => {
  const sfc = await readFile(SFC_URL, 'utf8')
  const toggleMenu = sfc.match(/function toggleMenu[\s\S]*?\n}/)?.[0]
  assert.ok(toggleMenu, 'toggleMenu must be defined')
  assert.match(toggleMenu, /descendantsOf\(menu\.key\)/)
  assert.doesNotMatch(toggleMenu, /visibleChildrenOf/)
  assert.match(toggleMenu, /ancestorsOf\(menu\.key\)/)
})

test('unselecting prunes ancestors only when they keep no selected descendant', async () => {
  const sfc = await readFile(SFC_URL, 'utf8')
  const hasSelectedDescendant = sfc.match(/function hasSelectedDescendant[\s\S]*?\n}/)?.[0]
  assert.ok(hasSelectedDescendant, 'hasSelectedDescendant must be defined')
  assert.match(hasSelectedDescendant, /descendantsOf\(menuKey\)/)
  const togglePermission = sfc.match(/function togglePermission[\s\S]*?\n}/)?.[0]
  assert.ok(togglePermission, 'togglePermission must be defined')
  assert.match(togglePermission, /hasSelectedDescendant\(ancestor\)/)
})

test('select visible selects only visible actions but still grants ancestors', async () => {
  const sfc = await readFile(SFC_URL, 'utf8')
  const selectVisible = sfc.match(/function selectVisible[\s\S]*?\n}/)?.[0]
  assert.ok(selectVisible, 'selectVisible must be defined')
  assert.match(selectVisible, /perm\.nodeType === 'ACTION'/)
  assert.match(selectVisible, /ancestorsOf\(key\)/)
})

test('recommended permissions can be applied explicitly and match backend defaults', async () => {
  const sfc = await readFile(SFC_URL, 'utf8')
  assert.match(sfc, /RECOMMENDED_PERMISSION_KEYS/)
  assert.match(sfc, /2: \['menu_sales_dashboard', 'view_sales_dashboard', 'view_team_dashboard'/)
  assert.match(sfc, /3: \['menu_sales_dashboard', 'view_sales_dashboard', 'view_team_dashboard'/)
  assert.match(sfc, /4: \['menu_sales_dashboard', 'view_sales_dashboard', 'menu_my_prospects'/)
  assert.match(sfc, /function applyRecommendedPermissions/)
  assert.match(sfc, /normalizedPermissionSet\(recommendedKeys\.value\)/)
  assert.match(sfc, /Apply Recommended/)
  assert.match(sfc, /Reset to Recommended/)
})

test('editing a role does not overwrite custom permission selection on load or level change', async () => {
  const sfc = await readFile(SFC_URL, 'utf8')
  const levelWatch = sfc.match(/watch\(\s*\(\) => form\.level[\s\S]*?\n\)/)?.[0]
  assert.ok(levelWatch, 'level watcher must be defined')
  assert.match(levelWatch, /if \(isEdit\.value/)
  assert.match(levelWatch, /selectedKeys\.value\.size > 0/)
  assert.match(sfc, /Saved custom selections are kept until you apply or reset explicitly/)
})

test('team dashboard capability does not create duplicate landing page route', async () => {
  const sfc = await readFile(SFC_URL, 'utf8')
  assert.match(sfc, /perm\.nodeType !== 'ACTION' \|\| !perm\.routePath/)
  assert.match(sfc, /perm\.key === 'view_team_dashboard'/)
})
