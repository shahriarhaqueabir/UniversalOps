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
  ChevronLeft,
  ChevronRight,
} from 'lucide-react'
import type { Page } from '../../App'
import { cn } from '../../lib/utils'

interface NavItem {
  id: Page
  label: string
  icon: React.ReactNode
}

const opsItems: NavItem[] = [
  { id: 'dashboard', label: 'Dashboard', icon: <LayoutDashboard size={22} /> },
  { id: 'sysops', label: 'System Ops', icon: <Monitor size={22} /> },
  { id: 'netops', label: 'Network Ops', icon: <Network size={22} /> },
  { id: 'secops', label: 'Security Ops', icon: <Shield size={22} /> },
  { id: 'devops', label: 'DevOps', icon: <Terminal size={22} /> },
  { id: 'aiops', label: 'AI Ops', icon: <Bot size={22} /> },
]

const toolsItems: NavItem[] = [
  { id: 'network-design', label: 'Network Design', icon: <GitBranch size={22} /> },
  { id: 'logs', label: 'Logs', icon: <ScrollText size={22} /> },
  { id: 'settings', label: 'Settings', icon: <Settings size={22} /> },
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
        collapsed ? 'w-[80px]' : 'w-[280px]',
        'border-[var(--color-border)] bg-[var(--color-sidebar)]'
      )}
    >
      {/* Brand section */}
      <div
        className={cn(
          'flex items-center border-b border-[var(--color-border)]',
          'gap-4 px-6 py-6'
        )}
      >
        {/* Gradient icon with glow */}
        <div className="relative flex-shrink-0">
          <div
            className="w-10 h-10 rounded-xl flex items-center justify-center"
            style={{
              background: 'linear-gradient(135deg, #7c6cff, #a78bff)',
              boxShadow: '0 0 16px rgba(124, 108, 255, 0.45)',
            }}
          >
            <span className="text-[var(--color-bg)] font-bold text-lg">H</span>
          </div>
        </div>

        {!collapsed && (
          <div className="flex flex-col min-w-0 flex-1">
            <span className="font-bold text-lg tracking-wider text-[var(--color-text)] truncate">
              HAWKWARD
            </span>
            <span className="text-xs text-[var(--color-text-faint)] truncate leading-tight uppercase font-medium">
              Ops Platform
            </span>
          </div>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto px-4 py-6 space-y-8">
        {/* Operations section */}
        <ul className={cn('space-y-2')}>
          {opsItems.map((item) => (
            <li key={item.id}>
              <button
                onClick={() => onNavigate(item.id)}
                className={cn(
                  'flex items-center w-full rounded-xl transition-all font-bold',
                  collapsed ? 'justify-center p-3' : 'gap-4 px-4 py-3 text-lg',
                  currentPage === item.id
                    ? 'bg-accent text-white shadow-lg'
                    : 'text-[var(--color-text-dim)] hover:bg-[var(--color-sidebar-hover)] hover:text-[var(--color-text)]'
                )}
                title={collapsed ? item.label : undefined}
              >
                <span className="flex-shrink-0">{item.icon}</span>
                {!collapsed && <span className="truncate">{item.label}</span>}
              </button>
            </li>
          ))}
        </ul>

        <div className="h-px bg-border mx-2 opacity-50" />

        {/* Tools section */}
        <ul className={cn('space-y-2')}>
          {toolsItems.map((item) => (
            <li key={item.id}>
              <button
                onClick={() => onNavigate(item.id)}
                className={cn(
                  'flex items-center w-full rounded-xl transition-all font-bold',
                  collapsed ? 'justify-center p-3' : 'gap-4 px-4 py-3 text-lg',
                  currentPage === item.id
                    ? 'bg-accent text-white shadow-lg'
                    : 'text-[var(--color-text-dim)] hover:bg-[var(--color-sidebar-hover)] hover:text-[var(--color-text)]'
                )}
                title={collapsed ? item.label : undefined}
              >
                <span className="flex-shrink-0">{item.icon}</span>
                {!collapsed && <span className="truncate">{item.label}</span>}
              </button>
            </li>
          ))}
        </ul>
      </nav>

      {/* Collapse toggle */}
      <div className="px-4 py-3 border-t border-[var(--color-border)]">
        <button
          onClick={() => setCollapsed(!collapsed)}
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          className="flex items-center justify-center w-full gap-2 px-4 py-3 rounded-xl text-[var(--color-text-dim)] hover:bg-[var(--color-sidebar-hover)] hover:text-[var(--color-text)] transition-all font-bold"
        >
          {collapsed ? <ChevronRight size={20} /> : <ChevronLeft size={20} />}
          {!collapsed && <span className="text-sm">Collapse</span>}
        </button>
      </div>

      {/* Footer */}
      <div className="px-[14px] py-3 border-t border-[var(--color-border)]">
        <p className="text-[11px] text-[var(--color-text-faint)]">Hawkward v1.3.0</p>
      </div>
    </aside>
  )
}
