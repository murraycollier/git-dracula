package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/murraycollier/git-dracula/internal/git"
	"github.com/murraycollier/git-dracula/internal/theme"
)

type focusPane int

const (
	paneFiles focusPane = iota
	paneDiff
)

// InteractiveModel is the Bubble Tea model for the interactive TUI
type InteractiveModel struct {
	client       *git.Client
	status       *git.StatusResult
	files        []git.FileStatus
	cursor       int
	focus        focusPane
	diffViewport viewport.Model
	styles       theme.Styles

	// Modals
	committing   bool
	commitInput  textinput.Model
	stashing     bool
	stashInput   textinput.Model
	showHelp     bool
	flashMessage string

	width  int
	height int
	ready  bool
}

// NewInteractiveModel creates a new Bubble Tea model
func NewInteractiveModel(client *git.Client, status *git.StatusResult) InteractiveModel {
	styles := theme.DefaultStyles()

	ci := textinput.New()
	ci.Placeholder = "Commit message (press Enter to commit, Esc to cancel)..."
	ci.Focus()
	ci.Prompt = styles.CommitHash.Render("commit ❯ ")
	ci.CharLimit = 150
	ci.Width = 60

	si := textinput.New()
	si.Placeholder = "Stash message (optional, press Enter to stash, Esc to cancel)..."
	si.Focus()
	si.Prompt = styles.StashBadge.Render("stash ❯ ")
	si.CharLimit = 150
	si.Width = 60

	m := InteractiveModel{
		client:      client,
		status:      status,
		files:       status.AllFiles(),
		styles:      styles,
		commitInput: ci,
		stashInput:  si,
		focus:       paneFiles,
	}

	return m
}

