package tasks

import (
	"errors"
	"fmt"

	"github.com/guild-labs/guild/pkg/spec"
	specvalidate "github.com/guild-labs/guild/pkg/spec/validate"
	"github.com/guild-labs/guild/server/internal/storage"
)

type Service struct {
	store storage.Store
}

func NewService(store storage.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Create(taskpack spec.Taskpack) error {
	if err := specvalidate.Taskpack(taskpack); err != nil {
		return err
	}
	if taskpack.InstitutionID != "" {
		var institution spec.Institution
		if err := s.store.Get("institutions", taskpack.InstitutionID, &institution); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return fmt.Errorf("institution_id %q does not exist", taskpack.InstitutionID)
			}
			return err
		}
	}
	if taskpack.ParentTaskpackID != "" {
		var parent spec.Taskpack
		if err := s.store.Get("taskpacks", taskpack.ParentTaskpackID, &parent); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return fmt.Errorf("parent_taskpack_id %q does not exist", taskpack.ParentTaskpackID)
			}
			return err
		}
	}
	for _, input := range taskpack.Inputs {
		var artifact spec.Artifact
		if err := s.store.Get("artifacts", input.ArtifactID, &artifact); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return fmt.Errorf("input artifact_id %q does not exist", input.ArtifactID)
			}
			return err
		}
	}
	return s.store.Put("taskpacks", taskpack.TaskpackID, taskpack)
}

func (s *Service) Get(id string) (spec.Taskpack, error) {
	var taskpack spec.Taskpack
	err := s.store.Get("taskpacks", id, &taskpack)
	return taskpack, err
}

func (s *Service) List() ([]spec.Taskpack, error) {
	var taskpacks []spec.Taskpack
	err := s.store.List("taskpacks", &taskpacks)
	return taskpacks, err
}

func (s *Service) Status() map[string]any {
	taskpacks, _ := s.List()
	return map[string]any{
		"service": "tasks",
		"state":   "ready",
		"count":   len(taskpacks),
	}
}
