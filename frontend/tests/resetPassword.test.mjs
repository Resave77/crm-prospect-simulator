import assert from 'node:assert/strict'
import test from 'node:test'
import { adminTemporaryPasswordError } from '../src/utils/admin.ts'

test('adminTemporaryPasswordError accepts a password meeting all rules', () => {
  assert.equal(adminTemporaryPasswordError('TempPass1'), null)
  assert.equal(adminTemporaryPasswordError('Passw0rd!2x'), null)
})

test('adminTemporaryPasswordError rejects passwords shorter than 8 characters', () => {
  const message = adminTemporaryPasswordError('Ab1cde')
  assert.match(message ?? '', /at least 8 characters/)
})

test('adminTemporaryPasswordError requires an uppercase letter', () => {
  assert.match(adminTemporaryPasswordError('abc12345') ?? '', /uppercase letter/)
})

test('adminTemporaryPasswordError requires a lowercase letter', () => {
  assert.match(adminTemporaryPasswordError('ABC12345') ?? '', /lowercase letter/)
})

test('adminTemporaryPasswordError requires a number', () => {
  assert.match(adminTemporaryPasswordError('Abcdefgh') ?? '', /number/)
})

test('adminTemporaryPasswordError rejects whitespace-only passwords', () => {
  assert.notEqual(adminTemporaryPasswordError('        '), null)
})
