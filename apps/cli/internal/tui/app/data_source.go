package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentij/lunie/apps/cli/internal/api"
	"github.com/gentij/lunie/apps/cli/internal/tui/data"
)

const apiPageSize = 100

type apiListSort struct {
	By    string
	Order string
}

type snapshotSortOptions struct {
	Workflows        apiListSort
	WorkflowVersions apiListSort
	Triggers         apiListSort
	Runs             apiListSort
	Events           apiListSort
	StepRuns         apiListSort
	Secrets          apiListSort
}

func defaultSnapshotSortOptions() snapshotSortOptions {
	return snapshotSortOptions{
		Workflows:        apiListSort{By: "updatedAt", Order: "desc"},
		WorkflowVersions: apiListSort{By: "version", Order: "desc"},
		Triggers:         apiListSort{By: "createdAt", Order: "desc"},
		Runs:             apiListSort{By: "createdAt", Order: "desc"},
		Events:           apiListSort{By: "receivedAt", Order: "desc"},
		StepRuns:         apiListSort{By: "createdAt", Order: "asc"},
		Secrets:          apiListSort{By: "createdAt", Order: "desc"},
	}
}

func normalizedSnapshotSortOptions(input snapshotSortOptions) snapshotSortOptions {
	defaults := defaultSnapshotSortOptions()
	merge := func(current apiListSort, fallback apiListSort) apiListSort {
		if strings.TrimSpace(current.By) == "" {
			current.By = fallback.By
		}
		if strings.TrimSpace(current.Order) == "" {
			current.Order = fallback.Order
		}
		current.Order = strings.ToLower(strings.TrimSpace(current.Order))
		if current.Order != "asc" && current.Order != "desc" {
			current.Order = fallback.Order
		}
		return current
	}

	return snapshotSortOptions{
		Workflows:        merge(input.Workflows, defaults.Workflows),
		WorkflowVersions: merge(input.WorkflowVersions, defaults.WorkflowVersions),
		Triggers:         merge(input.Triggers, defaults.Triggers),
		Runs:             merge(input.Runs, defaults.Runs),
		Events:           merge(input.Events, defaults.Events),
		StepRuns:         merge(input.StepRuns, defaults.StepRuns),
		Secrets:          merge(input.Secrets, defaults.Secrets),
	}
}

type snapshotLoadedMsg struct {
	store     data.Store
	apiStatus string
	err       error
}

type mutationResultMsg struct {
	successMessage string
	err            error
	refresh        bool
}

func fetchSnapshotCmd(client *api.Client, sortOptions snapshotSortOptions, delay time.Duration, fail bool) tea.Cmd {
	return func() tea.Msg {
		if delay > 0 {
			time.Sleep(delay)
		}
		if fail {
			return snapshotLoadedMsg{err: fmt.Errorf("simulated flaky network")}
		}
		store, err := loadStoreFromAPI(client, normalizedSnapshotSortOptions(sortOptions))
		if err != nil {
			return snapshotLoadedMsg{err: err, apiStatus: "OFFLINE"}
		}
		return snapshotLoadedMsg{store: store, apiStatus: "CONNECTED"}
	}
}

