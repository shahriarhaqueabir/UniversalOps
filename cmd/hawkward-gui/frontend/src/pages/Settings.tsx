import { useQuery } from '@tanstack/react-query'
import {
  Monitor,
  Activity,
  Network,
  Info,
  Moon,
  Sun,
  ExternalLink,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import * as Slider from '@radix-ui/react-slider'
import { useBackend } from '@/hooks/useBackend'
import { useThemeStore, useSettingsStore } from '@/stores/useSettingsStore'

// ── Section Card (aligned with Squib design system) ──

function Section({ title, icon, children }: { title: string; icon: React.ReactNode; children: React.ReactNode }) {
  return (
    <div className="bg-panel border border-border rounded-[24px] p-8 shadow-xl">
      <div className="flex items-center gap-4 mb-6 pb-4 border-b border-border/50">
        <div className="w-10 h-10 rounded-xl bg-accent-soft flex items-center justify-center text-accent">
          {icon}
        </div>
        <h2 className="text-sm font-black text-text uppercase tracking-[0.2em]">{title}</h2>
      </div>
      <div className="space-y-6">{children}</div>
    </div>
  )
}

// ── Setting Row ──

function SettingRow({ label, description, children }: { label: string; description?: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-4">
      <div className="min-w-0">
        <p className="text-base font-bold text-text">{label}</p>
        {description && <p className="text-sm text-text-faint mt-1">{description}</p>}
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
    <div className="h-full overflow-y-auto p-8 space-y-6 max-w-3xl">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-text flex items-center gap-4">
          <Monitor size={32} className="text-accent" /> Settings
        </h1>
        <p className="text-lg text-text-dim mt-2 font-medium">
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
            }}
            className="bg-panel-2 border border-border rounded-xl px-4 py-2.5 text-base font-bold text-text focus:outline-none focus:border-accent transition-colors"
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
              onValueChange={([v]: number[]) => { setPingCount(v) }}
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
            <span className="text-sm font-bold font-[JetBrains_Mono] text-text w-6 text-right">{pingCount}</span>
          </div>
        </SettingRow>

        <SettingRow
          label="DNS Timeout"
          description="Timeout in milliseconds for DNS lookups"
        >
          <div className="flex items-center gap-3 w-44">
            <Slider.Root
              value={[dnsTimeout]}
              onValueChange={([v]: number[]) => { setDnsTimeout(v) }}
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
            <span className="text-sm font-bold font-[JetBrains_Mono] text-text w-14 text-right">{dnsTimeout}ms</span>
          </div>
        </SettingRow>
      </Section>

      {/* ── About ── */}
      <Section title="About" icon={<Info size={20} />}>
        <div className="grid grid-cols-2 gap-6 text-base">
          <div>
            <p className="text-xs font-black text-text-faint uppercase tracking-[0.2em] mb-1">Application</p>
            <p className="text-text font-bold">{appInfo.name}</p>
          </div>
          <div>
            <p className="text-xs font-black text-text-faint uppercase tracking-[0.2em] mb-1">Version</p>
            <p className="text-text font-bold font-[JetBrains_Mono]">{appInfo.version}</p>
          </div>
          <div>
            <p className="text-xs font-black text-text-faint uppercase tracking-[0.2em] mb-1">Go Version</p>
            <p className="text-text font-bold font-[JetBrains_Mono]">{appInfo.go_version}</p>
          </div>
          <div>
            <p className="text-xs font-black text-text-faint uppercase tracking-[0.2em] mb-1">Uptime</p>
            <p className="text-text font-bold font-[JetBrains_Mono]">{appInfo.uptime}</p>
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
