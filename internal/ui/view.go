package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/murraycollier/git-dracula/internal/git"
	"github.com/murraycollier/git-dracula/internal/theme"
	"golang.org/x/term"
)

// ViewConfig options for rendering git-dracula output
type ViewConfig struct {
	Short    bool
	ShowDiff bool
	NoColor  bool
	Width    int
}

// RenderFull formats and renders the complete Dracula-themed status output
func RenderFull(res *git.StatusResult, client *git.Client, cfg ViewConfig, styles theme.Styles) string {
	width := cfg.Width
	if width <= 0 {
		if w, _, err := term.GetSize(0); err == nil && w > 20 {
			width = w
		} else {
			width = 90
		}
	}
	if width > 120 {
		width = 120
	}

	if cfg.Short {
		return RenderCompact(res, styles)
	}

	var sb strings.Builder

	// 1. Header Card
	sb.WriteString(renderHeaderCard(res, styles, width))
	sb.WriteString("\n\n")

	// 2. If working tree is clean
	if res.IsClean() {
		sb.WriteString(renderCleanState(res, styles, width))
		sb.WriteString("\n")
		return sb.String()
	}

	// 3. Summary Pills Bar
	sb.WriteString(renderSummaryBar(res, styles))
	sb.WriteString("\n\n")

	// 4. Conflicts Section (if any)
	if len(res.Conflicts) > 0 {
		sb.WriteString(renderSection(SectionConflictDef(res.Conflicts, styles), width))
		sb.WriteString("\n")
	}

	// 5. Staged Changes Section
	if len(res.Staged) > 0 {
		sb.WriteString(renderSection(SectionStagedDef(res.Staged, styles), width))
		sb.WriteString("\n")
	}

	// 6. Unstaged Changes Section
	if len(res.Unstaged) > 0 {
		sb.WriteString(renderSection(SectionUnstagedDef(res.Unstaged, styles), width))
		sb.WriteString("\n")
	}

	// 7. Untracked Files Section
	if len(res.Untracked) > 0 {
		sb.WriteString(renderSection(SectionUntrackedDef(res.Untracked, styles), width))
		sb.WriteString("\n")
	}

	// 8. Optional Inline Diffs
	if cfg.ShowDiff && client != nil {
		diffs := renderInlineDiffs(res, client, styles, width)
		if diffs != "" {
			sb.WriteString("\n")
			sb.WriteString(diffs)
			sb.WriteString("\n")
		}
	}

	// 9. Footer Hint
	sb.WriteString(renderFooter(styles))
	sb.WriteString("\n")

	return sb.String()
}

func renderHeaderCard(res *git.StatusResult, styles theme.Styles, width int) string {
	// Top Line: [Repo] [Branch] [Remote] [Stashes] [State]
	repoBadge := styles.RepoBadge.Render("󰊢 " + res.RepoName)

	var branchText string
	if res.Branch.Detached {
		branchText = styles.BranchBadge.Render("⎇ HEAD detached at " + res.Branch.Head)
	} else if res.Branch.Initial {
		branchText = styles.BranchBadge.Render(" " + res.Branch.Head + " (initial)")
	} else {
		branchText = styles.BranchBadge.Render(" " + res.Branch.Head)
	}

	var remoteText string
	if res.Branch.Upstream != "" {
		ahead := res.Branch.Ahead
		behind := res.Branch.Behind
		if ahead == 0 && behind == 0 {
			remoteText = styles.RemoteSynced.Render("✓ " + res.Branch.Upstream)
		} else if ahead > 0 && behind == 0 {
			remoteText = styles.RemoteAhead.Render(fmt.Sprintf("⇡%d %s", ahead, res.Branch.Upstream))
		} else if ahead == 0 && behind > 0 {
			remoteText = styles.RemoteBehind.Render(fmt.Sprintf("⇣%d %s", behind, res.Branch.Upstream))
		} else {
			remoteText = styles.RemoteDiverged.Render(fmt.Sprintf("⇡%d ⇣%d %s", ahead, behind, res.Branch.Upstream))
		}
	} else if !res.Branch.Initial {
		remoteText = styles.CommitMeta.Render("no upstream")
	}

	topParts := []string{repoBadge, branchText}
	if remoteText != "" {
		topParts = append(topParts, remoteText)
	}

	if res.StashCount > 0 {
		topParts = append(topParts, styles.StashBadge.Render(fmt.Sprintf("📦 %d stashed", res.StashCount)))
	}

	if res.State != git.StateNormal {
		topParts = append(topParts, styles.StateBadge.Render("⚠ "+string(res.State)))
	}

	topLine := strings.Join(topParts, "  ")

	// Bottom Line: Latest Commit Info
	var bottomLine string
	if res.HeadCommit != nil {
		hash := styles.CommitHash.Render("◈ " + res.HeadCommit.Hash)
		subject := styles.CommitSubject.Render(truncate(res.HeadCommit.Subject, 55))
		authorTime := styles.CommitMeta.Render(fmt.Sprintf("· %s (%s)", res.HeadCommit.Author, res.HeadCommit.RelativeTime))
		bottomLine = fmt.Sprintf("%s %s %s", hash, subject, authorTime)
	} else {
		bottomLine = styles.CommitMeta.Render("◈ No commits yet on this branch")
	}

	content := topLine + "\n" + bottomLine
	cardWidth := width - 4
	if cardWidth < 40 {
		cardWidth = 40
	}

	return styles.Card.Width(cardWidth).Render(content)
}

