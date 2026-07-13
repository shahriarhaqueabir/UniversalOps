import { ArrowRight } from 'lucide-react'

interface EmptyStateProps {
  icon: React.ReactNode
  title: string
  description: string
  action?: { label: string; onClick: () => void }
}

export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-20 px-8 text-center">
      <div className="w-20 h-20 rounded-2xl bg-[var(--color-panel-2)] border border-[var(--color-border)] flex items-center justify-center text-[var(--color-text-faint)] mb-6">
        {icon}
      </div>
      <h3 className="text-xl font-bold text-[var(--color-text)] mb-2">{title}</h3>
      <p className="text-sm text-[var(--color-text-dim)] max-w-md leading-relaxed mb-6">{description}</p>
      {action && (
        <button
          onClick={action.onClick}
          className="flex items-center gap-2 px-5 py-2.5 rounded-xl bg-[var(--color-accent)] text-white text-sm font-bold hover:opacity-90 transition-opacity"
        >
          {action.label} <ArrowRight size={16} />
        </button>
      )}
    </div>
  )
}
