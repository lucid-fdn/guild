package promotions

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

func (s *Service) Create(record spec.PromotionRecord) error {
	if err := specvalidate.PromotionRecord(record); err != nil {
		return err
	}
	var institution spec.Institution
	if err := s.store.Get("institutions", record.InstitutionID, &institution); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("institution_id %q does not exist", record.InstitutionID)
		}
		return err
	}
	var candidate spec.Artifact
	if err := s.store.Get("artifacts", record.CandidateRef.ArtifactID, &candidate); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("candidate_ref.artifact_id %q does not exist", record.CandidateRef.ArtifactID)
		}
		return err
	}
	return s.store.Put("promotion-records", record.PromotionID, record)
}

func (s *Service) Get(id string) (spec.PromotionRecord, error) {
	var record spec.PromotionRecord
	err := s.store.Get("promotion-records", id, &record)
	return record, err
}

func (s *Service) List() ([]spec.PromotionRecord, error) {
	var records []spec.PromotionRecord
	err := s.store.List("promotion-records", &records)
	return records, err
}

func (s *Service) Status() map[string]any {
	records, _ := s.List()
	return map[string]any{
		"service": "promotions",
		"state":   "ready",
		"count":   len(records),
	}
}
