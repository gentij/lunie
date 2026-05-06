package cli

import (
	"fmt"
	"strconv"
	"strings"
)

func parsePositiveIntArg(name string, raw string) (int, error) {
	trimmed := strings.TrimSpace(raw)
	value, err := strconv.Atoi(trimmed)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("invalid %s: %s", name, raw)
	}

	return value, nil
}
