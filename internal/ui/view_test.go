package ui

import (
	"strings"
	"testing"

	"github.com/murraycollier/git-dracula/internal/git"
	"github.com/murraycollier/git-dracula/internal/theme"
)

func TestRenderCleanState(t *testing.T) {
	styles := theme.DefaultStyles()
	status := &git.StatusResult{
		RepoName: "test-repo",
		Branch: git.BranchInfo{
			Head:     "main",
			Upstream: "origin/main",
		},
	}

	out := RenderFull(status, nil, ViewConfig{Width: 80}, styles)
	if !strings.Contains(out, "Working tree is spotless") {
		t.Errorf("expected clean message, got %s", out)
	}
	if !strings.Contains(out, "test-repo") {
		t.Errorf("expected repo name, got %s", out)
	}
}

func TestRenderDirtyState(t *testing.T) {
	styles := theme.DefaultStyles()
	status := &git.StatusResult{
		RepoName: "dracula-app",
		Branch: git.BranchInfo{
			Head:     "feature/tui",
			Upstream: "origin/feature/tui",
			Ahead:    2,
			Behind:   1,
		},
		HeadCommit: &git.CommitInfo{
			Hash:         "8a1b2c3",
			Subject:      "feat(cli): initial commit",
			Author:       "Murray Collier",
			RelativeTime: "5m ago",
		},
		StashCount: 1,
		Staged: []git.FileStatus{
			{Path: "main.go", StatusCode: "A", Additions: 40, Deletions: 0},
		},
		Unstaged: []git.FileStatus{
			{Path: "theme.go", StatusCode: "M", Additions: 10, Deletions: 5},
		},
		Untracked: []git.FileStatus{
			{Path: "README.md", StatusCode: "?"},
		},
	}

	out := RenderFull(status, nil, ViewConfig{Width: 90}, styles)

	if !strings.Contains(out, "dracula-app") {
		t.Errorf("expected dracula-app in output")
	}
	if !strings.Contains(out, "feature/tui") {
		t.Errorf("expected branch feature/tui in output")
	}
	if !strings.Contains(out, "STAGED CHANGES") {
		t.Errorf("expected staged section in output")
	}
	if !strings.Contains(out, "UNSTAGED CHANGES") {
		t.Errorf("expected unstaged section in output")
	}
	if !strings.Contains(out, "UNTRACKED FILES") {
		t.Errorf("expected untracked section in output")
	}
}

func TestRenderCompact(t *testing.T) {
	styles := theme.DefaultStyles()
	status := &git.StatusResult{
		RepoName: "dracula-app",
		Branch: git.BranchInfo{
			Head:     "main",
			Upstream: "origin/main",
		},
		Staged: []git.FileStatus{
			{Path: "staged.go", StatusCode: "M"},
		},
		Unstaged: []git.FileStatus{
			{Path: "unstaged.go", StatusCode: "M"},
		},
		Untracked: []git.FileStatus{
			{Path: "untracked.txt", StatusCode: "?"},
		},
	}

	out := RenderCompact(status, styles)
	if !strings.Contains(out, "## dracula-app") {
		t.Errorf("expected compact header, got: %s", out)
	}
	if !strings.Contains(out, "staged.go") || !strings.Contains(out, "unstaged.go") {
		t.Errorf("expected files in compact output, got: %s", out)
	}
}

func TestRenderDiffBar(t *testing.T) {
	styles := theme.DefaultStyles()
	bar := RenderDiffBar(10, 5, 10, styles)
	if bar == "" {
		t.Errorf("expected non-empty diff bar")
	}

	empty := RenderDiffBar(0, 0, 10, styles)
	if empty != "" {
		t.Errorf("expected empty diff bar for 0, 0")
	}
}

func TestHighlightDiff(t *testing.T) {
	styles := theme.DefaultStyles()
	rawDiff := "diff --git a/file.go b/file.go\n@@ -1,3 +1,4 @@\n context\n+addition\n-deletion\n"
	highlighted := HighlightDiff(rawDiff, styles, 80)
	if !strings.Contains(highlighted, "addition") || !strings.Contains(highlighted, "deletion") {
		t.Errorf("expected diff lines, got: %s", highlighted)
	}
}
