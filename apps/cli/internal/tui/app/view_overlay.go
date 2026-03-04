package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func renderOverlay(base string, modal string, m Model) string {
	basePlain := ansi.Strip(base)
	baseLines := padPlainLines(basePlain, m.width, m.height)

	modalW, modalH := ansiBlockSize(modal)
	if modalW < 1 {
		modalW = 1
	}
	if modalH < 1 {
		modalH = 1
	}
	if modalW > m.width {
		modalW = m.width
	}
	if modalH > m.height {
		modalH = m.height
	}
	modalBlock := lipgloss.Place(
		modalW,
		modalH,
		lipgloss.Left,
		lipgloss.Top,
		modal,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceBackground(m.theme.Surface),
	)
	modalLines := strings.Split(modalBlock, "\n")
	x := max((m.width-modalW)/2, 0)
	y := max((m.height-modalH)/2, 0)

	out := make([]string, 0, m.height)
	for row := 0; row < m.height; row++ {
		line := baseLines[row]
		if row < y || row >= y+modalH {
			out = append(out, m.styles.Dim.Render(line))
			continue
		}
		modalLine := ""
		modalIdx := row - y
		if modalIdx >= 0 && modalIdx < len(modalLines) {
			modalLine = ansi.Truncate(modalLines[modalIdx], modalW, "")
		}
		leftW := min(max(x, 0), m.width)
		rightStart := min(leftW+modalW, m.width)
		left := plainSlice(line, 0, leftW)
		right := plainSlice(line, rightStart, m.width)
		composed := m.styles.Dim.Render(left) + modalLine + m.styles.Dim.Render(right)
		out = append(out, composed)
	}

	return clampToViewport(strings.Join(out, "\n"), m.width, m.height)
}

func padPlainLines(content string, width int, height int) []string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, height)
	for i := 0; i < height; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		if width > 0 {
			r := []rune(line)
			if len(r) > width {
				line = string(r[:width])
			} else if len(r) < width {
				line += strings.Repeat(" ", width-len(r))
			}
		}
		result = append(result, line)
	}
	return result
}

func plainSlice(line string, start int, end int) string {
	r := []rune(line)
	if start < 0 {
		start = 0
	}
	if end < 0 {
		end = 0
	}
	if start > len(r) {
		start = len(r)
	}
	if end > len(r) {
		end = len(r)
	}
	if start >= end {
		return ""
	}
	return string(r[start:end])
}

func ansiBlockSize(content string) (int, int) {
	lines := strings.Split(content, "\n")
	maxWidth := 0
	for _, line := range lines {
		w := ansi.StringWidth(line)
		if w > maxWidth {
			maxWidth = w
		}
	}
	return maxWidth, len(lines)
}

func truncateLines(content string, width int) string {
	lines := strings.Split(content, "\n")
	for i := range lines {
		lines[i] = ansi.Truncate(lines[i], width, "")
	}
	return strings.Join(lines, "\n")
}

func sanitizeRenderable(content string) string {
	content = strings.ReplaceAll(content, "\r", "")
	return stripNonSGRANSI(content)
}

func stripNonSGRANSI(content string) string {
	var builder strings.Builder
	builder.Grow(len(content))
	for i := 0; i < len(content); i++ {
		ch := content[i]
		if ch != '\x1b' {
			builder.WriteByte(ch)
			continue
		}
		if i+1 >= len(content) {
			break
		}
		next := content[i+1]
		if next != '[' {
			continue
		}
		j := i + 2
		for j < len(content) {
			final := content[j]
			if final >= 0x40 && final <= 0x7E {
				if final == 'm' {
					builder.WriteString(content[i : j+1])
				}
				i = j
				break
			}
			j++
		}
		if j >= len(content) {
			break
		}
	}
	return builder.String()
}

func applyBackgroundLayer(content string, width int, height int, style lipgloss.Style) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	prefix, reset := backgroundCodes(style)
	lines := strings.Split(content, "\n")
	filled := make([]string, 0, height)
	for i := 0; i < height; i++ {
		line := ""
		if i < len(lines) {
			line = ansi.Truncate(lines[i], width, "")
		}
		if prefix != "" {
			line = strings.ReplaceAll(line, reset, reset+prefix)
			line = strings.ReplaceAll(line, "\x1b[m", "\x1b[m"+prefix)
			line = strings.ReplaceAll(line, "\x1b[49m", "\x1b[49m"+prefix)
			line = strings.ReplaceAll(line, "\x1b[39m", "\x1b[39m"+prefix)
		}
		pad := width - ansi.StringWidth(line)
		if pad < 0 {
			pad = 0
		}
		line = prefix + line + strings.Repeat(" ", pad) + reset
		filled = append(filled, line)
	}
	return strings.Join(filled, "\n")
}

func backgroundCodes(style lipgloss.Style) (string, string) {
	sample := style.Render("X")
	idx := strings.Index(sample, "X")
	if idx == -1 {
		return "", "\x1b[0m"
	}
	prefix := sample[:idx]
	suffix := sample[idx+1:]
	if suffix == "" {
		suffix = "\x1b[0m"
	}
	return prefix, suffix
}

func applyScanlines(content string, width int, color lipgloss.Color) string {
	lines := strings.Split(content, "\n")
	scanStyle := lipgloss.NewStyle().Foreground(color).Faint(true)
	scanChar := scanStyle.Render(".")
	for i, line := range lines {
		if i%2 == 1 {
			lines[i] = scanlineLine(line, width, scanChar)
		}
	}
	return strings.Join(lines, "\n")
}

func scanlineLine(line string, width int, scanChar string) string {
	if width < 1 {
		width = 1
	}
	var builder strings.Builder
	builder.Grow(len(line) + width)
	inEscape := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		if ch == '\x1b' {
			inEscape = true
			builder.WriteByte(ch)
			continue
		}
		if inEscape {
			builder.WriteByte(ch)
			if ch == 'm' {
				inEscape = false
			}
			continue
		}
		if ch == ' ' {
			builder.WriteString(scanChar)
		} else {
			builder.WriteByte(ch)
		}
	}
	padding := width - ansi.StringWidth(line)
	for i := 0; i < padding; i++ {
		builder.WriteString(scanChar)
	}
	return builder.String()
}

func clampToViewport(content string, width int, height int) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	lines := strings.Split(content, "\n")
	clamped := make([]string, 0, height)
	for i := 0; i < len(lines) && i < height; i++ {
		clamped = append(clamped, ansi.Truncate(lines[i], width, ""))
	}
	for len(clamped) < height {
		clamped = append(clamped, "")
	}
	return strings.Join(clamped, "\n")
}

func clampSection(content string, width int, height int) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	lines := strings.Split(content, "\n")
	clamped := make([]string, 0, height)
	for i := 0; i < len(lines) && i < height; i++ {
		clamped = append(clamped, ansi.Truncate(lines[i], width, ""))
	}
	for len(clamped) < height {
		clamped = append(clamped, "")
	}
	return strings.Join(clamped, "\n")
}
