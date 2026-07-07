import { useState, useEffect } from 'react'
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

// ── Section Card ──

function Section({ title, icon, children }: { title: string; icon: React.ReactNode; children: React.ReactNode }) {
  return (
    <div className="bg-card border border-border rounded-lg">
      <div className="flex items-center gap-2 px-4 py-3 border-b border-border">
        <span className="text-primary">{icon}</span>
        <h2 className="text-sm font-semibold text-text">{title}</h2>
      </div>
      <div className="p-4 space-y-4">{children}</div>
    </div>
  )
}

// ── Setting Row ──

function SettingRow({ label, description, children }: { label: string; description?: string; children: React.ReactNode }) {
  return (
    <div className="flex items-center justify-between gap-4">
      <div className="min-w-0">
        <p className="text-sm text-text">{label}</p>
        {description && <p className="text-xs text-muted mt-0.5">{description}</p>}
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
  goVersion: string
  uptime: string
}

const DEFAULT_APP_INFO: AppInfo = {
  name: 'Hawkward',
  version: '0.1.0',
  goVersion: 'go1.26.4',
  uptime: '--',
}

export function Settings() {
  const { call } = useBackend()

  // Theme
  const [darkMode, setDarkMode] = useState(true)

  // Collection / Refresh
  const [refreshInterval, setRefreshInterval] = useState(5000)

  // Network
  const [pingCount, setPingCount] = useState(4)
  const [dnsTimeout, setDnsTimeout] = useState(2000)

  // About
  const [appInfo, setAppInfo] = useState<AppInfo>(DEFAULT_APP_INFO)

  useEffect(() => {
    call('GetAppInfo').then((result: unknown) => {
      if (result) {
        const info = result as Partial<AppInfo>
        setAppInfo({
          name: info.name || DEFAULT_APP_INFO.name,
          version: info.version || DEFAULT_APP_INFO.version,
          goVersion: info.goVersion || DEFAULT_APP_INFO.goVersion,
          uptime: info.uptime || DEFAULT_APP_INFO.uptime,
        })
      }
    })
  }, [call])

  return (
    <div className="h-full overflow-y-auto p-6 max-w-3xl space-y-6">
      <h1 className="text-xl font-bold text-text">Settings</h1>

      {/* ── Theme ── */}
      <Section title="Theme" icon={<Monitor size={16} />}>
        <SettingRow label="Appearance" description="Switch between dark and light theme">
          <div className="flex items-center rounded-lg overflow-hidden border border-border">
            <button
              onClick={() => setDarkMode(true)}
              className={cn(
                'flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium transition-colors',
                darkMode
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-background text-muted hover:text-text',
              )}
            >
              <Moon size={13} /> Dark
            </button>
            <button
              onClick={() => setDarkMode(false)}
              className={cn(
                'flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium transition-colors',
                !darkMode
                  ? 'bg-primary text-primary-foreground'
                  : 'bg-background text-muted hover:text-text',
              )}
            >
              <Sun size={13} /> Light
            </button>
          </div>
        </SettingRow>
      </Section>

      {/* ── Collection Interval ── */}
      <Section title="Collection" icon={<Activity size={16} />}>
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
            className="bg-background border border-border rounded-lg px-3 py-1.5 text-sm text-text focus:outline-none focus:ring-1 focus:ring-primary"
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
      <Section title="Network" icon={<Network size={16} />}>
        <SettingRow
          label="Default Ping Count"
          description="Number of echo requests per ping (1–20)"
        >
          <div className="flex items-center gap-3 w-44">
            <Slider.Root
              value={[pingCount]}
              onValueChange={([v]) => setPingCount(v)}
              min={1}
              max={20}
              step={1}
              className="relative flex items-center flex-1 h-5 cursor-pointer"
            >
              <Slider.Track className="relative h-1.5 flex-1 rounded-full bg-background border border-border">
                <Slider.Range className="absolute h-full rounded-full bg-primary" />
              </Slider.Track>
              <Slider.Thumb className="block w-4 h-4 bg-text rounded-full shadow focus:outline-none focus:ring-2 focus:ring-primary" />
            </Slider.Root>
            <span className="text-xs font-mono text-text w-5 text-right">{pingCount}</span>
          </div>
        </SettingRow>

        <SettingRow
          label="DNS Timeout"
          description="Timeout in milliseconds for DNS lookups"
        >
          <div className="flex items-center gap-3 w-44">
            <Slider.Root
              value={[dnsTimeout]}
              onValueChange={([v]) => setDnsTimeout(v)}
              min={500}
              max={10000}
              step={100}
              className="relative flex items-center flex-1 h-5 cursor-pointer"
            >
              <Slider.Track className="relative h-1.5 flex-1 rounded-full bg-background border border-border">
                <Slider.Range className="absolute h-full rounded-full bg-primary" />
              </Slider.Track>
              <Slider.Thumb className="block w-4 h-4 bg-text rounded-full shadow focus:outline-none focus:ring-2 focus:ring-primary" />
            </Slider.Root>
            <span className="text-xs font-mono text-text w-10 text-right">{dnsTimeout}ms</span>
          </div>
        </SettingRow>
      </Section>

      {/* ── About ── */}
      <Section title="About" icon={<Info size={16} />}>
        <div className="grid grid-cols-2 gap-4 text-sm">
          <div>
            <p className="text-xs text-muted mb-0.5">Application</p>
            <p className="text-text font-medium">{appInfo.name}</p>
          </div>
          <div>
            <p className="text-xs text-muted mb-0.5">Version</p>
            <p className="text-text font-mono">{appInfo.version}</p>
          </div>
          <div>
            <p className="text-xs text-muted mb-0.5">Go Version</p>
            <p className="text-text font-mono">{appInfo.goVersion}</p>
          </div>
          <div>
            <p className="text-xs text-muted mb-0.5">Uptime</p>
            <p className="text-text font-mono">{appInfo.uptime}</p>
          </div>
        </div>
        <div className="pt-4 border-t border-border mt-4">
          <div className="flex items-center gap-4">
            <a href="#" className="flex items-center gap-1 text-xs text-primary hover:underline">
              <ExternalLink size={12} /> GitHub
            </a>
            <a href="#" className="flex items-center gap-1 text-xs text-primary hover:underline">
              <ExternalLink size={12} /> Documentation
            </a>
          </div>
        </div>
      </Section>
    </div>
  )
}
