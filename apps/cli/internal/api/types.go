package api

type Pagination struct {
	Page       int  `json:"page"`
	PageSize   int  `json:"pageSize"`
	Total      int  `json:"total"`
	TotalPages int  `json:"totalPages"`
	HasNext    bool `json:"hasNext"`
	HasPrev    bool `json:"hasPrev"`
}

type Paginated[T any] struct {
	Items      []T        `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type Workflow struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	IsActive        bool    `json:"isActive"`
	LatestVersionID *string `json:"latestVersionId"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

type WorkflowVersion struct {
	ID         string `json:"id"`
	WorkflowID string `json:"workflowId"`
	Version    int    `json:"version"`
	Definition any    `json:"definition"`
	CreatedAt  string `json:"createdAt"`
}

type Trigger struct {
	ID         string  `json:"id"`
	WorkflowID string  `json:"workflowId"`
	Type       string  `json:"type"`
	Name       *string `json:"name"`
	IsActive   bool    `json:"isActive"`
	Config     any     `json:"config"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
}

type Event struct {
	ID         string  `json:"id"`
	TriggerID  string  `json:"triggerId"`
	Type       *string `json:"type"`
	ExternalID *string `json:"externalId"`
	Payload    any     `json:"payload"`
	ReceivedAt string  `json:"receivedAt"`
	CreatedAt  string  `json:"createdAt"`
}

type WorkflowRun struct {
	ID                string  `json:"id"`
	WorkflowID        string  `json:"workflowId"`
	WorkflowVersionID string  `json:"workflowVersionId"`
	TriggerID         *string `json:"triggerId"`
	EventID           *string `json:"eventId"`
	Status            string  `json:"status"`
	Input             any     `json:"input"`
	Overrides         any     `json:"overrides"`
	Output            any     `json:"output"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
	StartedAt         *string `json:"startedAt"`
	FinishedAt        *string `json:"finishedAt"`
}

type StepRun struct {
	ID              string  `json:"id"`
	WorkflowRunID   string  `json:"workflowRunId"`
	StepKey         string  `json:"stepKey"`
	Status          string  `json:"status"`
	Attempt         int     `json:"attempt"`
	Input           any     `json:"input"`
	RequestOverride any     `json:"requestOverride"`
	Output          any     `json:"output"`
	Error           any     `json:"error"`
	Logs            any     `json:"logs"`
	LastErrorAt     *string `json:"lastErrorAt"`
	DurationMs      *int    `json:"durationMs"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
	StartedAt       *string `json:"startedAt"`
	FinishedAt      *string `json:"finishedAt"`
}

type Secret struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Value       string  `json:"value"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type WhoAmI struct {
	ID     string   `json:"id"`
	Name   string   `json:"name"`
	Scopes []string `json:"scopes"`
}

type Health struct {
	Status  string  `json:"status"`
	Version string  `json:"version"`
	Uptime  float64 `json:"uptime"`
	DB      struct {
		Ok bool `json:"ok"`
	} `json:"db"`
}
