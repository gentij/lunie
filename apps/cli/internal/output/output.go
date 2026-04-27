package output

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/gentij/lunie/apps/cli/internal/api"
	"golang.org/x/term"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
)

func PrintJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(os.Stdout, string(data))
	return err
}

func NewTableWriter() *tabwriter.Writer {
	return tabwriter.NewWriter(os.Stdout, 2, 4, 2, ' ', 0)
}

func BoolLabel(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func PrintPagination(meta api.Pagination) error {
	if meta.TotalPages <= 1 {
		return nil
	}

	sortLabel := ""
	if strings.TrimSpace(meta.SortBy) != "" && strings.TrimSpace(meta.SortOrder) != "" {
		sortLabel = fmt.Sprintf(" · Sort %s %s", meta.SortBy, strings.ToLower(meta.SortOrder))
	}

	_, err := fmt.Fprintf(
		os.Stdout,
		"Page %d/%d · Total %d · PageSize %d%s\n",
		meta.Page,
		meta.TotalPages,
		meta.Total,
		meta.PageSize,
		sortLabel,
	)
	return err
}

func PrintError(err error) {
	if err == nil {
		return
	}

	if apiErr := api.AsAPIError(err); apiErr != nil {
		fmt.Fprintf(os.Stderr, "ERROR %s: %s\n", apiErr.Code, apiErr.Message)
		if apiErr.Details != nil {
			data, marshalErr := json.MarshalIndent(apiErr.Details, "", "  ")
			if marshalErr == nil {
				fmt.Fprintln(os.Stderr, string(data))
			}
		}
		return
	}

	fmt.Fprintf(os.Stderr, "ERROR %s\n", err.Error())
}

var noColorOverride bool

func SetNoColor(value bool) {
	noColorOverride = value
}

func ColorStatus(status string) string {
	if !colorEnabled() {
		return status
	}

	switch strings.ToUpper(status) {
	case "SUCCEEDED", "SUCCESS":
		return colorize(status, "\x1b[32m")
	case "FAILED", "ERROR":
		return colorize(status, "\x1b[31m")
	case "RUNNING", "IN_PROGRESS":
		return colorize(status, "\x1b[33m")
	case "CANCELLED", "CANCELED":
		return colorize(status, "\x1b[90m")
	default:
		return status
	}
}

func colorEnabled() bool {
	if noColorOverride || os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func colorize(text string, color string) string {
	const reset = "\x1b[0m"
	return color + text + reset
}
