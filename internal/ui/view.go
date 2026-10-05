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

// RenderFull formats and renders the complete Dracula-themed status output matching the screenshot
func RenderFull(res *git.StatusResult, client *git.Client, cfg ViewConfig, styles theme.Styles) string {
	width := cfg.Width
	if width <= 0 {
		if w, _, err := term.GetSize(0); err == nil && w > 20 {
			width = w
		} else {
			width = 100
		}
	}
	if width > 140 {
		width = 140
	}

	if cfg.Short {
		return RenderCompact(res, styles)
	}

	// Size the card so the right-hand border margin lines up with the end of the commit message text
	cardWidth := width - 4
	if res.HeadCommit != nil && res.HeadCommit.Subject != "" {
		commitWidth := lipgloss.Width(res.HeadCommit.Subject)
		cardWidth = commitWidth + 4
		if cardWidth > width-2 {
			cardWidth = width - 2
		}
	}
	if cardWidth < 35 {
		cardWidth = 35
	}

	availWidth := cardWidth - 4
	numDots := 21
	if availWidth < 41 && availWidth >= 4 {
		numDots = (availWidth + 1) / 2
	}
	var dotStr string
	if numDots > 1 {
		dotStr = strings.Repeat("· ", numDots-1) + "·"
	} else {
		dotStr = "·"
	}
	dividerDots := styles.Divider.Render(dotStr)

	var contentLines []string

	// 1. Repo & Branch line (e.g. "homelab   on main")
	repo := styles.RepoBadge.Render(res.RepoName)
	on := styles.OnText.Render("on")
	branch := styles.BranchBadge.Render(res.Branch.Head)
	if res.Branch.Detached {
		on = styles.OnText.Render("detached at")
	} else if res.Branch.Initial {
		branch += " " + styles.OnText.Render("(initial)")
	}
	contentLines = append(contentLines, fmt.Sprintf("%s   %s %s", repo, on, branch))

	// 2. Upstream line (e.g. "origin/main   ✓ up to date")
	var upstreamParts []string
	if res.Branch.Upstream != "" {
		upstream := styles.CommitMeta.Render(res.Branch.Upstream)
		var syncText string
		ahead := res.Branch.Ahead
		behind := res.Branch.Behind
		if ahead == 0 && behind == 0 {
			syncText = styles.RemoteSynced.Render("✓ up to date")
		} else if ahead > 0 && behind == 0 {
			syncText = styles.RemoteAhead.Render(fmt.Sprintf("⇡%d ahead", ahead))
		} else if ahead == 0 && behind > 0 {
			syncText = styles.RemoteBehind.Render(fmt.Sprintf("⇣%d behind", behind))
		} else {
			syncText = styles.RemoteDiverged.Render(fmt.Sprintf("⇡%d ⇣%d diverged", ahead, behind))
		}
		upstreamParts = append(upstreamParts, fmt.Sprintf("%s   %s", upstream, syncText))
	} else if !res.Branch.Initial {
		upstreamParts = append(upstreamParts, styles.CommitMeta.Render("no upstream"))
	}

	if res.StashCount > 0 {
		upstreamParts = append(upstreamParts, styles.StashBadge.Render(fmt.Sprintf("📦 %d stashed", res.StashCount)))
	}
	if res.State != git.StateNormal {
		upstreamParts = append(upstreamParts, styles.StateBadge.Render("⚠ "+string(res.State)))
	}

	if len(upstreamParts) > 0 {
		contentLines = append(contentLines, strings.Join(upstreamParts, "   "))
	}

	// 3. Divider
	contentLines = append(contentLines, dividerDots)

	// 4. Commit hash, author, relative time (e.g. "fd18a0b   Murray Collier · 7 months ago")
	if res.HeadCommit != nil {
		hash := styles.CommitHash.Render(res.HeadCommit.Hash)
		author := styles.CommitAuthor.Render(res.HeadCommit.Author)
		dot := styles.CommitMeta.Render("·")
		relTime := styles.CommitMeta.Render(res.HeadCommit.RelativeTime)
		line4 := fmt.Sprintf("%s   %s  %s  %s", hash, author, dot, relTime)
		if lipgloss.Width(line4) > cardWidth-4 {
			line4 = fmt.Sprintf("%s  %s · %s", hash, author, relTime)
		}
		contentLines = append(contentLines, line4)

		// 5. Commit subject (e.g. "chore(tidy): Cleaning up temp files")
		contentLines = append(contentLines, styles.CommitSubject.Render(truncate(res.HeadCommit.Subject, cardWidth-4)))
	} else {
		contentLines = append(contentLines, styles.CommitMeta.Render("No commits yet"))
	}

	// 6. Divider
	contentLines = append(contentLines, dividerDots)

	// 7. Summary Pills (e.g. " staged 1   modified 0   untracked 1   conflicts 0")
	stagedPill := styles.PillStaged.Render(fmt.Sprintf("staged %d", len(res.Staged)))
	modifiedPill := styles.PillUnstaged.Render(fmt.Sprintf("modified %d", len(res.Unstaged)))
	untrackedPill := styles.PillUntracked.Render(fmt.Sprintf("untracked %d", len(res.Untracked)))
	conflictsPill := styles.PillConflicts.Render(fmt.Sprintf("conflicts %d", len(res.Conflicts)))

	var pillsLine string
	pillsFull := fmt.Sprintf("%s   %s   %s   %s", stagedPill, modifiedPill, untrackedPill, conflictsPill)
	pillsCompact := fmt.Sprintf("%s %s %s %s", stagedPill, modifiedPill, untrackedPill, conflictsPill)
	if lipgloss.Width(pillsFull) <= cardWidth-4 {
		pillsLine = pillsFull
	} else if lipgloss.Width(pillsCompact) <= cardWidth-4 {
		pillsLine = pillsCompact
	} else {
		pillsLine = fmt.Sprintf("%s  %s\n%s  %s", stagedPill, modifiedPill, untrackedPill, conflictsPill)
	}
	contentLines = append(contentLines, pillsLine)

	// 8. Divider
	contentLines = append(contentLines, dividerDots)

	// 9. File sections or clean message
	if res.IsClean() {
		contentLines = append(contentLines, "")
		contentLines = append(contentLines, styles.SectionStaged.Render("🦇 Working tree is spotless. Dracula approves!"))
		contentLines = append(contentLines, styles.CommitMeta.Render("Nothing to commit, working directory is clean."))
	} else {
		var sections []string

		// Conflicts
		if len(res.Conflicts) > 0 {
			var sb strings.Builder
			sb.WriteString(styles.SectionConflicts.Render("! conflicts") + "\n")
			for _, f := range res.Conflicts {
				sb.WriteString("  " + styles.SectionConflicts.Render("! ") + styles.FileName.Render(f.Path) + "\n")
			}
			sections = append(sections, strings.TrimRight(sb.String(), "\n"))
		}

		// Staged
		if len(res.Staged) > 0 {
			var sb strings.Builder
			sb.WriteString(styles.SectionStaged.Render("+ staged") + "\n")
			for _, f := range res.Staged {
				sym := "+"
				switch f.StatusCode {
				case "D":
					sym = "-"
				case "M":
					sym = "~"
				case "R":
					sym = "➜"
				}
				sb.WriteString("  " + styles.SectionStaged.Render(sym+" ") + styles.FileName.Render(f.Path) + "\n")
			}
			sections = append(sections, strings.TrimRight(sb.String(), "\n"))
		}

		// Modified (Unstaged)
		if len(res.Unstaged) > 0 {
			var sb strings.Builder
			sb.WriteString(styles.SectionUnstaged.Render("~ modified") + "\n")
			for _, f := range res.Unstaged {
				sym := "~"
				if f.StatusCode == "D" {
					sym = "-"
				}
				sb.WriteString("  " + styles.SectionUnstaged.Render(sym+" ") + styles.FileName.Render(f.Path) + "\n")
			}
			sections = append(sections, strings.TrimRight(sb.String(), "\n"))
		}

		// Untracked
		if len(res.Untracked) > 0 {
			var sb strings.Builder
			sb.WriteString(styles.SectionUntracked.Render("? untracked") + "\n")
			for _, f := range res.Untracked {
				sb.WriteString("  " + styles.SectionUntracked.Render("? ") + styles.FileName.Render(f.Path) + "\n")
			}
			sections = append(sections, strings.TrimRight(sb.String(), "\n"))
		}

		if len(sections) > 0 {
			// One blank line before sections, and exactly one blank line between sections
			contentLines = append(contentLines, "", strings.Join(sections, "\n\n"))
		}
	}

	// Optional Inline Diffs
	if cfg.ShowDiff && client != nil {
		diffs := renderInlineDiffs(res, client, styles, cardWidth-6)
		if diffs != "" {
			contentLines = append(contentLines, "", diffs)
		}
	}

	fullContent := strings.Join(contentLines, "\n")
	return styles.Card.Width(cardWidth).Render(fullContent) + "\n"
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

	cardWidth := width - 4
	if cardWidth < 40 {
		cardWidth = 40
	}

	// Bottom Line: Latest Commit Info
	var bottomLine string
	if res.HeadCommit != nil {
		hash := styles.CommitHash.Render("◈ " + res.HeadCommit.Hash)
		authorTimeStr := fmt.Sprintf("· %s (%s)", res.HeadCommit.Author, res.HeadCommit.RelativeTime)
		authorTime := styles.CommitMeta.Render(authorTimeStr)

		availSubject := cardWidth - lipgloss.Width(hash) - lipgloss.Width(authorTime) - 6
		if availSubject < 15 {
			availSubject = 15
		}
		subject := styles.CommitSubject.Render(truncate(res.HeadCommit.Subject, availSubject))
		bottomLine = fmt.Sprintf("%s %s %s", hash, subject, authorTime)
	} else {
		bottomLine = styles.CommitMeta.Render("◈ No commits yet on this branch")
	}

	content := topLine + "\n" + bottomLine
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
