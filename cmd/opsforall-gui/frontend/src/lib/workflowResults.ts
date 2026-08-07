/**
 * Standardized Execution Result helpers for the Workflow Center.
 *
 * The backend wraps every workflow step payload in a StepResult envelope
 * (see internal/common/workflow.go) with stable snake_case keys. These helpers
 * decode that envelope for rendering: status badge, summary line, meta footer
 * (duration / item count / timestamp), and a table-friendly row view.
 */

// StepResult mirrors internal/common.StepResult json tags.
export interface StepResult {
  status: 'success' | 'error'
  summary?: string
  items?: number
  duration_ns?: number
  timestamp?: string
  data?: any
  error?: string
}

export function isStepResult(r: any): r is StepResult {
  return (
    r !== null &&
    typeof r === 'object' &&
    typeof r.status === 'string' &&
    ('data' in r || 'summary' in r || 'error' in r || 'duration_ns' in r)
  )
}

// Shell payload nested inside envelope.data. devops.ShellResult has no json
// tags, so it serializes with PascalCase keys — detect it the same way the
// old duck-typing did, but scoped to the envelope's data field.
export interface ShellPayload {
  Command: string
  Output: string
  ExitCode: number
  Duration: number
}

export function isShellPayload(d: any): d is ShellPayload {
  return (
    d !== null &&
    typeof d === 'object' &&
    typeof d.Command === 'string' &&
    typeof d.Output === 'string' &&
    typeof d.ExitCode === 'number'
  )
}

export function formatDurationNs(ns?: number): string {
  if (!ns || ns <= 0) return ''
  if (ns < 1_000_000) return '<1ms'
  const ms = Math.floor(ns / 1_000_000)
  if (ms < 1000) return `${ms}ms`
  if (ms < 60_000) return `${(ms / 1000).toFixed(1)}s`
  const min = Math.floor(ms / 60_000)
  const sec = Math.floor((ms % 60_000) / 1000)
  return `${min}m ${sec}s`
}

export function formatTimestamp(ts?: string): string {
  if (!ts) return ''
  const d = new Date(ts)
  if (Number.isNaN(d.getTime())) return ''
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
}

export interface ResultRow {
  key: string
  value: any
}

/**
 * Flattens an arbitrary payload into key/value rows for a table or grid.
 * - Array of objects -> one row per entry per field, prefixed with the index.
 * - Plain object      -> one row per key.
 * - Scalar / string   -> single row.
 */
export function resultToRows(data: any): ResultRow[] {
  if (data === null || data === undefined) return []

  if (Array.isArray(data)) {
    if (data.length === 0) return []
    const rows: ResultRow[] = []
    data.forEach((item, i) => {
      if (item !== null && typeof item === 'object' && !Array.isArray(item)) {
        const entries = Object.entries(item)
        if (entries.length === 0) {
          rows.push({ key: `#${i + 1}`, value: item })
          return
        }
        for (const [k, v] of entries) {
          rows.push({ key: `#${i + 1} · ${k}`, value: v })
        }
      } else {
        rows.push({ key: `#${i + 1}`, value: item })
      }
    })
    return rows
  }

  if (typeof data === 'object') {
    return Object.entries(data).map(([k, v]) => ({ key: k, value: v }))
  }

  return [{ key: 'value', value: data }]
}

/** Renders a scalar value; nested objects/arrays are JSON-stringified. */
export function formatValue(v: any): string {
  if (v === null || v === undefined) return ''
  if (typeof v === 'string') return v
  if (typeof v === 'number' || typeof v === 'boolean') return String(v)
  return JSON.stringify(v, null, 2)
}

/** Extracts the one-line interpretation for a step envelope. */
export function getResultSummary(r: StepResult): string {
  if (r.status === 'error') return r.error || 'Step failed'
  return r.summary || 'Completed'
}
