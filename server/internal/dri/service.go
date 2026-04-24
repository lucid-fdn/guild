package dri

import (
	"errors"
	"fmt"

	"github.com/lucid-fdn/guild/pkg/spec"
	specvalidate "github.com/lucid-fdn/guild/pkg/spec/validate"
	"github.com/lucid-fdn/guild/server/internal/storage"
)

type Service struct {
	store storage.Store
}

func NewService(store storage.Store) *Service {
	return &Service{store: store}
}

func (s *Service) Create(binding spec.DriBinding) error {
	if err := specvalidate.DriBinding(binding); err != nil {
		return err
	}
	var taskpack spec.Taskpack
	if err := s.store.Get("taskpacks", binding.TaskpackID, &taskpack); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("taskpack_id %q does not exist", binding.TaskpackID)
		}
		return err
	}
	return s.store.Put("dri-bindings", binding.DriBindingID, binding)
}

func (s *Service) Get(id string) (spec.DriBinding, error) {
	var binding spec.DriBinding
	err := s.store.Get("dri-bindings", id, &binding)
	return binding, err
}

func (s *Service) List() ([]spec.DriBinding, error) {
	var bindings []spec.DriBinding
	err := s.store.List("dri-bindings", &bindings)
	return bindings, err
}

func (s *Service) Status() map[string]any {
	bindings, _ := s.List()
	return map[string]any{
		"service": "dri",
		"state":   "ready",
		"count":   len(bindings),
	}
}
