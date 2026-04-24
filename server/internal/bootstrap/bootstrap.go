package bootstrap

import (
	_ "embed"
	"encoding/json"

	"github.com/lucid-fdn/guild/pkg/spec"
	"github.com/lucid-fdn/guild/server/internal/storage"
)

//go:embed fixtures/taskpack.json
var taskpackFixture []byte

//go:embed fixtures/institution.json
var institutionFixture []byte

//go:embed fixtures/dri-binding.json
var driFixture []byte

//go:embed fixtures/artifact.json
var artifactFixture []byte

//go:embed fixtures/promotion-record.json
var promotionFixture []byte

//go:embed fixtures/governance-policy.json
var governancePolicyFixture []byte

//go:embed fixtures/approval-request.json
var approvalRequestFixture []byte

//go:embed fixtures/promotion-gate.json
var promotionGateFixture []byte

//go:embed fixtures/commons-entry.json
var commonsEntryFixture []byte

func SeedIfEmpty(store storage.Store) error {
	var existing []spec.Taskpack
	if err := store.List("taskpacks", &existing); err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	var taskpack spec.Taskpack
	var institution spec.Institution
	var binding spec.DriBinding
	var artifact spec.Artifact
	var promotion spec.PromotionRecord
	var policy spec.GovernancePolicy
	var approval spec.ApprovalRequest
	var gate spec.PromotionGate
	var commons spec.CommonsEntry

	if err := json.Unmarshal(institutionFixture, &institution); err != nil {
		return err
	}
	if err := json.Unmarshal(taskpackFixture, &taskpack); err != nil {
		return err
	}
	if err := json.Unmarshal(driFixture, &binding); err != nil {
		return err
	}
	if err := json.Unmarshal(artifactFixture, &artifact); err != nil {
		return err
	}
	if err := json.Unmarshal(promotionFixture, &promotion); err != nil {
		return err
	}
	if err := json.Unmarshal(governancePolicyFixture, &policy); err != nil {
		return err
	}
	if err := json.Unmarshal(approvalRequestFixture, &approval); err != nil {
		return err
	}
	if err := json.Unmarshal(promotionGateFixture, &gate); err != nil {
		return err
	}
	if err := json.Unmarshal(commonsEntryFixture, &commons); err != nil {
		return err
	}

	if err := store.Put("institutions", institution.InstitutionID, institution); err != nil {
		return err
	}
	if err := store.Put("governance-policies", policy.PolicyID, policy); err != nil {
		return err
	}
	if err := store.Put("taskpacks", taskpack.TaskpackID, taskpack); err != nil {
		return err
	}
	if err := store.Put("dri-bindings", binding.DriBindingID, binding); err != nil {
		return err
	}
	if err := store.Put("artifacts", artifact.ArtifactID, artifact); err != nil {
		return err
	}
	if err := store.Put("promotion-records", promotion.PromotionID, promotion); err != nil {
		return err
	}
	if err := store.Put("approval-requests", approval.ApprovalID, approval); err != nil {
		return err
	}
	if err := store.Put("promotion-gates", gate.GateID, gate); err != nil {
		return err
	}
	if err := store.Put("commons-entries", commons.CommonsEntryID, commons); err != nil {
		return err
	}
	return nil
}
