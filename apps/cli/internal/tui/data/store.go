package data

import (
	"strconv"
	"time"
)

func MockStore(now time.Time) Store {
	workflows := []Workflow{
		{ID: "wf_analytics", Key: "daily-digest", Name: "daily-digest", Active: true, LatestVersion: 3, UpdatedAt: now.Add(-4 * time.Hour)},
		{ID: "wf_sync", Key: "sync-crm", Name: "sync-crm", Active: true, LatestVersion: 2, UpdatedAt: now.Add(-2 * time.Hour)},
		{ID: "wf_ops", Key: "ops-healthcheck", Name: "ops-healthcheck", Active: true, LatestVersion: 4, UpdatedAt: now.Add(-30 * time.Minute)},
		{ID: "wf_marketing", Key: "segment-import", Name: "segment-import", Active: false, LatestVersion: 1, UpdatedAt: now.Add(-36 * time.Hour)},
		{ID: "wf_billing", Key: "invoice-reminders", Name: "invoice-reminders", Active: true, LatestVersion: 5, UpdatedAt: now.Add(-12 * time.Hour)},
		{ID: "wf_security", Key: "rotate-keys", Name: "rotate-keys", Active: true, LatestVersion: 2, UpdatedAt: now.Add(-6 * time.Hour)},
		{ID: "wf_exports", Key: "weekly-export", Name: "weekly-export", Active: false, LatestVersion: 1, UpdatedAt: now.Add(-72 * time.Hour)},
	}

	versions := []WorkflowVersion{
		{ID: "v1", WorkflowID: "wf_analytics", Version: 1, CreatedAt: now.Add(-72 * time.Hour), DefinitionJSON: sampleWorkflowJSON("daily-digest", 1)},
		{ID: "v2", WorkflowID: "wf_analytics", Version: 2, CreatedAt: now.Add(-48 * time.Hour), DefinitionJSON: sampleWorkflowJSON("daily-digest", 2)},
		{ID: "v3", WorkflowID: "wf_analytics", Version: 3, CreatedAt: now.Add(-24 * time.Hour), DefinitionJSON: sampleWorkflowJSON("daily-digest", 3)},
		{ID: "v4", WorkflowID: "wf_sync", Version: 1, CreatedAt: now.Add(-96 * time.Hour), DefinitionJSON: sampleWorkflowJSON("sync-crm", 1)},
		{ID: "v5", WorkflowID: "wf_sync", Version: 2, CreatedAt: now.Add(-8 * time.Hour), DefinitionJSON: sampleWorkflowJSON("sync-crm", 2)},
		{ID: "v6", WorkflowID: "wf_ops", Version: 4, CreatedAt: now.Add(-2 * time.Hour), DefinitionJSON: sampleWorkflowJSON("ops-healthcheck", 4)},
		{ID: "v7", WorkflowID: "wf_marketing", Version: 1, CreatedAt: now.Add(-36 * time.Hour), DefinitionJSON: sampleWorkflowJSON("segment-import", 1)},
		{ID: "v8", WorkflowID: "wf_billing", Version: 5, CreatedAt: now.Add(-12 * time.Hour), DefinitionJSON: sampleWorkflowJSON("invoice-reminders", 5)},
		{ID: "v9", WorkflowID: "wf_security", Version: 2, CreatedAt: now.Add(-6 * time.Hour), DefinitionJSON: sampleWorkflowJSON("rotate-keys", 2)},
		{ID: "v10", WorkflowID: "wf_exports", Version: 1, CreatedAt: now.Add(-72 * time.Hour), DefinitionJSON: sampleWorkflowJSON("weekly-export", 1)},
	}

	triggers := []Trigger{
		{ID: "trg_1", WorkflowID: "wf_analytics", Key: "daily-0700", Type: "cron", Name: "daily-07:00", Active: true, CreatedAt: now.Add(-10 * time.Hour), ConfigJSON: `{"schedule":"0 7 * * *"}`},
		{ID: "trg_2", WorkflowID: "wf_sync", Key: "salesforce-push", Type: "webhook", Name: "salesforce-push", Active: true, CreatedAt: now.Add(-24 * time.Hour), ConfigJSON: `{"endpoint":"/hooks/salesforce"}`},
		{ID: "trg_3", WorkflowID: "wf_ops", Key: "every-5m", Type: "cron", Name: "every-5m", Active: true, CreatedAt: now.Add(-6 * time.Hour), ConfigJSON: `{"schedule":"*/5 * * * *"}`},
		{ID: "trg_4", WorkflowID: "wf_marketing", Key: "ad-hoc", Type: "manual", Name: "ad-hoc", Active: false, CreatedAt: now.Add(-36 * time.Hour), ConfigJSON: `{"owner":"growth"}`},
		{ID: "trg_5", WorkflowID: "wf_billing", Key: "hourly", Type: "cron", Name: "hourly", Active: true, CreatedAt: now.Add(-18 * time.Hour), ConfigJSON: `{"schedule":"0 * * * *"}`},
		{ID: "trg_6", WorkflowID: "wf_security", Key: "rotate-now", Type: "manual", Name: "rotate-now", Active: true, CreatedAt: now.Add(-12 * time.Hour), ConfigJSON: `{"severity":"high"}`},
		{ID: "trg_7", WorkflowID: "wf_exports", Key: "weekly-monday", Type: "cron", Name: "weekly-monday", Active: false, CreatedAt: now.Add(-72 * time.Hour), ConfigJSON: `{"schedule":"0 6 * * 1"}`},
	}

	runs := []WorkflowRun{
		{ID: "run_1024", WorkflowID: "wf_analytics", Number: 1024, Status: "SUCCEEDED", TriggerType: "cron", StartedAt: now.Add(-55 * time.Minute), Duration: 4 * time.Minute, InputJSON: `{"range":"24h"}`, OutputJSON: `{"rows":142,"status":"ok"}`},
		{ID: "run_1025", WorkflowID: "wf_analytics", Number: 1025, Status: "FAILED", TriggerType: "cron", StartedAt: now.Add(-3 * time.Hour), Duration: 2 * time.Minute, InputJSON: `{"range":"24h"}`, OutputJSON: `{}`, ErrorJSON: `{"error":"s3 upload failed","code":"S3_502"}`},
		{ID: "run_1026", WorkflowID: "wf_sync", Number: 1026, Status: "RUNNING", TriggerType: "webhook", StartedAt: now.Add(-8 * time.Minute), Duration: 0, InputJSON: `{"batch":"crm-91"}`, OutputJSON: `{}`},
		{ID: "run_1027", WorkflowID: "wf_ops", Number: 1027, Status: "SUCCEEDED", TriggerType: "cron", StartedAt: now.Add(-12 * time.Minute), Duration: 1 * time.Minute, InputJSON: `{"targets":8}`, OutputJSON: `{"healthy":8}`},
		{ID: "run_1028", WorkflowID: "wf_marketing", Number: 1028, Status: "QUEUED", TriggerType: "manual", StartedAt: now.Add(-4 * time.Minute), Duration: 0, InputJSON: `{"segment":"retarget"}`, OutputJSON: `{}`},
		{ID: "run_1029", WorkflowID: "wf_billing", Number: 1029, Status: "FAILED", TriggerType: "cron", StartedAt: now.Add(-90 * time.Minute), Duration: 3 * time.Minute, InputJSON: `{"count":51}`, OutputJSON: `{}`, ErrorJSON: `{"error":"rate limited","retry_in":"15m"}`},
		{ID: "run_1030", WorkflowID: "wf_security", Number: 1030, Status: "SUCCEEDED", TriggerType: "manual", StartedAt: now.Add(-6 * time.Hour), Duration: 6 * time.Minute, InputJSON: `{"rotate":"kms"}`, OutputJSON: `{"rotated":3}`},
		{ID: "run_1031", WorkflowID: "wf_exports", Number: 1031, Status: "SUCCEEDED", TriggerType: "cron", StartedAt: now.Add(-26 * time.Hour), Duration: 12 * time.Minute, InputJSON: `{"target":"s3"}`, OutputJSON: `{"files":12}`},
		{ID: "run_1032", WorkflowID: "wf_sync", Number: 1032, Status: "SUCCEEDED", TriggerType: "webhook", StartedAt: now.Add(-5 * time.Hour), Duration: 7 * time.Minute, InputJSON: `{"batch":"crm-90"}`, OutputJSON: `{"synced":128}`},
		{ID: "run_1033", WorkflowID: "wf_ops", Number: 1033, Status: "SUCCEEDED", TriggerType: "cron", StartedAt: now.Add(-30 * time.Minute), Duration: 1 * time.Minute, InputJSON: `{"targets":8}`, OutputJSON: `{"healthy":7,"degraded":1}`},
		{ID: "run_1034", WorkflowID: "wf_billing", Number: 1034, Status: "RUNNING", TriggerType: "cron", StartedAt: now.Add(-2 * time.Minute), Duration: 0, InputJSON: `{"count":34}`, OutputJSON: `{}`},
		{ID: "run_1035", WorkflowID: "wf_security", Number: 1035, Status: "QUEUED", TriggerType: "manual", StartedAt: now.Add(-1 * time.Minute), Duration: 0, InputJSON: `{"rotate":"db"}`, OutputJSON: `{}`},
	}

	stepRuns := []StepRun{}
	for _, run := range runs {
		stepRuns = append(stepRuns, buildStepRuns(now, run.ID, run.Status)...)
	}

	events := []Event{
		{ID: "evt_2001", TriggerID: "trg_2", Type: "webhook", ReceivedAt: now.Add(-8 * time.Minute), RunID: strPtr("run_1026"), PayloadJSON: `{"account":"acme","count":12}`, Metadata: `{"source":"salesforce"}`},
		{ID: "evt_2002", TriggerID: "trg_3", Type: "cron", ReceivedAt: now.Add(-12 * time.Minute), RunID: strPtr("run_1027"), PayloadJSON: `{"schedule":"*/5 * * * *"}`, Metadata: `{"node":"ops-1"}`},
		{ID: "evt_2003", TriggerID: "trg_1", Type: "cron", ReceivedAt: now.Add(-55 * time.Minute), RunID: strPtr("run_1024"), PayloadJSON: `{"schedule":"0 7 * * *"}`, Metadata: `{"timezone":"UTC"}`},
		{ID: "evt_2004", TriggerID: "trg_5", Type: "cron", ReceivedAt: now.Add(-2 * time.Minute), RunID: strPtr("run_1034"), PayloadJSON: `{"schedule":"0 * * * *"}`, Metadata: `{"queue":"billing"}`},
		{ID: "evt_2005", TriggerID: "trg_6", Type: "manual", ReceivedAt: now.Add(-6 * time.Hour), RunID: strPtr("run_1030"), PayloadJSON: `{"operator":"secops"}`, Metadata: `{"priority":"high"}`},
		{ID: "evt_2006", TriggerID: "trg_4", Type: "manual", ReceivedAt: now.Add(-4 * time.Hour), RunID: nil, PayloadJSON: `{"segment":"retarget"}`, Metadata: `{"note":"skipped"}`},
	}

	secrets := []Secret{
		{ID: "sec_01", Name: "slack_token", Description: "Slack bot token", CreatedAt: now.Add(-240 * time.Hour)},
		{ID: "sec_02", Name: "stripe_key", Description: "Stripe secret key", CreatedAt: now.Add(-120 * time.Hour)},
		{ID: "sec_03", Name: "s3_backup", Description: "S3 backup credentials", CreatedAt: now.Add(-96 * time.Hour)},
		{ID: "sec_04", Name: "sendgrid_api", Description: "SendGrid API token", CreatedAt: now.Add(-200 * time.Hour)},
	}

	lastUsed := now.Add(-2 * time.Hour)
	tokens := []ApiToken{
		{ID: "tok_01", Name: "deploy-bot", Scopes: []string{"runs:write", "workflows:read"}, CreatedAt: now.Add(-400 * time.Hour), LastUsedAt: &lastUsed, Revoked: false},
		{ID: "tok_02", Name: "ops-dashboard", Scopes: []string{"runs:read", "events:read"}, CreatedAt: now.Add(-300 * time.Hour), LastUsedAt: nil, Revoked: false},
		{ID: "tok_03", Name: "legacy-token", Scopes: []string{"*"}, CreatedAt: now.Add(-900 * time.Hour), LastUsedAt: nil, Revoked: true},
	}

	return Store{
		Workflows:        workflows,
		WorkflowVersions: versions,
		Triggers:         triggers,
		Runs:             runs,
		StepRuns:         stepRuns,
		Events:           events,
		Secrets:          secrets,
		ApiTokens:        tokens,
	}
}

