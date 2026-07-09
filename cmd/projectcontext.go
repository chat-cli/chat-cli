/*
Copyright © 2024 Micah Walter
*/
package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/chat-cli/chat-cli/utils"
)

// maxContextFileSize is the size (in bytes) at which project-context file
// content is truncated before being used as the system prompt (BR9).
const maxContextFileSize = 32 * 1024

// defaultContextFilenames is the default precedence list (BR5) used when the
// context-files config key is unset.
var defaultContextFilenames = []string{"AGENTS.md", "CLAUDE.md", ".github/copilot-instructions.md"}

// resolveContextFilenames implements BR11-BR13: parses a comma-separated
// context-files config value into an ordered candidate list, trimming
// whitespace and dropping empty entries. When configSet is false (key unset),
// the default list is used. When configSet is true and configValue is empty
// or trims down to zero entries, an empty list is returned (BR12 disable).
func resolveContextFilenames(configValue string, configSet bool) []string {
	if !configSet {
		return defaultContextFilenames
	}
	if configValue == "" {
		return []string{}
	}

	parts := strings.Split(configValue, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// findProjectContextFile implements the Phase A / Phase B algorithm from
// business-logic-model.md: it walks up from cwd (stat-only, capped at
// maxWalkUpLevels) looking for a .git boundary, then checks candidates (in
// precedence order, BR3) at cwd first and, if no match there, at the
// boundary directory (BR4). It returns the matched path and which candidate
// string matched (so callers can exclude it and keep searching, per BR8).
func findProjectContextFile(cwd string, candidates []string) (path string, matchedCandidate string, ok bool) {
	if len(candidates) == 0 {
		return "", "", false
	}

	if path, matched, found := matchCandidateInDir(cwd, candidates); found {
		return path, matched, true
	}

	boundary := utils.FindGitBoundary(cwd)
	if boundary == "" || boundary == cwd {
		return "", "", false
	}

	if path, matched, found := matchCandidateInDir(boundary, candidates); found {
		return path, matched, true
	}

	return "", "", false
}

// matchCandidateInDir checks candidates (in order) against dir, returning
// the first one that resolves to a regular, readable file (BR2/BR3).
func matchCandidateInDir(dir string, candidates []string) (path string, matchedCandidate string, ok bool) {
	for _, candidate := range candidates {
		candidatePath := filepath.Join(dir, candidate)
		if !isContextFileMatch(candidatePath) {
			continue
		}
		return candidatePath, candidate, true
	}
	return "", "", false
}

// isContextFileMatch reports whether path is a regular file or a symlink to
// one (BR2). Directories and other non-file entries are rejected.
func isContextFileMatch(path string) bool {
	f, err := os.Open(path) // nolint:gosec // path is from a fixed candidate list, not user-supplied
	if err != nil {
		return false
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode().IsRegular()
}

// formatProjectContextDisplayPath returns a cwd-relative path for the notice
// (BR16), falling back to the basename when relative resolution fails.
func formatProjectContextDisplayPath(cwd, sourcePath string) string {
	if rel, err := filepath.Rel(cwd, sourcePath); err == nil && rel != "" && rel != "." {
		return rel
	}
	return filepath.Base(sourcePath)
}

// loadProjectContext implements BR7-BR10: reads path, trims surrounding
// whitespace, and truncates to maxContextFileSize if needed. A read error
// (permission denied, race-condition deletion, path is a directory, etc.)
// is returned as err for the caller to treat as "no match" (BR10) - this
// function does not decide search-continuation policy.
func loadProjectContext(path string) (content string, truncated bool, originalSize int, err error) {
	data, err := os.ReadFile(path) // nolint:gosec // path is resolved from a fixed, known candidate list, not user-supplied
	if err != nil {
		return "", false, 0, err
	}

	originalSize = len(data)
	trimmed := strings.TrimSpace(string(data))

	if len(trimmed) > maxContextFileSize {
		return trimmed[:maxContextFileSize], true, originalSize, nil
	}

	return trimmed, false, originalSize, nil
}

// resolveAndLoadProjectContext is the composition root tying discovery and
// loading together, including BR8's rule that an empty-after-trim match (or
// an unreadable one, BR10) is treated as no match and search continues with
// the next candidate.
func resolveAndLoadProjectContext(cwd string, candidates []string) (content string, sourcePath string, truncated bool, found bool) {
	remaining := append([]string(nil), candidates...)

	for len(remaining) > 0 {
		path, matched, ok := findProjectContextFile(cwd, remaining)
		if !ok {
			return "", "", false, false
		}

		loaded, trunc, _, err := loadProjectContext(path)
		if err == nil && loaded != "" {
			return loaded, path, trunc, true
		}

		remaining = removeString(remaining, matched)
	}

	return "", "", false, false
}

func removeString(list []string, target string) []string {
	result := make([]string, 0, len(list))
	for _, s := range list {
		if s != target {
			result = append(result, s)
		}
	}
	return result
}
