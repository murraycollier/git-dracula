package git

import (
	"bufio"
	"strconv"
	"strings"
)

// ParsePorcelainV2 parses the raw stdout of `git status --porcelain=v2 --branch --ahead-behind --untracked-files=all`
func ParsePorcelainV2(output string) *StatusResult {
	res := &StatusResult{
		Staged:    make([]FileStatus, 0),
		Unstaged:  make([]FileStatus, 0),
		Untracked: make([]FileStatus, 0),
		Conflicts: make([]FileStatus, 0),
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		// Header lines
		if strings.HasPrefix(line, "# ") {
			parseHeaderLine(line[2:], &res.Branch)
			continue
		}

		// Entries
		switch line[0] {
		case '1': // Ordinary change
			parseEntry1(line, res)
		case '2': // Renamed or copied
			parseEntry2(line, res)
		case 'u': // Unmerged / conflict
			parseEntryConflict(line, res)
		case '?': // Untracked
			parseEntryUntracked(line, res)
		}
	}

	return res
}

func parseHeaderLine(line string, branch *BranchInfo) {
	parts := strings.SplitN(line, " ", 2)
	if len(parts) < 2 {
		return
	}
	key := parts[0]
	val := parts[1]

	switch key {
	case "branch.oid":
		branch.OID = val
		if val == "(initial)" {
			branch.Initial = true
		}
	case "branch.head":
		branch.Head = val
		if val == "(detached)" {
			branch.Detached = true
		}
	case "branch.upstream":
		branch.Upstream = val
	case "branch.ab":
		// Format: +ahead -behind
		fields := strings.Fields(val)
		for _, f := range fields {
			if strings.HasPrefix(f, "+") {
				if n, err := strconv.Atoi(f[1:]); err == nil {
					branch.Ahead = n
				}
			} else if strings.HasPrefix(f, "-") {
				if n, err := strconv.Atoi(f[1:]); err == nil {
					branch.Behind = n
				}
			}
		}
	}
}

// Format: 1 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <path>
func parseEntry1(line string, res *StatusResult) {
	fields := strings.Split(line, " ")
	if len(fields) < 9 {
		return
	}

	xy := fields[1]
	if len(xy) < 2 {
		return
	}

	// Path starts at field index 8 and can contain spaces if unquoted
	path := cleanPath(strings.Join(fields[8:], " "))

	x := xy[0] // Staged
	y := xy[1] // Unstaged

	// Staged change
	if x != '.' {
		res.Staged = append(res.Staged, FileStatus{
			Path:        path,
			Section:     SectionStaged,
			StatusCode:  string(x),
			StatusLabel: statusLabel(x),
		})
	}

	// Unstaged change
	if y != '.' {
		res.Unstaged = append(res.Unstaged, FileStatus{
			Path:        path,
			Section:     SectionUnstaged,
			StatusCode:  string(y),
			StatusLabel: statusLabel(y),
		})
	}
}

// Format: 2 <XY> <sub> <mH> <mI> <mW> <hH> <hI> <Xscore> <path><TAB><origPath>
func parseEntry2(line string, res *StatusResult) {
	// First split tab to isolate origPath
	tabParts := strings.SplitN(line, "\t", 2)
	var origPath string
	mainPart := tabParts[0]
	if len(tabParts) > 1 {
		origPath = cleanPath(tabParts[1])
	}

	fields := strings.Split(mainPart, " ")
	if len(fields) < 10 {
		return
	}

	xy := fields[1]
	if len(xy) < 2 {
		return
	}

	path := cleanPath(strings.Join(fields[9:], " "))
	x := xy[0]
	y := xy[1]

	if x != '.' {
		res.Staged = append(res.Staged, FileStatus{
			Path:        path,
			OrigPath:    origPath,
			Section:     SectionStaged,
			StatusCode:  string(x),
			StatusLabel: statusLabel(x),
		})
	}

	if y != '.' {
		res.Unstaged = append(res.Unstaged, FileStatus{
			Path:        path,
			OrigPath:    origPath,
			Section:     SectionUnstaged,
			StatusCode:  string(y),
			StatusLabel: statusLabel(y),
		})
	}
}

// Format: u <XY> <sub> <m1> <m2> <m3> <mW> <h1> <h2> <h3> <path>
func parseEntryConflict(line string, res *StatusResult) {
	fields := strings.Split(line, " ")
	if len(fields) < 11 {
		return
	}

	xy := fields[1]
	path := cleanPath(strings.Join(fields[10:], " "))

	res.Conflicts = append(res.Conflicts, FileStatus{
		Path:        path,
		Section:     SectionConflict,
		StatusCode:  xy,
		StatusLabel: "conflict",
	})
}

// Format: ? <path>
func parseEntryUntracked(line string, res *StatusResult) {
	if len(line) < 3 {
		return
	}
	path := cleanPath(strings.TrimSpace(line[2:]))
	res.Untracked = append(res.Untracked, FileStatus{
		Path:        path,
		Section:     SectionUntracked,
		StatusCode:  "?",
		StatusLabel: "untracked",
	})
}

func statusLabel(code byte) string {
	switch code {
	case 'M':
		return "modified"
	case 'A':
		return "added"
	case 'D':
		return "deleted"
	case 'R':
		return "renamed"
	case 'C':
		return "copied"
	case 'T':
		return "type changed"
	case 'U':
		return "unmerged"
	default:
		return "changed"
	}
}

func cleanPath(p string) string {
	p = strings.TrimSpace(p)
	if strings.HasPrefix(p, "\"") && strings.HasSuffix(p, "\"") && len(p) >= 2 {
		p = p[1 : len(p)-1]
		// Handle basic escaped quotes
		p = strings.ReplaceAll(p, "\\\"", "\"")
		p = strings.ReplaceAll(p, "\\\\", "\\")
	}
	return p
}

// DiffStatEntry holds diff numstat numbers for a file
type DiffStatEntry struct {
	Additions int
	Deletions int
	IsBinary  bool
}

// ParseNumstat parses output of `git diff --numstat` or `git diff --cached --numstat`
func ParseNumstat(output string) map[string]DiffStatEntry {
	stats := make(map[string]DiffStatEntry)
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}

		addStr := parts[0]
		delStr := parts[1]
		path := parts[2]

		var entry DiffStatEntry
		if addStr == "-" && delStr == "-" {
			entry.IsBinary = true
		} else {
			entry.Additions, _ = strconv.Atoi(addStr)
			entry.Deletions, _ = strconv.Atoi(delStr)
		}

		// Handle rename format in numstat: "old => new" or "prefix/{old => new}/suffix"
		cleanP := normalizeNumstatPath(path)
		stats[cleanP] = entry
		if cleanP != path {
			stats[path] = entry
		}
	}

	return stats
}

func normalizeNumstatPath(p string) string {
	if strings.Contains(p, " => ") {
		// e.g. "dir/{old.go => new.go}" -> "dir/new.go"
		if strings.Contains(p, "{") && strings.Contains(p, "}") {
			start := strings.Index(p, "{")
			end := strings.Index(p, "}")
			prefix := p[:start]
			suffix := p[end+1:]
			middle := p[start+1 : end]
			arrowParts := strings.Split(middle, " => ")
			if len(arrowParts) == 2 {
				return prefix + strings.TrimSpace(arrowParts[1]) + suffix
			}
		}

		// e.g. "old.go => new.go"
		arrowParts := strings.Split(p, " => ")
		if len(arrowParts) == 2 {
			return strings.TrimSpace(arrowParts[1])
		}
	}
	return p
}
