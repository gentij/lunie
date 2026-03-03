package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func AsAPIError(err error) *APIError {
	if err == nil {
		return nil
	}
	if apiErr, ok := err.(*APIError); ok {
		return apiErr
	}
	return nil
}

type Envelope struct {
	Ok         bool            `json:"ok"`
	StatusCode int             `json:"statusCode"`
	Path       string          `json:"path"`
	Timestamp  string          `json:"timestamp"`
	Data       json.RawMessage `json:"data"`
	Error      *APIError       `json:"error,omitempty"`
}

func NewClient(baseURL string, token string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) GetJSON(path string, out any) error {
	return c.doJSON(http.MethodGet, path, nil, out)
}

func (c *Client) PostJSON(path string, body any, out any) error {
	return c.doJSON(http.MethodPost, path, body, out)
}

func (c *Client) PatchJSON(path string, body any, out any) error {
	return c.doJSON(http.MethodPatch, path, body, out)
}

func (c *Client) DeleteJSON(path string, out any) error {
	return c.doJSON(http.MethodDelete, path, nil, out)
}

type validateResponse struct {
	Valid                bool                `json:"valid"`
	Issues               []any               `json:"issues"`
	InferredDependencies map[string][]string `json:"inferredDependencies"`
	ExecutionBatches     [][]string          `json:"executionBatches"`
	ReferencedSecrets    []string            `json:"referencedSecrets"`
}

func (c *Client) ListWorkflows(page int, pageSize int) (Paginated[Workflow], error) {
	var result Paginated[Workflow]
	path := fmt.Sprintf("/workflows?page=%d&pageSize=%d", page, pageSize)
	return result, c.GetJSON(path, &result)
}

func (c *Client) GetWorkflow(id string) (Workflow, error) {
	var result Workflow
	return result, c.GetJSON("/workflows/"+id, &result)
}

func (c *Client) CreateWorkflow(name string, definition any) (Workflow, error) {
	var result Workflow
	payload := map[string]any{"name": name, "definition": definition}
	return result, c.PostJSON("/workflows", payload, &result)
}

func (c *Client) UpdateWorkflow(id string, patch map[string]any) (Workflow, error) {
	var result Workflow
	return result, c.PatchJSON("/workflows/"+id, patch, &result)
}

func (c *Client) DeleteWorkflow(id string) (Workflow, error) {
	var result Workflow
	return result, c.DeleteJSON("/workflows/"+id, &result)
}

func (c *Client) RunWorkflow(id string, input any, overrides any) (map[string]string, error) {
	var result struct {
		WorkflowRunID string `json:"workflowRunId"`
		Status        string `json:"status"`
	}
	payload := map[string]any{"input": input, "overrides": overrides}
	err := c.PostJSON("/workflows/"+id+"/run", payload, &result)
	if err != nil {
		return nil, err
	}
	return map[string]string{"workflowRunId": result.WorkflowRunID, "status": result.Status}, nil
}

func (c *Client) ValidateWorkflow(id string, definition any) (validateResponse, error) {
	var result validateResponse
	payload := map[string]any{"definition": definition}
	return result, c.PostJSON("/workflows/"+id+"/versions/validate", payload, &result)
}

func (c *Client) ListWorkflowVersions(workflowID string, page int, pageSize int) (Paginated[WorkflowVersion], error) {
	var result Paginated[WorkflowVersion]
	path := fmt.Sprintf("/workflows/%s/versions?page=%d&pageSize=%d", workflowID, page, pageSize)
	return result, c.GetJSON(path, &result)
}

func (c *Client) GetWorkflowVersion(workflowID string, version string) (WorkflowVersion, error) {
	var result WorkflowVersion
	return result, c.GetJSON("/workflows/"+workflowID+"/versions/"+version, &result)
}

func (c *Client) CreateWorkflowVersion(workflowID string, definition any) (WorkflowVersion, error) {
	var result WorkflowVersion
	payload := map[string]any{"definition": definition}
	return result, c.PostJSON("/workflows/"+workflowID+"/versions", payload, &result)
}

func (c *Client) ListTriggers(workflowID string, page int, pageSize int) (Paginated[Trigger], error) {
	var result Paginated[Trigger]
	path := fmt.Sprintf("/workflows/%s/triggers?page=%d&pageSize=%d", workflowID, page, pageSize)
	return result, c.GetJSON(path, &result)
}

func (c *Client) GetTrigger(workflowID string, triggerID string) (Trigger, error) {
	var result Trigger
	return result, c.GetJSON("/workflows/"+workflowID+"/triggers/"+triggerID, &result)
}

func (c *Client) CreateTrigger(workflowID string, payload map[string]any) (Trigger, error) {
	var result Trigger
	return result, c.PostJSON("/workflows/"+workflowID+"/triggers", payload, &result)
}

