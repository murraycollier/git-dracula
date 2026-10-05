package theme

import "github.com/charmbracelet/lipgloss"

// Styles defines the Lipgloss styling for git-dracula
type Styles struct {
	// Card / Frame
	Card           lipgloss.Style
	CleanCard      lipgloss.Style
	WarningCard    lipgloss.Style

	// Header elements
	RepoBadge      lipgloss.Style
	BranchBadge    lipgloss.Style
	RemoteSynced   lipgloss.Style
	RemoteAhead    lipgloss.Style
	RemoteBehind   lipgloss.Style
	RemoteDiverged lipgloss.Style
	CommitHash     lipgloss.Style
	CommitSubject  lipgloss.Style
	CommitMeta     lipgloss.Style
	StashBadge     lipgloss.Style
	StateBadge     lipgloss.Style

	// Summary Pills
	PillStaged     lipgloss.Style
	PillUnstaged   lipgloss.Style
	PillUntracked  lipgloss.Style
	PillConflicts  lipgloss.Style
	PillCount      lipgloss.Style

	// Section Headers
	SectionStaged    lipgloss.Style
	SectionUnstaged  lipgloss.Style
	SectionUntracked lipgloss.Style
	SectionConflicts lipgloss.Style
	SectionCount     lipgloss.Style

	// Status Badges (File rows)
	BadgeAdd       lipgloss.Style
	BadgeMod       lipgloss.Style
	BadgeDel       lipgloss.Style
	BadgeRen       lipgloss.Style
	BadgeUnt       lipgloss.Style
	BadgeCnf       lipgloss.Style
	BadgeType      lipgloss.Style

	// Path components
	DirName        lipgloss.Style
	FileName       lipgloss.Style
	Arrow          lipgloss.Style

	// Diff Stats
	StatAdd        lipgloss.Style
	StatDel        lipgloss.Style
	StatZero       lipgloss.Style
	StatBarAdd     lipgloss.Style
	StatBarDel     lipgloss.Style

	// Diff Viewer
	DiffHeader     lipgloss.Style
	DiffHunk       lipgloss.Style
	DiffAddLine    lipgloss.Style
	DiffDelLine    lipgloss.Style
	DiffContext    lipgloss.Style

	// Footer / Hints
	FooterHint     lipgloss.Style
	KeyBadge       lipgloss.Style
	KeyDesc        lipgloss.Style

	// Interactive TUI
	Cursor         lipgloss.Style
	ActiveRow      lipgloss.Style
	InactiveRow    lipgloss.Style
	CheckStaged    lipgloss.Style
	CheckUnstaged  lipgloss.Style
	CheckUntracked lipgloss.Style
	PaneTitle      lipgloss.Style
	PaneBorder     lipgloss.Style
}

