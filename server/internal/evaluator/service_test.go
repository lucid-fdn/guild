package evaluator

import (
	"testing"

	"github.com/guild-labs/guild/pkg/spec"
	"github.com/guild-labs/guild/server/internal/artifacts"
	"github.com/guild-labs/guild/server/internal/dri"
	"github.com/guild-labs/guild/server/internal/promotions"
	"github.com/guild-labs/guild/server/internal/storage"
	"github.com/guild-labs/guild/server/internal/tasks"
)

func TestEvaluationJobRunsReplaySuiteAndCreatesPromotionCandidate(t *testing.T) {
	store, err := storage.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	taskService := tasks.NewService(store)
	driService := dri.NewService(store)
	artifactService := artifacts.NewService(store, nil)
	promotionService := promotions.NewService(store)
	service := NewService(store, taskService, driService, artifactService, promotionService)

	institution := spec.Institution{
		SchemaVersion: "v1alpha1",
		InstitutionID: "5d3dca03-89a0-4fb0-99ee-5f39ef5a6f0c",
		Name:          "Test institution",
		CreatedAt:     "2026-04-24T10:00:00Z",
	}
	if err := store.Put("institutions", institution.InstitutionID, institution); err != nil {
		t.Fatal(err)
	}
	taskpack := spec.Taskpack{
		SchemaVersion: "v1alpha1",
		TaskpackID:    "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
		InstitutionID: institution.InstitutionID,
		Title:         "Replay task",
		Objective:     "Evaluate a replay suite.",
		TaskType:      "analysis",
		Priority:      "medium",
		RequestedBy: spec.ActorRef{
			ActorID:     "52a203e8-0a20-4f0b-9b3d-78631e00d6ab",
			ActorType:   "human",
			DisplayName: "Operator",
		},
		ContextBudget: spec.ContextBudget{
			MaxInputTokens:  4000,
			MaxOutputTokens: 1500,
			ContextStrategy: "artifact_refs_first",
		},
		Permissions: spec.Permissions{ApprovalMode: "ask"},
		Acceptance: []spec.AcceptanceCriteria{
			{CriterionID: "replay-suite", Description: "Run replay suite.", Required: true},
		},
		CreatedAt: "2026-04-24T10:00:00Z",
	}
	if err := taskService.Create(taskpack); err != nil {
		t.Fatal(err)
	}

	job, err := service.Enqueue(ReplaySuite{
		SuiteID:     "test-suite",
		Name:        "Test suite",
		TaskpackIDs: []string{taskpack.TaskpackID},
		MetricName:  "catch_rate",
		Before:      0.4,
		After:       0.8,
		Direction:   "higher_is_better",
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := service.RunJob(job.EvaluationJobID)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != "succeeded" {
		t.Fatalf("expected succeeded, got %q", result.Status)
	}
	if result.BenchmarkArtifactID == "" || result.CandidateArtifactID == "" || result.PromotionRecordID == "" {
		t.Fatalf("expected result artifact and promotion ids, got %+v", result)
	}

	artifacts, err := artifactService.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(artifacts) != 2 {
		t.Fatalf("expected 2 artifacts, got %d", len(artifacts))
	}
	promotions, err := promotionService.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(promotions) != 1 {
		t.Fatalf("expected 1 promotion record, got %d", len(promotions))
	}
	if promotions[0].Decision != "needs_human_review" {
		t.Fatalf("expected needs_human_review, got %q", promotions[0].Decision)
	}
}
