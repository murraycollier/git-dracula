# 🧛 git-dracula

A gorgeous, **Dracula-themed** Git status and interactive staging tool built with the **Charm Go suite** ([Lipgloss](https://github.com/charmbracelet/lipgloss), [Bubble Tea](https://github.com/charmbracelet/bubbletea), and [Bubbles](https://github.com/charmbracelet/bubbles)).

Designed as a drop-in, beautiful replacement for `git status` that brings rich visual hierarchy, live diff previews, and interactive staging to your terminal.

---

```
╭────────────────────────────────────────────────────────────────────────────╮
│  󰊢 git-dracula    feature/tui  ✓ origin/feature/tui          📦 1 stashed  │
│ ◈ 8ba12fd feat(ui): dynamic truncation for header card · Murray (2m ago)   │
╰────────────────────────────────────────────────────────────────────────────╯

   ● 2 Staged    ✎ 1 Modified    ? 4 Untracked 

  ✔ STAGED CHANGES  (2 to be committed)
    + ADD   cmd/git-dracula/main.go                     +125    -0  ■■■■■■■■
    ~ MOD   internal/git/parser.go                       +24    -3  ■■■■■■■░

  ✎ UNSTAGED CHANGES  (1 not staged)
    ~ MOD   internal/ui/view.go                          +13    -7  ■■■■■■░░

  ? UNTRACKED FILES  (4 files)
    ? NEW   README.md
    ? NEW   Makefile
    ? NEW   LICENSE
    ? NEW   .gitignore

  💡 Tip: Run  git-dracula -i  for interactive staging & diffs ·  -s  short ·  -d  diffs
```

When your working tree is clean:

```
╭──────────────────────────────────────────────────────────────────────────╮
│                                                                          │
│             🦇  Working tree is spotless. Dracula approves!              │
│                Nothing to commit, working directory is clean.            │
│                                                                          │
╰──────────────────────────────────────────────────────────────────────────╯
```

---

## ✨ Features

- **Dracula Palette**: Faithfully designed using official [Dracula Theme](https://draculatheme.com) hex specifications:
  - 🟣 **Purple** (`#BD93F9`) & 🌸 **Pink** (`#FF79C6`) — Repository badges, commit hashes, untracked files
  - 🟢 **Green** (`#50FA7B`) — Staged changes, additions, clean state
  - 🟡 **Yellow** (`#F1FA8C`) & 🟠 **Orange** (`#FFB86C`) — Unstaged modifications, stashes
  - 🔵 **Cyan** (`#8BE9FD`) — Current branch, diff hunks, renames
  - 🔴 **Red** (`#FF5555`) — Deletions, merge conflicts, repo alerts
  - 🩶 **Comment / Selection** (`#6272A4` / `#44475A`) — Muted metadata, paths, cursor highlight
- **Header Overview Card**:
  - Current repository name and branch (with detached HEAD & initial branch indicators)
  - Remote tracking status (`✓ synced`, `⇡ ahead`, `⇣ behind`, `⇡⇣ diverged`)
  - Stash count badge (`📦 N stashed`)
  - Repository state warnings (`⚠ MERGING`, `⚠ REBASING`, `⚠ CHERRY-PICKING`, `⚠ BISECTING`)
  - Latest commit hash, subject, author, and relative time
- **Status Summary Pills**: Glanceable counter pills for staged, modified, untracked, and conflict files.
- **Visual Diffstat Histogram Bars**: Compact graphical bar (`■■■■░░`) showing additions vs deletions per file.
- **Interactive Staging & Diff Mode (`-i`)**:
  - Live split-view TUI powered by **Bubble Tea**
  - Instant syntax-highlighted Dracula diff viewport on the right
  - Stage / unstage files with <kbd>Space</kbd> or <kbd>s</kbd>
  - Stage all (<kbd>a</kbd>) / Unstage all (<kbd>u</kbd>)
  - Commit prompt modal (<kbd>c</kbd>)
  - Stash modal (<kbd>S</kbd>)
  - Scrollable diff preview with <kbd>Tab</kbd> or <kbd>d</kbd>
- **Native Git Subcommand**: Automatically recognized by git as `git dracula` when installed in `PATH`.
- **Fast & Lightweight**: Concurrent execution of numstats, diffs, and log queries completes in < 15ms.

---

## 🚀 Installation

### Option 1: Using Make (Recommended)

Installs directly to `~/.local/bin` (or custom `PREFIX`):

```bash
git clone https://github.com/murraycollier/git-dracula.git
cd git-dracula
make install
```

### Option 2: Go Install

```bash
go install github.com/murraycollier/git-dracula/cmd/git-dracula@latest
```

Ensure `$(go env GOPATH)/bin` or `~/.local/bin` is in your `$PATH`.

---

## 🎮 Usage

### Standard Git Status

Run `git-dracula` directly or as a Git subcommand:

```bash
git-dracula
# or
git dracula
```

### Interactive Staging TUI (`-i`)

Launch the interactive split-pane browser:

```bash
git-dracula -i
# or
git dracula -i
# or
git dracula interactive
```

### Compact Status (`-s`)

Produces a concise one-line-per-file representation:

```bash
git dracula -s
```

### Inline Diffs (`-d`)

Shows syntax-highlighted Dracula diffs directly below each changed file:

```bash
git dracula -d
```

### Target Another Directory (`-C`)

```bash
git dracula -C /path/to/repo
```

---

## ⌨️ Interactive Keybindings

| Key | Action |
|-----|--------|
| <kbd>j</kbd> / <kbd>↓</kbd> | Move cursor down |
| <kbd>k</kbd> / <kbd>↑</kbd> | Move cursor up |
| <kbd>Space</kbd> / <kbd>s</kbd> | Toggle Stage / Unstage highlighted file |
| <kbd>a</kbd> | Stage all changes (`git add -A`) |
| <kbd>u</kbd> | Unstage all changes (`git restore --staged .`) |
| <kbd>Tab</kbd> / <kbd>d</kbd> | Switch focus between Files and Diff Viewport |
| <kbd>PgUp</kbd> / <kbd>PgDn</kbd> | Scroll diff viewport |
| <kbd>c</kbd> | Open Commit modal (prompts for commit message) |
| <kbd>S</kbd> | Open Stash modal |
| <kbd>r</kbd> | Refresh repository status |
| <kbd>?</kbd> | Toggle help cheatsheet |
| <kbd>q</kbd> / <kbd>Esc</kbd> / <kbd>Ctrl+C</kbd> | Quit interactive mode |

---

## 🔧 Git Aliases Configuration

You can configure Git to use `git-dracula` for your daily status commands:

```bash
# Make 'git st' use git-dracula
git config --global alias.st dracula

# Make 'git sti' launch interactive staging
git config --global alias.sti "dracula -i"

# Replace standard 'git status' (optional alias)
git config --global alias.status dracula
```

---

## 🛠️ Built With Charm

- [Lipgloss](https://github.com/charmbracelet/lipgloss) — Style definitions, Dracula color palettes, layout boxes, and tables
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — The Elm Architecture for interactive terminal apps
- [Bubbles](https://github.com/charmbracelet/bubbles) — Viewport scrolling and interactive commit message text input

---

## 📄 License

[MIT License](LICENSE) © 2026 Murray Collier
