import { Terminal } from 'lucide-react'

interface CommandOutputProps {
  output: string
  title?: string
}

export function CommandOutput({ output, title }: CommandOutputProps) {
  return (
    <div className="bg-[var(--color-panel-3)] border border-[var(--color-border)] rounded-xl overflow-hidden">
      {title && (
        <div className="flex items-center gap-2 px-5 py-2 border-b border-[var(--color-border)] bg-[var(--color-panel-2)]">
          <Terminal size={14} className="text-[var(--color-accent)]" />
          <span className="text-xs font-bold text-[var(--color-text-faint)] uppercase">{title}</span>
        </div>
      )}
      <pre className="p-5 text-sm font-mono text-[var(--color-text-dim)] overflow-x-auto whitespace-pre-wrap">
        {output || 'No output'}
      </pre>
    </div>
  )
}
