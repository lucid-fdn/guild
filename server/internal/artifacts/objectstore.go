package artifacts

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/guild-labs/guild/pkg/spec"
)

type LocalObjectStore struct {
	root string
}

func NewLocalObjectStore(root string) (*LocalObjectStore, error) {
	if err := os.MkdirAll(filepath.Join(root, "artifacts"), 0o755); err != nil {
		return nil, err
	}
	return &LocalObjectStore{root: root}, nil
}

func (s *LocalObjectStore) PutArtifactMetadata(artifact spec.Artifact) error {
	dir := filepath.Join(s.root, "artifacts", artifact.ArtifactID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "metadata.json"), append(data, '\n'), 0o644)
}

func (s *LocalObjectStore) Status() map[string]any {
	return map[string]any{
		"driver": "local",
		"root":   s.root,
	}
}