func (m InteractiveModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InteractiveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		// Configure viewport for diffs
		diffWidth := m.width/2 - 4
		if diffWidth < 30 {
			diffWidth = 30
		}
		diffHeight := m.height - 8
		if diffHeight < 5 {
			diffHeight = 5
		}

		if !m.ready {
			m.diffViewport = viewport.New(diffWidth, diffHeight)
			m.ready = true
		} else {
			m.diffViewport.Width = diffWidth
			m.diffViewport.Height = diffHeight
		}

		m.updateDiffContent()

	case tea.KeyMsg:
		// Modal handling: Commit
		if m.committing {
			switch msg.String() {
			case "esc":
				m.committing = false
				m.commitInput.Reset()
				return m, nil
			case "enter":
				val := strings.TrimSpace(m.commitInput.Value())
				if val != "" {
					err := m.client.Commit(val)
					if err != nil {
						m.flashMessage = m.styles.BadgeDel.Render(" Error ") + " " + err.Error()
					} else {
						m.flashMessage = m.styles.BadgeAdd.Render(" Committed ") + " " + truncate(val, 40)
						m.committing = false
						m.commitInput.Reset()
						m.refreshStatus()
					}
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.commitInput, cmd = m.commitInput.Update(msg)
				return m, cmd
			}
		}

		// Modal handling: Stash
		if m.stashing {
			switch msg.String() {
			case "esc":
				m.stashing = false
				m.stashInput.Reset()
				return m, nil
			case "enter":
				val := strings.TrimSpace(m.stashInput.Value())
				err := m.client.Stash(val)
				if err != nil {
					m.flashMessage = m.styles.BadgeDel.Render(" Error ") + " " + err.Error()
				} else {
					m.flashMessage = m.styles.StashBadge.Render(" Stashed ") + " changes saved"
					m.stashing = false
					m.stashInput.Reset()
					m.refreshStatus()
				}
				return m, nil
			default:
				var cmd tea.Cmd
				m.stashInput, cmd = m.stashInput.Update(msg)
				return m, cmd
			}
		}

		// Modal handling: Help
		if m.showHelp {
			if msg.String() == "?" || msg.String() == "esc" || msg.String() == "q" {
				m.showHelp = false
			}
			return m, nil
		}

		// Normal mode keybindings
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "?":
			m.showHelp = true
			return m, nil

		case "tab":
			if m.focus == paneFiles {
				m.focus = paneDiff
			} else {
				m.focus = paneFiles
			}
			return m, nil

		case "up", "k":
			if m.focus == paneFiles {
				if m.cursor > 0 {
					m.cursor--
					m.updateDiffContent()
				}
			} else {
				m.diffViewport.LineUp(1)
			}

		case "down", "j":
			if m.focus == paneFiles {
				if m.cursor < len(m.files)-1 {
					m.cursor++
					m.updateDiffContent()
				}
			} else {
				m.diffViewport.LineDown(1)
			}

		case "pageup":
			m.diffViewport.ViewUp()

		case "pagedown":
			m.diffViewport.ViewDown()

		case " ", "s": // Stage or Unstage current file
			if len(m.files) > 0 && m.cursor < len(m.files) {
				file := m.files[m.cursor]
				if file.Section == git.SectionStaged {
					if err := m.client.UnstageFile(file.Path); err == nil {
						m.flashMessage = m.styles.BadgeMod.Render(" Unstaged ") + " " + file.Path
					}
				} else {
					if err := m.client.StageFile(file.Path); err == nil {
						m.flashMessage = m.styles.BadgeAdd.Render(" Staged ") + " " + file.Path
					}
				}
				m.refreshStatus()
			}

		case "a": // Stage all
			if err := m.client.StageAll(); err == nil {
				m.flashMessage = m.styles.BadgeAdd.Render(" Staged All ") + " All changes staged"
				m.refreshStatus()
			}

		case "u": // Unstage all
			if err := m.client.UnstageAll(); err == nil {
				m.flashMessage = m.styles.BadgeMod.Render(" Unstaged All ") + " Reset staged index"
				m.refreshStatus()
			}

		case "c": // Commit
			if len(m.status.Staged) == 0 {
				m.flashMessage = m.styles.BadgeDel.Render(" Notice ") + " No staged changes to commit"
			} else {
				m.committing = true
				m.commitInput.Focus()
				return m, textinput.Blink
			}

		case "S": // Stash
			m.stashing = true
			m.stashInput.Focus()
			return m, textinput.Blink

		case "r": // Refresh
			m.refreshStatus()
			m.flashMessage = m.styles.SectionStaged.Render("Refreshed status")
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *InteractiveModel) refreshStatus() {
	if newStatus, err := m.client.GetStatus(); err == nil {
		m.status = newStatus
		m.files = newStatus.AllFiles()
		if m.cursor >= len(m.files) {
			m.cursor = len(m.files) - 1
		}
		if m.cursor < 0 {
			m.cursor = 0
		}
		m.updateDiffContent()
	}
}

func (m *InteractiveModel) updateDiffContent() {
	if len(m.files) == 0 || m.cursor >= len(m.files) {
		m.diffViewport.SetContent(m.styles.CommitMeta.Render("No file selected or tree is clean."))
		return
	}

	curFile := m.files[m.cursor]
	isStaged := curFile.Section == git.SectionStaged
	diff, err := m.client.GetDiff(curFile.Path, isStaged)
	if err != nil || strings.TrimSpace(diff) == "" {
		m.diffViewport.SetContent(m.styles.CommitMeta.Render("No diff content for " + curFile.Path))
		return
	}

	highlighted := HighlightDiff(diff, m.styles, m.diffViewport.Width-2)
	m.diffViewport.SetContent(highlighted)
}

func (m InteractiveModel) View() string {
	if !m.ready {
		return "Initializing Dracula..."
	}

	if m.showHelp {
		return m.renderHelpModal()
	}

	if m.committing {
		return m.renderCommitModal()
	}

	if m.stashing {
		return m.renderStashModal()
	}

	var sb strings.Builder

	// 1. Header Mini Card
	header := renderHeaderCard(m.status, m.styles, m.width)
	sb.WriteString(header + "\n")

	// 2. Summary Bar
	summary := renderSummaryBar(m.status, m.styles)
	sb.WriteString(summary + "\n\n")

	// 3. Main Split View: Files list (Left) vs Diff Viewport (Right)
	leftWidth := m.width/2 - 2
	rightWidth := m.width - leftWidth - 4
	if leftWidth < 30 {
		leftWidth = 30
	}
	if rightWidth < 30 {
		rightWidth = 30
	}

	filesView := m.renderFilesPane(leftWidth)
	diffView := m.renderDiffPane(rightWidth)

	split := lipgloss.JoinHorizontal(lipgloss.Top, filesView, "  ", diffView)
	sb.WriteString(split + "\n")

	// 4. Flash Message / Status bar
	if m.flashMessage != "" {
		sb.WriteString("  " + m.flashMessage + "\n")
	} else {
		sb.WriteString("\n")
	}

	// 5. Interactive Footer Help Bar
	sb.WriteString(m.renderInteractiveFooter())

	return sb.String()
}

func (m InteractiveModel) renderFilesPane(width int) string {
	var sb strings.Builder

	titleStyle := m.styles.PaneTitle
	if m.focus == paneFiles {
		titleStyle = titleStyle.Copy().Background(theme.Pink).Foreground(theme.DarkerBg)
	}
	sb.WriteString(titleStyle.Render(" FILES ") + "\n")

	if len(m.files) == 0 {
		clean := m.styles.SectionStaged.Render("✓ Working tree clean")
		sb.WriteString("  " + clean + "\n")
		return m.styles.PaneBorder.Width(width).Render(sb.String())
	}

	var currentSection git.FileSection = -1
	for idx, file := range m.files {
		// Section separator header
		if file.Section != currentSection {
			currentSection = file.Section
			secHeader := git.SectionHeader(currentSection)
			var hStyle lipgloss.Style
			switch currentSection {
			case git.SectionStaged:
				hStyle = m.styles.SectionStaged
			case git.SectionUnstaged:
				hStyle = m.styles.SectionUnstaged
			case git.SectionUntracked:
				hStyle = m.styles.SectionUntracked
			case git.SectionConflict:
				hStyle = m.styles.SectionConflicts
			}
			sb.WriteString("\n " + hStyle.Render(secHeader) + "\n")
		}

		// Cursor
		cursor := "  "
		isCurrent := idx == m.cursor
		if isCurrent {
			cursor = m.styles.Cursor.Render("▸ ")
		}

		// Checkbox / State symbol
		var check string
		switch file.Section {
		case git.SectionStaged:
			check = m.styles.CheckStaged.Render("[✔]")
		case git.SectionUnstaged:
			check = m.styles.CheckUnstaged.Render("[ ]")
		case git.SectionUntracked:
			check = m.styles.CheckUntracked.Render("[?]")
		case git.SectionConflict:
			check = m.styles.BadgeCnf.Render("[!]")
		}

		badge := file.StatusCode
		path := file.Path
		if len(path) > width-20 {
			path = "..." + path[len(path)-(width-23):]
		}

		rowContent := fmt.Sprintf("%s%s %s %s", cursor, check, badge, path)
		if isCurrent {
			rowContent = m.styles.ActiveRow.Width(width - 4).Render(rowContent)
		} else {
			rowContent = m.styles.InactiveRow.Render(rowContent)
		}

		sb.WriteString(rowContent + "\n")
	}

	paneBorder := m.styles.PaneBorder
	if m.focus == paneFiles {
		paneBorder = paneBorder.Copy().BorderForeground(theme.Pink)
	}

	return paneBorder.Width(width).Height(m.height - 11).Render(sb.String())
}

func (m InteractiveModel) renderDiffPane(width int) string {
	var sb strings.Builder

	titleStyle := m.styles.PaneTitle
	if m.focus == paneDiff {
		titleStyle = titleStyle.Copy().Background(theme.Cyan).Foreground(theme.DarkerBg)
	}

	var curFileName string
	if len(m.files) > 0 && m.cursor < len(m.files) {
		curFileName = " — " + m.files[m.cursor].Path
	}
	sb.WriteString(titleStyle.Render(" DIFF PREVIEW"+curFileName) + "\n")
	sb.WriteString(m.diffViewport.View())

	paneBorder := m.styles.PaneBorder
	if m.focus == paneDiff {
		paneBorder = paneBorder.Copy().BorderForeground(theme.Cyan)
	}

	return paneBorder.Width(width).Height(m.height - 11).Render(sb.String())
}

func (m InteractiveModel) renderInteractiveFooter() string {
	keys := []string{
		m.styles.KeyBadge.Render("space/s") + " " + m.styles.KeyDesc.Render("stage/unstage"),
		m.styles.KeyBadge.Render("a") + " " + m.styles.KeyDesc.Render("stage all"),
		m.styles.KeyBadge.Render("u") + " " + m.styles.KeyDesc.Render("unstage all"),
		m.styles.KeyBadge.Render("c") + " " + m.styles.KeyDesc.Render("commit"),
		m.styles.KeyBadge.Render("S") + " " + m.styles.KeyDesc.Render("stash"),
		m.styles.KeyBadge.Render("tab") + " " + m.styles.KeyDesc.Render("switch pane"),
		m.styles.KeyBadge.Render("?") + " " + m.styles.KeyDesc.Render("help"),
		m.styles.KeyBadge.Render("q") + " " + m.styles.KeyDesc.Render("quit"),
	}

	return "  " + strings.Join(keys, "  ")
}

func (m InteractiveModel) renderCommitModal() string {
	var sb strings.Builder
	sb.WriteString(m.styles.RepoBadge.Render(" GIT COMMIT ") + "\n\n")
	sb.WriteString(m.styles.SectionStaged.Render(fmt.Sprintf("%d staged changes ready to be committed\n\n", len(m.status.Staged))))
	sb.WriteString(m.commitInput.View() + "\n\n")
	sb.WriteString(m.styles.CommitMeta.Render("Press [Enter] to commit, [Esc] to cancel"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.styles.Card.Padding(2, 4).Render(sb.String()))
}

func (m InteractiveModel) renderStashModal() string {
	var sb strings.Builder
	sb.WriteString(m.styles.StashBadge.Render(" GIT STASH ") + "\n\n")
	sb.WriteString(m.styles.CommitSubject.Render("Save changes to stash stack\n\n"))
	sb.WriteString(m.stashInput.View() + "\n\n")
	sb.WriteString(m.styles.CommitMeta.Render("Press [Enter] to stash, [Esc] to cancel"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.styles.Card.Padding(2, 4).Render(sb.String()))
}

func (m InteractiveModel) renderHelpModal() string {
	var sb strings.Builder
	sb.WriteString(m.styles.RepoBadge.Render(" GIT DRACULA — KEYBOARD CHEATSHEET ") + "\n\n")

	cheatsheet := [][]string{
		{"j / ↓", "Move cursor down"},
		{"k / ↑", "Move cursor up"},
		{"space / s", "Toggle Stage / Unstage selected file"},
		{"a", "Stage all changes (git add -A)"},
		{"u", "Unstage all files (git restore --staged .)"},
		{"tab / d", "Toggle focus between Files and Diff Viewport"},
		{"pgup / pgdown", "Scroll diff viewer"},
		{"c", "Commit staged changes with prompt"},
		{"S", "Stash changes with optional message"},
		{"r", "Refresh status"},
		{"?", "Toggle this help modal"},
		{"q / esc / ctrl+c", "Quit back to shell"},
	}

	for _, item := range cheatsheet {
		key := m.styles.KeyBadge.Render(item[0])
		desc := m.styles.KeyDesc.Render(item[1])
		sb.WriteString(fmt.Sprintf("  %-25s  %s\n", key, desc))
	}

	sb.WriteString("\n" + m.styles.CommitMeta.Render("Press [?] or [Esc] to return"))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center,
		m.styles.Card.Padding(2, 4).Render(sb.String()))
}
