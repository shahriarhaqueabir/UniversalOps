import { useState } from 'react'
import { X, Download, FileText, FileJson, FileCode, AlignLeft } from 'lucide-react'
import { cn } from '@/lib/utils'

interface ExportDialogProps {
  open: boolean
  onClose: () => void
  data: Record<string, unknown>[]
  defaultFormat?: 'csv' | 'json' | 'markdown' | 'txt'
  filename?: string
  columns?: { key: string; label: string }[]
  title?: string
}

const formats = [
  { id: 'csv' as const, label: 'CSV', icon: FileText },
  { id: 'json' as const, label: 'JSON', icon: FileJson },
  { id: 'markdown' as const, label: 'Markdown', icon: FileCode },
  { id: 'txt' as const, label: 'TXT', icon: AlignLeft },
]

function formatData(
  data: Record<string, unknown>[],
  format: 'csv' | 'json' | 'markdown' | 'txt',
  columns?: { key: string; label: string }[],
): string {
  if (data.length === 0) return ''

  const keys = columns?.map(c => c.key) || Object.keys(data[0])
  const labels = columns?.map(c => c.label) || keys

  switch (format) {
    case 'csv': {
      const header = labels.join(',')
      const rows = data.map(row =>
        keys.map(k => {
          const v = row[k]
          const s = v === null || v === undefined ? '' : String(v)
          return s.includes(',') || s.includes('"') ? `"${s.replace(/"/g, '""')}"` : s
        }).join(','),
      )
      return [header, ...rows].join('\n')
    }
    case 'json':
      return JSON.stringify(data, null, 2)
    case 'markdown': {
      const header = `| ${labels.join(' | ')} |`
      const separator = `| ${labels.map(() => '---').join(' | ')} |`
      const rows = data.map(row =>
        `| ${keys.map(k => {
          const v = row[k]
          return v === null || v === undefined ? '' : String(v)
        }).join(' | ')} |`,
      )
      return [header, separator, ...rows].join('\n')
    }
    case 'txt': {
      return data
        .map(row =>
          keys
            .map(k => {
              const v = row[k]
              return `${k}: ${v === null || v === undefined ? '' : String(v)}`
            })
            .join('\n'),
        )
        .join('\n\n---\n\n')
    }
  }
}

export function ExportDialog({
  open,
  onClose,
  data,
  defaultFormat = 'csv',
  filename = 'export',
  columns,
  title = 'Export Data',
}: ExportDialogProps) {
  const [format, setFormat] = useState<'csv' | 'json' | 'markdown' | 'txt'>(defaultFormat)
  const [copied, setCopied] = useState(false)

  const formatted = formatData(data, format, columns)
  const lines = formatted.split('\n')

  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(formatted)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // fallback
    }
  }

  const handleSave = () => {
    const ext = format === 'markdown' ? 'md' : format
    const blob = new Blob([formatted], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${filename}.${ext}`
    a.click()
    URL.revokeObjectURL(url)
  }

  if (!open) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
      <div className="bg-panel border border-border rounded-xl shadow-2xl w-[640px] max-h-[80vh] flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          <div className="flex items-center gap-2">
            <Download size={18} className="text-accent" />
            <h2 className="text-lg font-semibold text-text">{title}</h2>
          </div>
          <button
            onClick={onClose}
            className="p-1 rounded hover:bg-sidebar-hover text-text-faint hover:text-text transition-colors"
          >
            <X size={18} />
          </button>
        </div>

        {/* Format Tabs */}
        <div className="flex gap-1 p-4 pb-2 border-b border-border">
          {formats.map(f => (
            <button
              key={f.id}
              onClick={() => setFormat(f.id)}
              className={cn(
                'flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-lg transition-colors',
                format === f.id
                  ? 'bg-accent text-white'
                  : 'text-text-faint hover:text-text hover:bg-sidebar-hover',
              )}
            >
              <f.icon size={14} />
              {f.label}
            </button>
          ))}
        </div>

        {/* Preview */}
        <div className="flex-1 overflow-y-auto p-4">
          <div className="bg-panel-2 border border-border rounded-lg p-3 font-[JetBrains_Mono] text-xs leading-relaxed max-h-80 overflow-y-auto">
            {lines.length > 200
              ? [...lines.slice(0, 100), '... (truncated) ...', ...lines.slice(-100)].join('\n')
              : formatted}
          </div>
          <div className="mt-2 text-xs text-text-faint">
            {lines.length} lines | {new Blob([formatted]).size.toLocaleString()} bytes
          </div>
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between p-4 border-t border-border">
          <button
            onClick={handleCopy}
            className="flex items-center gap-1.5 px-3 py-2 text-sm text-text-faint hover:text-text border border-border rounded-lg hover:bg-sidebar-hover transition-colors"
          >
            {copied ? '✓ Copied!' : 'Copy to Clipboard'}
          </button>
          <div className="flex items-center gap-2">
            <button
              onClick={onClose}
              className="px-4 py-2 text-sm text-text-faint border border-border rounded-lg hover:bg-sidebar-hover transition-colors"
            >
              Cancel
            </button>
            <button
              onClick={handleSave}
              className="flex items-center gap-1.5 px-4 py-2 text-sm bg-accent text-white rounded-lg hover:bg-accent/90 transition-colors"
            >
              <Download size={14} />
              Save to File
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
