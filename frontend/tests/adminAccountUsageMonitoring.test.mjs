import test from 'node:test'
import assert from 'node:assert/strict'
import fs from 'node:fs'

test('Admin Account Detail renders monitoring labels and honest unknown states', () => {
  const source = fs.readFileSync(new URL('../src/views/Admin/Accounts/AdminAccountDetailView.vue', import.meta.url), 'utf8')
  for (const label of ['Request Provider Tercatat', 'Kuota Gratis', 'Penggunaan', 'Estimasi Biaya', 'Aktivitas', 'Cache', 'Hit Provider', 'Status', 'Detail']) {
    assert.match(source, new RegExp(label))
  }
  assert.match(source, /Belum dikonfigurasi/)
  assert.match(source, /provider_hit_count/)
  assert.match(source, /cache_status/)
})
