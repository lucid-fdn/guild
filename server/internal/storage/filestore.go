package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
)

var ErrNotFound = errors.New("not found")

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type FileStore struct {
	root string
	mu   sync.RWMutex
}

func NewFileStore(root string) (*FileStore, error) {
	for _, dir := range []string{"institutions", "taskpacks", "dri-bindings", "artifacts", "promotion-records", "evaluation-jobs", "governance-policies", "approval-requests", "promotion-gates", "commons-entries"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			return nil, err
		}
	}
	return &FileStore{root: root}, nil
}

func (s *FileStore) Put(collection, id string, value any) error {
	if err := validateKey(collection, id); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	path := filepath.Join(s.root, collection, id+".json")
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func (s *FileStore) Get(collection, id string, dest any) error {
	if err := validateKey(collection, id); err != nil {
		return err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := filepath.Join(s.root, collection, id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}
	return json.Unmarshal(data, dest)
}

func (s *FileStore) List(collection string, dest any) error {
	if !isAllowedCollection(collection) {
		return fmt.Errorf("unknown collection %q", collection)
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Join(s.root, collection)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	items := make([]json.RawMessage, 0, len(names))
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		items = append(items, data)
	}

	payload, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, dest)
}

func validateKey(collection, id string) error {
	if !isAllowedCollection(collection) {
		return fmt.Errorf("unknown collection %q", collection)
	}
	if !uuidPattern.MatchString(id) {
		return fmt.Errorf("id %q must be a UUID", id)
	}
	return nil
}

func isAllowedCollection(collection string) bool {
	switch collection {
	case "institutions", "taskpacks", "dri-bindings", "artifacts", "promotion-records", "evaluation-jobs", "governance-policies", "approval-requests", "promotion-gates", "commons-entries":
		return true
	default:
		return false
	}
}
