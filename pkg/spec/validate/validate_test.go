package validate

import (
	"strings"
	"testing"

	"github.com/guild-labs/guild/pkg/spec"
)

func TestTaskpackRejectsSpecViolations(t *testing.T) {
	taskpack := validTaskpack()

	cases := []struct {
		name    string
		mutate  func(*spec.Taskpack)
		wantErr string
	}{
		{
			name:    "schema version",
			mutate:  func(taskpack *spec.Taskpack) { taskpack.SchemaVersion = "v9" },
			wantErr: "schema_version",
		},
		{
			name:    "task type enum",
			mutate:  func(taskpack *spec.Taskpack) { taskpack.TaskType = "meeting" },
			wantErr: "task_type",
		},
		{
			name:    "token minimum",
			mutate:  func(taskpack *spec.Taskpack) { taskpack.ContextBudget.MaxInputTokens = 128 },
			wantErr: "max_input_tokens",
		},
		{
			name:    "timestamp",
			mutate:  func(taskpack *spec.Taskpack) { taskpack.CreatedAt = "tomorrow" },
			wantErr: "created_at",
		},
		{
			name:    "actor UUID",
			mutate:  func(taskpack *spec.Taskpack) { taskpack.RequestedBy.ActorID = "../escape" },
			wantErr: "requested_by.actor_id",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := taskpack
			tc.mutate(&payload)

			err := Taskpack(payload)
			if err == nil {
				t.Fatalf("expected validation error containing %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestArtifactRejectsSpecViolations(t *testing.T) {
	artifact := validArtifact()

	cases := []struct {
		name    string
		mutate  func(*spec.Artifact)
		wantErr string
	}{
		{
			name:    "kind enum",
			mutate:  func(artifact *spec.Artifact) { artifact.Kind = "memo" },
			wantErr: "kind",
		},
		{
			name:    "storage uri",
			mutate:  func(artifact *spec.Artifact) { artifact.Storage.URI = "not a uri" },
			wantErr: "storage.uri",
		},
		{
			name:    "sha256",
			mutate:  func(artifact *spec.Artifact) { artifact.Storage.SHA256 = "abc" },
			wantErr: "storage.sha256",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			payload := artifact
			tc.mutate(&payload)

			err := Artifact(payload)
			if err == nil {
				t.Fatalf("expected validation error containing %q", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestPromotionRecordRejectsSpecViolations(t *testing.T) {
	record := validPromotionRecord()
	record.Decision = "rubber_stamp"

	err := PromotionRecord(record)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "decision") {
		t.Fatalf("expected decision error, got %q", err.Error())
	}
}

func TestReplayBundleRejectsCrossReferenceViolations(t *testing.T) {
	bundle := spec.ReplayBundle{
		SchemaVersion: "v1alpha1",
		Taskpack:      validTaskpack(),
		DriBindings: []spec.DriBinding{
			validDriBinding(),
		},
		Artifacts: []spec.Artifact{
			validArtifact(),
		},
		PromotionRecords: []spec.PromotionRecord{
			validPromotionRecord(),
		},
	}

	if err := ReplayBundle(bundle); err != nil {
		t.Fatalf("expected valid replay bundle, got %v", err)
	}

	bundle.PromotionRecords[0].CandidateRef.ArtifactID = "11111111-1111-1111-1111-111111111111"
	err := ReplayBundle(bundle)
	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "candidate_ref.artifact_id") {
		t.Fatalf("expected candidate ref error, got %q", err.Error())
	}
}

func validTaskpack() spec.Taskpack {
	return spec.Taskpack{
		SchemaVersion: "v1alpha1",
		TaskpackID:    "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
		Title:         "Audit retry path",
		Objective:     "Find retry edge cases.",
		TaskType:      "analysis",
		Priority:      "high",
		RequestedBy: spec.ActorRef{
			ActorID:     "52a203e8-0a20-4f0b-9b3d-78631e00d6ab",
			ActorType:   "human",
			DisplayName: "Quentin",
		},
		ContextBudget: spec.ContextBudget{
			MaxInputTokens:     12000,
			MaxOutputTokens:    2500,
			ContextStrategy:    "artifact_refs_first",
			MaxArtifactsInline: 2,
		},
		Permissions: spec.Permissions{
			ApprovalMode: "ask",
		},
		Acceptance: []spec.AcceptanceCriteria{
			{
				CriterionID: "root-cause",
				Description: "Find one likely duplicate side-effect path.",
				Required:    true,
			},
		},
		CreatedAt: "2026-04-24T10:00:00Z",
	}
}

func validDriBinding() spec.DriBinding {
	return spec.DriBinding{
		SchemaVersion: "v1alpha1",
		DriBindingID:  "19887415-bb68-438b-9599-0b07b5f13e97",
		TaskpackID:    "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
		Owner: spec.ActorRef{
			ActorID:     "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
			ActorType:   "agent",
			DisplayName: "payments-dri",
		},
		Status:    "assigned",
		CreatedAt: "2026-04-24T10:01:00Z",
	}
}

func validArtifact() spec.Artifact {
	return spec.Artifact{
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
