package app

import (
	"testing"

	"github.com/charmbracelet/bubbles/list"
	"github.com/gentij/lunie/apps/cli/internal/tui/styles"
)

func TestBuildPalette_RecentActionsDedupedAgainstBaseItems(t *testing.T) {
	state := paletteBuildState{View: ViewDashboard, Profile: NetworkNormal}
	recentAction := paletteAction{Kind: paletteGoToView, View: ViewWorkflows}

	palette := buildPalette(styles.DefaultTheme(), []paletteAction{recentAction}, state)
	items := palette.Items()

	if got := countPaletteAction(items, recentAction); got != 1 {
		t.Fatalf("expected one workflows navigation action after dedupe, got %d", got)
	}

	if key := firstNonSectionActionKey(items); key != paletteActionKey(recentAction) {
		t.Fatalf("expected recent action to appear first, got key %q", key)
	}
}

func TestBuildPalette_RecentDedupeDoesNotRemoveOtherActions(t *testing.T) {
	state := paletteBuildState{View: ViewDashboard, Profile: NetworkNormal}
	recentAction := paletteAction{Kind: paletteGoToView, View: ViewWorkflows}
	otherAction := paletteAction{Kind: paletteGoToView, View: ViewDashboard}

	palette := buildPalette(styles.DefaultTheme(), []paletteAction{recentAction}, state)
	items := palette.Items()

	if got := countPaletteAction(items, otherAction); got != 1 {
		t.Fatalf("expected dashboard navigation action to remain, got %d", got)
	}
}

func countPaletteAction(items []list.Item, action paletteAction) int {
	count := 0
	wantKey := paletteActionKey(action)
	for _, item := range items {
		paletteItemValue, ok := item.(paletteItem)
		if !ok || paletteItemValue.Section {
			continue
		}
		if paletteActionKey(paletteItemValue.Action) == wantKey {
			count++
		}
	}
	return count
}

func firstNonSectionActionKey(items []list.Item) string {
	for _, item := range items {
		paletteItemValue, ok := item.(paletteItem)
		if !ok || paletteItemValue.Section {
			continue
		}
		return paletteActionKey(paletteItemValue.Action)
	}
	return ""
}
