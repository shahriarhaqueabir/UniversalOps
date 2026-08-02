/**
 * Local-first export utilities (CSV + PDF via print).
 *
 * No external dependencies — CSV is generated client-side and downloaded as a
 * Blob; PDF uses the browser's native print-to-PDF dialog on a printable view.
 * This keeps the app 100% local with zero telemetry.
 */

/** Escape a single CSV field per RFC 4180. */
export function csvEscape(value: unknown): string {
  if (value === null || value === undefined) return ''
  const s = String(value)
  if (/[",\n\r]/.test(s)) {
    return `"${s.replace(/"/g, '""')}"`
  }
  return s
}

/** Convert an array of flat objects into a CSV string with a header row. */
export function toCSV(rows: Record<string, unknown>[]): string {
  if (rows.length === 0) return ''
  const headers = Object.keys(rows[0])
  const lines = [headers.map(csvEscape).join(',')]
  for (const row of rows) {
    lines.push(headers.map((h) => csvEscape(row[h])).join(','))
  }
  return lines.join('\r\n')
}

/** Trigger a client-side file download from a string payload. */
export function downloadText(filename: string, content: string, mimeType = 'text/csv'): void {
  const blob = new Blob([content], { type: `${mimeType};charset=utf-8` })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

/** Build a safe filename from a report id (strip path-hostile characters). */
export function safeFilename(id: string): string {
  return id.replace(/[^a-zA-Z0-9_-]/g, '_')
}

/** Flatten a report record (including parsed data_json) into CSV rows. */
export function reportToCSVRows(report: {
  id: string
  timestamp: string
  type: string
  score: number
  data_json?: string | null
}): Record<string, unknown>[] {
  const base: Record<string, unknown> = {
    id: report.id,
    timestamp: report.timestamp,
    type: report.type,
    score: report.score,
  }
  if (!report.data_json) return [base]

  let parsed: unknown
  try {
    parsed = JSON.parse(report.data_json)
  } catch {
    return [base]
  }

  // Flatten nested objects into dot-notation columns.
  const flat: Record<string, unknown> = {}
  const walk = (obj: unknown, prefix: string) => {
    if (Array.isArray(obj)) {
      obj.forEach((item, i) => walk(item, `${prefix}[${i}]`))
      return
    }
    if (obj && typeof obj === 'object') {
      for (const [k, v] of Object.entries(obj as Record<string, unknown>)) {
        walk(v, prefix ? `${prefix}.${k}` : k)
      }
      return
    }
    flat[prefix] = obj
  }
  walk(parsed, '')
  return [{ ...base, ...flat }]
}