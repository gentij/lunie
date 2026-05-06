package data

import "time"

type Workflow struct {
	ID            string
	Key           string
	Name          string
	Active        bool
	LatestVersion int
	UpdatedAt     time.Time
}

type WorkflowVersion struct {
	ID             string
	WorkflowID     string
	Version        int
	CreatedAt      time.Time
	DefinitionJSON string
}

type Trigger struct {
	ID         string
	WorkflowID string
	Key        string
	Type       string
	Name       string
	Active     bool
	CreatedAt  time.Time
	ConfigJSON string
}

type Event struct {
	ID          string
	TriggerID   string
	Type        string
	ReceivedAt  time.Time
	RunID       *string
	PayloadJSON string
	Metadata    string
}

type WorkflowRun struct {
	ID          string
	WorkflowID  string
	Number      int
	Status      string
	TriggerType string
	StartedAt   time.Time
	Duration    time.Duration
	InputJSON   string
	OutputJSON  string
	ErrorJSON   string
}

type StepRun struct {
	ID        string
	RunID     string
	StepKey   string
	Status    string
	StartedAt time.Time
	Duration  time.Duration
	Log       string
}

func (s StepRun) FilterValue() string { return s.StepKey }

type Secret struct {
	ID          string
	Name        string
	Description string
	CreatedAt   time.Time
}

type ApiToken struct {
	ID         string
	Name       string
	Scopes     []string
	CreatedAt  time.Time
	LastUsedAt *time.Time
	Revoked    bool
}

type Store struct {
	Workflows        []Workflow
	WorkflowVersions []WorkflowVersion
	Triggers         []Trigger
	Runs             []WorkflowRun
	StepRuns         []StepRun
	Events           []Event
	Secrets          []Secret
	ApiTokens        []ApiToken
}
