package ui

import (
	"bufio"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/murraycollier/git-dracula/internal/theme"
)

// HighlightDiff formats unified git diff output with Dracula theme colors
func HighlightDiff(diff string, styles theme.Styles, width int) string {
	if strings.TrimSpace(diff) == "" {
		return styles.CommitMeta.Render("No diff content available.")
	}

	scanner := bufio.NewScanner(strings.NewReader(diff))
	var out strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		var styledLine string

		switch {
		case strings.HasPrefix(line, "diff --git"),
			strings.HasPrefix(line, "index "),
			strings.HasPrefix(line, "--- "),
			strings.HasPrefix(line, "+++ "),
			strings.HasPrefix(line, "old mode"),
			strings.HasPrefix(line, "new mode"),
			strings.HasPrefix(line, "deleted file mode"),
			strings.HasPrefix(line, "new file mode"):
			styledLine = styles.DiffHeader.Render(line)

		case strings.HasPrefix(line, "@@"):
			styledLine = styles.DiffHunk.Render(line)

		case strings.HasPrefix(line, "+"):
			if width > 0 {
				styledLine = styles.DiffAddLine.Width(width).Render(line)
			} else {
				styledLine = styles.DiffAddLine.Render(line)
			}

		case strings.HasPrefix(line, "-"):
			if width > 0 {
				styledLine = styles.DiffDelLine.Width(width).Render(line)
			} else {
				styledLine = styles.DiffDelLine.Render(line)
			}

		case strings.HasPrefix(line, "\\ No newline at end of file"):
			styledLine = styles.CommitMeta.Italic(true).Render(line)

		default:
			styledLine = styles.DiffContext.Render(line)
		}

		out.WriteString(styledLine)
		out.WriteString("\n")
	}

	return strings.TrimRight(out.String(), "\n")
}

// RenderDiffBar generates a compact visual histogram bar for additions and deletions (like git diffstat)
func RenderDiffBar(additions, deletions int, maxBarWidth int, styles theme.Styles) string {
	total := additions + deletions
	if total == 0 {
		return ""
	}

	if maxBarWidth <= 0 {
		maxBarWidth = 10
	}

	var addBlocks, delBlocks int
	if total <= maxBarWidth {
		addBlocks = additions
		delBlocks = deletions
	} else {
		addBlocks = (additions * maxBarWidth) / total
		delBlocks = (deletions * maxBarWidth) / total
		if addBlocks == 0 && additions > 0 {
			addBlocks = 1
		}
		if delBlocks == 0 && deletions > 0 {
			delBlocks = 1
		}
		if addBlocks+delBlocks > maxBarWidth {
			if additions > deletions {
				addBlocks = maxBarWidth - delBlocks
			} else {
				delBlocks = maxBarWidth - addBlocks
			}
		}
	}

	var b strings.Builder
	if addBlocks > 0 {
		b.WriteString(styles.StatBarAdd.Render(strings.Repeat("■", addBlocks)))
	}
	if delBlocks > 0 {
		b.WriteString(styles.StatBarDel.Render(strings.Repeat("■", delBlocks)))
	}

	return b.String()
}

// FormatFileStat formats the +add -del diffstat string with colored numbers
func FormatFileStat(additions, deletions int, isBinary bool, styles theme.Styles) string {
	if isBinary {
		return styles.CommitMeta.Render(" [binary]")
	}
	if additions == 0 && deletions == 0 {
		return ""
	}

	var parts []string
	if additions > 0 {
		parts = append(parts, styles.StatAdd.Render(lipgloss.NewStyle().Width(5).Align(lipgloss.Right).Render("+"+stringNum(additions))))
	} else {
		parts = append(parts, styles.StatZero.Render(lipgloss.NewStyle().Width(5).Align(lipgloss.Right).Render("+0")))
	}

	if deletions > 0 {
		parts = append(parts, styles.StatDel.Render(lipgloss.NewStyle().Width(5).Align(lipgloss.Right).Render("-"+stringNum(deletions))))
	} else {
		parts = append(parts, styles.StatZero.Render(lipgloss.NewStyle().Width(5).Align(lipgloss.Right).Render("-0")))
	}

	return strings.Join(parts, " ")
}

func stringNum(n int) string {
	if n < 1000 {
		return strings.TrimSpace(lipgloss.NewStyle().Render(stringFromInt(n)))
	}
	return lipgloss.NewStyle().Render(stringFromInt(n/1000) + "k")
}

func stringFromInt(n int) string {
	var digits []byte
	if n == 0 {
		return "0"
	}
	for n > 0 {
		digits = append([]byte{byte('0' + (n % 10))}, digits...)
		n /= 10
	}
	return string(digits)
}
