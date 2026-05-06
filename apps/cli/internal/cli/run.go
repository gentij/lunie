package cli

import (
	"fmt"
	"os"

	"github.com/gentij/lunie/apps/cli/internal/output"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Manage workflow runs",
}

var runListPage int
var runListPageSize int
var runListSortBy string
var runListSortOrder string

func init() {
	listCmd := &cobra.Command{
		Use:   "list <workflow-key>",
		Short: "List workflow runs",
		Args:  cobra.ExactArgs(1),
		RunE:  runList,
	}
	listCmd.Flags().IntVar(&runListPage, "page", 1, "Page number")
	listCmd.Flags().IntVar(&runListPageSize, "page-size", 25, "Page size")
	listCmd.Flags().StringVar(&runListSortBy, "sort-by", "createdAt", "Sort field (createdAt|updatedAt)")
	listCmd.Flags().StringVar(&runListSortOrder, "sort-order", "desc", "Sort order (asc|desc)")

	getCmd := &cobra.Command{
		Use:   "get <workflow-key> <run-number>",
		Short: "Get a workflow run",
		Args:  cobra.ExactArgs(2),
		RunE:  runGet,
	}

	runCmd.AddCommand(listCmd)
	runCmd.AddCommand(getCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	workflowKey := args[0]
	result, err := ctx.Client.ListWorkflowRunsByKey(workflowKey, runListPage, runListPageSize, runListSortBy, runListSortOrder)
	if err != nil {
		return err
	}

	if IsJSON(ctx) {
		return output.PrintJSON(result)
	}

	if ctx.Quiet {
		for _, item := range result.Items {
			fmt.Fprintln(os.Stdout, item.Number)
		}
		return nil
	}

	rows := make([][]string, 0, len(result.Items))
	for _, item := range result.Items {
		rows = append(rows, []string{fmt.Sprintf("%d", item.Number), output.ColorStatus(item.Status), item.WorkflowVersionID, item.CreatedAt})
	}
	if err := output.PrintListTable([]string{"RUN", "STATUS", "VERSION", "CREATED"}, rows); err != nil {
		return err
	}
	return output.PrintPagination(result.Pagination)
}

func runGet(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	workflowKey := args[0]
	runNumber, err := parsePositiveIntArg("run number", args[1])
	if err != nil {
		return err
	}

	result, err := ctx.Client.GetWorkflowRunByNumber(workflowKey, runNumber)
	if err != nil {
		return err
	}

	if IsJSON(ctx) {
		return output.PrintJSON(result)
	}
	if ctx.Quiet {
		fmt.Fprintln(os.Stdout, result.Number)
		return nil
	}

	started := ""
	if result.StartedAt != nil {
		started = *result.StartedAt
	}
	finished := ""
	if result.FinishedAt != nil {
		finished = *result.FinishedAt
	}

	return output.PrintKVTable([][2]string{
		{"number", fmt.Sprintf("%d", result.Number)},
		{"id", result.ID},
		{"workflowId", result.WorkflowID},
		{"workflowVersionId", result.WorkflowVersionID},
		{"status", output.ColorStatus(result.Status)},
		{"createdAt", result.CreatedAt},
		{"startedAt", started},
		{"finishedAt", finished},
	})
}