func (c *Client) UpdateTrigger(workflowID string, triggerID string, patch map[string]any) (Trigger, error) {
	var result Trigger
	return result, c.PatchJSON("/workflows/"+workflowID+"/triggers/"+triggerID, patch, &result)
}

func (c *Client) DeleteTrigger(workflowID string, triggerID string) (Trigger, error) {
	var result Trigger
	return result, c.DeleteJSON("/workflows/"+workflowID+"/triggers/"+triggerID, &result)
}

func (c *Client) ListWorkflowRuns(workflowID string, page int, pageSize int) (Paginated[WorkflowRun], error) {
	var result Paginated[WorkflowRun]
	path := fmt.Sprintf("/workflows/%s/runs?page=%d&pageSize=%d", workflowID, page, pageSize)
	return result, c.GetJSON(path, &result)
}

func (c *Client) GetWorkflowRun(workflowID string, runID string) (WorkflowRun, error) {
	var result WorkflowRun
	return result, c.GetJSON("/workflows/"+workflowID+"/runs/"+runID, &result)
}

func (c *Client) ListStepRuns(workflowID string, runID string, page int, pageSize int) (Paginated[StepRun], error) {
	var result Paginated[StepRun]
	path := fmt.Sprintf("/workflows/%s/runs/%s/steps?page=%d&pageSize=%d", workflowID, runID, page, pageSize)
	return result, c.GetJSON(path, &result)
}

func (c *Client) GetStepRun(workflowID string, runID string, stepID string) (StepRun, error) {
	var result StepRun
	return result, c.GetJSON("/workflows/"+workflowID+"/runs/"+runID+"/steps/"+stepID, &result)
}

func (c *Client) ListEvents(workflowID string, triggerID string, page int, pageSize int) (Paginated[Event], error) {
	var result Paginated[Event]
	path := fmt.Sprintf("/workflows/%s/triggers/%s/events?page=%d&pageSize=%d", workflowID, triggerID, page, pageSize)
	return result, c.GetJSON(path, &result)
}

func (c *Client) GetEvent(workflowID string, triggerID string, eventID string) (Event, error) {
	var result Event
	path := "/workflows/" + workflowID + "/triggers/" + triggerID + "/events/" + eventID
	return result, c.GetJSON(path, &result)
}

func (c *Client) ListSecrets(page int, pageSize int) (Paginated[Secret], error) {
	var result Paginated[Secret]
	path := fmt.Sprintf("/secrets?page=%d&pageSize=%d", page, pageSize)
	return result, c.GetJSON(path, &result)
}

func (c *Client) GetSecret(id string) (Secret, error) {
	var result Secret
	return result, c.GetJSON("/secrets/"+id, &result)
}

func (c *Client) CreateSecret(payload map[string]any) (Secret, error) {
	var result Secret
	return result, c.PostJSON("/secrets", payload, &result)
}

func (c *Client) UpdateSecret(id string, patch map[string]any) (Secret, error) {
	var result Secret
	return result, c.PatchJSON("/secrets/"+id, patch, &result)
}

func (c *Client) DeleteSecret(id string) (Secret, error) {
	var result Secret
	return result, c.DeleteJSON("/secrets/"+id, &result)
}

func (c *Client) GetHealth() (Health, error) {
	var result Health
	if err := c.GetJSON("/health", &result); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) WhoAmI() (WhoAmI, error) {
	var result WhoAmI
	if err := c.GetJSON("/auth/whoami", &result); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) doJSON(method string, path string, body any, out any) error {
	fullURL, err := c.buildURL(path)
	if err != nil {
		return err
	}

	var bodyReader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(payload)
	}

	req, err := http.NewRequest(method, fullURL, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if strings.TrimSpace(c.Token) != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return fmt.Errorf("empty response body (status %d)", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "application/json") {
		snippet := string(data)
		if len(snippet) > 200 {
			snippet = snippet[:200] + "..."
		}
		return fmt.Errorf(
			"unexpected content type %q (status %d): %s",
			contentType,
			resp.StatusCode,
			snippet,
		)
	}

	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return err
	}

	if !env.Ok {
		if env.Error != nil {
			return env.Error
		}
		return errors.New("request failed")
	}

	if out != nil && len(env.Data) > 0 {
		if err := json.Unmarshal(env.Data, out); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) buildURL(path string) (string, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path, nil
	}

	base, err := url.Parse(c.BaseURL + "/")
	if err != nil {
		return "", err
	}

	ref, err := url.Parse(strings.TrimLeft(path, "/"))
	if err != nil {
		return "", err
	}

	return base.ResolveReference(ref).String(), nil
}
