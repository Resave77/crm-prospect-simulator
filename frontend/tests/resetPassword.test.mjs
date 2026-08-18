import assert from 'node:assert/strict'
import test from 'node:test'
import { adminTemporaryPasswordError } from '../src/utils/admin.ts'

test('adminTemporaryPasswordError accepts a password meeting all rules', () => {
  assert.equal(adminTemporaryPasswordError('TempPass1'), null)
  assert.equal(adminTemporaryPasswordError('Passw0rd!2x'), null)
})

test('adminTemporaryPasswordError rejects passwords shorter than 6 characters', () => {
  const message = adminTemporaryPasswordError('abc12')
  assert.match(message ?? '', /at least 6 characters/)
})

test('adminTemporaryPasswordError accepts password123 without character classes', () => {
  assert.equal(adminTemporaryPasswordError('password123'), null)
})

test('adminTemporaryPasswordError rejects whitespace-only passwords', () => {
  assert.notEqual(adminTemporaryPasswordError('     '), null)
})
