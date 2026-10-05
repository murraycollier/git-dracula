package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/murraycollier/git-dracula/internal/git"
	"github.com/murraycollier/git-dracula/internal/theme"
	"github.com/murraycollier/git-dracula/internal/ui"
)

const version = "1.0.0"

func main() {
	var (
		flagInteractive = flag.Bool("i", false, "Launch interactive Bubble Tea TUI")
		flagInteractiveLong = flag.Bool("interactive", false, "Launch interactive Bubble Tea TUI")
		flagShort       = flag.Bool("s", false, "Show compact status format")
		flagShortLong   = flag.Bool("short", false, "Show compact status format")
		flagDiff        = flag.Bool("d", false, "Show inline diffs")
		flagDiffLong    = flag.Bool("diff", false, "Show inline diffs")
		flagDir         = flag.String("C", "", "Run as if git was started in <path>")
		flagVersion     = flag.Bool("v", false, "Show version")
		flagVersionLong = flag.Bool("version", false, "Show version")
		flagHelp        = flag.Bool("h", false, "Show help")
		flagHelpLong    = flag.Bool("help", false, "Show help")
	)

	flag.Usage = func() {
		styles := theme.DefaultStyles()
		header := styles.RepoBadge.Render(" 🧛 GIT-DRACULA ") + " " + styles.CommitMeta.Render("v"+version)
		fmt.Fprintf(os.Stderr, "\n%s\n", header)
		fmt.Fprintf(os.Stderr, "%s\n\n", styles.CommitSubject.Render("A gorgeous, Dracula-themed Git status and staging tool built with Charm."))
		fmt.Fprintf(os.Stderr, "%s\n", styles.SectionStaged.Render("USAGE:"))
		fmt.Fprintf(os.Stderr, "  git-dracula [options]\n")
		fmt.Fprintf(os.Stderr, "  git dracula [options]    (when in PATH)\n\n")
		fmt.Fprintf(os.Stderr, "%s\n", styles.SectionStaged.Render("OPTIONS:"))
		fmt.Fprintf(os.Stderr, "  -i, --interactive   Launch interactive staging & diff browser (Bubble Tea)\n")
		fmt.Fprintf(os.Stderr, "  -s, --short         Compact status representation (similar to git status -s)\n")
		fmt.Fprintf(os.Stderr, "  -d, --diff          Show inline Dracula-highlighted diffs\n")
		fmt.Fprintf(os.Stderr, "  -C <path>           Run in specified repository directory\n")
		fmt.Fprintf(os.Stderr, "  -v, --version       Print version information\n")
		fmt.Fprintf(os.Stderr, "  -h, --help          Show this help message\n\n")
		fmt.Fprintf(os.Stderr, "%s\n", styles.SectionStaged.Render("INTERACTIVE KEYBINDINGS:"))
		fmt.Fprintf(os.Stderr, "  j / k, ↑ / ↓        Navigate files\n")
		fmt.Fprintf(os.Stderr, "  space / s           Toggle stage / unstage selected file\n")
		fmt.Fprintf(os.Stderr, "  a / u               Stage all / Unstage all\n")
		fmt.Fprintf(os.Stderr, "  tab / d             Focus diff viewport / scroll\n")
		fmt.Fprintf(os.Stderr, "  c                   Commit staged changes (prompts for message)\n")
		fmt.Fprintf(os.Stderr, "  S                   Stash changes\n")
		fmt.Fprintf(os.Stderr, "  r                   Refresh repository status\n")
		fmt.Fprintf(os.Stderr, "  ?                   Toggle help cheatsheet\n")
		fmt.Fprintf(os.Stderr, "  q / esc             Quit\n\n")
	}

	flag.Parse()

	if *flagVersion || *flagVersionLong {
		fmt.Printf("git-dracula v%s\n", version)
		os.Exit(0)
	}

	if *flagHelp || *flagHelpLong {
		flag.Usage()
		os.Exit(0)
	}

	// Check if positional argument is "interactive" or "diff"
	args := flag.Args()
	isInteractive := *flagInteractive || *flagInteractiveLong
	isShort := *flagShort || *flagShortLong
	isDiff := *flagDiff || *flagDiffLong

	if len(args) > 0 {
		switch args[0] {
		case "interactive", "ui", "tui":
			isInteractive = true
		case "diff":
			isDiff = true
		case "short":
			isShort = true
		}
	}

	client := git.NewClient(*flagDir)
	styles := theme.DefaultStyles()

	if !client.IsInsideWorkTree() {
		errBox := styles.WarningCard.Width(65).Render(
			styles.SectionConflicts.Render("✖ fatal: not a git repository (or any of the parent directories)\n") +
				styles.CommitMeta.Render("Run 'git init' to initialize a new repository, or cd into a git project."),
		)
		fmt.Fprintf(os.Stderr, "\n%s\n\n", errBox)
		os.Exit(128)
	}

	status, err := client.GetStatus()
	if err != nil {
		fmt.Fprintf(os.Stderr, "git-dracula: %v\n", err)
		os.Exit(1)
	}

	if isInteractive {
		model := ui.NewInteractiveModel(client, status)
		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running interactive TUI: %v\n", err)
			os.Exit(1)
		}
		return
	}

	cfg := ui.ViewConfig{
		Short:    isShort,
		ShowDiff: isDiff,
	}

	output := ui.RenderFull(status, client, cfg, styles)
	fmt.Print(output)
}
