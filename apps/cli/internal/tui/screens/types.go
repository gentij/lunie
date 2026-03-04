package screens

type ViewID string

const (
	ViewDashboard ViewID = "dashboard"
	ViewWorkflows ViewID = "workflows"
	ViewRuns      ViewID = "runs"
	ViewTriggers  ViewID = "triggers"
	ViewEvents    ViewID = "events"
	ViewSecrets   ViewID = "secrets"
	ViewTokens    ViewID = "tokens"
)

type ContextTab int

const (
	ContextTabOverview ContextTab = iota
	ContextTabJSON
	ContextTabSteps
	ContextTabLogs
)
