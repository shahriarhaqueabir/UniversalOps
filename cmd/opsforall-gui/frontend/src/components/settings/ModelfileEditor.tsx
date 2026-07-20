import { useState, useEffect } from 'react'
import { useBackend } from '@/hooks/useBackend'
import { useConfigStore } from '@/stores/useConfigStore'
import { FileCode, Loader2, Info, AlertTriangle, FileUp } from 'lucide-react'
import { cn } from '@/lib/utils'
import { toast } from 'sonner'

/**
 * ModelfileEditor — Raw text editor for the AI's neural core.
 * Allows deep customization of the companion's system instructions and parameters.
 */
export function ModelfileEditor() {
  const { call } = useBackend()
  const { stagedChanges, stageChange } = useConfigStore()
  const [content, setContent] = useState<string>('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchModelfile = async () => {
      try {
        const res = await call('AIOps.GetModelfile')
        setContent(res as string)
      } catch (err) {
        console.error('Failed to load Modelfile:', err)
      } finally {
        setLoading(false)
      }
    }
    fetchModelfile()
  }, [call])

  // Sync with staged changes if already present
  useEffect(() => {
    if (stagedChanges.has('modelfile')) {
      setContent(stagedChanges.get('modelfile'))
    }
  }, [stagedChanges])

  const handleChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    const newVal = e.target.value
    setContent(newVal)
    stageChange('modelfile', newVal)
  }

  const handleLoadFile = async () => {
    try {
      const path = await call('App.OpenFileDialog', 'Select Modelfile', ['Modelfile|*.modelfile', 'All Files|*.*'])
      if (path) {
        // Since we can't easily read arbitrary files via App.OpenFileDialog result in this specific bridge
        // without a dedicated reader, we'll assume the user wants to load the CONTENT.
        // Actually, let's add a ReadFile binding to App to make this useful.
        const fileContent = await call('App.ReadTextFile', path)
        if (fileContent) {
          setContent(fileContent as string)
          stageChange('modelfile', fileContent)
          toast.success('Modelfile loaded into editor')
        }
      }
    } catch (err: any) {
      toast.error(err?.message || 'Failed to load file')
    }
  }

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center p-12 bg-[var(--color-panel-2)] rounded-2xl border border-[var(--color-border)] border-dashed">
        <Loader2 size={24} className="text-[var(--color-accent)] animate-spin mb-3" />
        <p className="text-xs text-[var(--color-text-dim)] font-bold uppercase tracking-widest">Loading Neural Core...</p>
      </div>
    )
  }

  const isModified = stagedChanges.has('modelfile')

  return (
    <div className="space-y-4 animate-in fade-in duration-500">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-violet-400/10 text-violet-400 flex items-center justify-center border border-violet-400/20">
            <FileCode size={16} />
          </div>
          <div>
            <p className="text-xs font-black uppercase tracking-tight text-[var(--color-text)]">Modelfile Editor</p>
            <p className="text-[10px] text-[var(--color-text-dim)] mt-0.5">Edit the raw persona definition and parameters</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={handleLoadFile}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-white/5 border border-white/10 text-[9px] font-black uppercase text-white hover:bg-white/10 transition-all"
          >
            <FileUp size={12} /> Import File
          </button>
          {isModified && (
            <div className="flex items-center gap-1.5 px-2 py-1 rounded-md bg-violet-400/10 border border-violet-400/20 text-violet-400 text-[9px] font-black uppercase animate-pulse">
              <AlertTriangle size={10} /> Staged
            </div>
          )}
        </div>
      </div>

      <div className="relative group">
        <textarea
          value={content}
          onChange={handleChange}
          spellCheck={false}
          className={cn(
            "w-full h-64 bg-[var(--color-bg)] border border-[var(--color-border)] rounded-2xl p-5 text-[11px] font-mono leading-relaxed text-[var(--color-text)] focus:outline-none focus:border-violet-400 focus:ring-1 focus:ring-violet-400/20 transition-all resize-none shadow-inner",
            isModified && "border-violet-400/50"
          )}
          placeholder="# Modelfile instructions..."
        />
        <div className="absolute top-4 right-4 opacity-30 group-hover:opacity-100 transition-opacity">
          <Info size={14} className="text-[var(--color-text-faint)]" />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="p-4 rounded-xl bg-violet-400/5 border border-violet-400/10">
          <p className="text-[9px] font-black text-violet-400 uppercase tracking-widest mb-1">PRO-TIP: SYSTEM PROMPT</p>
          <p className="text-[10px] text-[var(--color-text-dim)] leading-tight">
            Use the <code>SYSTEM</code> command to define the tone, name, and expertise level of your companion.
          </p>
        </div>
        <div className="p-4 rounded-xl bg-violet-400/5 border border-violet-400/10">
          <p className="text-[9px] font-black text-violet-400 uppercase tracking-widest mb-1">PRO-TIP: PARAMETERS</p>
          <p className="text-[10px] text-[var(--color-text-dim)] leading-tight">
            Set <code>PARAMETER temperature 0.1</code> for factual ops responses or <code>0.7</code> for creative analysis.
          </p>
        </div>
      </div>
    </div>
  )
}
