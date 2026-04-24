package artifacts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucid-fdn/guild/pkg/spec"
	"github.com/lucid-fdn/guild/server/internal/storage"
)

func TestCreateRejectsMissingTaskpack(t *testing.T) {
	store, err := storage.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(store, nil)

	err = service.Create(spec.Artifact{
		SchemaVersion: "v1alpha1",
		ArtifactID:    "5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
		TaskpackID:    "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
		Kind:          "review",
		Title:         "Webhook retry audit findings",
		Producer: spec.ActorRef{
			ActorID:     "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
			ActorType:   "agent",
			DisplayName: "payments-dri",
		},
		Storage: spec.ArtifactStorage{
			URI:      "s3://guild/artifacts/findings.md",
			MimeType: "text/markdown",
		},
		Version:   1,
		CreatedAt: "2026-04-24T10:15:00Z",
	})

	if err == nil {
		t.Fatal("expected missing taskpack error")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected referential integrity error, got %q", err.Error())
	}
}

func TestCreateMirrorsArtifactMetadataToObjectStore(t *testing.T) {
	store, err := storage.NewFileStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	taskpack := spec.Taskpack{
		SchemaVersion: "v1alpha1",
		TaskpackID:    "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
		Title:         "Artifact object mirror",
		Objective:     "Persist artifact metadata.",
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
			{CriterionID: "object-mirror", Description: "Mirror artifact metadata.", Required: true},
		},
		CreatedAt: "2026-04-24T10:00:00Z",
	}
	if err := store.Put("taskpacks", taskpack.TaskpackID, taskpack); err != nil {
		t.Fatal(err)
	}
	objectDir := t.TempDir()
	objectStore, err := NewLocalObjectStore(objectDir)
	if err != nil {
		t.Fatal(err)
	}
	service := NewService(store, objectStore)

	artifact := spec.Artifact{
		SchemaVersion: "v1alpha1",
		ArtifactID:    "5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
		TaskpackID:    taskpack.TaskpackID,
		Kind:          "benchmark_result",
		Title:         "Benchmark result",
		Producer: spec.ActorRef{
			ActorID:     "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
			ActorType:   "agent",
			DisplayName: "evaluator",
		},
		Storage: spec.ArtifactStorage{
			URI:      "guild://benchmarks/result.json",
			MimeType: "application/json",
		},
		Version:   1,
		CreatedAt: "2026-04-24T10:15:00Z",
	}
	if err := service.Create(artifact); err != nil {
		t.Fatal(err)
	}

	metadataPath := filepath.Join(objectDir, "artifacts", artifact.ArtifactID, "metadata.json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"kind": "benchmark_result"`) {
		t.Fatalf("expected mirrored metadata, got %s", string(data))
	}
}
