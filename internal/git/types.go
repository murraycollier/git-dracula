package git

import (
	"path/filepath"
	"strings"
)

// FileSection represents the staging status section
type FileSection int

const (
	SectionStaged FileSection = iota
	SectionUnstaged
	SectionUntracked
	SectionConflict
)

func (s FileSection) String() string {
	switch s {
	case SectionStaged:
		return "Staged"
	case SectionUnstaged:
		return "Unstaged"
	case SectionUntracked:
		return "Untracked"
	case SectionConflict:
		return "Conflict"
	default:
		return "Unknown"
	}
}

// FileStatus represents a single changed or untracked file
type FileStatus struct {
	Path        string      `json:"path"`
	OrigPath    string      `json:"orig_path,omitempty"`
	Section     FileSection `json:"section"`
	StatusCode  string      `json:"status_code"`  // "A", "M", "D", "R", "?", "UU", etc.
	StatusLabel string      `json:"status_label"` // "added", "modified", etc.
	Additions   int         `json:"additions"`
	Deletions   int         `json:"deletions"`
	IsBinary    bool        `json:"is_binary"`
}

// Dir returns the directory portion of the path (with trailing slash)
func (f FileStatus) Dir() string {
	dir := filepath.Dir(f.Path)
	if dir == "." {
		return ""
	}
	return dir + "/"
}

// Base returns the filename portion of the path
func (f FileStatus) Base() string {
	return filepath.Base(f.Path)
}

// BranchInfo contains branch details and tracking information
type BranchInfo struct {
	Head     string `json:"head"`
	OID      string `json:"oid"`
	Upstream string `json:"upstream,omitempty"`
	Ahead    int    `json:"ahead"`
	Behind   int    `json:"behind"`
	Detached bool   `json:"detached"`
	Initial  bool   `json:"initial"`
}

// CommitInfo holds information about the latest commit on HEAD
type CommitInfo struct {
	Hash         string `json:"hash"`
	FullHash     string `json:"full_hash"`
	Subject      string `json:"subject"`
	Author       string `json:"author"`
	RelativeTime string `json:"relative_time"`
}

// RepoState indicates special states like merge, rebase, etc.
type RepoState string

const (
	StateNormal     RepoState = ""
	StateMerging    RepoState = "MERGING"
	StateRebasing   RepoState = "REBASING"
	StateCherryPick RepoState = "CHERRY-PICKING"
	StateBisecting  RepoState = "BISECTING"
	StateReverting  RepoState = "REVERTING"
)

// StatusResult aggregates complete repository status
type StatusResult struct {
	RepoName    string        `json:"repo_name"`
	RepoRoot    string        `json:"repo_root"`
	Branch      BranchInfo    `json:"branch"`
	HeadCommit  *CommitInfo   `json:"head_commit,omitempty"`
	StashCount  int           `json:"stash_count"`
	State       RepoState     `json:"state,omitempty"`
	StateDetail string        `json:"state_detail,omitempty"`

	Staged    []FileStatus `json:"staged"`
	Unstaged  []FileStatus `json:"unstaged"`
	Untracked []FileStatus `json:"untracked"`
	Conflicts []FileStatus `json:"conflicts"`
}

// IsClean returns true if working tree has no staged, unstaged, untracked, or conflict files
func (r *StatusResult) IsClean() bool {
	return len(r.Staged) == 0 && len(r.Unstaged) == 0 && len(r.Untracked) == 0 && len(r.Conflicts) == 0
}

// TotalChanges returns count of all file items across all sections
func (r *StatusResult) TotalChanges() int {
	return len(r.Staged) + len(r.Unstaged) + len(r.Untracked) + len(r.Conflicts)
}

// AllFiles returns all files in a unified ordered slice for interactive navigation
func (r *StatusResult) AllFiles() []FileStatus {
	all := make([]FileStatus, 0, r.TotalChanges())
	all = append(all, r.Conflicts...)
	all = append(all, r.Staged...)
	all = append(all, r.Unstaged...)
	all = append(all, r.Untracked...)
	return all
}

// SectionHeader returns a human readable representation of the section
func SectionHeader(sec FileSection) string {
	switch sec {
	case SectionStaged:
		return "✔ STAGED CHANGES"
	case SectionUnstaged:
		return "✎ UNSTAGED CHANGES"
	case SectionUntracked:
		return "? UNTRACKED FILES"
	case SectionConflict:
		return "✖ MERGE CONFLICTS"
	default:
		return strings.ToUpper(sec.String())
	}
}
