package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestSecretCreatePayloadOmitsEmptyDescription(t *testing.T) {
	secretCreateName = "API_KEY"
	secretCreateValue = "super-secret"
	secretCreateDescription = ""
	t.Cleanup(func() {
		secretCreateName = ""
		secretCreateValue = ""
		secretCreateDescription = ""
	})

	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&secretCreateDescription, "description", "", "Secret description")
	if err := cmd.Flags().Set("description", ""); err != nil {
		t.Fatalf("set description flag: %v", err)
	}

	payload := secretCreatePayload(cmd)

	if _, ok := payload["description"]; ok {
		t.Fatalf("expected empty description to be omitted, got %#v", payload)
	}
	if payload["name"] != secretCreateName {
		t.Fatalf("expected name %q, got %#v", secretCreateName, payload["name"])
	}
	if payload["value"] != secretCreateValue {
		t.Fatalf("expected value %q, got %#v", secretCreateValue, payload["value"])
	}
}

func TestSecretCreatePayloadIncludesDescriptionWhenProvided(t *testing.T) {
	description := "Smoke test secret"
	secretCreateName = "API_KEY"
	secretCreateValue = "super-secret"
	secretCreateDescription = ""
	t.Cleanup(func() {
		secretCreateName = ""
		secretCreateValue = ""
		secretCreateDescription = ""
	})

	cmd := &cobra.Command{}
	cmd.Flags().StringVar(&secretCreateDescription, "description", "", "Secret description")
	if err := cmd.Flags().Set("description", description); err != nil {
		t.Fatalf("set description flag: %v", err)
	}

	payload := secretCreatePayload(cmd)

	if got, ok := payload["description"]; !ok || got != description {
		t.Fatalf("expected description %q, got %#v", description, payload)
	}
}
