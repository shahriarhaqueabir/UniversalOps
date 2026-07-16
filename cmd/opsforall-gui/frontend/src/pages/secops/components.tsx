import { Shield } from 'lucide-react'

export function ExpertInsight({ title, content }: { title: string; content: string }) {
  return (
    <div className="bg-accent-soft border border-accent/20 rounded-2xl p-8 flex gap-8 animate-in slide-in-from-left-4 duration-500">
      <div className="w-14 h-14 rounded-2xl bg-accent flex items-center justify-center shrink-0 shadow-xl">
        <Shield size={32} className="text-white" />
      </div>
      <div>
        <h4 className="text-2xl font-bold text-text mb-2">{title}</h4>
        <p className="text-text-dim text-xl leading-relaxed">{content}</p>
      </div>
    </div>
  )
}
