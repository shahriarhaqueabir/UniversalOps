import { useState } from 'react'
import { motion } from 'motion/react'
import {
  LayoutDashboard,
  Monitor,
  Network,
  Shield,
  Terminal,
  Bot,
  FileText,
  Bell,
  ScrollText,
  Settings,
  ChevronLeft,
  ChevronRight,
  Library,
} from 'lucide-react'
import { type Page } from '../../stores'
import { cn } from '../../lib/utils'

interface NavItem {
  id: Page
  label: string
  icon: React.ReactNode
  shortcut?: number
}

const opsItems: NavItem[] = [
  { id: 'dashboard', label: 'Dashboard', icon: <LayoutDashboard size={20} />, shortcut: 1 },
  { id: 'sysops', label: 'System Ops', icon: <Monitor size={20} />, shortcut: 2 },
  { id: 'workflows', label: 'Workflow Library', icon: <Library size={20} />, shortcut: 3 },
  { id: 'netops', label: 'Network Ops', icon: <Network size={20} />, shortcut: 4 },
  { id: 'secops', label: 'Security Ops', icon: <Shield size={20} />, shortcut: 5 },
  { id: 'devops', label: 'DevOps', icon: <Terminal size={20} />, shortcut: 6 },
  { id: 'aiops', label: 'AI Ops', icon: <Bot size={20} />, shortcut: 7 },
]

