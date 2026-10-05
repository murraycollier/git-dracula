package git

import (
	"testing"
)

func TestParsePorcelainV2(t *testing.T) {
	input := `# branch.oid 1234567890abcdef1234567890abcdef12345678
# branch.head feature/dracula-theme
# branch.upstream origin/feature/dracula-theme
# branch.ab +3 -1
1 .M N... 100644 100644 100644 ce013625030ba8dba906f756967f9e9ca394464a ce013625030ba8dba906f756967f9e9ca394464a internal/ui/view.go
1 A. N... 000000 100644 100644 0000000000000000000000000000000000000000 1234567890123456789012345678901234567890 cmd/git-dracula/main.go
1 MM N... 100644 100644 100644 1111111111111111111111111111111111111111 2222222222222222222222222222222222222222 README.md
2 R. N... 100644 100644 100644 aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb R100 new_name.go	old_name.go
u UU N... 100644 100644 100644 100644 cccccccccccccccccccccccccccccccccccccccc dddddddddddddddddddddddddddddddddddddddd eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee conflict.go
? untracked_file.txt
`

	res := ParsePorcelainV2(input)

	if res.Branch.Head != "feature/dracula-theme" {
		t.Errorf("expected branch head feature/dracula-theme, got %s", res.Branch.Head)
	}
	if res.Branch.Upstream != "origin/feature/dracula-theme" {
		t.Errorf("expected upstream origin/feature/dracula-theme, got %s", res.Branch.Upstream)
	}
	if res.Branch.Ahead != 3 {
		t.Errorf("expected ahead 3, got %d", res.Branch.Ahead)
	}
	if res.Branch.Behind != 1 {
		t.Errorf("expected behind 1, got %d", res.Branch.Behind)
	}

	// Staged:
	// - cmd/git-dracula/main.go (A)
	// - README.md (M)
	// - new_name.go (R)
	if len(res.Staged) != 3 {
		t.Fatalf("expected 3 staged files, got %d", len(res.Staged))
	}
	if res.Staged[0].Path != "cmd/git-dracula/main.go" || res.Staged[0].StatusCode != "A" {
		t.Errorf("unexpected staged file: %+v", res.Staged[0])
	}
	if res.Staged[1].Path != "README.md" || res.Staged[1].StatusCode != "M" {
		t.Errorf("unexpected staged file: %+v", res.Staged[1])
	}
	if res.Staged[2].Path != "new_name.go" || res.Staged[2].OrigPath != "old_name.go" || res.Staged[2].StatusCode != "R" {
		t.Errorf("unexpected staged rename: %+v", res.Staged[2])
	}

	// Unstaged:
	// - internal/ui/view.go (M)
	// - README.md (M)
	if len(res.Unstaged) != 2 {
		t.Fatalf("expected 2 unstaged files, got %d", len(res.Unstaged))
	}
	if res.Unstaged[0].Path != "internal/ui/view.go" {
		t.Errorf("unexpected unstaged file: %+v", res.Unstaged[0])
	}
	if res.Unstaged[1].Path != "README.md" {
		t.Errorf("unexpected unstaged file: %+v", res.Unstaged[1])
	}

	// Untracked:
	if len(res.Untracked) != 1 || res.Untracked[0].Path != "untracked_file.txt" {
		t.Errorf("unexpected untracked: %+v", res.Untracked)
	}

	// Conflicts:
	if len(res.Conflicts) != 1 || res.Conflicts[0].Path != "conflict.go" {
		t.Errorf("unexpected conflicts: %+v", res.Conflicts)
	}

	if res.IsClean() {
		t.Errorf("expected not clean")
	}

	if res.TotalChanges() != 7 {
		t.Errorf("expected 7 total changes, got %d", res.TotalChanges())
	}
}

func TestParseNumstat(t *testing.T) {
	input := "15\t3\tfile1.go\n-\t-\timage.png\n42\t0\tdir/{old.go => new.go}\n"
	stats := ParseNumstat(input)

	if s, ok := stats["file1.go"]; !ok || s.Additions != 15 || s.Deletions != 3 || s.IsBinary {
		t.Errorf("unexpected stat for file1.go: %+v", s)
	}
	if s, ok := stats["image.png"]; !ok || !s.IsBinary {
		t.Errorf("unexpected stat for image.png: %+v", s)
	}
	if s, ok := stats["dir/new.go"]; !ok || s.Additions != 42 || s.Deletions != 0 {
		t.Errorf("unexpected stat for dir/new.go: %+v", s)
	}
}

func TestCleanPath(t *testing.T) {
	if got := cleanPath("\"quoted path.txt\""); got != "quoted path.txt" {
		t.Errorf("cleanPath failed: got %s", got)
	}
	if got := cleanPath("normal.txt"); got != "normal.txt" {
		t.Errorf("cleanPath failed: got %s", got)
	}
}