func renderSummaryBar(res *git.StatusResult, styles theme.Styles) string {
	var pills []string

	if len(res.Staged) > 0 {
		pills = append(pills, styles.PillStaged.Render(fmt.Sprintf("● %d Staged", len(res.Staged))))
	}
	if len(res.Unstaged) > 0 {
		pills = append(pills, styles.PillUnstaged.Render(fmt.Sprintf("✎ %d Modified", len(res.Unstaged))))
	}
	if len(res.Untracked) > 0 {
		pills = append(pills, styles.PillUntracked.Render(fmt.Sprintf("? %d Untracked", len(res.Untracked))))
	}
	if len(res.Conflicts) > 0 {
		pills = append(pills, styles.PillConflicts.Render(fmt.Sprintf("✖ %d Conflicts", len(res.Conflicts))))
	}

	return "  " + strings.Join(pills, "  ")
}

func renderCleanState(res *git.StatusResult, styles theme.Styles, width int) string {
	cardWidth := width - 6
	if cardWidth < 50 {
		cardWidth = 50
	}

	var content strings.Builder
	bat := lipgloss.NewStyle().Foreground(theme.Purple).Bold(true).Render("🦇")
	headline := styles.SectionStaged.Render("Working tree is spotless. Dracula approves!")
	subtext := styles.CommitMeta.Render("Nothing to commit, working directory is clean.")

	content.WriteString(fmt.Sprintf("%s  %s\n", bat, headline))
	content.WriteString(fmt.Sprintf("    %s", subtext))

	return styles.CleanCard.Width(cardWidth).Render(content.String())
}

// SectionDef defines styling and data for a file status section
type SectionDef struct {
	Title     string
	Header    string
	Count     int
	Files     []git.FileStatus
	BadgeFunc func(f git.FileStatus) string
}

func SectionConflictDef(files []git.FileStatus, styles theme.Styles) SectionDef {
	return SectionDef{
		Title:  "Conflicts",
		Header: styles.SectionConflicts.Render(fmt.Sprintf("✖ MERGE CONFLICTS  %s", styles.SectionCount.Render(fmt.Sprintf("(%d)", len(files))))),
		Count:  len(files),
		Files:  files,
		BadgeFunc: func(f git.FileStatus) string {
			return styles.BadgeCnf.Render("! CNF")
		},
	}
}

func SectionStagedDef(files []git.FileStatus, styles theme.Styles) SectionDef {
	return SectionDef{
		Title:  "Staged",
		Header: styles.SectionStaged.Render(fmt.Sprintf("✔ STAGED CHANGES  %s", styles.SectionCount.Render(fmt.Sprintf("(%d to be committed)", len(files))))),
		Count:  len(files),
		Files:  files,
		BadgeFunc: func(f git.FileStatus) string {
			switch f.StatusCode {
			case "A":
				return styles.BadgeAdd.Render("+ ADD")
			case "M":
				return styles.BadgeMod.Render("~ MOD")
			case "D":
				return styles.BadgeDel.Render("- DEL")
			case "R":
				return styles.BadgeRen.Render("➜ REN")
			case "T":
				return styles.BadgeType.Render("! TYP")
			default:
				return styles.BadgeMod.Render("• " + f.StatusCode)
			}
		},
	}
}

func SectionUnstagedDef(files []git.FileStatus, styles theme.Styles) SectionDef {
	return SectionDef{
		Title:  "Unstaged",
		Header: styles.SectionUnstaged.Render(fmt.Sprintf("✎ UNSTAGED CHANGES  %s", styles.SectionCount.Render(fmt.Sprintf("(%d not staged)", len(files))))),
		Count:  len(files),
		Files:  files,
		BadgeFunc: func(f git.FileStatus) string {
			switch f.StatusCode {
			case "M":
				return styles.BadgeMod.Render("~ MOD")
			case "D":
				return styles.BadgeDel.Render("- DEL")
			case "T":
				return styles.BadgeType.Render("! TYP")
			default:
				return styles.BadgeMod.Render("• " + f.StatusCode)
			}
		},
	}
}

func SectionUntrackedDef(files []git.FileStatus, styles theme.Styles) SectionDef {
	return SectionDef{
		Title:  "Untracked",
		Header: styles.SectionUntracked.Render(fmt.Sprintf("? UNTRACKED FILES  %s", styles.SectionCount.Render(fmt.Sprintf("(%d files)", len(files))))),
		Count:  len(files),
		Files:  files,
		BadgeFunc: func(f git.FileStatus) string {
			return styles.BadgeUnt.Render("? NEW")
		},
	}
}