const toolsItems: NavItem[] = [
  { id: 'reports', label: 'Reports', icon: <FileText size={20} />, shortcut: 8 },
  { id: 'alerts', label: 'Alerts', icon: <Bell size={20} />, shortcut: 9 },
  { id: 'logs', label: 'Logs', icon: <ScrollText size={20} />, shortcut: 10 },
  { id: 'settings', label: 'Settings', icon: <Settings size={20} />, shortcut: 11 },
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
        'flex flex-col transition-all duration-200 border-r wails-drag',
        collapsed ? 'w-[80px]' : 'w-[280px]',
        'border-[var(--color-border)] bg-[var(--color-sidebar)]/80 backdrop-blur-md'
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
        <div className="relative flex-shrink-0 group wails-no-drag">
          <div
            className="w-10 h-10 rounded-xl flex items-center justify-center transition-transform duration-700 group-hover:rotate-12"
            style={{
              background: 'linear-gradient(135deg, var(--color-accent), var(--color-accent-2))',
              boxShadow: '0 0 16px var(--color-accent-soft)',
            }}
          >
            <LayoutDashboard size={24} className="text-white fill-white" />
          </div>
        </div>

        {!collapsed && (
          <div className="flex flex-col min-w-0 flex-1 wails-no-drag">
            <span className="font-bold text-lg tracking-wider text-[var(--color-text)] truncate">
              UNIVERSAL-OPS
            </span>
            <span className="text-xs text-[var(--color-text-faint)] truncate leading-tight uppercase font-medium">
              The all-in-one Operations dashboard
            </span>
          </div>
        )}
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto px-3 py-4 space-y-6 wails-no-drag" role="navigation" aria-label="Main navigation">
        {/* Operations section */}
        <ul className="space-y-1">
          {opsItems.map((item) => (
            <li key={item.id} className="relative">
              <button
                onClick={() => onNavigate(item.id)}
                data-automation-id={`main-tab-${item.id}`}
                className={cn(
                  'group flex items-center w-full rounded-xl transition-all duration-150 active:scale-[0.95]',
                  collapsed ? 'justify-center p-3' : 'gap-3 px-3 py-2.5 text-sm',
                  currentPage === item.id
                    ? 'bg-accent text-white shadow-lg shadow-accent/20 font-black'
                    : 'text-text-dim hover:bg-[var(--color-sidebar-hover)] hover:text-text font-bold'
                )}
                title={collapsed ? item.label : undefined}
                aria-current={currentPage === item.id ? 'page' : undefined}
              >
                <span className={cn("flex-shrink-0", currentPage === item.id ? "text-white" : "text-accent")}>{item.icon}</span>
                {!collapsed && (
                  <span className="flex items-center justify-between flex-1 min-w-0">
                    <span className="truncate">{item.label}</span>
                    {item.shortcut && (
                      <kbd className={cn(
                        "text-[9px] font-black font-[Geist_Mono] border rounded px-1.5 py-0.5 opacity-0 group-hover:opacity-100 transition-opacity shrink-0",
                        currentPage === item.id ? "bg-white/20 border-white/30 text-white" : "bg-panel-3 border-border text-text-faint"
                      )}>
                        {item.shortcut}
                      </kbd>
                    )}
                  </span>
                )}
              </button>
              {currentPage === item.id && !collapsed && (
                <motion.div
                  layoutId="sidebar-active"
                  className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-5 bg-[var(--color-accent)] rounded-r"
                  transition={{ type: 'spring', stiffness: 400, damping: 30 }}
                />
              )}
            </li>
          ))}
        </ul>

        <div className="h-px bg-[var(--color-border)] mx-2 opacity-50" />

        {/* Tools section */}
        <ul className="space-y-1">
          {toolsItems.map((item) => (
            <li key={item.id} className="relative">
              <button
                onClick={() => onNavigate(item.id)}
                data-automation-id={`main-tab-${item.id}`}
                className={cn(
                  'group flex items-center w-full rounded-xl transition-all duration-150 active:scale-[0.95]',
                  collapsed ? 'justify-center p-3' : 'gap-3 px-3 py-2.5 text-sm',
                  currentPage === item.id
                    ? 'bg-accent text-white shadow-lg shadow-accent/20 font-black'
                    : 'text-text-dim hover:bg-[var(--color-sidebar-hover)] hover:text-text font-bold'
                )}
                title={collapsed ? item.label : undefined}
                aria-current={currentPage === item.id ? 'page' : undefined}
              >
                <span className={cn("flex-shrink-0", currentPage === item.id ? "text-white" : "text-accent")}>{item.icon}</span>
                {!collapsed && (
                  <span className="flex items-center justify-between flex-1 min-w-0">
                    <span className="truncate">{item.label}</span>
                    {item.shortcut && (
                      <kbd className={cn(
                        "text-[9px] font-black font-[Geist_Mono] border rounded px-1.5 py-0.5 opacity-0 group-hover:opacity-100 transition-opacity shrink-0",
                        currentPage === item.id ? "bg-white/20 border-white/30 text-white" : "bg-panel-3 border-border text-text-faint"
                      )}>
                        {item.shortcut}
                      </kbd>
                    )}
                  </span>
                )}
              </button>
              {currentPage === item.id && !collapsed && (
                <motion.div
                  layoutId="sidebar-active"
                  className="absolute left-0 top-1/2 -translate-y-1/2 w-0.5 h-5 bg-[var(--color-accent)] rounded-r"
                  transition={{ type: 'spring', stiffness: 400, damping: 30 }}
                />
              )}
            </li>
          ))}
        </ul>
      </nav>

      {/* Collapse toggle */}
      <div className="px-3 py-3 border-t border-[var(--color-border)] wails-no-drag">
        <button
          onClick={() => setCollapsed(!collapsed)}
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          className="flex items-center justify-center w-full gap-2 px-3 py-2.5 rounded-xl text-[var(--color-text-dim)] hover:bg-[var(--color-sidebar-hover)] hover:text-[var(--color-text)] transition-all font-medium text-sm"
        >
          {collapsed ? <ChevronRight size={18} /> : <ChevronLeft size={18} />}
          {!collapsed && <span>Collapse</span>}
        </button>
      </div>

      {/* Footer */}
      <div className="px-3 pb-4 pt-1 wails-no-drag">
        <p className="text-[11px] text-[var(--color-text-faint)]">Universal-Ops v1.3.1</p>
      </div>
    </aside>
  )
}
