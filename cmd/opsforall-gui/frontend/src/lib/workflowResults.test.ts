import { describe, expect, it } from 'vitest'
import {
  formatDurationNs,
  formatTimestamp,
  formatValue,
  getResultSummary,
  isShellPayload,
  isStepResult,
  resultToRows,
  shellPayloadOf,
} from './workflowResults'

describe('isStepResult', () => {
  it('recognizes a standardized envelope', () => {
    expect(isStepResult({ status: 'success', summary: '2 records', data: [] })).toBe(true)
  })
  it('rejects raw payloads and non-objects', () => {
    expect(isStepResult({ Command: 'x', Output: 'y', ExitCode: 0 })).toBe(false)
    expect(isStepResult('hello')).toBe(false)
    expect(isStepResult(null)).toBe(false)
    expect(isStepResult(undefined)).toBe(false)
  })
})

describe('isShellPayload', () => {
  it('recognizes the PascalCase shell payload nested in envelope.data', () => {
    expect(isShellPayload({ Command: 'Get-Process', Output: 'a\nb', ExitCode: 0, Duration: 5 })).toBe(true)
  })
  it('rejects envelopes and non-shell shapes', () => {
    expect(isShellPayload({ status: 'success', data: [] })).toBe(false)
    expect(isShellPayload('nope')).toBe(false)
  })
})

describe('shellPayloadOf', () => {
  it('returns the payload for shell results', () => {
    const payload = shellPayloadOf({ Command: 'Get-Process', Output: 'a\nb', ExitCode: 0, Duration: 5 })
    expect(payload?.ExitCode).toBe(0)
    expect(payload?.Output).toBe('a\nb')
  })
  it('returns null for non-shell shapes', () => {
    expect(shellPayloadOf({ status: 'success', data: [] })).toBeNull()
    expect(shellPayloadOf('nope')).toBeNull()
    expect(shellPayloadOf(null)).toBeNull()
    expect(shellPayloadOf(undefined)).toBeNull()
  })
})

describe('formatDurationNs', () => {
  it('formats ms, seconds, and minutes', () => {
    expect(formatDurationNs(500_000)).toBe('<1ms')
    expect(formatDurationNs(12_000_000)).toBe('12ms')
    expect(formatDurationNs(1_500_000_000)).toBe('1.5s')
    expect(formatDurationNs(125_000_000_000)).toBe('2m 5s')
  })
  it('returns empty for missing/zero values', () => {
    expect(formatDurationNs(undefined)).toBe('')
    expect(formatDurationNs(0)).toBe('')
  })
})

describe('formatTimestamp', () => {
  it('formats an RFC3339 timestamp', () => {
    expect(formatTimestamp('2026-08-07T12:34:56Z')).toMatch(/\d{2}:\d{2}:\d{2}/)
  })
  it('returns empty for garbage', () => {
    expect(formatTimestamp('not-a-date')).toBe('')
    expect(formatTimestamp(undefined)).toBe('')
  })
})

describe('resultToRows', () => {
  it('flattens an array of objects with index prefixes', () => {
    const rows = resultToRows([{ name: 'svc1', status: 'running' }, { name: 'svc2' }])
    expect(rows).toEqual([
      { key: '#1 · name', value: 'svc1' },
      { key: '#1 · status', value: 'running' },
      { key: '#2 · name', value: 'svc2' },
    ])
  })
  it('flattens a plain object', () => {
    expect(resultToRows({ cpu: 12, mem: 80 })).toEqual([
      { key: 'cpu', value: 12 },
      { key: 'mem', value: 80 },
    ])
  })
  it('wraps a scalar in a single row', () => {
    expect(resultToRows('done')).toEqual([{ key: 'value', value: 'done' }])
  })
  it('returns [] for empty arrays, null, and undefined', () => {
    expect(resultToRows([])).toEqual([])
    expect(resultToRows(null)).toEqual([])
    expect(resultToRows(undefined)).toEqual([])
  })
})

describe('formatValue', () => {
  it('passes strings through and stringifies nested objects', () => {
    expect(formatValue('plain')).toBe('plain')
    expect(formatValue(42)).toBe('42')
    expect(formatValue({ a: 1 })).toContain('"a": 1')
  })
  it('returns empty for null/undefined', () => {
    expect(formatValue(null)).toBe('')
    expect(formatValue(undefined)).toBe('')
  })
})

describe('getResultSummary', () => {
  it('uses summary for success and error message for failures', () => {
    expect(getResultSummary({ status: 'success', summary: '3 services' })).toBe('3 services')
    expect(getResultSummary({ status: 'error', error: 'boom' })).toBe('boom')
    expect(getResultSummary({ status: 'error' })).toBe('Step failed')
    expect(getResultSummary({ status: 'success' })).toBe('Completed')
  })
})