// DefaultStyles returns an instantiated set of Dracula Lipgloss styles
func DefaultStyles() Styles {
	return Styles{
		// Card / Frame
		Card: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Purple).
			Padding(0, 1),

		CleanCard: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Green).
			Padding(1, 2).
			Align(lipgloss.Center),

		WarningCard: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Red).
			Padding(1, 2),

		// Header elements
		RepoBadge: lipgloss.NewStyle().
			Background(Purple).
			Foreground(DarkerBg).
			Bold(true).
			Padding(0, 1),

		BranchBadge: lipgloss.NewStyle().
			Foreground(Cyan).
			Bold(true),

		RemoteSynced: lipgloss.NewStyle().
			Foreground(Green),

		RemoteAhead: lipgloss.NewStyle().
			Foreground(Yellow).
			Bold(true),

		RemoteBehind: lipgloss.NewStyle().
			Foreground(Orange).
			Bold(true),

		RemoteDiverged: lipgloss.NewStyle().
			Foreground(Red).
			Bold(true),

		CommitHash: lipgloss.NewStyle().
			Foreground(Pink).
			Bold(true),

		CommitSubject: lipgloss.NewStyle().
			Foreground(Foreground),

		CommitMeta: lipgloss.NewStyle().
			Foreground(Comment),

		StashBadge: lipgloss.NewStyle().
			Background(CurrentLine).
			Foreground(Orange).
			Bold(true).
			Padding(0, 1),

		StateBadge: lipgloss.NewStyle().
			Background(Red).
			Foreground(Foreground).
			Bold(true).
			Padding(0, 1),

		// Summary Pills
		PillStaged: lipgloss.NewStyle().
			Background(Green).
			Foreground(DarkerBg).
			Bold(true).
			Padding(0, 1),

		PillUnstaged: lipgloss.NewStyle().
			Background(Yellow).
			Foreground(DarkerBg).
			Bold(true).
			Padding(0, 1),

		PillUntracked: lipgloss.NewStyle().
			Background(Pink).
			Foreground(DarkerBg).
			Bold(true).
			Padding(0, 1),

		PillConflicts: lipgloss.NewStyle().
			Background(Red).
			Foreground(Foreground).
			Bold(true).
			Padding(0, 1),

		PillCount: lipgloss.NewStyle().
			Foreground(Foreground).
			Bold(true),

		// Section Headers
		SectionStaged: lipgloss.NewStyle().
			Foreground(Green).
			Bold(true),

		SectionUnstaged: lipgloss.NewStyle().
			Foreground(Yellow).
			Bold(true),

		SectionUntracked: lipgloss.NewStyle().
			Foreground(Pink).
			Bold(true),

		SectionConflicts: lipgloss.NewStyle().
			Foreground(Red).
			Bold(true),

		SectionCount: lipgloss.NewStyle().
			Foreground(Comment),

		// Status Badges
		BadgeAdd: lipgloss.NewStyle().
			Background(Green).
			Foreground(DarkerBg).
			Bold(true).
			Width(6).
			Align(lipgloss.Center),

		BadgeMod: lipgloss.NewStyle().
			Background(Yellow).
			Foreground(DarkerBg).
			Bold(true).
			Width(6).
			Align(lipgloss.Center),

		BadgeDel: lipgloss.NewStyle().
			Background(Red).
			Foreground(Foreground).
			Bold(true).
			Width(6).
			Align(lipgloss.Center),

		BadgeRen: lipgloss.NewStyle().
			Background(Purple).
			Foreground(DarkerBg).
			Bold(true).
			Width(6).
			Align(lipgloss.Center),

		BadgeUnt: lipgloss.NewStyle().
			Background(Pink).
			Foreground(DarkerBg).
			Bold(true).
			Width(6).
			Align(lipgloss.Center),

		BadgeCnf: lipgloss.NewStyle().
			Background(Red).
			Foreground(Foreground).
			Bold(true).
			Width(6).
			Align(lipgloss.Center),

		BadgeType: lipgloss.NewStyle().
			Background(Cyan).
			Foreground(DarkerBg).
			Bold(true).
			Width(6).
			Align(lipgloss.Center),

		// Path components
		DirName: lipgloss.NewStyle().
			Foreground(Comment),

		FileName: lipgloss.NewStyle().
			Foreground(Foreground).
			Bold(true),

		Arrow: lipgloss.NewStyle().
			Foreground(Cyan),

		// Diff Stats
		StatAdd: lipgloss.NewStyle().
			Foreground(Green).
			Bold(true),

		StatDel: lipgloss.NewStyle().
			Foreground(Red).
			Bold(true),

		StatZero: lipgloss.NewStyle().
			Foreground(Comment),

		StatBarAdd: lipgloss.NewStyle().
			Foreground(Green),

		StatBarDel: lipgloss.NewStyle().
			Foreground(Red),

		// Diff Viewer
		DiffHeader: lipgloss.NewStyle().
			Foreground(Purple).
			Bold(true),

		DiffHunk: lipgloss.NewStyle().
			Foreground(Cyan),

		DiffAddLine: lipgloss.NewStyle().
			Foreground(Green).
			Background(lipgloss.Color("#1e3b2b")),

		DiffDelLine: lipgloss.NewStyle().
			Foreground(Red).
			Background(lipgloss.Color("#3d1e24")),

		DiffContext: lipgloss.NewStyle().
			Foreground(Foreground),

		// Footer / Hints
		FooterHint: lipgloss.NewStyle().
			Foreground(Comment),

		KeyBadge: lipgloss.NewStyle().
			Background(CurrentLine).
			Foreground(Pink).
			Bold(true).
			Padding(0, 1),

		KeyDesc: lipgloss.NewStyle().
			Foreground(Foreground),

		// Interactive TUI
		Cursor: lipgloss.NewStyle().
			Foreground(Pink).
			Bold(true),

		ActiveRow: lipgloss.NewStyle().
			Background(CurrentLine).
			Foreground(Foreground),

		InactiveRow: lipgloss.NewStyle(),

		CheckStaged: lipgloss.NewStyle().
			Foreground(Green).
			Bold(true),

		CheckUnstaged: lipgloss.NewStyle().
			Foreground(Yellow),

		CheckUntracked: lipgloss.NewStyle().
			Foreground(Pink),

		PaneTitle: lipgloss.NewStyle().
			Background(Purple).
			Foreground(DarkerBg).
			Bold(true).
			Padding(0, 1),

		PaneBorder: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Comment),
	}
}
