package artifacts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/lucid-fdn/guild/pkg/spec"
	specvalidate "github.com/lucid-fdn/guild/pkg/spec/validate"
	"github.com/lucid-fdn/guild/server/internal/storage"
)

type Service struct {
	store       storage.Store
	objectStore ObjectStore
}

type ObjectStore interface {
	PutArtifactMetadata(spec.Artifact) error
	Status() map[string]any
}

func NewService(store storage.Store, objectStore ObjectStore) *Service {
	return &Service{store: store, objectStore: objectStore}
}

func (s *Service) Create(artifact spec.Artifact) error {
	if err := specvalidate.Artifact(artifact); err != nil {
		return err
	}
	var taskpack spec.Taskpack
	if err := s.store.Get("taskpacks", artifact.TaskpackID, &taskpack); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("taskpack_id %q does not exist", artifact.TaskpackID)
		}
		return err
	}
	for _, parentID := range artifact.ParentArtifactIDs {
		var parent spec.Artifact
		if err := s.store.Get("artifacts", parentID, &parent); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return fmt.Errorf("parent_artifact_id %q does not exist", parentID)
			}
			return err
		}
	}
	if artifact.Storage.SHA256 == "" || artifact.Storage.SizeBytes == 0 {
		artifact = withArtifactMetadataDigest(artifact)
	}
	if s.objectStore != nil {
		if err := s.objectStore.PutArtifactMetadata(artifact); err != nil {
			return err
		}
	}
	return s.store.Put("artifacts", artifact.ArtifactID, artifact)
}

func (s *Service) Get(id string) (spec.Artifact, error) {
	var artifact spec.Artifact
	err := s.store.Get("artifacts", id, &artifact)
	return artifact, err
}

func (s *Service) List() ([]spec.Artifact, error) {
	var artifacts []spec.Artifact
	err := s.store.List("artifacts", &artifacts)
	return artifacts, err
}

func (s *Service) ListByTaskpack(taskpackID string) ([]spec.Artifact, error) {
	artifacts, err := s.List()
	if err != nil {
		return nil, err
	}
	filtered := make([]spec.Artifact, 0)
	for _, artifact := range artifacts {
		if artifact.TaskpackID == taskpackID {
			filtered = append(filtered, artifact)
		}
	}
	return filtered, nil
}

func (s *Service) Status() map[string]any {
	artifacts, _ := s.List()
	return map[string]any{
		"service":      "artifacts",
		"state":        "ready",
		"count":        len(artifacts),
		"object_store": objectStoreStatus(s.objectStore),
	}
}

func withArtifactMetadataDigest(artifact spec.Artifact) spec.Artifact {
	data, err := json.Marshal(artifact)
	if err != nil {
		return artifact
	}
	sum := sha256.Sum256(data)
	if artifact.Storage.SHA256 == "" {
		artifact.Storage.SHA256 = hex.EncodeToString(sum[:])
	}
	if artifact.Storage.SizeBytes == 0 {
		artifact.Storage.SizeBytes = int64(len(data))
	}
	return artifact
}

func objectStoreStatus(store ObjectStore) map[string]any {
	if store == nil {
		return map[string]any{"driver": "disabled"}
	}
	return store.Status()
}
