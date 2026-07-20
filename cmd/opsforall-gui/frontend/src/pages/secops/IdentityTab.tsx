import { useQuery } from '@tanstack/react-query'
import { Users, ShieldAlert, UserCheck, Lock, Clock } from 'lucide-react'
import { useBackend } from '@/hooks/useBackend'
import { useSettingsStore } from '@/stores/useSettingsStore'
import { SectionBriefing } from '@/components/ui/SectionBriefing'
import { MiniStat } from '@/components/ui/MiniStat'
import { StatusBadge } from '@/components/ui/StatusBadge'
import { Panel } from '@/components/ui/Panel'
import type { UserInfo, PasswordPolicy, FailedLogin, LockedAccount } from '@/types'

export function IdentityTab() {
  const { call } = useBackend()
  const { refreshInterval } = useSettingsStore()

  const { data: users = [] } = useQuery<UserInfo[]>({
    queryKey: ['secops-users'],
    queryFn: async () => (await call('SecOps.GetUsers') as UserInfo[]) || [],
    refetchInterval: refreshInterval,
  })

  const { data: policy } = useQuery<PasswordPolicy>({
    queryKey: ['secops-password-policy'],
    queryFn: async () => await call('SecOps.GetPasswordPolicy') as PasswordPolicy,
    refetchInterval: refreshInterval,
  })

  const { data: failedLogins = [] } = useQuery<FailedLogin[]>({
    queryKey: ['secops-failed-logins'],
    queryFn: async () => (await call('SecOps.GetFailedLogins') as FailedLogin[]) || [],
    refetchInterval: refreshInterval,
  })

  const { data: lockouts = [] } = useQuery<LockedAccount[]>({
    queryKey: ['secops-lockouts'],
    queryFn: async () => (await call('SecOps.GetAccountLockouts') as LockedAccount[]) || [],
    refetchInterval: refreshInterval,
  })

  const adminCount = users.filter(u => u.is_admin && u.is_enabled).length
  const totalCount = users.length

  return (
    <div className="space-y-8 animate-in fade-in duration-500">
      <SectionBriefing
        title="Identity & Access"
        objective="Audit user accounts, password policy, and access control. Apply the Principle of Least Privilege (PoLP) — only grant admin access when necessary."
        checklist={['User account inventory', 'Admin account count', 'Password policy strength', 'Failed login monitoring', 'Account lockout status']}
      />

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        <MiniStat label="Total Accounts" value={totalCount} icon={<Users size={24} />} variant="default" />
        <MiniStat label="Active Admins" value={adminCount} icon={<ShieldAlert size={24} />} variant={adminCount <= 2 ? 'success' : 'warning'} />
        <MiniStat label="Failed Logins" value={failedLogins.length} icon={<Lock size={24} />} variant={failedLogins.length === 0 ? 'success' : 'danger'} />
        <MiniStat label="Locked Accounts" value={lockouts.length} icon={<Clock size={24} />} variant={lockouts.length === 0 ? 'success' : 'warning'} />
      </div>

      {/* Password Policy */}
      {policy && (
        <Panel padding="lg" category="security">
          <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6">Password Policy</h3>
          <div className="grid grid-cols-5 gap-6">
            <div className="bg-panel-2 border border-border rounded-xl p-4">
              <p className="text-xs font-bold text-text-faint uppercase tracking-wider mb-1">Max Age</p>
              <p className="text-xl font-bold text-text tabular-nums">{policy.max_age} days</p>
            </div>
            <div className="bg-panel-2 border border-border rounded-xl p-4">
              <p className="text-xs font-bold text-text-faint uppercase tracking-wider mb-1">Min Length</p>
              <p className="text-xl font-bold text-text tabular-nums">{policy.min_length}</p>
            </div>
            <div className="bg-panel-2 border border-border rounded-xl p-4">
              <p className="text-xs font-bold text-text-faint uppercase tracking-wider mb-1">Complexity</p>
              <StatusBadge status={policy.complexity ? 'enabled' : 'disabled'} />
            </div>
            <div className="bg-panel-2 border border-border rounded-xl p-4">
              <p className="text-xs font-bold text-text-faint uppercase tracking-wider mb-1">Lockout Threshold</p>
              <p className="text-xl font-bold text-text tabular-nums">{policy.lockout_threshold || 'None'}</p>
            </div>
            <div className="bg-panel-2 border border-border rounded-xl p-4">
              <p className="text-xs font-bold text-text-faint uppercase tracking-wider mb-1">Lockout Duration</p>
              <p className="text-xl font-bold text-text tabular-nums">{policy.lockout_duration} min</p>
            </div>
          </div>
        </Panel>
      )}

      {/* User Accounts */}
      <Panel padding="lg" category="security">
        <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6">User Accounts</h3>
        <div className="grid grid-cols-2 gap-6">
          {users.map(user => (
            <div key={user.username} className="bg-panel-2 border border-border rounded-xl p-6 flex items-center gap-6">
              <div className={`w-12 h-12 rounded-xl flex items-center justify-center border shadow-inner ${user.is_admin ? 'bg-warning/10 border-warning/30 text-warning' : 'bg-accent/10 border-accent/30 text-accent'}`}>
                {user.is_admin ? <ShieldAlert size={24} /> : <UserCheck size={24} />}
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-3 mb-1">
                  <span className="text-sm font-bold text-text truncate">{user.username}</span>
                  {user.is_admin && <StatusBadge status="Admin" />}
                  {!user.is_enabled && <StatusBadge status="Disabled" />}
                </div>
                <p className="text-xs text-text-faint truncate">{user.full_name || 'System Account'}</p>
              </div>
            </div>
          ))}
        </div>
      </Panel>

      {/* Failed Logins */}
      {failedLogins.length > 0 && (
        <Panel padding="lg" category="security">
          <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6 text-danger">Failed Login Attempts</h3>
          <div className="space-y-2">
            {failedLogins.map((fl, i) => (
              <div key={i} className="flex items-center justify-between bg-panel-2 border border-border rounded-xl px-5 py-3">
                <span className="text-sm font-bold text-text">{fl.username}</span>
                <span className="text-sm text-text-dim">{fl.source_ip}</span>
                <span className="text-sm font-bold text-danger tabular-nums">{fl.count} attempts</span>
              </div>
            ))}
          </div>
        </Panel>
      )}

      {/* Locked Accounts */}
      {lockouts.length > 0 && (
        <Panel padding="lg" category="security">
          <h3 className="text-lg font-bold text-text uppercase tracking-widest mb-6 text-warning">Locked Accounts</h3>
          <div className="space-y-2">
            {lockouts.map((l, i) => (
              <div key={i} className="flex items-center justify-between bg-panel-2 border border-border rounded-xl px-5 py-3">
                <span className="text-sm font-bold text-text">{l.username}</span>
                <span className="text-sm text-text-dim">Locked since: {l.locked_since || 'Unknown'}</span>
              </div>
            ))}
          </div>
        </Panel>
      )}
    </div>
  )
}
