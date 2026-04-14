package cli

import (
	"fmt"
	"os"

	"github.com/gentij/taskforge/apps/cli/internal/api"
	"github.com/gentij/taskforge/apps/cli/internal/output"
	"github.com/spf13/cobra"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secrets",
}

var secretListPage int
var secretListPageSize int
var secretListSortBy string
var secretListSortOrder string
var secretCreateName string
var secretCreateValue string
var secretCreateDescription string
var secretUpdateName string
var secretUpdateValue string
var secretUpdateDescription string

func init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List secrets",
		RunE:  secretList,
	}
	listCmd.Flags().IntVar(&secretListPage, "page", 1, "Page number")
	listCmd.Flags().IntVar(&secretListPageSize, "page-size", 25, "Page size")
	listCmd.Flags().StringVar(&secretListSortBy, "sort-by", "createdAt", "Sort field (createdAt|updatedAt)")
	listCmd.Flags().StringVar(&secretListSortOrder, "sort-order", "desc", "Sort order (asc|desc)")

	getCmd := &cobra.Command{
		Use:   "get <secret-id>",
		Short: "Get a secret",
		Args:  cobra.ExactArgs(1),
		RunE:  secretGet,
	}

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a secret",
		RunE:  secretCreate,
	}
	createCmd.Flags().StringVar(&secretCreateName, "name", "", "Secret name")
	createCmd.Flags().StringVar(&secretCreateValue, "value", "", "Secret value")
	createCmd.Flags().StringVar(&secretCreateDescription, "description", "", "Secret description")
	_ = createCmd.MarkFlagRequired("name")
	_ = createCmd.MarkFlagRequired("value")

	updateCmd := &cobra.Command{
		Use:   "update <secret-id>",
		Short: "Update a secret",
		Args:  cobra.ExactArgs(1),
		RunE:  secretUpdate,
	}
	updateCmd.Flags().StringVar(&secretUpdateName, "name", "", "Secret name")
	updateCmd.Flags().StringVar(&secretUpdateValue, "value", "", "Secret value")
	updateCmd.Flags().StringVar(&secretUpdateDescription, "description", "", "Secret description")

	deleteCmd := &cobra.Command{
		Use:   "delete <secret-id>",
		Short: "Delete a secret",
		Args:  cobra.ExactArgs(1),
		RunE:  secretDelete,
	}

	secretCmd.AddCommand(listCmd)
	secretCmd.AddCommand(getCmd)
	secretCmd.AddCommand(createCmd)
	secretCmd.AddCommand(updateCmd)
	secretCmd.AddCommand(deleteCmd)
}

func secretList(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	result, err := ctx.Client.ListSecrets(secretListPage, secretListPageSize, secretListSortBy, secretListSortOrder)
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
		rows = append(rows, []string{item.ID, item.Name, item.CreatedAt, item.UpdatedAt})
	}
	if err := output.PrintListTable([]string{"ID", "NAME", "CREATED", "UPDATED"}, rows); err != nil {
		return err
	}
	return output.PrintPagination(result.Pagination)
}

func secretGet(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	result, err := ctx.Client.GetSecret(args[0])
	if err != nil {
		return err
	}

	return printSecret(ctx, result)
}

func secretCreate(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	payload := secretCreatePayload(cmd)

	result, err := ctx.Client.CreateSecret(payload)
	if err != nil {
		return err
	}

	return printSecret(ctx, result)
}

func secretCreatePayload(cmd *cobra.Command) map[string]any {
	payload := map[string]any{
		"name":  secretCreateName,
		"value": secretCreateValue,
	}

	if cmd.Flags().Changed("description") && secretCreateDescription != "" {
		payload["description"] = secretCreateDescription
	}

	return payload
}

func secretUpdate(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	patch := map[string]any{}
	if secretUpdateName != "" {
		patch["name"] = secretUpdateName
	}
	if secretUpdateValue != "" {
		patch["value"] = secretUpdateValue
	}
	if cmd.Flags().Changed("description") {
		patch["description"] = secretUpdateDescription
	}
	if len(patch) == 0 {
		return fmt.Errorf("no fields to update")
	}

	result, err := ctx.Client.UpdateSecret(args[0], patch)
	if err != nil {
		return err
	}

	return printSecret(ctx, result)
}

func secretDelete(cmd *cobra.Command, args []string) error {
	ctx := GetContext(cmd.Context())
	if ctx == nil {
		return fmt.Errorf("missing context")
	}

	result, err := ctx.Client.DeleteSecret(args[0])
	if err != nil {
		return err
	}

	return printSecret(ctx, result)
}

func printSecret(ctx *Context, result api.Secret) error {
	if IsJSON(ctx) {
		return output.PrintJSON(result)
	}
	if ctx.Quiet {
		fmt.Fprintln(os.Stdout, result.ID)
		return nil
	}

	description := ""
	if result.Description != nil {
		description = *result.Description
	}

	return output.PrintKVTable([][2]string{
		{"id", result.ID},
		{"name", result.Name},
		{"description", description},
		{"createdAt", result.CreatedAt},
		{"updatedAt", result.UpdatedAt},
	})
}