func loadStoreFromAPI(client *api.Client, sortOptions snapshotSortOptions) (data.Store, error) {
	if client == nil {
		return data.Store{}, fmt.Errorf("api client is not configured")
	}

	apiWorkflows, err := listAllWorkflows(client, sortOptions.Workflows)
	if err != nil {
		return data.Store{}, err
	}

	versionsByWorkflow := map[string][]api.WorkflowVersion{}
	triggersByWorkflow := map[string][]api.Trigger{}
	runsByWorkflow := map[string][]api.WorkflowRun{}
	eventsByTrigger := map[string][]api.Event{}
	stepsByRun := map[string][]api.StepRun{}

	for _, wf := range apiWorkflows {
		versions, err := listAllWorkflowVersions(client, wf.ID, sortOptions.WorkflowVersions)
		if err != nil {
			return data.Store{}, err
		}
		versionsByWorkflow[wf.ID] = versions

		triggers, err := listAllTriggers(client, wf.ID, sortOptions.Triggers)
		if err != nil {
			return data.Store{}, err
		}
		triggersByWorkflow[wf.ID] = triggers

		runs, err := listAllWorkflowRuns(client, wf.ID, sortOptions.Runs)
		if err != nil {
			return data.Store{}, err
		}
		runsByWorkflow[wf.ID] = runs

		for _, trigger := range triggers {
			events, err := listAllEvents(client, wf.ID, trigger.ID, sortOptions.Events)
			if err != nil {
				return data.Store{}, err
			}
			eventsByTrigger[trigger.ID] = events
		}

		for _, run := range runs {
			steps, err := listAllStepRuns(client, wf.ID, run.ID, sortOptions.StepRuns)
			if err != nil {
				return data.Store{}, err
			}
			stepsByRun[run.ID] = steps
		}
	}

	secrets, err := listAllSecrets(client, sortOptions.Secrets)
	if err != nil {
		return data.Store{}, err
	}

	store := data.Store{}
	triggerTypeByID := map[string]string{}
	runByEventID := map[string]string{}
	runIndexByID := map[string]int{}

	for _, wf := range apiWorkflows {
		versions := versionsByWorkflow[wf.ID]
		latestVersion := 0
		for _, version := range versions {
			if version.Version > latestVersion {
				latestVersion = version.Version
			}
			store.WorkflowVersions = append(store.WorkflowVersions, data.WorkflowVersion{
				ID:             version.ID,
				WorkflowID:     version.WorkflowID,
				Version:        version.Version,
				CreatedAt:      parseTime(version.CreatedAt),
				DefinitionJSON: stringifyJSON(version.Definition, "{}"),
			})
		}

		for _, trigger := range triggersByWorkflow[wf.ID] {
			name := "(unnamed)"
			if trigger.Name != nil && strings.TrimSpace(*trigger.Name) != "" {
				name = *trigger.Name
			}
			triggerTypeByID[trigger.ID] = trigger.Type
			store.Triggers = append(store.Triggers, data.Trigger{
				ID:         trigger.ID,
				WorkflowID: trigger.WorkflowID,
				Key:        trigger.Key,
				Type:       strings.ToLower(trigger.Type),
				Name:       name,
				Active:     trigger.IsActive,
				CreatedAt:  parseTime(trigger.CreatedAt),
				ConfigJSON: stringifyJSON(trigger.Config, "{}"),
			})
		}

		for _, run := range runsByWorkflow[wf.ID] {
			if run.EventID != nil && strings.TrimSpace(*run.EventID) != "" {
				runByEventID[*run.EventID] = run.ID
			}
			triggerType := "manual"
			if run.TriggerID != nil {
				if t, ok := triggerTypeByID[*run.TriggerID]; ok {
					triggerType = strings.ToLower(t)
				}
			}
			startedAt := parseNullableTime(run.StartedAt)
			if startedAt.IsZero() {
				startedAt = parseTime(run.CreatedAt)
			}
			store.Runs = append(store.Runs, data.WorkflowRun{
				ID:          run.ID,
				WorkflowID:  run.WorkflowID,
				Number:      run.Number,
				Status:      run.Status,
				TriggerType: triggerType,
				StartedAt:   startedAt,
				Duration:    durationFromTimes(run.StartedAt, run.FinishedAt),
				InputJSON:   stringifyJSON(run.Input, "{}"),
				OutputJSON:  stringifyJSON(run.Output, "{}"),
				ErrorJSON:   "",
			})
			runIndexByID[run.ID] = len(store.Runs) - 1
		}

		for _, trigger := range triggersByWorkflow[wf.ID] {
			for _, event := range eventsByTrigger[trigger.ID] {
				var runID *string
				if linkedRunID, ok := runByEventID[event.ID]; ok {
					rid := linkedRunID
					runID = &rid
				}
				eventType := "event"
				if event.Type != nil && strings.TrimSpace(*event.Type) != "" {
					eventType = strings.ToLower(*event.Type)
				}
				metadata := map[string]any{}
				if event.ExternalID != nil && strings.TrimSpace(*event.ExternalID) != "" {
					metadata["externalId"] = *event.ExternalID
				}
				store.Events = append(store.Events, data.Event{
					ID:          event.ID,
					TriggerID:   event.TriggerID,
					Type:        eventType,
					ReceivedAt:  parseTime(event.ReceivedAt),
					RunID:       runID,
					PayloadJSON: stringifyJSON(event.Payload, "{}"),
					Metadata:    stringifyJSON(metadata, "{}"),
				})
			}
		}
	}

	for runID, steps := range stepsByRun {
		for _, step := range steps {
			store.StepRuns = append(store.StepRuns, data.StepRun{
				ID:        step.ID,
				RunID:     step.WorkflowRunID,
				StepKey:   step.StepKey,
				Status:    step.Status,
				StartedAt: parseNullableOrCreated(step.StartedAt, step.CreatedAt),
				Duration:  durationFromStep(step),
				Log:       stringifyLog(step.Logs),
			})
			if step.Status == "FAILED" && step.Error != nil {
				if idx, ok := runIndexByID[runID]; ok && strings.TrimSpace(store.Runs[idx].ErrorJSON) == "" {
					store.Runs[idx].ErrorJSON = stringifyJSON(step.Error, "{}")
				}
			}
		}
	}

	for _, secret := range secrets {
		description := ""
		if secret.Description != nil {
			description = *secret.Description
		}
		store.Secrets = append(store.Secrets, data.Secret{
			ID:          secret.ID,
			Name:        secret.Name,
			Description: description,
			CreatedAt:   parseTime(secret.CreatedAt),
		})
	}

	for _, wf := range apiWorkflows {
		latestVersion := 0
		for _, version := range versionsByWorkflow[wf.ID] {
			if version.Version > latestVersion {
				latestVersion = version.Version
			}
		}
		store.Workflows = append(store.Workflows, data.Workflow{
			ID:            wf.ID,
			Key:           wf.Key,
			Name:          wf.Name,
			Active:        wf.IsActive,
			LatestVersion: latestVersion,
			UpdatedAt:     parseTime(wf.UpdatedAt),
		})
	}

	return store, nil
}

