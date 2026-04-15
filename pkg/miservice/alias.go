package miservice

import (
	"os"
	"path/filepath"
	"strings"
)

// AliasStore manages custom device name aliases.
type AliasStore struct {
	filePath string
	Aliases  map[string]string `json:"aliases"`
}

// NewAliasStore creates a new AliasStore with the default path ~/.micli/aliases.json.
func NewAliasStore() *AliasStore {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return &AliasStore{
		filePath: filepath.Join(homeDir, ".micli", "aliases.json"),
		Aliases:  make(map[string]string),
	}
}

// Load reads aliases from the JSON file.
func (s *AliasStore) Load() error {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &s.Aliases)
}

// Save writes aliases to the JSON file.
func (s *AliasStore) Save() error {
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s.Aliases, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

// Add adds or updates an alias mapping.
func (s *AliasStore) Add(alias, did string) {
	s.Aliases[alias] = did
}

// Remove removes an alias mapping.
func (s *AliasStore) Remove(alias string) {
	delete(s.Aliases, alias)
}

// Resolve resolves an input string to a device DID.
// Priority: exact DID match (if digit) > exact alias match > fuzzy name match.
func (s *AliasStore) Resolve(input string, devices []*DeviceInfo) (string, error) {
	if input == "" {
		return "", ErrDeviceNotFound
	}

	// If input is all digits, treat as DID directly
	if isAllDigits(input) {
		for _, d := range devices {
			if d.Did == input {
				return input, nil
			}
		}
		return "", ErrDeviceNotFound
	}

	// Check alias map for exact match
	if did, ok := s.Aliases[input]; ok {
		for _, d := range devices {
			if d.Did == did {
				return did, nil
			}
		}
		// Alias points to a DID not in current device list, return it anyway
		return did, nil
	}

	// Fuzzy match: strings.Contains on device name
	var matched []*DeviceInfo
	for _, d := range devices {
		if strings.Contains(d.Name, input) {
			matched = append(matched, d)
		}
	}
	if len(matched) == 1 {
		return matched[0].Did, nil
	}
	if len(matched) > 1 {
		names := make([]string, len(matched))
		for i, d := range matched {
			names[i] = d.Name
		}
		return "", &ErrAmbiguousDevice{Input: input, Matches: names}
	}

	return "", ErrDeviceNotFound
}

func isAllDigits(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// ErrAmbiguousDevice is returned when a partial name matches multiple devices.
type ErrAmbiguousDevice struct {
	Input   string
	Matches []string
}

func (e *ErrAmbiguousDevice) Error() string {
	return "multiple devices match '" + e.Input + "': " + strings.Join(e.Matches, ", ")
}
