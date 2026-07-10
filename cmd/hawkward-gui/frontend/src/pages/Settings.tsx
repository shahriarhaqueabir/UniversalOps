import { useQuery } from '@tanstack/react-query'
import {
  Monitor,
  Activity,
  Network,
  Info,
  Moon,
  Sun,
  ExternalLink,
  RotateCcw,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import * as Slider from '@radix-ui/react-slider'
import { useBackend } from '@/hooks/useBackend'
import { useThemeStore, useSettingsStore } from '@/stores/useSettingsStore'
import { toast } from 'sonner'

// ── Section Card (aligned with Squib design system) ──

function Section({ title, icon, children }: { title: string; icon: React.ReactNode; children: React.ReactNode }) {
  return (
    <div className="bg-[var(--color-panel)] border border-[var(--color-border)] rounded-[var(--radius-lg)] p-6 shadow-lg">
      <div className="flex items-center gap-3 mb-4 pb-3 border-b border-[var(--color-border)]/50">
        <div className="w-8 h-8 rounded-lg bg-[var(--color-accent-soft)] flex items-center justify-center text-[var(--color-accent)]">
          {icon}
        </div>
        <h2 className="text-xs font-semibold text-[var(--color-text)] uppercase tracking-wider">{title}</h2>
      </div>
      <div className="space-y-4">{children}</div>
    </div>
  )
}

// ── Setting Row ──

function SettingRow({ label, description, children }: { label: string; description?: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-4">
      <div className="min-w-0">
        <p className="text-sm font-medium text-[var(--color-text)]">{label}</p>
        {description && <p className="text-xs text-[var(--color-text-faint)] mt-0.5">{description}</p>}
      </div>
      <div className="shrink-0">{children}</div>
    </div>
  )
}

// ══════════════════════════════════
//  Settings Page
// ══════════════════════════════════

const intervalOptions = [
  { value: 1000, label: '1s' },
  { value: 3000, label: '3s' },
  { value: 5000, label: '5s' },
  { value: 10000, label: '10s' },
  { value: 30000, label: '30s' },
]

const DEFAULT_SETTINGS = {
  refreshInterval: 5000,
  pingCount: 4,
  dnsTimeout: 2000,
}

interface AppInfo {
  name: string
  version: string
  go_version: string
  uptime: string
}

const DEFAULT_APP_INFO: AppInfo = {
  name: 'Hawkward',
  version: '1.3.0',
  go_version: 'go1.26.5',
  uptime: '--',
}

export function Settings() {
  const { call } = useBackend()

  // Theme — from zustand store
  const { theme, toggle } = useThemeStore()

  // Settings — from zustand store (auto-persisted to localStorage)
  const { refreshInterval, pingCount, dnsTimeout, setRefreshInterval, setPingCount, setDnsTimeout } = useSettingsStore()

  // About — via react-query
  const { data: appInfo = DEFAULT_APP_INFO } = useQuery<AppInfo>({
    queryKey: ['app-info'],
    queryFn: async () => {
      try {
        const result = await call('App.GetAppInfo') as Partial<AppInfo>
        if (result) {
          return {
            name: result.name || DEFAULT_APP_INFO.name,
            version: result.version || DEFAULT_APP_INFO.version,
            go_version: result.go_version || DEFAULT_APP_INFO.go_version,
            uptime: result.uptime || DEFAULT_APP_INFO.uptime,
          }
        }
      } catch {
        console.warn('[Settings] Failed to fetch AppInfo, using defaults')
      }
      return DEFAULT_APP_INFO
    },
    staleTime: 60000,
  })

  const isDark = theme === 'dark'

  return (
    <div className="h-full overflow-y-auto p-6 space-y-5 max-w-3xl">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-[var(--color-text)] flex items-center gap-3">
          <Monitor size={24} className="text-[var(--color-accent)]" /> Settings
        </h1>
        <p className="text-sm text-[var(--color-text-dim)] mt-1">
          Configure theme, collection intervals, and network parameters.
        </p>
      </div>

      {/* ── Theme ── */}
      <Section title="Theme" icon={<Monitor size={20} />}>
        <SettingRow label="Appearance" description="Switch between dark and light theme">
          <div className="flex items-center rounded-xl overflow-hidden border border-border">
            <button
              onClick={() => { if (!isDark) toggle() }}
              className={cn(
                'flex items-center gap-2 px-4 py-2 text-sm font-bold transition-all',
                isDark
                  ? 'bg-accent text-white shadow-lg'
                  : 'bg-panel-2 text-text-dim hover:text-text',
              )}
            >
              <Moon size={16} /> Dark
            </button>
            <button
              onClick={() => { if (isDark) toggle() }}
              className={cn(
                'flex items-center gap-2 px-4 py-2 text-sm font-bold transition-all',
                !isDark
                  ? 'bg-accent text-white shadow-lg'
                  : 'bg-panel-2 text-text-dim hover:text-text',
              )}
            >
              <Sun size={16} /> Light
            </button>
          </div>
        </SettingRow>
      </Section>

      {/* ── Collection Interval ── */}
      <Section title="Collection" icon={<Activity size={20} />}>
        <SettingRow
          label="Refresh Interval"
          description="How often the dashboard and metrics refresh"
        >
          <select
            value={refreshInterval}
            onChange={(e) => {
              const val = Number(e.target.value)
              setRefreshInterval(val)
              call('PipelineAPI.UpdateSettings', val, 0)
              toast.success(`Refresh interval set to ${val / 1000}s`)
            }}
            className="bg-[var(--color-panel-2)] border border-[var(--color-border)] rounded-lg px-3 py-2 text-sm font-medium text-[var(--color-text)] focus:outline-none focus:border-[var(--color-accent)] transition-colors"
          >
            {intervalOptions.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </SettingRow>
      </Section>

      {/* ── Network ── */}
      <Section title="Network" icon={<Network size={20} />}>
        <SettingRow
          label="Default Ping Count"
          description="Number of echo requests per ping (1–20)"
        >
          <div className="flex items-center gap-3 w-44">
            <Slider.Root
              value={[pingCount]}
              onValueChange={([v]: number[]) => { setPingCount(v); toast.success(`Ping count set to ${v}`) }}
              min={1}
              max={20}
              step={1}
              className="relative flex items-center flex-1 h-5 cursor-pointer"
            >
              <Slider.Track className="relative h-1.5 flex-1 rounded-full bg-panel-2 border border-border">
                <Slider.Range className="absolute h-full rounded-full bg-accent" />
              </Slider.Track>
              <Slider.Thumb className="block w-4 h-4 bg-text rounded-full shadow-lg focus:outline-none focus:ring-2 focus:ring-accent" />
            </Slider.Root>
            <span className="text-xs font-semibold font-[Geist_Mono] text-[var(--color-text)] w-6 text-right tabular-nums">{pingCount}</span>
          </div>
        </SettingRow>

        <SettingRow
          label="DNS Timeout"
          description="Timeout in milliseconds for DNS lookups"
        >
          <div className="flex items-center gap-3 w-44">
            <Slider.Root
              value={[dnsTimeout]}
              onValueChange={([v]: number[]) => { setDnsTimeout(v); toast.success(`DNS timeout set to ${v}ms`) }}
              min={500}
              max={10000}
              step={100}
              className="relative flex items-center flex-1 h-5 cursor-pointer"
            >
              <Slider.Track className="relative h-1.5 flex-1 rounded-full bg-panel-2 border border-border">
                <Slider.Range className="absolute h-full rounded-full bg-accent" />
              </Slider.Track>
              <Slider.Thumb className="block w-4 h-4 bg-text rounded-full shadow-lg focus:outline-none focus:ring-2 focus:ring-accent" />
            </Slider.Root>
            <span className="text-xs font-semibold font-[Geist_Mono] text-[var(--color-text)] w-14 text-right tabular-nums">{dnsTimeout}ms</span>
          </div>
        </SettingRow>
      </Section>

      {/* ── Reset ── */}
      <Section title="Management" icon={<Monitor size={20} />}>
        <div className="space-y-4">
          <p className="text-sm text-text-faint">
            Reset all settings to their factory defaults.
          </p>
          <button
            onClick={() => {
              setRefreshInterval(DEFAULT_SETTINGS.refreshInterval)
              setPingCount(DEFAULT_SETTINGS.pingCount)
              setDnsTimeout(DEFAULT_SETTINGS.dnsTimeout)
              // Overwrite localStorage directly
              localStorage.setItem('hawkward_refreshInterval', JSON.stringify(DEFAULT_SETTINGS.refreshInterval))
              localStorage.setItem('hawkward_pingCount', JSON.stringify(DEFAULT_SETTINGS.pingCount))
              localStorage.setItem('hawkward_dnsTimeout', JSON.stringify(DEFAULT_SETTINGS.dnsTimeout))
              // Also push the new interval to backend
              call('PipelineAPI.UpdateSettings', DEFAULT_SETTINGS.refreshInterval, 0)
              toast.success('All settings reset to defaults')
            }}
            className="px-5 py-2.5 bg-[var(--color-danger)]/10 hover:bg-[var(--color-danger)]/20 text-[var(--color-danger)] text-sm font-semibold rounded-lg border border-[var(--color-danger)]/30 hover:border-[var(--color-danger)]/50 transition-all"
          >
            <RotateCcw size={16} className="inline mr-2" />
            Reset to Defaults
          </button>
        </div>
      </Section>

      {/* ── About ── */}
      <Section title="About" icon={<Info size={20} />}>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <p className="text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-wider mb-1">Application</p>
            <p className="text-[var(--color-text)] font-medium">{appInfo.name}</p>
          </div>
          <div>
            <p className="text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-wider mb-1">Version</p>
            <p className="text-[var(--color-text)] font-medium font-[Geist_Mono]">{appInfo.version}</p>
          </div>
          <div>
            <p className="text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-wider mb-1">Go Version</p>
            <p className="text-[var(--color-text)] font-medium font-[Geist_Mono]">{appInfo.go_version}</p>
          </div>
          <div>
            <p className="text-xs font-semibold text-[var(--color-text-faint)] uppercase tracking-wider mb-1">Uptime</p>
            <p className="text-[var(--color-text)] font-medium font-[Geist_Mono]">{appInfo.uptime}</p>
          </div>
        </div>
        <div className="pt-6 border-t border-border/50 mt-6">
          <div className="flex items-center gap-6">
            <a href="https://github.com/shahriarhaqueabir/AllOpsFull" target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 text-sm font-bold text-accent hover:text-accent-2 transition-colors">
              <ExternalLink size={16} /> GitHub
            </a>
            <a href="https://github.com/shahriarhaqueabir/AllOpsFull#readme" target="_blank" rel="noopener noreferrer" className="flex items-center gap-2 text-sm font-bold text-accent hover:text-accent-2 transition-colors">
              <ExternalLink size={16} /> Documentation
            </a>
          </div>
        </div>
      </Section>
    </div>
  )
}
