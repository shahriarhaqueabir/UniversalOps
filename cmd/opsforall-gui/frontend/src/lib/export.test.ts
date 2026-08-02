import { describe, it, expect } from 'vitest'
import { csvEscape, toCSV, reportToCSVRows, safeFilename } from '@/lib/export'

describe('csvEscape', () => {
  it('returns empty string for null/undefined', () => {
    expect(csvEscape(null)).toBe('')
    expect(csvEscape(undefined)).toBe('')
  })

  it('passes through simple values', () => {
    expect(csvEscape('cpu')).toBe('cpu')
    expect(csvEscape(90)).toBe('90')
  })

  it('quotes fields containing commas, quotes, or newlines', () => {
    expect(csvEscape('a,b')).toBe('"a,b"')
    expect(csvEscape('say "hi"')).toBe('"say ""hi"""')
    expect(csvEscape('line1\nline2')).toBe('"line1\nline2"')
  })
})

describe('toCSV', () => {
  it('returns empty string for no rows', () => {
    expect(toCSV([])).toBe('')
  })

  it('writes a header row and data rows', () => {
    const csv = toCSV([
      { metric: 'cpu.percent', value: 90 },
      { metric: 'mem.used', value: 50 },
    ])
    expect(csv).toBe('metric,value\r\ncpu.percent,90\r\nmem.used,50')
  })

  it('escapes values in data rows', () => {
    const csv = toCSV([{ note: 'a,b', n: 1 }])
    expect(csv).toBe('note,n\r\n"a,b",1')
  })
})

describe('safeFilename', () => {
  it('strips path-hostile characters', () => {
    expect(safeFilename('report:health/1')).toBe('report_health_1')
    expect(safeFilename('abc-123')).toBe('abc-123')
  })
})

describe('reportToCSVRows', () => {
  it('returns a single base row when no data_json', () => {
    const rows = reportToCSVRows({ id: 'r1', timestamp: 't', type: 'health', score: 90 })
    expect(rows).toHaveLength(1)
    expect(rows[0]).toMatchObject({ id: 'r1', type: 'health', score: 90 })
  })

  it('flattens nested data_json into dot-notation columns', () => {
    const rows = reportToCSVRows({
      id: 'r1',
      timestamp: 't',
      type: 'health',
      score: 90,
      data_json: JSON.stringify({ checks: [{ name: 'cpu', status: 'pass' }], score: 90 }),
    })
    expect(rows).toHaveLength(1)
    expect(rows[0]).toMatchObject({
      id: 'r1',
      'checks[0].name': 'cpu',
      'checks[0].status': 'pass',
      score: 90,
    })
  })

  it('falls back to base row on invalid JSON', () => {
    const rows = reportToCSVRows({
      id: 'r1',
      timestamp: 't',
      type: 'health',
      score: 90,
      data_json: 'not json',
    })
    expect(rows).toHaveLength(1)
    expect(rows[0]).toMatchObject({ id: 'r1', score: 90 })
  })
})