import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  Library, ChevronRight,
  Zap, Terminal, Search,
  RefreshCw
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { useBackend } from '@/hooks/useBackend'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { toast } from 'sonner'

interface WorkflowStep {
  id: string
  label: string
  description: string
  command: string
  expected_outcome: string
}

interface Workflow {
  id: string
  name: string
  description: string
  why: string
  risks: string[]
  typical_values: string
  steps: WorkflowStep[]
}

export function WorkflowCenter() {
  const { call } = useBackend()
  const [selectedWf, setSelectedWf] = useState<Workflow | null>(null)
  const [activeStepIdx, setActiveStepIdx] = useState<number | null>(null)
  const [search, setSearch] = useState('')

  const { data: workflows = [], isLoading } = useQuery<Workflow[]>({
    queryKey: ['workflows'],
    queryFn: async () => (await call('WorkflowAPI.ListWorkflows')) as Workflow[] || []
  })

  const handleExecute = async () => {
    if (!selectedWf) return
    const id = toast.loading(`Executing ${selectedWf.name}...`)
    try {
      const result = await call('WorkflowAPI.ExecuteWorkflow', selectedWf.id) as Workflow
      setSelectedWf(result)
      toast.success(`${selectedWf.name} completed`, { id })
    } catch (err: any) {
      toast.error(err?.message || 'Execution failed', { id })
    }
  }

  const filtered = workflows.filter(w =>
    w.name.toLowerCase().includes(search.toLowerCase()) ||
    w.description.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className="flex flex-col h-full bg-[var(--color-bg)] animate-in fade-in duration-500 overflow-hidden">
      {/* Header */}
      <div className="py-8 border-b border-[var(--color-border)] bg-[var(--color-panel-2)]/50 px-10 flex items-center justify-between">
        <div>
          <div className="flex items-center gap-3 mb-2">
            <div className="w-8 h-8 rounded-lg bg-accent/10 flex items-center justify-center text-accent border border-accent/20">
               <Library size={18} />
            </div>
            <h1 className="text-sm font-black text-[var(--color-text)] uppercase tracking-[0.25em]">Ops Library</h1>
          </div>
          <p className="text-3xl font-bold text-[var(--color-text)] tracking-tight">Workflow Center</p>
          <p className="text-[var(--color-text-dim)] text-xs font-semibold uppercase tracking-widest mt-2">Reusable operational sequences for high-fidelity system diagnostics</p>
        </div>
      </div>

      <div className="flex-1 flex overflow-hidden">
        {/* Left: Library List */}
        <div className="w-80 border-r border-[var(--color-border)] bg-[var(--color-panel)] flex flex-col">
          <div className="p-4 border-b border-[var(--color-border)]/50">
            <div className="relative group">
              <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-text-faint group-focus-within:text-accent transition-colors" />
              <input
                type="text"
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Search workflows..."
                className="w-full bg-[var(--color-bg)] border border-border rounded-xl pl-9 pr-4 py-2 text-xs font-bold text-text focus:outline-none focus:border-accent transition-all"
              />
            </div>
          </div>
          <div className="flex-1 overflow-y-auto p-4 space-y-2">
            {isLoading ? (
               <div className="flex flex-col items-center justify-center py-10 opacity-30">
                 <RefreshCw size={24} className="animate-spin mb-2" />
                 <p className="text-[10px] font-black uppercase tracking-widest">Loading Library...</p>
               </div>
            ) : filtered.map(wf => (
              <button
                key={wf.id}
                onClick={() => { setSelectedWf(wf); setActiveStepIdx(null); }}
                className={cn(
                  "w-full text-left p-4 rounded-2xl transition-all border group",
                  selectedWf?.id === wf.id
                    ? "bg-accent/10 border-accent/30 shadow-lg"
                    : "bg-panel-2/50 border-transparent hover:border-white/10 hover:bg-panel-2"
                )}
              >
                <p className={cn("text-xs font-black uppercase tracking-widest mb-1", selectedWf?.id === wf.id ? "text-accent" : "text-white")}>{wf.name}</p>
                <p className="text-[10px] text-text-dim font-medium leading-relaxed line-clamp-2">{wf.description}</p>
              </button>
            ))}
          </div>
        </div>

        {/* Right: Selected Workflow Workspace */}
        <div className="flex-1 overflow-y-auto bg-[var(--color-bg)]/30 relative">
          {!selectedWf ? (
            <div className="flex flex-col items-center justify-center h-full opacity-20 pointer-events-none">
              <Library size={120} className="text-accent mb-6" />
              <p className="text-xl font-black uppercase tracking-[0.3em] text-white">Select a workflow</p>
              <p className="text-sm font-bold text-text-dim mt-2 tracking-widest">Select an entry from the library to begin discovery</p>
            </div>
          ) : (
            <div className="p-10 space-y-8 max-w-5xl mx-auto animate-in fade-in slide-in-from-right-4 duration-500">
              <SectionBriefing
                title={selectedWf.name}
                objective={selectedWf.description}
                checklist={selectedWf.steps.map(s => s.label)}
                why={selectedWf.why}
                risks={selectedWf.risks}
                typicalValues={selectedWf.typical_values}
                recommendedActions={["Run 'Dry Run' review first", "Verify network isolation if required"]}
              />

              <div className="space-y-6">
                <div className="flex items-center justify-between">
                  <h4 className="text-[10px] font-black uppercase tracking-[0.25em] text-text-faint flex items-center gap-2">
                    <Terminal size={14} /> Execution Pipeline (Dry Run)
                  </h4>
                  <div className="px-3 py-1 rounded-full bg-success/10 border border-success/30 text-success text-[9px] font-black uppercase tracking-widest">
                    Safe to Review
                  </div>
                </div>

                <div className="space-y-4">
                  {selectedWf.steps.map((step, idx) => (
                    <div
                      key={step.id}
                      className={cn(
                        "p-6 rounded-[2rem] border transition-all relative overflow-hidden group cursor-pointer",
                        activeStepIdx === idx ? "bg-panel-2 border-accent/50 shadow-2xl scale-[1.01]" : "bg-panel-2/40 border-white/5 hover:border-white/10"
                      )}
                      onClick={() => setActiveStepIdx(idx)}
                    >
                      <div className="flex items-start justify-between gap-6 relative z-10">
                        <div className="flex gap-5 min-w-0">
                          <div className={cn(
                            "w-10 h-10 rounded-xl flex items-center justify-center shrink-0 border transition-colors",
                            activeStepIdx === idx ? "bg-accent text-white border-white/20 shadow-[0_0_20px_rgba(var(--color-accent-rgb),0.4)]" : "bg-panel-3 text-text-faint border-white/5"
                          )}>
                            <span className="text-sm font-black italic">{idx + 1}</span>
                          </div>
                          <div className="min-w-0">
                            <p className="text-sm font-black text-white uppercase tracking-widest">{step.label}</p>
                            <p className="text-[11px] text-text-dim mt-1 font-medium leading-relaxed">{step.description}</p>
                          </div>
                        </div>
                        <div className="text-right shrink-0">
                          <p className="text-[9px] font-black uppercase text-accent tracking-tighter mb-1">Expected Outcome</p>
                          <p className="text-[10px] text-success font-bold italic">{step.expected_outcome}</p>
                        </div>
                      </div>

                      {activeStepIdx === idx && (
                        <div className="mt-6 pt-6 border-t border-white/5 animate-in slide-in-from-top-2 duration-300">
                           <p className="text-[9px] font-black uppercase text-text-faint tracking-widest mb-2 flex items-center gap-2">
                             <Zap size={10} className="text-warning" /> Primary Native Command
                           </p>
                           <div className="bg-black/40 rounded-xl p-4 font-mono text-xs text-warning border border-warning/20 shadow-inner">
                             {step.command}
                           </div>
                        </div>
                      )}
                    </div>
                  ))}
                </div>

                <div className="pt-10 flex gap-4">
                  <button
                    onClick={handleExecute}
                    className="flex-1 py-6 bg-accent text-white rounded-[2rem] font-black uppercase tracking-[0.3em] shadow-[0_0_50px_rgba(var(--color-accent-rgb),0.3)] hover:scale-[1.02] active:scale-[0.98] transition-all flex items-center justify-center gap-4 group"
                  >
                    Execute Workflow <ChevronRight size={24} className="group-hover:translate-x-2 transition-transform" />
                  </button>
                  <button
                    onClick={() => setSelectedWf(null)}
                    className="px-10 py-6 bg-panel-3 border border-border text-text-dim rounded-[2rem] font-black uppercase tracking-[0.2em] hover:text-white transition-all"
                  >
                    Clear
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
