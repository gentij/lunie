package cli

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/gentij/lunie/apps/cli/internal/api"
	"github.com/gentij/lunie/apps/cli/internal/output"
	"github.com/spf13/cobra"
)

var triggerCmd = &cobra.Command{
	Use:   "trigger",
	Short: "Manage triggers",
}

var triggerListPage int
var triggerListPageSize int
var triggerListSortBy string
var triggerListSortOrder string
var triggerCreateType string
var triggerCreateName string
var triggerCreateIsActive bool
var triggerCreateConfig string
var triggerUpdateName string
var triggerUpdateIsActive bool
var triggerUpdateConfig string
var triggerWebhookPublicBase string

func init() {
	listCmd := &cobra.Command{
		Use:   "list <workflow-id>",
		Short: "List triggers",
		Args:  cobra.ExactArgs(1),
		RunE:  triggerList,
	}
	listCmd.Flags().IntVar(&triggerListPage, "page", 1, "Page number")
	listCmd.Flags().IntVar(&triggerListPageSize, "page-size", 25, "Page size")
	listCmd.Flags().StringVar(&triggerListSortBy, "sort-by", "createdAt", "Sort field (createdAt|updatedAt)")
	listCmd.Flags().StringVar(&triggerListSortOrder, "sort-order", "desc", "Sort order (asc|desc)")

	getCmd := &cobra.Command{
		Use:   "get <workflow-id> <trigger-id>",
		Short: "Get a trigger",
		Args:  cobra.ExactArgs(2),
		RunE:  triggerGet,
	}

	createCmd := &cobra.Command{
		Use:   "create <workflow-id>",
		Short: "Create a trigger",
		Args:  cobra.ExactArgs(1),
		RunE:  triggerCreate,
	}
	createCmd.Flags().StringVar(&triggerCreateType, "type", "", "Trigger type (MANUAL|CRON|WEBHOOK)")
	createCmd.Flags().StringVar(&triggerCreateName, "name", "", "Trigger name")
	createCmd.Flags().BoolVar(&triggerCreateIsActive, "is-active", true, "Trigger active state")
	createCmd.Flags().StringVar(&triggerCreateConfig, "config", "", "Path to config JSON")
	_ = createCmd.MarkFlagRequired("type")

	updateCmd := &cobra.Command{
		Use:   "update <workflow-id> <trigger-id>",
		Short: "Update a trigger",
		Args:  cobra.ExactArgs(2),
		RunE:  triggerUpdate,
	}
	updateCmd.Flags().StringVar(&triggerUpdateName, "name", "", "Trigger name")
	updateCmd.Flags().BoolVar(&triggerUpdateIsActive, "is-active", false, "Set trigger active state")
	updateCmd.Flags().StringVar(&triggerUpdateConfig, "config", "", "Path to config JSON")

	deleteCmd := &cobra.Command{
		Use:   "delete <workflow-id> <trigger-id>",
		Short: "Delete a trigger (soft)",
		Args:  cobra.ExactArgs(2),
		RunE:  triggerDelete,
	}

	webhookCmd := &cobra.Command{
		Use:   "webhook",
		Short: "Webhook utilities",
	}

	rotateKeyCmd := &cobra.Command{
		Use:   "rotate-key <workflow-id> <trigger-id>",
		Short: "Rotate webhook key and print webhook URL",
		Args:  cobra.ExactArgs(2),
		RunE:  triggerWebhookRotateKey,
	}
	rotateKeyCmd.Flags().StringVar(&triggerWebhookPublicBase, "public-base", "", "Public API base URL override")
	webhookCmd.AddCommand(rotateKeyCmd)

	triggerCmd.AddCommand(listCmd)
	triggerCmd.AddCommand(getCmd)
	triggerCmd.AddCommand(createCmd)
	triggerCmd.AddCommand(updateCmd)
	triggerCmd.AddCommand(deleteCmd)
	triggerCmd.AddCommand(webhookCmd)
}

func triggerList(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	workflowID := args[0]
	result, err := ctx.Client.ListTriggers(workflowID, triggerListPage, triggerListPageSize, triggerListSortBy, triggerListSortOrder)
	if err != nil {
		return err
	}

	if IsJSON(ctx) {
		return output.PrintJSON(result)
	}

	if ctx.Quiet {
		for _, item := range result.Items {
			fmt.Fprintln(os.Stdout, item.ID)
		}
		return nil
	}

	rows := make([][]string, 0, len(result.Items))
	for _, item := range result.Items {
		rows = append(rows, []string{item.ID, item.Type, triggerNameValue(item.Name), output.BoolLabel(item.IsActive)})
	}
	if err := output.PrintListTable([]string{"ID", "TYPE", "NAME", "ACTIVE"}, rows); err != nil {
		return err
	}
	return output.PrintPagination(result.Pagination)
}