func strPtr(value string) *string {
	return &value
}

func buildStepRuns(now time.Time, runID string, status string) []StepRun {
	steps := []string{"ingest", "transform", "load"}
	items := make([]StepRun, 0, len(steps))
	for i, step := range steps {
		items = append(items, StepRun{
			ID:        runID + "_" + step,
			RunID:     runID,
			StepKey:   step,
			Status:    stepStatus(status, i),
			StartedAt: now.Add(time.Duration(-i-1) * time.Minute),
			Duration:  time.Duration(i+1) * time.Minute,
			Log:       sampleLog(runID, step),
		})
	}
	return items
}

func stepStatus(runStatus string, index int) string {
	switch runStatus {
	case "FAILED":
		if index == 1 {
			return "FAILED"
		}
		return "SUCCEEDED"
	case "RUNNING":
		if index == 2 {
			return "RUNNING"
		}
		return "SUCCEEDED"
	default:
		return "SUCCEEDED"
	}
}

func sampleLog(runID string, step string) string {
	return "run=" + runID + " step=" + step + "\n" +
		"info starting\n" +
		"info processing batch\n" +
		"info done"
}

func sampleWorkflowJSON(name string, version int) string {
	return `{"workflow":"` + name + `","version":` + itoa(version) + `,"steps":[{"id":"ingest"},{"id":"transform"},{"id":"load"}]}`
}

func itoa(value int) string {
	return strconv.Itoa(value)
}
