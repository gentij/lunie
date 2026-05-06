package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/gentij/lunie/apps/cli/internal/tui/data"
)

func workflowName(store *data.Store, workflowID string) string {
	if wf, ok := workflowByID(store, workflowID); ok {
		return wf.Name
	}
	return workflowID
}

func workflowKey(store *data.Store, workflowID string) string {
	if wf, ok := workflowByID(store, workflowID); ok && strings.TrimSpace(wf.Key) != "" {
		return wf.Key
	}
	return workflowID
}

func triggerKey(store *data.Store, triggerID string) string {
	if trg, ok := triggerByID(store, triggerID); ok && strings.TrimSpace(trg.Key) != "" {
		return trg.Key
	}
	return triggerID
}

func runLabel(store *data.Store, runID string) string {
	if run, ok := runByID(store, runID); ok {
		return fmt.Sprintf("#%d", run.Number)
	}
	return runID
}

func workflowByID(store *data.Store, workflowID string) (data.Workflow, bool) {
	for _, wf := range store.Workflows {
		if wf.ID == workflowID {
			return wf, true
		}
	}
	return data.Workflow{}, false
}

func runByID(store *data.Store, runID string) (data.WorkflowRun, bool) {
	for _, run := range store.Runs {
		if run.ID == runID {
			return run, true
		}
	}
	return data.WorkflowRun{}, false
}

func triggerByID(store *data.Store, triggerID string) (data.Trigger, bool) {
	for _, trg := range store.Triggers {
		if trg.ID == triggerID {
			return trg, true
		}
	}
	return data.Trigger{}, false
}

func eventByID(store *data.Store, eventID string) (data.Event, bool) {
	for _, evt := range store.Events {
		if evt.ID == eventID {
			return evt, true
		}
	}
	return data.Event{}, false
}

func secretByID(store *data.Store, secretID string) (data.Secret, bool) {
	for _, sec := range store.Secrets {
		if sec.ID == secretID {
			return sec, true
		}
	}
	return data.Secret{}, false
}

func tokenByID(store *data.Store, tokenID string) (data.ApiToken, bool) {
	for _, tok := range store.ApiTokens {
		if tok.ID == tokenID {
			return tok, true
		}
	}
	return data.ApiToken{}, false
}

func countTriggers(store *data.Store, workflowID string) int {
	count := 0
	for _, trg := range store.Triggers {
		if trg.WorkflowID == workflowID {
			count++
		}
	}
	return count
}

func runsForWorkflow(store *data.Store, workflowID string) []data.WorkflowRun {
	items := []data.WorkflowRun{}
	for _, run := range store.Runs {
		if run.WorkflowID == workflowID {
			items = append(items, run)
		}
	}
	return items
}

func lastRunForWorkflow(store *data.Store, workflowID string) *data.WorkflowRun {
	for _, run := range store.Runs {
		if run.WorkflowID == workflowID {
			return &run
		}
	}
	return nil
}

func lastRunStatus(store *data.Store, workflowID string) string {
	if run := lastRunForWorkflow(store, workflowID); run != nil {
		return normalizeStatus(run.Status)
	}
	return "QUEUED"
}

func versionsForWorkflow(store *data.Store, workflowID string) []data.WorkflowVersion {
	items := []data.WorkflowVersion{}
	for _, v := range store.WorkflowVersions {
		if v.WorkflowID == workflowID {
			items = append(items, v)
		}
	}
	return items
}

func triggersForWorkflow(store *data.Store, workflowID string) []data.Trigger {
	items := []data.Trigger{}
	for _, t := range store.Triggers {
		if t.WorkflowID == workflowID {
			items = append(items, t)
		}
	}
	return items
}

func eventsForTrigger(store *data.Store, triggerID string) []data.Event {
	items := []data.Event{}
	for _, evt := range store.Events {
		if evt.TriggerID == triggerID {
			items = append(items, evt)
		}
	}
	return items
}

func stepsForRun(store *data.Store, runID string) []data.StepRun {
	items := []data.StepRun{}
	for _, step := range store.StepRuns {
		if step.RunID == runID {
			items = append(items, step)
		}
	}
	return items
}

func latestDefinition(versions []data.WorkflowVersion) string {
	if len(versions) == 0 {
		return "{}"
	}
	latest := versions[0]
	for _, v := range versions {
		if v.Version > latest.Version {
			latest = v
		}
	}
	return latest.DefinitionJSON
}

func activeLabel(active bool) string {
	if active {
		return "ACTIVE"
	}
	return "INACTIVE"
}

func normalizeStatus(status string) string {
	status = strings.TrimSpace(strings.ToUpper(status))
	if status == "" {
		return "UNKNOWN"
	}
	return status
}

func formatDuration(d time.Duration) string {
	if d <= 0 {
		return "-"
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}
