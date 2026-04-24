package promotions

import (
	"strings"
	"testing"

	"github.com/guild-labs/guild/pkg/spec"
	"github.com/guild-labs/guild/server/internal/storage"
)

func TestCreateRejectsMissingInstitution(t *testing.T) {
	store, err := storage.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(store)

	err = service.Create(validPromotionRecord())

	if err == nil {
		t.Fatal("expected missing institution error")
	}
	if !strings.Contains(err.Error(), "institution_id") {
		t.Fatalf("expected institution integrity error, got %q", err.Error())
	}
}

func TestCreateRejectsMissingCandidateArtifact(t *testing.T) {
	store, err := storage.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Put("institutions", "5d3dca03-89a0-4fb0-99ee-5f39ef5a6f0c", spec.Institution{
		SchemaVersion: "v1alpha1",
		InstitutionID: "5d3dca03-89a0-4fb0-99ee-5f39ef5a6f0c",
		Name:          "Guild Bootstrap Institution",
		CreatedAt:     "2026-04-24T09:55:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	service := NewService(store)

	err = service.Create(validPromotionRecord())

	if err == nil {
		t.Fatal("expected missing candidate artifact error")
	}
	if !strings.Contains(err.Error(), "candidate_ref.artifact_id") {
		t.Fatalf("expected candidate artifact integrity error, got %q", err.Error())
	}
}

func validPromotionRecord() spec.PromotionRecord {
	return spec.PromotionRecord{
		SchemaVersion: "v1alpha1",
		PromotionID:   "b2ddb0dd-b29c-4a28-b1ba-e9a2f8ff23fb",
		InstitutionID: "5d3dca03-89a0-4fb0-99ee-5f39ef5a6f0c",
		CandidateKind: "review_heuristic",
		CandidateRef: spec.ArtifactRef{
			ArtifactID: "5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
			Kind:       "review",
			URI:        "s3://guild/artifacts/findings.md",
			Version:    1,
		},
		Decision:  "accepted",
		DecidedAt: "2026-04-24T11:00:00Z",
	}
}
