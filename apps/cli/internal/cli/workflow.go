package cli

import (
	"fmt"
	"os"

	"github.com/gentij/lunie/apps/cli/internal/api"
	"github.com/gentij/lunie/apps/cli/internal/output"
	"github.com/spf13/cobra"
)

var workflowCmd = &cobra.Command{
	Use:   "workflow",
	Short: "Manage workflows",
}

var workflowListPage int
var workflowListPageSize int
var workflowListSortBy string
var workflowListSortOrder string
var workflowCreateName string
var workflowCreateDefinition string
var workflowUpdateName string
var workflowUpdateIsActive bool
var workflowRunInput string
var workflowRunOverrides string
var workflowValidateDefinition string

func init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List workflows",
		RunE:  workflowList,
	}
	listCmd.Flags().IntVar(&workflowListPage, "page", 1, "Page number")
	listCmd.Flags().IntVar(&workflowListPageSize, "page-size", 25, "Page size")
	listCmd.Flags().StringVar(&workflowListSortBy, "sort-by", "createdAt", "Sort field (createdAt|updatedAt)")
	listCmd.Flags().StringVar(&workflowListSortOrder, "sort-order", "desc", "Sort order (asc|desc)")

	getCmd := &cobra.Command{
		Use:   "get <workflow-id>",
		Short: "Get a workflow",
		Args:  cobra.ExactArgs(1),
		RunE:  workflowGet,
	}

	workflowCmd.AddCommand(listCmd)
	workflowCmd.AddCommand(getCmd)
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a workflow",
		RunE:  workflowCreate,
	}
	createCmd.Flags().StringVar(&workflowCreateName, "name", "", "Workflow name")
	createCmd.Flags().StringVar(&workflowCreateDefinition, "definition", "", "Path to definition JSON")
	_ = createCmd.MarkFlagRequired("name")
	_ = createCmd.MarkFlagRequired("definition")

	updateCmd := &cobra.Command{
		Use:   "update <workflow-id>",
		Short: "Update a workflow",
		Args:  cobra.ExactArgs(1),
		RunE:  workflowUpdate,
	}
	updateCmd.Flags().StringVar(&workflowUpdateName, "name", "", "Workflow name")
	updateCmd.Flags().BoolVar(&workflowUpdateIsActive, "is-active", false, "Set workflow active state")

	deleteCmd := &cobra.Command{
		Use:   "delete <workflow-id>",
		Short: "Delete a workflow (soft)",
		Args:  cobra.ExactArgs(1),
		RunE:  workflowDelete,
	}

	runCmd := &cobra.Command{
		Use:   "run <workflow-id>",
		Short: "Run a workflow",
		Args:  cobra.ExactArgs(1),
		RunE:  workflowRun,
	}
	runCmd.Flags().StringVar(&workflowRunInput, "input", "", "Path to input JSON (overrides workflow definition input defaults)")
	runCmd.Flags().StringVar(&workflowRunOverrides, "overrides", "", "Path to HTTP step request overrides JSON")

	workflowCmd.AddCommand(createCmd)
	workflowCmd.AddCommand(updateCmd)
	workflowCmd.AddCommand(deleteCmd)
	workflowCmd.AddCommand(runCmd)
	workflowCmd.AddCommand(workflowVersionCmd)

	validateCmd := &cobra.Command{
		Use:   "validate <workflow-id>",
		Short: "Validate a workflow definition",
		Args:  cobra.ExactArgs(1),
		RunE:  workflowValidate,
	}
	validateCmd.Flags().StringVar(&workflowValidateDefinition, "definition", "", "Path to definition JSON")
	_ = validateCmd.MarkFlagRequired("definition")
	workflowCmd.AddCommand(validateCmd)
}

func workflowList(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	result, err := ctx.Client.ListWorkflows(workflowListPage, workflowListPageSize, workflowListSortBy, workflowListSortOrder)
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
		latest := ""
		if item.LatestVersionID != nil {
			latest = *item.LatestVersionID
		}
		rows = append(rows, []string{item.ID, item.Name, fmt.Sprintf("%t", item.IsActive), latest})
	}

	return output.PrintListTable([]string{"ID", "NAME", "ACTIVE", "LATEST_VERSION"}, rows)
}

func workflowGet(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	result, err := ctx.Client.GetWorkflow(args[0])
	if err != nil {
		return err
	}

	return printWorkflow(ctx, result)
}

func workflowCreate(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	definition, err := readJSONFile(workflowCreateDefinition)
	if err != nil {
		return err
	}

	result, err := ctx.Client.CreateWorkflow(workflowCreateName, definition)
	if err != nil {
		return err
	}

	return printWorkflow(ctx, result)
}

func workflowUpdate(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	patch := map[string]any{}
	if workflowUpdateName != "" {
		patch["name"] = workflowUpdateName
	}
	if cmd.Flags().Changed("is-active") {
		patch["isActive"] = workflowUpdateIsActive
	}
	if len(patch) == 0 {
		return fmt.Errorf("no fields to update")
	}

	result, err := ctx.Client.UpdateWorkflow(args[0], patch)
	if err != nil {
		return err
	}

	return printWorkflow(ctx, result)
}

func workflowDelete(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	result, err := ctx.Client.DeleteWorkflow(args[0])
	if err != nil {
		return err
	}

	return printWorkflow(ctx, result)
}

func workflowRun(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	input, err := readOptionalJSONFile(workflowRunInput)
	if err != nil {
		return err
	}

	overrides, err := readOptionalJSONFile(workflowRunOverrides)
	if err != nil {
		return err
	}

	result, err := ctx.Client.RunWorkflow(args[0], input, overrides)
	if err != nil {
		return err
	}

	if IsJSON(ctx) {
		return output.PrintJSON(result)
	}
	if ctx.Quiet {
		fmt.Fprintln(os.Stdout, result["workflowRunId"])
		return nil
	}

	return output.PrintKVTable([][2]string{{"workflowRunId", result["workflowRunId"]}, {"status", result["status"]}})
}

func workflowValidate(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	definition, err := readJSONFile(workflowValidateDefinition)
	if err != nil {
		return err
	}

	result, err := ctx.Client.ValidateWorkflow(args[0], definition)
	if err != nil {
		return err
	}

	if IsJSON(ctx) {
		return output.PrintJSON(result)
	}
	if ctx.Quiet {
		fmt.Fprintln(os.Stdout, result.Valid)
		return nil
	}

	issueCount := 0
	if result.Issues != nil {
		issueCount = len(result.Issues)
	}

	return output.PrintKVTable([][2]string{{"valid", fmt.Sprintf("%t", result.Valid)}, {"issues", fmt.Sprintf("%d", issueCount)}})
}

func printWorkflow(ctx *Context, result api.Workflow) error {
	if IsJSON(ctx) {
		return output.PrintJSON(result)
	}

	if ctx.Quiet {
		fmt.Fprintln(os.Stdout, result.ID)
		return nil
	}

	latest := ""
	if result.LatestVersionID != nil {
		latest = *result.LatestVersionID
	}

	return output.PrintKVTable([][2]string{
		{"id", result.ID},
		{"name", result.Name},
		{"isActive", fmt.Sprintf("%t", result.IsActive)},
		{"latestVersionId", latest},
		{"createdAt", result.CreatedAt},
		{"updatedAt", result.UpdatedAt},
	})
}
