import { useState } from 'react'
import { useConfigStore } from '@/stores/useConfigStore'
import { Rocket, ChevronRight } from 'lucide-react'
import { ReviewModal } from './ReviewModal'

/**
 * DeploymentBar — A sticky footer that appears when there are pending settings changes.
 */
export function DeploymentBar() {
  const { stagedChanges } = useConfigStore()
  const [showReview, setShowReview] = useState(false)

  if (stagedChanges.size === 0) return null

  return (
    <>
      <div className="fixed bottom-8 left-1/2 -translate-x-1/2 z-30 animate-in slide-in-from-bottom-8 duration-500">
        <div className="bg-[var(--color-accent)] text-white px-6 py-4 rounded-[2rem] shadow-[0_20px_50px_rgba(var(--color-accent-rgb),0.3)] flex items-center gap-8 border border-white/20 backdrop-blur-md">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 rounded-full bg-white/20 flex items-center justify-center">
              <Rocket size={16} />
            </div>
            <div>
              <p className="text-[10px] font-black uppercase tracking-widest leading-none opacity-80">
                Shadow State Active
              </p>
              <p className="text-sm font-bold mt-0.5">
                {stagedChanges.size} {stagedChanges.size === 1 ? 'Change' : 'Changes'} Pending
              </p>
            </div>
          </div>

          <button
            onClick={() => setShowReview(true)}
            className="group flex items-center gap-2 bg-white text-[var(--color-accent)] px-5 py-2.5 rounded-full text-xs font-black uppercase tracking-widest hover:scale-105 transition-all active:scale-95"
          >
            Review & Deploy
            <ChevronRight size={14} className="group-hover:translate-x-0.5 transition-transform" />
          </button>
        </div>
      </div>

      <ReviewModal isOpen={showReview} onOpenChange={setShowReview} />
    </>
  )
}
