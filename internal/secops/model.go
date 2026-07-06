package secops

import (
	tea "charm.land/bubbletea/v2"
)

// Model represents the SecOps (Security Operations) layer.
type Model struct {
	// UI state
	tabIndex int
	ready    bool
	loading  bool

	// Firewall data
	firewallRules []FirewallRule
	fwErr         error

	// User data
	users    []UserInfo
	groups   []string
	usersErr error

	// Listening port data
	listeningPorts []ListeningPort
	lpErr          error

	// Defender data
	defenderStatus *DefenderStatus
	defErr         error

	// Scheduled task data
	scheduledTasks []ScheduledTask
	tasksErr       error

	// Security event data
	securityEvents []SecurityEvent
	eventsErr      error

	// Workflow state
	workflowReport string
	showReport     bool
}

// NewModel creates a new SecOps model.
func NewModel() *Model {
	return &Model{
		tabIndex: 0,
	}
}

// Init initializes the SecOps layer by loading all data sources.
func (m *Model) Init() tea.Cmd {
	m.loading = true
	return tea.Batch(
		loadFirewallCmd(),
		loadUsersCmd(),
		loadListeningCmd(),
		loadDefenderCmd(),
		loadTasksCmd(),
		loadEventsCmd(),
	)
}

// ---- Result message types ----

// FirewallResult is sent after firewall rules are collected.
type FirewallResult struct {
	Rules []FirewallRule
	Err   error
}

// UsersResult is sent after user accounts are collected.
type UsersResult struct {
	Users  []UserInfo
	Groups []string
	Err    error
}

// ListeningResult is sent after listening ports are collected.
type ListeningResult struct {
	Ports []ListeningPort
	Err   error
}

// DefenderResult is sent after Defender status is collected.
type DefenderResult struct {
	Status *DefenderStatus
	Err    error
}

// TasksResult is sent after scheduled tasks are collected.
type TasksResult struct {
	Tasks []ScheduledTask
	Err   error
}

// EventsResult is sent after security events are collected.
type EventsResult struct {
	Events []SecurityEvent
	Err    error
}

// ---- Load commands ----

func loadFirewallCmd() tea.Cmd {
	return func() tea.Msg {
		rules, err := GetFirewallRules()
		return FirewallResult{Rules: rules, Err: err}
	}
}

func loadUsersCmd() tea.Cmd {
	return func() tea.Msg {
		users, err := GetUsers()
		groups, _ := GetGroups()
		return UsersResult{Users: users, Groups: groups, Err: err}
	}
}

func loadListeningCmd() tea.Cmd {
	return func() tea.Msg {
		ports, err := GetListeningPorts()
		return ListeningResult{Ports: ports, Err: err}
	}
}

func loadDefenderCmd() tea.Cmd {
	return func() tea.Msg {
		status, err := GetDefenderStatus()
		return DefenderResult{Status: status, Err: err}
	}
}

func loadTasksCmd() tea.Cmd {
	return func() tea.Msg {
		tasks, err := GetScheduledTasks()
		return TasksResult{Tasks: tasks, Err: err}
	}
}

func loadEventsCmd() tea.Cmd {
	return func() tea.Msg {
		events, err := GetSecurityEvents()
		return EventsResult{Events: events, Err: err}
	}
}

// ---- Getters ----

// Ready returns true if data has been collected at least once.
func (m *Model) Ready() bool {
	return m.ready
}

// TabIndex returns the current active tab.
func (m *Model) TabIndex() int {
	return m.tabIndex
}

// Loading returns true if data is currently being collected.
func (m *Model) Loading() bool {
	return m.loading
}

// Error returns the first non-nil error across all data sources.
func (m *Model) Error() error {
	for _, err := range []error{m.fwErr, m.usersErr, m.lpErr, m.defErr, m.tasksErr, m.eventsErr} {
		if err != nil {
			return err
		}
	}
	return nil
}

// FirewallRules returns the collected firewall rules.
func (m *Model) FirewallRules() []FirewallRule {
	return m.firewallRules
}

// Users returns the collected user accounts.
func (m *Model) Users() []UserInfo {
	return m.users
}

// Groups returns the collected security groups.
func (m *Model) Groups() []string {
	return m.groups
}

// ListeningPorts returns the collected listening ports.
func (m *Model) ListeningPorts() []ListeningPort {
	return m.listeningPorts
}

// DefenderStatus returns the collected Defender status.
func (m *Model) DefenderStatus() *DefenderStatus {
	return m.defenderStatus
}

// ScheduledTasks returns the collected scheduled tasks.
func (m *Model) ScheduledTasks() []ScheduledTask {
	return m.scheduledTasks
}

// SecurityEvents returns the collected security events.
func (m *Model) SecurityEvents() []SecurityEvent {
	return m.securityEvents
}

// FirewallError returns the firewall fetch error.
func (m *Model) FirewallError() error {
	return m.fwErr
}

// UsersError returns the users fetch error.
func (m *Model) UsersError() error {
	return m.usersErr
}

// ListeningError returns the listening ports fetch error.
func (m *Model) ListeningError() error {
	return m.lpErr
}

// DefenderError returns the defender status fetch error.
func (m *Model) DefenderError() error {
	return m.defErr
}

// TasksError returns the scheduled tasks fetch error.
func (m *Model) TasksError() error {
	return m.tasksErr
}

// EventsError returns the security events fetch error.
func (m *Model) EventsError() error {
	return m.eventsErr
}

// ShowReport returns true if the workflow report is currently being displayed.
func (m *Model) ShowReport() bool {
	return m.showReport
}

// WorkflowReport returns the last generated workflow report content.
func (m *Model) WorkflowReport() string {
	return m.workflowReport
}
