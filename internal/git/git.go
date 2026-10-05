package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// Client handles git repository interactions
type Client struct {
	RepoDir string
}

// NewClient creates a new git client for repoDir. If repoDir is empty, current directory is used.
func NewClient(repoDir string) *Client {
	if repoDir == "" {
		repoDir = "."
	}
	return &Client{RepoDir: repoDir}
}

// run runs a git command in RepoDir and returns stdout and error
func (c *Client) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = c.RepoDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

// IsInsideWorkTree checks if current directory is inside a Git working tree
func (c *Client) IsInsideWorkTree() bool {
	out, err := c.run("rev-parse", "--is-inside-work-tree")
	return err == nil && strings.TrimSpace(out) == "true"
}

// GetRepoRoot returns the top level directory of the repository
func (c *Client) GetRepoRoot() (string, error) {
	out, err := c.run("rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// GetGitDir returns the .git directory path
func (c *Client) GetGitDir() (string, error) {
	out, err := c.run("rev-parse", "--git-dir")
	if err != nil {
		return "", err
	}
	p := strings.TrimSpace(out)
	if !filepath.IsAbs(p) {
		root, _ := c.GetRepoRoot()
		p = filepath.Join(root, p)
	}
	return p, nil
}

// GetStatus returns the full status result for the repository
func (c *Client) GetStatus() (*StatusResult, error) {
	if !c.IsInsideWorkTree() {
		return nil, fmt.Errorf("not a git repository (or any of the parent directories)")
	}

	root, err := c.GetRepoRoot()
	if err != nil {
		return nil, err
	}

	rawStatus, err := c.run("status", "--porcelain=v2", "--branch", "--ahead-behind", "--untracked-files=all")
	if err != nil {
		return nil, fmt.Errorf("failed to get git status: %w", err)
	}

	result := ParsePorcelainV2(rawStatus)
	result.RepoRoot = root
	result.RepoName = filepath.Base(root)

	// Concurrently fetch numstats, commit, stash, and state
	var wg sync.WaitGroup
	var cachedStats, unstagedStats map[string]DiffStatEntry
	var commit *CommitInfo
	var stashCount int
	var state RepoState
	var stateDetail string

	// Cached numstat
	wg.Add(1)
	go func() {
		defer wg.Done()
		if out, err := c.run("diff", "--cached", "--numstat"); err == nil {
			cachedStats = ParseNumstat(out)
		}
	}()

	// Unstaged numstat
	wg.Add(1)
	go func() {
		defer wg.Done()
		if out, err := c.run("diff", "--numstat"); err == nil {
			unstagedStats = ParseNumstat(out)
		}
	}()

	// Last commit
	wg.Add(1)
	go func() {
		defer wg.Done()
		commit = c.getLatestCommit()
	}()

	// Stash count
	wg.Add(1)
	go func() {
		defer wg.Done()
		stashCount = c.getStashCount()
	}()

	// Repo state
	wg.Add(1)
	go func() {
		defer wg.Done()
		state, stateDetail = c.getRepoState()
	}()

	wg.Wait()

	// Attach stats to files
	for i := range result.Staged {
		p := result.Staged[i].Path
		if stat, ok := cachedStats[p]; ok {
			result.Staged[i].Additions = stat.Additions
			result.Staged[i].Deletions = stat.Deletions
			result.Staged[i].IsBinary = stat.IsBinary
		}
	}

	for i := range result.Unstaged {
		p := result.Unstaged[i].Path
		if stat, ok := unstagedStats[p]; ok {
			result.Unstaged[i].Additions = stat.Additions
			result.Unstaged[i].Deletions = stat.Deletions
			result.Unstaged[i].IsBinary = stat.IsBinary
		}
	}

	result.HeadCommit = commit
	result.StashCount = stashCount
	result.State = state
	result.StateDetail = stateDetail

	return result, nil
}

func (c *Client) getLatestCommit() *CommitInfo {
	out, err := c.run("log", "-1", "--pretty=format:%h%x00%H%x00%s%x00%an%x00%cr")
	if err != nil || len(out) == 0 {
		return nil
	}

	parts := strings.Split(out, "\x00")
	if len(parts) < 5 {
		return nil
	}

	return &CommitInfo{
		Hash:         parts[0],
		FullHash:     parts[1],
		Subject:      parts[2],
		Author:       parts[3],
		RelativeTime: parts[4],
	}
}

func (c *Client) getStashCount() int {
	out, err := c.run("stash", "list")
	if err != nil || len(strings.TrimSpace(out)) == 0 {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	return len(lines)
}

func (c *Client) getRepoState() (RepoState, string) {
	gitDir, err := c.GetGitDir()
	if err != nil {
		return StateNormal, ""
	}

	if exists(filepath.Join(gitDir, "MERGE_HEAD")) {
		return StateMerging, "Merge in progress"
	}

	if exists(filepath.Join(gitDir, "rebase-merge")) || exists(filepath.Join(gitDir, "rebase-apply")) {
		return StateRebasing, "Rebase in progress"
	}

	if exists(filepath.Join(gitDir, "CHERRY_PICK_HEAD")) {
		return StateCherryPick, "Cherry-pick in progress"
	}

	if exists(filepath.Join(gitDir, "BISECT_LOG")) {
		return StateBisecting, "Bisecting"
	}

	if exists(filepath.Join(gitDir, "REVERT_HEAD")) {
		return StateReverting, "Revert in progress"
	}

	return StateNormal, ""
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// StageFile stages a file
func (c *Client) StageFile(path string) error {
	_, err := c.run("add", "--", path)
	return err
}

// UnstageFile unstages a file
func (c *Client) UnstageFile(path string) error {
	// Try `git restore --staged -- <path>` first, fallback to `git reset HEAD -- <path>`
	_, err := c.run("restore", "--staged", "--", path)
	if err != nil {
		_, err = c.run("reset", "HEAD", "--", path)
	}
	return err
}

// StageAll stages all changes
func (c *Client) StageAll() error {
	_, err := c.run("add", "-A")
	return err
}

// UnstageAll unstages all files
func (c *Client) UnstageAll() error {
	_, err := c.run("restore", "--staged", ".")
	if err != nil {
		_, err = c.run("reset", "HEAD")
	}
	return err
}

// DiscardChanges discards working tree changes for a file
func (c *Client) DiscardChanges(file FileStatus) error {
	if file.Section == SectionUntracked {
		root, _ := c.GetRepoRoot()
		fullPath := filepath.Join(root, file.Path)
		return os.RemoveAll(fullPath)
	}
	_, err := c.run("restore", "--", file.Path)
	if err != nil {
		_, err = c.run("checkout", "--", file.Path)
	}
	return err
}

// GetDiff returns colored or raw diff for a file
func (c *Client) GetDiff(path string, staged bool) (string, error) {
	args := []string{"diff"}
	if staged {
		args = append(args, "--cached")
	}
	args = append(args, "--", path)

	out, err := c.run(args...)
	if err != nil {
		return "", err
	}
	if len(out) == 0 && !staged {
		// If untracked file, display whole file content as additions
		root, _ := c.GetRepoRoot()
		fullPath := filepath.Join(root, path)
		if data, readErr := os.ReadFile(fullPath); readErr == nil {
			var b strings.Builder
			b.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", path, path))
			b.WriteString(fmt.Sprintf("--- /dev/null\n+++ b/%s\n", path))
			lines := strings.Split(string(data), "\n")
			b.WriteString(fmt.Sprintf("@@ -0,0 +1,%d @@\n", len(lines)))
			for _, line := range lines {
				b.WriteString("+" + line + "\n")
			}
			return b.String(), nil
		}
	}
	return out, nil
}

// Commit creates a commit with the given message
func (c *Client) Commit(msg string) error {
	if strings.TrimSpace(msg) == "" {
		return fmt.Errorf("commit message cannot be empty")
	}
	_, err := c.run("commit", "-m", msg)
	return err
}

// Stash saves unstaged and staged changes into stash
func (c *Client) Stash(msg string) error {
	args := []string{"stash", "push"}
	if strings.TrimSpace(msg) != "" {
		args = append(args, "-m", msg)
	}
	_, err := c.run(args...)
	return err
}
