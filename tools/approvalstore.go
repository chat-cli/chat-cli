package tools

import (
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const approvalStoreFilename = "tool-approvals.yaml"

// approvalStoreFile is the on-disk shape of the persisted "always" tier:
// a map from absolute repository root to a list of "toolName:patternKey"
// entries approved for that repository. Uses ":" (not the in-memory "\x00"
// key) so the file stays human-readable/hand-editable, per
// business-logic-model.md.
type approvalStoreFile struct {
	Repos map[string][]string `yaml:"repos"`
}

// persistedEntry formats toolName+patternKey for on-disk storage.
func persistedEntry(toolName, patternKey string) string {
	return toolName + ":" + patternKey
}

// ApprovalStore tracks granted sticky approvals for destructive tool calls
// at two tiers: session (in-memory, this process only) and always
// (persisted, scoped per git repository - see NewApprovalStore's repoRoot
// parameter and RecordAlways).
type ApprovalStore struct {
	configPath string
	repoRoot   string
	session    map[string]bool
	always     map[string]bool // this repo's always-approvals only, loaded at construction
}

// NewApprovalStore creates an ApprovalStore. configPath is the directory
// chat-cli's other config/data lives in (fm.ConfigPath) - the persisted
// "always" tier is stored there. repoRoot is the current git repository's
// root (from utils.FindGitBoundary), or "" if not inside one - "always"
// approvals are unavailable in that case (RecordAlways is a safe no-op).
//
// A missing or malformed store file degrades to zero always-approvals
// rather than an error - consistent with #88's precedent for tolerating
// unreadable local files.
func NewApprovalStore(configPath, repoRoot string) (*ApprovalStore, error) {
	s := &ApprovalStore{
		configPath: configPath,
		repoRoot:   repoRoot,
		session:    make(map[string]bool),
		always:     make(map[string]bool),
	}

	if repoRoot == "" {
		return s, nil
	}

	data, err := os.ReadFile(s.storePath()) // #nosec G304 - configPath is chat-cli's own config directory
	if err != nil {
		return s, nil // missing file: zero approvals, not an error
	}

	var file approvalStoreFile
	if err := yaml.Unmarshal(data, &file); err != nil {
		return s, nil // malformed file: zero approvals, not an error
	}

	for _, entry := range file.Repos[repoRoot] {
		toolName, patternKey, ok := strings.Cut(entry, ":")
		if !ok {
			continue // malformed entry, skip rather than fail the whole load
		}
		s.always[approvalKey(toolName, patternKey)] = true
	}

	return s, nil
}

func (s *ApprovalStore) storePath() string {
	return filepath.Join(s.configPath, approvalStoreFilename)
}

// approvalKey joins toolName and patternKey with a NUL byte, which can't
// appear in either component, avoiding any delimiter-collision edge case a
// printable separator like ":" could hit if a pattern key ever contained one.
func approvalKey(toolName, patternKey string) string {
	return toolName + "\x00" + patternKey
}

// IsApproved reports whether toolName+patternKey has a matching approval in
// either tier (session checked first, then always).
func (s *ApprovalStore) IsApproved(toolName, patternKey string) bool {
	key := approvalKey(toolName, patternKey)
	return s.session[key] || s.always[key]
}

// CanRecordAlways reports whether this store has a repository root to scope
// "always" approvals to (BR10) - false when chat isn't running inside a git
// repository, in which case "always" should not be offered as a choice.
func (s *ApprovalStore) CanRecordAlways() bool {
	return s.repoRoot != ""
}

// RecordSession grants an in-memory-only approval for the rest of this
// process's lifetime.
func (s *ApprovalStore) RecordSession(toolName, patternKey string) {
	s.session[approvalKey(toolName, patternKey)] = true
}

// RecordAlways grants an approval persisted to disk, scoped to this store's
// repoRoot. A safe no-op (no error, no file written) when repoRoot is empty
// (not inside a git repository) - there's no meaningful root to scope it to.
func (s *ApprovalStore) RecordAlways(toolName, patternKey string) error {
	if s.repoRoot == "" {
		return nil
	}

	s.always[approvalKey(toolName, patternKey)] = true
	entry := persistedEntry(toolName, patternKey)

	file := approvalStoreFile{Repos: make(map[string][]string)}

	if data, err := os.ReadFile(s.storePath()); err == nil { // #nosec G304 - configPath is chat-cli's own config directory
		_ = yaml.Unmarshal(data, &file) // best-effort merge; a malformed existing file is overwritten below
		if file.Repos == nil {
			file.Repos = make(map[string][]string)
		}
	}

	entries := file.Repos[s.repoRoot]
	found := false
	for _, e := range entries {
		if e == entry {
			found = true
			break
		}
	}
	if !found {
		entries = append(entries, entry)
	}
	file.Repos[s.repoRoot] = entries

	out, err := yaml.Marshal(file)
	if err != nil {
		return err
	}

	return os.WriteFile(s.storePath(), out, 0600)
}
