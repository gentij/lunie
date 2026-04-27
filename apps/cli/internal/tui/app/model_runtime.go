package app

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gentij/lunie/apps/cli/internal/api"
)

func pulseTick() tea.Cmd {
	return tea.Tick(650*time.Millisecond, func(time.Time) tea.Msg {
		return pulseMsg{}
	})
}

func (m Model) profileDelay() time.Duration {
	switch m.networkProfile {
	case NetworkFast:
		return 90 * time.Millisecond
	case NetworkSlow:
		return 900 * time.Millisecond
	case NetworkFlaky:
		return 550 * time.Millisecond
	default:
		return 240 * time.Millisecond
	}
}

func (m Model) profileShouldFail(manual bool) bool {
	if m.networkProfile != NetworkFlaky {
		return false
	}
	if manual {
		return m.refreshCount%4 == 0
	}
	return m.refreshCount%3 == 0
}

func (m Model) profileRefreshEvery() time.Duration {
	switch m.networkProfile {
	case NetworkFast:
		return 1200 * time.Millisecond
	case NetworkSlow:
		return 4 * time.Second
	case NetworkFlaky:
		return 2500 * time.Millisecond
	default:
		return 2 * time.Second
	}
}

func (m *Model) setNetworkProfile(profile NetworkProfile) {
	m.networkProfile = profile
	m.refreshEvery = m.profileRefreshEvery()
	m.lastRefresh = time.Now()
}

func networkProfileLabel(profile NetworkProfile) string {
	switch profile {
	case NetworkFast:
		return "FAST"
	case NetworkSlow:
		return "SLOW"
	case NetworkFlaky:
		return "FLAKY"
	default:
		return "NORMAL"
	}
}

func (m *Model) startMockRefresh(manual bool) {
	m.refreshPending = true
	m.refreshCount++
	m.mainState = SurfaceRefreshing
	m.contextState = SurfaceRefreshing
	if manual {
		m.toast = ToastState{}
	}
}

func (m *Model) syncSurfaceStates() {
	if !m.uiReady {
		return
	}
	if m.mainState != SurfaceRefreshing && m.mainState != SurfaceError && m.mainState != SurfaceStale {
		if len(m.filteredRows) == 0 {
			m.mainState = SurfaceEmpty
		} else {
			m.mainState = SurfaceSuccess
		}
	}
	if m.contextState != SurfaceRefreshing && m.contextState != SurfaceError && m.contextState != SurfaceStale {
		content := strings.TrimSpace(m.contextViewport.View())
		if content == "" {
			m.contextState = SurfaceEmpty
		} else {
			m.contextState = SurfaceSuccess
		}
	}
}

func (m *Model) pushToast(level ToastLevel, message string) tea.Cmd {
	m.toast.ID++
	m.toast.Active = true
	m.toast.Level = level
	m.toast.Message = message
	id := m.toast.ID
	return tea.Tick(2400*time.Millisecond, func(time.Time) tea.Msg {
		return toastClearMsg{id: id}
	})
}

func (m *Model) canRetry() bool {
	return m.mainState == SurfaceError || m.mainState == SurfaceStale || m.contextState == SurfaceError || m.contextState == SurfaceStale
}

func mutationErrorMessage(err error) string {
	if err == nil {
		return "Action failed"
	}
	if apiErr := api.AsAPIError(err); apiErr != nil {
		code := strings.TrimSpace(apiErr.Code)
		msg := strings.TrimSpace(apiErr.Message)
		if code != "" && msg != "" {
			return code + ": " + msg
		}
		if msg != "" {
			return msg
		}
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return "Action failed"
	}
	return msg
}