func triggerGet(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	workflowID := args[0]
	triggerID := args[1]
	result, err := ctx.Client.GetTrigger(workflowID, triggerID)
	if err != nil {
		return err
	}

	return printTrigger(ctx, result)
}

func triggerCreate(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	workflowID := args[0]
	configValue, err := readOptionalJSONFile(triggerCreateConfig)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"type":     strings.ToUpper(triggerCreateType),
		"name":     triggerCreateName,
		"isActive": triggerCreateIsActive,
		"config":   configValue,
	}

	result, err := ctx.Client.CreateTrigger(workflowID, payload)
	if err != nil {
		return err
	}

	return printTrigger(ctx, result)
}

func triggerUpdate(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	workflowID := args[0]
	triggerID := args[1]

	patch := map[string]any{}
	if triggerUpdateName != "" {
		patch["name"] = triggerUpdateName
	}
	if cmd.Flags().Changed("is-active") {
		patch["isActive"] = triggerUpdateIsActive
	}
	if strings.TrimSpace(triggerUpdateConfig) != "" {
		configValue, err := readJSONFile(triggerUpdateConfig)
		if err != nil {
			return err
		}
		patch["config"] = configValue
	}
	if len(patch) == 0 {
		return fmt.Errorf("no fields to update")
	}

	result, err := ctx.Client.UpdateTrigger(workflowID, triggerID, patch)
	if err != nil {
		return err
	}

	return printTrigger(ctx, result)
}

func triggerDelete(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	workflowID := args[0]
	triggerID := args[1]
	result, err := ctx.Client.DeleteTrigger(workflowID, triggerID)
	if err != nil {
		return err
	}

	return printTrigger(ctx, result)
}

func triggerWebhookRotateKey(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	workflowID := args[0]
	triggerID := args[1]

	result, err := ctx.Client.RotateTriggerWebhookKey(workflowID, triggerID)
	if err != nil {
		return err
	}

	webhookURL, err := buildWebhookURL(
		ctx.Client.BaseURL,
		workflowID,
		triggerID,
		result.WebhookKey,
		triggerWebhookPublicBase,
	)
	if err != nil {
		return err
	}

	if IsJSON(ctx) {
		return output.PrintJSON(map[string]string{
			"workflowId": workflowID,
			"triggerId":  triggerID,
			"webhookKey": result.WebhookKey,
			"webhookUrl": webhookURL,
		})
	}

	if ctx.Quiet {
		fmt.Fprintln(os.Stdout, webhookURL)
		return nil
	}

	return output.PrintKVTable([][2]string{
		{"workflowId", workflowID},
		{"triggerId", triggerID},
		{"webhookKey", result.WebhookKey},
		{"webhookUrl", webhookURL},
	})
}

func printTrigger(ctx *Context, result api.Trigger) error {
	if IsJSON(ctx) {
		return output.PrintJSON(result)
	}
	if ctx.Quiet {
		fmt.Fprintln(os.Stdout, result.ID)
		return nil
	}

	configData, err := json.Marshal(result.Config)
	configValue := ""
	if err == nil {
		configValue = string(configData)
	}

	return output.PrintKVTable([][2]string{
		{"id", result.ID},
		{"workflowId", result.WorkflowID},
		{"type", result.Type},
		{"name", triggerNameValue(result.Name)},
		{"isActive", output.BoolLabel(result.IsActive)},
		{"config", configValue},
		{"createdAt", result.CreatedAt},
		{"updatedAt", result.UpdatedAt},
	})
}

func triggerNameValue(name *string) string {
	if name == nil {
		return ""
	}
	return *name
}

func buildWebhookURL(apiBase string, workflowID string, triggerID string, webhookKey string, publicBase string) (string, error) {
	apiBase = strings.TrimSpace(apiBase)
	apiParsed, err := url.Parse(apiBase)
	if err != nil {
		return "", err
	}
	if apiParsed.Scheme == "" || apiParsed.Host == "" {
		return "", fmt.Errorf("invalid API base URL: %s", apiBase)
	}

	base := strings.TrimSpace(publicBase)
	if base == "" {
		base = apiBase
	}

	parsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("invalid base URL: %s", base)
	}

	basePath := strings.TrimRight(parsed.Path, "/")
	if strings.TrimSpace(publicBase) != "" && (basePath == "" || basePath == "/") {
		basePath = strings.TrimRight(apiParsed.Path, "/")
	}

	parsed.Path = fmt.Sprintf(
		"%s/hooks/%s/%s/%s",
		basePath,
		url.PathEscape(workflowID),
		url.PathEscape(triggerID),
		url.PathEscape(webhookKey),
	)
	parsed.RawQuery = ""
	parsed.Fragment = ""

	return parsed.String(), nil
}
