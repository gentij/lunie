package cli

import (
	"fmt"
	"os"

	"github.com/gentij/lunie/apps/cli/internal/output"
	"github.com/spf13/cobra"
)

var stepCmd = &cobra.Command{
	Use:   "step",
	Short: "Manage step runs",
}

var stepListPage int
var stepListPageSize int
var stepListSortBy string
var stepListSortOrder string

func init() {
	listCmd := &cobra.Command{
		Use:   "list <workflow-key> <run-number>",
		Short: "List step runs",
		Args:  cobra.ExactArgs(2),
		RunE:  stepList,
	}
	listCmd.Flags().IntVar(&stepListPage, "page", 1, "Page number")
	listCmd.Flags().IntVar(&stepListPageSize, "page-size", 25, "Page size")
	listCmd.Flags().StringVar(&stepListSortBy, "sort-by", "createdAt", "Sort field (createdAt|updatedAt)")
	listCmd.Flags().StringVar(&stepListSortOrder, "sort-order", "asc", "Sort order (asc|desc)")

	getCmd := &cobra.Command{
		Use:   "get <workflow-key> <run-number> <step-key>",
		Short: "Get a step run",
		Args:  cobra.ExactArgs(3),
		RunE:  stepGet,
	}

	stepCmd.AddCommand(listCmd)
	stepCmd.AddCommand(getCmd)
}

func stepList(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	workflowKey := args[0]
	runNumber, err := parsePositiveIntArg("run number", args[1])
	if err != nil {
		return err
	}

	result, err := ctx.Client.ListStepRunsByWorkflowKeyAndRunNumber(workflowKey, runNumber, stepListPage, stepListPageSize, stepListSortBy, stepListSortOrder)
	if err != nil {
		return err
	}

	if IsJSON(ctx) {
		return output.PrintJSON(result)
	}

	if ctx.Quiet {
		for _, item := range result.Items {
			fmt.Fprintln(os.Stdout, item.StepKey)
		}
		return nil
	}

	rows := make([][]string, 0, len(result.Items))
	for _, item := range result.Items {
		started := ""
		if item.StartedAt != nil {
			started = *item.StartedAt
		}
		rows = append(rows, []string{item.StepKey, item.Status, started})
	}
	if err := output.PrintListTable([]string{"STEP_KEY", "STATUS", "STARTED"}, rows); err != nil {
		return err
	}
	return output.PrintPagination(result.Pagination)
}

func stepGet(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	workflowKey := args[0]
	runNumber, err := parsePositiveIntArg("run number", args[1])
	if err != nil {
		return err
	}

	stepKey := args[2]
	result, err := ctx.Client.GetStepRunByStepKey(workflowKey, runNumber, stepKey)
	if err != nil {
		return err
	}

	if IsJSON(ctx) {
		return output.PrintJSON(result)
	}
	if ctx.Quiet {
		fmt.Fprintln(os.Stdout, result.StepKey)
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
		{"workflowRunId", result.WorkflowRunID},
		{"stepKey", result.StepKey},
		{"id", result.ID},
		{"status", output.ColorStatus(result.Status)},
		{"startedAt", started},
		{"finishedAt", finished},
	})
}
