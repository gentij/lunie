package cli

import (
	"fmt"
	"os"

	"github.com/gentij/lunie/apps/cli/internal/api"
	"github.com/gentij/lunie/apps/cli/internal/output"
	"github.com/spf13/cobra"
)

var workflowVersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Manage workflow versions",
}

var workflowVersionListPage int
var workflowVersionListPageSize int
var workflowVersionListSortBy string
var workflowVersionListSortOrder string
var workflowVersionCreateDefinition string

func init() {
	listCmd := &cobra.Command{
		Use:   "list <workflow-id>",
		Short: "List workflow versions",
		Args:  cobra.ExactArgs(1),
		RunE:  workflowVersionList,
	}
	listCmd.Flags().IntVar(&workflowVersionListPage, "page", 1, "Page number")
	listCmd.Flags().IntVar(&workflowVersionListPageSize, "page-size", 25, "Page size")
	listCmd.Flags().StringVar(&workflowVersionListSortBy, "sort-by", "version", "Sort field (version|createdAt)")
	listCmd.Flags().StringVar(&workflowVersionListSortOrder, "sort-order", "desc", "Sort order (asc|desc)")

	getCmd := &cobra.Command{
		Use:   "get <workflow-id> <version>",
		Short: "Get a workflow version",
		Args:  cobra.ExactArgs(2),
		RunE:  workflowVersionGet,
	}

	createCmd := &cobra.Command{
		Use:   "create <workflow-id>",
		Short: "Create a workflow version",
		Args:  cobra.ExactArgs(1),
		RunE:  workflowVersionCreate,
	}
	createCmd.Flags().StringVar(&workflowVersionCreateDefinition, "definition", "", "Path to definition JSON")
	_ = createCmd.MarkFlagRequired("definition")

	workflowVersionCmd.AddCommand(listCmd)
	workflowVersionCmd.AddCommand(getCmd)
	workflowVersionCmd.AddCommand(createCmd)
}

func workflowVersionList(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	workflowID := args[0]
	result, err := ctx.Client.ListWorkflowVersions(workflowID, workflowVersionListPage, workflowVersionListPageSize, workflowVersionListSortBy, workflowVersionListSortOrder)
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
		rows = append(rows, []string{fmt.Sprintf("%d", item.Version), item.ID, item.CreatedAt})
	}
	if err := output.PrintListTable([]string{"VERSION", "ID", "CREATED"}, rows); err != nil {
		return err
	}
	return output.PrintPagination(result.Pagination)
}

func workflowVersionGet(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	workflowID := args[0]
	version := args[1]
	result, err := ctx.Client.GetWorkflowVersion(workflowID, version)
	if err != nil {
		return err
	}

	return printWorkflowVersion(ctx, result)
}

func workflowVersionCreate(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	workflowID := args[0]
	definition, err := readJSONFile(workflowVersionCreateDefinition)
	if err != nil {
		return err
	}

	result, err := ctx.Client.CreateWorkflowVersion(workflowID, definition)
	if err != nil {
		return err
	}

	return printWorkflowVersion(ctx, result)
}

func printWorkflowVersion(ctx *Context, result api.WorkflowVersion) error {
	if IsJSON(ctx) {
		return output.PrintJSON(result)
	}
	if ctx.Quiet {
		fmt.Fprintln(os.Stdout, result.ID)
		return nil
	}

	return output.PrintKVTable([][2]string{
		{"id", result.ID},
		{"workflowId", result.WorkflowID},
		{"version", fmt.Sprintf("%d", result.Version)},
		{"createdAt", result.CreatedAt},
	})
}
