import { useState, useRef } from 'react'
import {
  Send,
  Trash2,
  Bot,
  User,
  FileText,
  RefreshCw,
  AlertTriangle,
  Activity,
  Clock,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { mockReportContent } from '@/lib/mockData'
import * as Tabs from '@radix-ui/react-tabs'

type TabId = 'ai-chat' | 'reports' | 'anomalies'
type ReportType = 'health' | 'network' | 'security' | 'combined'

// ── Inline helpers ──

function ChatBubble({ role, content }: { role: 'user' | 'assistant'; content: string }) {
  return (
    <div className={cn('flex gap-3 max-w-[80%]', role === 'user' ? 'ml-auto flex-row-reverse' : '')}>
      <div
        className={cn(
          'w-8 h-8 rounded-lg flex items-center justify-center shrink-0',
          role === 'user' ? 'bg-primary/20' : 'bg-[#a78bfa]/20',
        )}
      >
        {role === 'user' ? (
          <User size={14} className="text-primary" />
        ) : (
          <Bot size={14} className="text-[#a78bfa]" />
        )}
      </div>
      <div>
        <div
          className={cn(
            'rounded-2xl px-4 py-3 text-sm',
            role === 'user'
              ? 'bg-primary text-primary-foreground rounded-br-md'
              : 'bg-card border border-border text-text rounded-bl-md',
          )}
        >
          <div className="whitespace-pre-wrap leading-relaxed">{content}</div>
        </div>
      </div>
    </div>
  )
}

function StatusBadge({ status }: { status: string }) {
  const colors: Record<string, string> = {
    critical: 'bg-danger/20 text-danger',
    high: 'bg-warning/20 text-warning',
    medium: 'bg-warning/20 text-warning',
    low: 'bg-primary/20 text-primary',
    warning: 'bg-warning/20 text-warning',
    info: 'bg-primary/20 text-primary',
    healthy: 'bg-success/20 text-success',
    warning2: 'bg-warning/20 text-warning',
    needs_review: 'bg-warning/20 text-warning',
    pass: 'bg-success/20 text-success',
    operational: 'bg-success/20 text-success',
  }
  const key = status.toLowerCase().replace(/\s+/g, '_')
  return (
    <span className={cn('px-2 py-0.5 rounded-full text-xs font-medium', colors[key] || 'bg-muted/20 text-muted')}>
      {status}
    </span>
  )
}

// ── Mock data ──

interface ChatMsg {
  id: string
  role: 'user' | 'assistant'
  content: string
  timestamp: string
}

const initialMessages: ChatMsg[] = [
  {
    id: 'chat-1',
    role: 'user',
    content: 'Can you analyze the current system health and identify any issues?',
    timestamp: new Date(Date.now() - 300000).toISOString(),
  },
  {
    id: 'chat-2',
    role: 'assistant',
    content: "I've analyzed the system metrics. Everything looks operational with no critical issues detected.\n\n**Summary:**\n- CPU: 42% (normal)\n- Memory: 58% (normal)\n- Disk: 55% (normal)\n- Network: 2.1 MB/s\n\nNo immediate action required.",
    timestamp: new Date(Date.now() - 290000).toISOString(),
  },
]

const aiResponses = [
  "I'll analyze that data for you. Based on the current system metrics, everything looks operational with no critical issues detected.",
  "Here's what I found: CPU usage is at 42%, memory at 58%, and disk at 55%. All within normal ranges. The network is performing well with 2.1 MB/s throughput.",
  'Let me check the logs. I can see there was a failed login attempt from 203.0.113.42 about 15 minutes ago. This could be worth investigating further.',
  "Based on the anomaly detection model, I'm seeing normal patterns across all metrics. No statistically significant deviations detected in the last hour.",
  "I've generated a quick summary:\n\n**System Health**: Good\n**Security Events**: 1 warning (failed login)\n**Network**: All interfaces up\n**Recommendation**: No immediate action required.",
]

interface AnomalyItem {
  time: string
  metric: string
  value: number
  expected: number
  deviation: number
  severity: 'critical' | 'high' | 'medium' | 'low'
}

function generateMockAnomalies(): AnomalyItem[] {
  const metrics = ['CPU', 'Memory', 'Network RX', 'Network TX', 'Disk I/O']
  const severities: Array<AnomalyItem['severity']> = ['critical', 'high', 'medium', 'low']
  const now = Date.now()
  return Array.from({ length: 12 }, (_, i) => {
    const severity = severities[Math.floor(Math.random() * severities.length)]
    const base = 30 + Math.random() * 40
    const anomaly = Math.random() < 0.4 ? 50 + Math.random() * 80 : 0
    return {
      time: new Date(now - (12 - i) * 60000 * 5).toLocaleTimeString(),
      metric: metrics[Math.floor(Math.random() * metrics.length)],
      value: base + anomaly,
      expected: base,
      deviation: anomaly > 0 ? anomaly / base * 100 : Math.random() * 5,
      severity: anomaly > 70 ? 'critical' : anomaly > 40 ? 'high' : severity,
    }
  }).filter(a => a.deviation > 3).slice(0, 8)
}

const reportTemplatesList: { id: ReportType; label: string; description: string }[] = [
  { id: 'health', label: 'System Health', description: 'CPU, memory, disk, and process overview' },
  { id: 'network', label: 'Network Audit', description: 'Interfaces, connections, DNS, and ports' },
  { id: 'security', label: 'Security Audit', description: 'Events, firewall, and threat analysis' },
  { id: 'combined', label: 'Combined Report', description: 'Full operations summary across all domains' },
]

// ══════════════════════════════════════════════
//  AIOps Page
// ══════════════════════════════════════════════

export function AIOps() {
  const [activeTab, setActiveTab] = useState<TabId>('ai-chat')

  return (
    <div className="flex h-full">
      <Tabs.Root value={activeTab} onValueChange={(v) => setActiveTab(v as TabId)} className="flex-1 flex flex-col min-w-0">
        <Tabs.List className="flex border-b border-border bg-background">
          <Tabs.Trigger
            value="ai-chat"
            className={cn(
              'flex items-center gap-2 px-4 py-3 text-sm border-b-2 border-transparent transition-colors',
              activeTab === 'ai-chat' ? 'border-primary text-text' : 'text-muted hover:text-text',
            )}
          >
            <Bot size={16} />
            Chat
          </Tabs.Trigger>
          <Tabs.Trigger
            value="reports"
            className={cn(
              'flex items-center gap-2 px-4 py-3 text-sm border-b-2 border-transparent transition-colors',
              activeTab === 'reports' ? 'border-primary text-text' : 'text-muted hover:text-text',
            )}
          >
            <FileText size={16} />
            Reports
          </Tabs.Trigger>
          <Tabs.Trigger
            value="anomalies"
            className={cn(
              'flex items-center gap-2 px-4 py-3 text-sm border-b-2 border-transparent transition-colors',
              activeTab === 'anomalies' ? 'border-primary text-text' : 'text-muted hover:text-text',
            )}
          >
            <Activity size={16} />
            Anomalies
          </Tabs.Trigger>
        </Tabs.List>

        <div className="flex-1 overflow-hidden">
          <Tabs.Content value="ai-chat" className="h-full">
            <ChatTab />
          </Tabs.Content>
          <Tabs.Content value="reports" className="h-full">
            <ReportsTab />
          </Tabs.Content>
          <Tabs.Content value="anomalies" className="h-full">
            <AnomaliesTab />
          </Tabs.Content>
        </div>
      </Tabs.Root>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Chat Tab
// ══════════════════════════════════════════════

function ChatTab() {
  const [messages, setMessages] = useState<ChatMsg[]>(initialMessages)
  const [input, setInput] = useState('')
  const [isTyping, setIsTyping] = useState(false)
  const chatRef = useRef<HTMLDivElement>(null)

  const handleSend = () => {
    if (!input.trim() || isTyping) return

    const userMsg: ChatMsg = {
      id: `user-${Date.now()}`,
      role: 'user',
      content: input,
      timestamp: new Date().toISOString(),
    }
    setMessages(prev => [...prev, userMsg])
    setInput('')
    setIsTyping(true)

    setTimeout(() => {
      const aiMsg: ChatMsg = {
        id: `ai-${Date.now()}`,
        role: 'assistant',
        content: aiResponses[Math.floor(Math.random() * aiResponses.length)],
        timestamp: new Date().toISOString(),
      }
      setMessages(prev => [...prev, aiMsg])
      setIsTyping(false)
    }, 1000 + Math.random() * 1500)
  }

  const clearChat = () => {
    setMessages([])
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="flex h-full">
      <div className="flex-1 flex flex-col min-w-0">
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-border">
          <div className="flex items-center gap-2">
            <Bot size={16} className="text-[#a78bfa]" />
            <span className="text-sm font-medium text-text">AI Assistant</span>
          </div>
          <button
            onClick={clearChat}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs text-muted border border-border rounded-lg hover:bg-sidebar-hover transition-colors"
          >
            <Trash2 size={12} /> Clear Chat
          </button>
        </div>

        {/* Messages */}
        <div ref={chatRef} className="flex-1 overflow-y-auto p-4 space-y-4">
          {messages.length === 0 && (
            <div className="flex flex-col items-center justify-center h-full text-muted">
              <Bot size={40} className="mb-3 opacity-30" />
              <p className="text-sm">Start a conversation with the AI assistant</p>
            </div>
          )}
          {messages.map((msg) => (
            <ChatBubble key={msg.id} role={msg.role} content={msg.content} />
          ))}
          {isTyping && (
            <div className="flex gap-3 max-w-[80%]">
              <div className="w-8 h-8 rounded-lg bg-[#a78bfa]/20 flex items-center justify-center shrink-0">
                <Bot size={14} className="text-[#a78bfa]" />
              </div>
              <div className="bg-card border border-border rounded-2xl rounded-bl-md px-4 py-3">
                <div className="flex items-center gap-1">
                  <span className="w-1.5 h-1.5 bg-[#a78bfa] rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                  <span className="w-1.5 h-1.5 bg-[#a78bfa] rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                  <span className="w-1.5 h-1.5 bg-[#a78bfa] rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Input */}
        <div className="border-t border-border p-4">
          <div className="flex items-center gap-2">
            <textarea
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              placeholder="Type a message... (Enter to send, Shift+Enter for newline)"
              rows={1}
              className="flex-1 bg-[#0f172a] border border-border rounded-lg px-4 py-2.5 text-sm text-text placeholder-muted focus:outline-none focus:border-primary resize-none"
              style={{ minHeight: 40, maxHeight: 120 }}
            />
            <button
              onClick={handleSend}
              disabled={!input.trim() || isTyping}
              className="flex items-center gap-1.5 px-4 py-2.5 text-sm bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              <Send size={14} />
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Reports Tab
// ══════════════════════════════════════════════

function ReportsTab() {
  const [reportType, setReportType] = useState<ReportType>('health')
  const [reportContent, setReportContent] = useState('')
  const [recentReports, setRecentReports] = useState<{ type: ReportType; date: string }[]>([])

  const generateReport = (type: ReportType) => {
    setReportType(type)
    const content = mockReportContent(type)
    setReportContent(content)
    setRecentReports(prev => {
      const updated = [{ type, date: new Date().toLocaleString() }, ...prev]
      return updated.slice(0, 5)
    })
  }

  return (
    <div className="flex h-full p-4 gap-4">
      {/* Left: Template cards */}
      <div className="w-64 shrink-0 space-y-3">
        <h3 className="text-sm font-medium text-text mb-2">Report Templates</h3>
        {reportTemplatesList.map((t) => (
          <div
            key={t.id}
            className={cn(
              'bg-card border rounded-lg p-4 transition-colors',
              reportType === t.id ? 'border-primary/50 bg-primary/5' : 'border-border',
            )}
          >
            <h4 className="text-sm font-medium text-text mb-1">{t.label}</h4>
            <p className="text-xs text-muted mb-3">{t.description}</p>
            <button
              onClick={() => generateReport(t.id)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-xs bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
            >
              <RefreshCw size={12} />
              Generate
            </button>
          </div>
        ))}

        {recentReports.length > 0 && (
          <div className="bg-card border border-border rounded-lg p-4 mt-4">
            <h3 className="text-sm font-medium text-text mb-3">Recent Reports</h3>
            <div className="space-y-1.5">
              {recentReports.map((r, i) => (
                <div key={i} className="flex items-center gap-2 text-xs text-muted">
                  <Clock size={10} />
                  <span className="truncate capitalize">{r.type}</span>
                  <span className="shrink-0">{r.date}</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Right: Report content */}
      <div className="flex-1 flex flex-col min-w-0">
        <div className="flex items-center justify-between mb-3">
          <h3 className="text-sm font-medium text-text capitalize">{reportType} Report</h3>
        </div>
        <div className="flex-1 bg-[#0b1120] border border-border rounded-lg overflow-y-auto p-6 font-[JetBrains_Mono] text-sm leading-relaxed">
          {reportContent ? (
            reportContent.split('\n').map((line, i) => {
              if (line.startsWith('# ')) {
                return <h1 key={i} className="text-xl font-bold text-text mb-4 font-sans">{line.replace('# ', '')}</h1>
              }
              if (line.startsWith('## ')) {
                return <h2 key={i} className="text-lg font-bold text-text mt-6 mb-2 font-sans">{line.replace('## ', '')}</h2>
              }
              if (line.startsWith('| ') && line.endsWith('|')) {
                const isSep = line.includes('---')
                if (isSep) return null
                const cells = line.split('|').filter(c => c.trim()).map(c => c.trim())
                return (
                  <div key={i} className="flex text-sm border-b border-border/30 py-1">
                    {cells.map((c, j) => (
                      <span key={j} className="flex-1 px-2">{c}</span>
                    ))}
                  </div>
                )
              }
              if (line.startsWith('- ')) {
                return <p key={i} className="text-sm text-text ml-4 my-1">• {line.slice(2)}</p>
              }
              if (line.startsWith('> ')) {
                return <p key={i} className="text-sm italic text-muted border-l-2 border-border pl-3 my-2">{line.slice(2)}</p>
              }
              if (line.trim() === '') return <div key={i} className="h-2" />
              return <p key={i} className="text-sm text-text leading-relaxed">{line}</p>
            })
          ) : (
            <div className="flex items-center justify-center h-full text-muted text-sm">
              Select a template and click Generate to create a report
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

// ══════════════════════════════════════════════
//  Anomalies Tab
// ══════════════════════════════════════════════

function AnomaliesTab() {
  const [anomalies, setAnomalies] = useState<AnomalyItem[]>(() => generateMockAnomalies())
  const [metricFilter, setMetricFilter] = useState('ALL')
  const [autoRefresh, setAutoRefresh] = useState(false)

  const metrics = ['ALL', 'CPU', 'Memory', 'Network RX', 'Network TX', 'Disk I/O']

  const filtered = metricFilter === 'ALL'
    ? anomalies
    : anomalies.filter(a => a.metric === metricFilter)

  return (
    <div className="flex flex-col h-full p-4">
      {/* Controls */}
      <div className="flex items-center gap-3 mb-4 flex-wrap">
        <div className="flex items-center gap-2">
          <span className="text-xs text-muted">Metric:</span>
          <select
            value={metricFilter}
            onChange={(e) => setMetricFilter(e.target.value)}
            className="bg-[#0f172a] border border-border rounded-lg px-3 py-1.5 text-sm text-text focus:outline-none focus:border-primary"
          >
            {metrics.map(m => (
              <option key={m} value={m}>{m}</option>
            ))}
          </select>
        </div>
        {autoRefresh && (
          <span className="text-xs text-primary animate-pulse">● Live</span>
        )}
        <div className="flex-1" />
        <label className="flex items-center gap-2 text-xs text-muted cursor-pointer">
          <input
            type="checkbox"
            checked={autoRefresh}
            onChange={(e) => setAutoRefresh(e.target.checked)}
            className="rounded border-border bg-[#0f172a]"
          />
          Auto-refresh
        </label>
        <button
          onClick={() => setAnomalies(generateMockAnomalies())}
          className="flex items-center gap-1.5 px-3 py-1.5 text-xs text-muted border border-border rounded-lg hover:bg-sidebar-hover transition-colors"
        >
          <RefreshCw size={12} /> Refresh
        </button>
      </div>

      {/* Summary */}
      <div className="flex items-center gap-4 mb-3 text-xs">
        <span className="text-muted">{filtered.length} anomalies</span>
        {['critical', 'high', 'medium', 'low'].map(sev => {
          const count = filtered.filter(a => a.severity === sev).length
          if (count === 0) return null
          return (
            <span key={sev} className="flex items-center gap-1">
              <span className={cn(
                'w-1.5 h-1.5 rounded-full',
                sev === 'critical' ? 'bg-danger' : sev === 'high' || sev === 'medium' ? 'bg-warning' : 'bg-primary',
              )} />
              <span className={cn(
                'font-medium',
                sev === 'critical' ? 'text-danger' : sev === 'high' || sev === 'medium' ? 'text-warning' : 'text-primary',
              )}>
                {count}
              </span>
              <span className="text-muted">{sev}</span>
            </span>
          )
        })}
      </div>

      {/* Table */}
      <div className="flex-1 bg-[#0b1120] border border-border rounded-lg overflow-y-auto min-h-0">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-border text-xs text-muted">
              <th className="text-left px-4 py-2.5 font-medium">Timestamp</th>
              <th className="text-left px-4 py-2.5 font-medium">Metric</th>
              <th className="text-right px-4 py-2.5 font-medium">Value</th>
              <th className="text-right px-4 py-2.5 font-medium">Expected</th>
              <th className="text-right px-4 py-2.5 font-medium">Deviation</th>
              <th className="text-left px-4 py-2.5 font-medium">Severity</th>
            </tr>
          </thead>
          <tbody>
            {filtered.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-12 text-center text-muted text-sm">
                  <div className="flex flex-col items-center gap-2">
                    <AlertTriangle size={20} className="opacity-30" />
                    No anomalies detected
                  </div>
                </td>
              </tr>
            )}
            {filtered.map((a, i) => (
              <tr key={i} className="border-b border-border/30 hover:bg-white/5 transition-colors">
                <td className="px-4 py-2.5 text-muted text-xs font-[JetBrains_Mono]">{a.time}</td>
                <td className="px-4 py-2.5 text-text">{a.metric}</td>
                <td className="px-4 py-2.5 text-right font-[JetBrains_Mono] text-xs text-text">{a.value.toFixed(1)}</td>
                <td className="px-4 py-2.5 text-right font-[JetBrains_Mono] text-xs text-muted">{a.expected.toFixed(1)}</td>
                <td className="px-4 py-2.5 text-right font-[JetBrains_Mono] text-xs">
                  <span className={cn(
                    a.severity === 'critical' ? 'text-danger' : 'text-warning',
                  )}>
                    +{a.deviation.toFixed(1)}%
                  </span>
                </td>
                <td className="px-4 py-2.5">
                  <StatusBadge status={a.severity} />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
