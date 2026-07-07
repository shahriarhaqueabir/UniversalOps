import { useState } from 'react'
import {
  LayoutDashboard,
  Monitor,
  Network,
  Shield,
  Terminal,
  Bot,
  GitBranch,
  ScrollText,
  Settings,
  PanelLeftClose,
  PanelLeft,
} from 'lucide-react'
import type { Page } from '../../App'
import { cn } from '../../lib/utils'

interface NavItem {
  id: Page
  label: string
  icon: React.ReactNode
}

const opsItems: NavItem[] = [
  { id: 'dashboard', label: 'Dashboard', icon: <LayoutDashboard size={18} /> },
  { id: 'sysops', label: 'System Ops', icon: <Monitor size={18} /> },
  { id: 'netops', label: 'Network Ops', icon: <Network size={18} /> },
  { id: 'secops', label: 'Security Ops', icon: <Shield size={18} /> },
  { id: 'devops', label: 'DevOps', icon: <Terminal size={18} /> },
  { id: 'aiops', label: 'AI Ops', icon: <Bot size={18} /> },
]

const toolsItems: NavItem[] = [
  { id: 'network-design', label: 'Network Design', icon: <GitBranch size={18} /> },
  { id: 'logs', label: 'Logs', icon: <ScrollText size={18} /> },
  { id: 'settings', label: 'Settings', icon: <Settings size={18} /> },
]

interface SidebarProps {
  currentPage: Page
  onNavigate: (page: Page) => void
}

export function Sidebar({ currentPage, onNavigate }: SidebarProps) {
  const [collapsed, setCollapsed] = useState(false)

  return (
    <aside
      className={cn(
        'flex flex-col transition-all duration-200 border-r',
        collapsed ? 'w-16' : 'w-[252px]',
        'border-[var(--color-border)] bg-[var(--color-sidebar)]'
      )}
    >
      {/* Brand section */}
      <div
        className={cn(
          'flex items-center border-b border-[var(--color-border)]',
          collapsed ? 'justify-center px-3 py-4' : 'gap-3 px-4 py-[14px]'
        )}
      >
        {/* Gradient icon with glow */}
        <div className="relative flex-shrink-0">
          <div
            className="w-[34px] h-[34px] rounded-xl flex items-center justify-center"
            style={{
              background: 'linear-gradient(135deg, #7c6cff, #a78bff)',
              boxShadow: '0 0 14px rgba(124, 108, 255, 0.45)',
            }}
          >
            <span className="text-[var(--color-bg)] font-bold text-sm">H</span>
          </div>
        </div>

        {!collapsed && (
          <div className="flex flex-col min-w-0 flex-1">
            <span className="font-bold text-sm tracking-wide text-[var(--color-text)] truncate">
              HAWKWARD
            </span>
            <span className="text-[10.5px] text-[var(--color-text-faint)] truncate leading-tight">
              Operations Platform
            </span>
          </div>
        )}

        {/* Collapse toggle */}
        {!collapsed && (
          <button
            onClick={() => setCollapsed(true)}
            className="flex-shrink-0 text-[var(--color-text-faint)] hover:text-[var(--color-text)] transition-colors"
            aria-label="Collapse sidebar"
          >
            <PanelLeftClose size={16} />
          </button>
        )}
        {collapsed && (
          <button
            onClick={() => setCollapsed(false)}
            className="text-[var(--color-text-faint)] hover:text-[var(--color-text)] transition-colors"
            aria-label="Expand sidebar"
          >
            <PanelLeft size={16} />
          </button>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto px-3 py-3 space-y-5">
        {/* Operations section */}
        {!collapsed && (
          <div className="text-[10.5px] uppercase tracking-[0.08em] text-[var(--color-text-faint)] px-[6px] mb-2">
            Operations
          </div>
        )}
        <ul className={cn('space-y-[2px]', collapsed && 'space-y-[3px]')}>
          {opsItems.map((item) => (
            <li key={item.id}>
              <button
                onClick={() => onNavigate(item.id)}
                className={cn(
                  'flex items-center w-full rounded-lg transition-colors',
                  collapsed ? 'justify-center p-[10px]' : 'gap-[11px] px-[10px] py-[8.5px]',
                  currentPage === item.id
                    ? 'bg-[var(--color-accent-soft)] text-[var(--color-accent-2)]'
                    : 'text-[var(--color-text-dim)] hover:bg-[var(--color-panel-2)] hover:text-[var(--color-text)]'
                )}
                title={collapsed ? item.label : undefined}
              >
                <span className="flex-shrink-0">{item.icon}</span>
                {!collapsed && <span className="text-sm truncate">{item.label}</span>}
              </button>
            </li>
          ))}
        </ul>

        {/* Tools section */}
        {!collapsed && (
          <div className="text-[10.5px] uppercase tracking-[0.08em] text-[var(--color-text-faint)] px-[6px] mb-2">
            Tools
          </div>
        )}
        <ul className={cn('space-y-[2px]', collapsed && 'space-y-[3px]')}>
          {toolsItems.map((item) => (
            <li key={item.id}>
              <button
                onClick={() => onNavigate(item.id)}
                className={cn(
                  'flex items-center w-full rounded-lg transition-colors',
                  collapsed ? 'justify-center p-[10px]' : 'gap-[11px] px-[10px] py-[8.5px]',
                  currentPage === item.id
                    ? 'bg-[var(--color-accent-soft)] text-[var(--color-accent-2)]'
                    : 'text-[var(--color-text-dim)] hover:bg-[var(--color-panel-2)] hover:text-[var(--color-text)]'
                )}
                title={collapsed ? item.label : undefined}
              >
                <span className="flex-shrink-0">{item.icon}</span>
                {!collapsed && <span className="text-sm truncate">{item.label}</span>}
              </button>
            </li>
          ))}
        </ul>
      </nav>

      {/* Footer */}
      {!collapsed && (
        <div className="px-[14px] py-3 border-t border-[var(--color-border)]">
          <p className="text-[11px] text-[var(--color-text-faint)]">Hawkward v0.1.0</p>
        </div>
      )}
    </aside>
  )
}
