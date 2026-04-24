package institution

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

func (s *Service) CreatePolicy(policy spec.GovernancePolicy) error {
	if err := specvalidate.GovernancePolicy(policy); err != nil {
		return err
	}
	if err := s.requireInstitution(policy.InstitutionID); err != nil {
		return err
	}
	return s.store.Put("governance-policies", policy.PolicyID, policy)
}

func (s *Service) ListPolicies() ([]spec.GovernancePolicy, error) {
	var items []spec.GovernancePolicy
	err := s.store.List("governance-policies", &items)
	return items, err
}

func (s *Service) CreateApproval(request spec.ApprovalRequest) error {
	if err := specvalidate.ApprovalRequest(request); err != nil {
		return err
	}
	var taskpack spec.Taskpack
	if err := s.store.Get("taskpacks", request.TaskpackID, &taskpack); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("taskpack_id %q does not exist", request.TaskpackID)
		}
		return err
	}
	if request.PolicyID != "" {
		var policy spec.GovernancePolicy
		if err := s.store.Get("governance-policies", request.PolicyID, &policy); err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return fmt.Errorf("policy_id %q does not exist", request.PolicyID)
			}
			return err
		}
	}
	return s.store.Put("approval-requests", request.ApprovalID, request)
}

func (s *Service) ListApprovals() ([]spec.ApprovalRequest, error) {
	var items []spec.ApprovalRequest
	err := s.store.List("approval-requests", &items)
	return items, err
}

func (s *Service) CreateGate(gate spec.PromotionGate) error {
	if err := specvalidate.PromotionGate(gate); err != nil {
		return err
	}
	if err := s.requireInstitution(gate.InstitutionID); err != nil {
		return err
	}
	return s.store.Put("promotion-gates", gate.GateID, gate)
}

func (s *Service) ListGates() ([]spec.PromotionGate, error) {
	var items []spec.PromotionGate
	err := s.store.List("promotion-gates", &items)
	return items, err
}

func (s *Service) CreateCommonsEntry(entry spec.CommonsEntry) error {
	if err := specvalidate.CommonsEntry(entry); err != nil {
		return err
	}
	if err := s.requireInstitution(entry.InstitutionID); err != nil {
		return err
	}
	var promotion spec.PromotionRecord
	if err := s.store.Get("promotion-records", entry.PromotionRecordID, &promotion); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("promotion_record_id %q does not exist", entry.PromotionRecordID)
		}
		return err
	}
	var artifact spec.Artifact
	if err := s.store.Get("artifacts", entry.ArtifactRef.ArtifactID, &artifact); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("artifact_ref.artifact_id %q does not exist", entry.ArtifactRef.ArtifactID)
		}
		return err
	}
	return s.store.Put("commons-entries", entry.CommonsEntryID, entry)
}

func (s *Service) ListCommonsEntries() ([]spec.CommonsEntry, error) {
	var items []spec.CommonsEntry
	err := s.store.List("commons-entries", &items)
	return items, err
}

func (s *Service) Status() map[string]any {
	commons, _ := s.ListCommonsEntries()
	approvals, _ := s.ListApprovals()
	return map[string]any{
		"service":   "institution",
		"state":     "ready",
		"commons":   len(commons),
		"approvals": len(approvals),
	}
}

func (s *Service) requireInstitution(id string) error {
	var institution spec.Institution
	if err := s.store.Get("institutions", id, &institution); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return fmt.Errorf("institution_id %q does not exist", id)
		}
		return err
	}
	return nil
}