func listAllWorkflows(client *api.Client, sortSpec apiListSort) ([]api.Workflow, error) {
	items := make([]api.Workflow, 0)
	page := 1
	for {
		result, err := client.ListWorkflows(page, apiPageSize, sortSpec.By, sortSpec.Order)
		if err != nil {
			return nil, err
		}
		items = append(items, result.Items...)
		if !result.Pagination.HasNext {
			break
		}
		page++
	}
	return items, nil
}

func listAllWorkflowVersions(client *api.Client, workflowID string, sortSpec apiListSort) ([]api.WorkflowVersion, error) {
	items := make([]api.WorkflowVersion, 0)
	page := 1
	for {
		result, err := client.ListWorkflowVersions(workflowID, page, apiPageSize, sortSpec.By, sortSpec.Order)
		if err != nil {
			return nil, err
		}
		items = append(items, result.Items...)
		if !result.Pagination.HasNext {
			break
		}
		page++
	}
	return items, nil
}

func listAllTriggers(client *api.Client, workflowID string, sortSpec apiListSort) ([]api.Trigger, error) {
	items := make([]api.Trigger, 0)
	page := 1
	for {
		result, err := client.ListTriggers(workflowID, page, apiPageSize, sortSpec.By, sortSpec.Order)
		if err != nil {
			return nil, err
		}
		items = append(items, result.Items...)
		if !result.Pagination.HasNext {
			break
		}
		page++
	}
	return items, nil
}

func listAllWorkflowRuns(client *api.Client, workflowID string, sortSpec apiListSort) ([]api.WorkflowRun, error) {
	items := make([]api.WorkflowRun, 0)
	page := 1
	for {
		result, err := client.ListWorkflowRuns(workflowID, page, apiPageSize, sortSpec.By, sortSpec.Order)
		if err != nil {
			return nil, err
		}
		items = append(items, result.Items...)
		if !result.Pagination.HasNext {
			break
		}
		page++
	}
	return items, nil
}

func listAllEvents(client *api.Client, workflowID string, triggerID string, sortSpec apiListSort) ([]api.Event, error) {
	items := make([]api.Event, 0)
	page := 1
	for {
		result, err := client.ListEvents(workflowID, triggerID, page, apiPageSize, sortSpec.By, sortSpec.Order)
		if err != nil {
			return nil, err
		}
		items = append(items, result.Items...)
		if !result.Pagination.HasNext {
			break
		}
		page++
	}
	return items, nil
}

func listAllStepRuns(client *api.Client, workflowID string, runID string, sortSpec apiListSort) ([]api.StepRun, error) {
	items := make([]api.StepRun, 0)
	page := 1
	for {
		result, err := client.ListStepRuns(workflowID, runID, page, apiPageSize, sortSpec.By, sortSpec.Order)
		if err != nil {
			return nil, err
		}
		items = append(items, result.Items...)
		if !result.Pagination.HasNext {
			break
		}
		page++
	}
	return items, nil
}

func listAllSecrets(client *api.Client, sortSpec apiListSort) ([]api.Secret, error) {
	items := make([]api.Secret, 0)
	page := 1
	for {
		result, err := client.ListSecrets(page, apiPageSize, sortSpec.By, sortSpec.Order)
		if err != nil {
			return nil, err
		}
		items = append(items, result.Items...)
		if !result.Pagination.HasNext {
			break
		}
		page++
	}
	return items, nil
}

func parseTime(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	if ts, err := time.Parse(time.RFC3339Nano, value); err == nil {
		return ts
	}
	if ts, err := time.Parse(time.RFC3339, value); err == nil {
		return ts
	}
	return time.Time{}
}

func parseNullableTime(value *string) time.Time {
	if value == nil {
		return time.Time{}
	}
	return parseTime(*value)
}

func parseNullableOrCreated(started *string, created string) time.Time {
	ts := parseNullableTime(started)
	if ts.IsZero() {
		return parseTime(created)
	}
	return ts
}

func durationFromTimes(started *string, finished *string) time.Duration {
	start := parseNullableTime(started)
	end := parseNullableTime(finished)
	if start.IsZero() || end.IsZero() {
		return 0
	}
	if end.Before(start) {
		return 0
	}
	return end.Sub(start)
}

func durationFromStep(step api.StepRun) time.Duration {
	if step.DurationMs != nil && *step.DurationMs > 0 {
		return time.Duration(*step.DurationMs) * time.Millisecond
	}
	return durationFromTimes(step.StartedAt, step.FinishedAt)
}

func stringifyJSON(value any, fallback string) string {
	if value == nil {
		return fallback
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return fallback
	}
	if strings.TrimSpace(string(encoded)) == "" {
		return fallback
	}
	return string(encoded)
}

func stringifyLog(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	default:
		encoded, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return ""
		}
		return string(encoded)
	}
}
