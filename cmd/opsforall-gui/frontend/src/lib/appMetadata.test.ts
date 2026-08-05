import { describe, it, expect } from 'vitest'
import { buildDocumentMetadata } from './appMetadata'

describe('buildDocumentMetadata', () => {
  it('builds a page-specific title and description', () => {
    const meta = buildDocumentMetadata('sysops')

    expect(meta.title).toBe('System Operations · Universal-Ops')
    expect(meta.description).toContain('CPU, memory, disk, process, and hardware telemetry')
    expect(meta.keywords).toContain('system operations')
  })

  it('includes active alert counts in the alerts title and description', () => {
    const meta = buildDocumentMetadata('alerts', 3)

    expect(meta.title).toBe('Alerts Dashboard (3) · Universal-Ops')
    expect(meta.description).toContain('3 active alerts require attention')
  })
})