func renderSection(sec SectionDef, width int) string {
	var sb strings.Builder
	sb.WriteString("  " + sec.Header + "\n")

	styles := theme.DefaultStyles()

	for _, file := range sec.Files {
		badge := sec.BadgeFunc(file)

		dir := styles.DirName.Render(file.Dir())
		base := styles.FileName.Render(file.Base())
		filePath := dir + base

		if file.OrigPath != "" {
			arrow := styles.Arrow.Render(" ➜ ")
			orig := styles.DirName.Render(file.OrigPath)
			filePath = orig + arrow + filePath
		}

		statStr := FormatFileStat(file.Additions, file.Deletions, file.IsBinary, styles)
		barStr := RenderDiffBar(file.Additions, file.Deletions, 8, styles)

		// Calculate padding
		// 4 spaces + badge(6) + 2 spaces + path + 2 spaces + stat(11) + 2 spaces + bar(8)
		leftPart := fmt.Sprintf("    %s  %s", badge, filePath)
		rightPart := ""
		if statStr != "" {
			rightPart = statStr
			if barStr != "" {
				rightPart += "  " + barStr
			}
		}

		leftLen := lipgloss.Width(leftPart)
		rightLen := lipgloss.Width(rightPart)
		avail := width - 4

		if rightPart != "" && avail > leftLen+rightLen+2 {
			padding := strings.Repeat(" ", avail-leftLen-rightLen)
			sb.WriteString(leftPart + padding + rightPart + "\n")
		} else {
			sb.WriteString(leftPart)
			if rightPart != "" {
				sb.WriteString("  " + rightPart)
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func renderInlineDiffs(res *git.StatusResult, client *git.Client, styles theme.Styles, width int) string {
	var sb strings.Builder
	banner := styles.SectionStaged.Render("── INLINE DIFFS ───────────────────────────────────────────────────────────")
	sb.WriteString("  " + banner + "\n\n")

	// Render staged diffs
	for _, file := range res.Staged {
		if diff, err := client.GetDiff(file.Path, true); err == nil && strings.TrimSpace(diff) != "" {
			header := styles.SectionStaged.Render("● Staged: " + file.Path)
			sb.WriteString("  " + header + "\n")
			sb.WriteString(indent(HighlightDiff(diff, styles, width-6), "    ") + "\n\n")
		}
	}

	// Render unstaged diffs
	for _, file := range res.Unstaged {
		if diff, err := client.GetDiff(file.Path, false); err == nil && strings.TrimSpace(diff) != "" {
			header := styles.SectionUnstaged.Render("✎ Unstaged: " + file.Path)
			sb.WriteString("  " + header + "\n")
			sb.WriteString(indent(HighlightDiff(diff, styles, width-6), "    ") + "\n\n")
		}
	}

	return sb.String()
}

func renderFooter(styles theme.Styles) string {
	hint := styles.FooterHint.Render("💡 Tip: Run ") +
		styles.KeyBadge.Render("git-dracula -i") +
		styles.FooterHint.Render(" for interactive staging & diffs · ") +
		styles.KeyBadge.Render("-s") +
		styles.FooterHint.Render(" short · ") +
		styles.KeyBadge.Render("-d") +
		styles.FooterHint.Render(" diffs")

	return "  " + hint
}

// RenderCompact provides a sleek one-line-per-file format (like git status -s but Dracula-enhanced)
func RenderCompact(res *git.StatusResult, styles theme.Styles) string {
	var sb strings.Builder

	// Header branch line
	branchPill := styles.BranchBadge.Render(" " + res.Branch.Head)
	sb.WriteString("## " + res.RepoName + " " + branchPill)
	if res.Branch.Upstream != "" {
		if res.Branch.Ahead > 0 || res.Branch.Behind > 0 {
			sb.WriteString(fmt.Sprintf(" [%s +%d -%d]", res.Branch.Upstream, res.Branch.Ahead, res.Branch.Behind))
		} else {
			sb.WriteString(" [" + res.Branch.Upstream + "]")
		}
	}
	sb.WriteString("\n")

	for _, f := range res.Staged {
		badge := styles.BadgeAdd.Width(4).Render(f.StatusCode + " ")
		sb.WriteString(fmt.Sprintf("  %s %s\n", badge, f.Path))
	}
	for _, f := range res.Unstaged {
		badge := styles.BadgeMod.Width(4).Render(" " + f.StatusCode)
		sb.WriteString(fmt.Sprintf("  %s %s\n", badge, f.Path))
	}
	for _, f := range res.Untracked {
		badge := styles.BadgeUnt.Width(4).Render("??")
		sb.WriteString(fmt.Sprintf("  %s %s\n", badge, f.Path))
	}
	for _, f := range res.Conflicts {
		badge := styles.BadgeCnf.Width(4).Render("UU")
		sb.WriteString(fmt.Sprintf("  %s %s\n", badge, f.Path))
	}

	return sb.String()
}

func indent(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i := range lines {
		if lines[i] != "" {
			lines[i] = prefix + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
