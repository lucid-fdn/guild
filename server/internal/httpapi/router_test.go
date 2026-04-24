package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lucid-fdn/guild/pkg/spec"
	"github.com/lucid-fdn/guild/server/internal/config"
	"github.com/lucid-fdn/guild/server/internal/evaluator"
	"github.com/lucid-fdn/guild/server/internal/storage"
)

func TestPostTaskpackRejectsUnknownFields(t *testing.T) {
	router := newTestRouter()

	request := httptest.NewRequest(http.MethodPost, "/api/v1/taskpacks", strings.NewReader(`{
		"schema_version": "v1alpha1",
		"taskpack_id": "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
		"title": "Audit retry path",
		"objective": "Find retry edge cases.",
		"task_type": "analysis",
		"priority": "high",
		"requested_by": {
			"actor_id": "52a203e8-0a20-4f0b-9b3d-78631e00d6ab",
			"actor_type": "human",
			"display_name": "Quentin"
		},
		"context_budget": {
			"max_input_tokens": 12000,
			"max_output_tokens": 2500,
			"context_strategy": "artifact_refs_first"
		},
		"permissions": {
			"allow_network": false,
			"allow_shell": false,
			"allow_external_write": false,
			"approval_mode": "ask"
		},
		"acceptance_criteria": [
			{
				"criterion_id": "root-cause",
				"description": "Find one likely duplicate side-effect path."
			}
		],
		"created_at": "2026-04-24T10:00:00Z",
		"surprise": true
	}`))
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
	if !strings.Contains(response.Body.String(), "unknown field") {
		t.Fatalf("expected unknown field error, got %s", response.Body.String())
	}
}

func TestItemRoutesRejectMalformedIDs(t *testing.T) {
	router := newTestRouter()

	cases := []string{
		"/api/v1/taskpacks/not-a-uuid",
		"/api/v1/taskpacks/not-a-uuid/artifacts",
		"/api/v1/dri-bindings/not-a-uuid",
		"/api/v1/artifacts/not-a-uuid",
		"/api/v1/promotion-records/not-a-uuid",
	}

	for _, path := range cases {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))

			if response.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d: %s", response.Code, response.Body.String())
			}
			if !strings.Contains(response.Body.String(), "UUID") {
				t.Fatalf("expected UUID error, got %s", response.Body.String())
			}
		})
	}
}

func TestItemRoutesRejectUnsupportedMethods(t *testing.T) {
	router := newTestRouter()

	cases := []string{
		"/api/v1/taskpacks/4e1fe00c-6303-453c-8cb6-2c34f84896e4",
		"/api/v1/dri-bindings/19887415-bb68-438b-9599-0b07b5f13e97",
		"/api/v1/artifacts/5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
		"/api/v1/promotion-records/b2ddb0dd-b29c-4a28-b1ba-e9a2f8ff23fb",
	}

	for _, path := range cases {
		t.Run(path, func(t *testing.T) {
			response := httptest.NewRecorder()
			router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, path, nil))

			if response.Code != http.StatusMethodNotAllowed {
				t.Fatalf("expected 405, got %d", response.Code)
			}
		})
	}
}

func TestReplayBundleIncludesTaskOwnershipArtifactsAndPromotions(t *testing.T) {
	router := newTestRouter()
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/replay/taskpacks/4e1fe00c-6303-453c-8cb6-2c34f84896e4", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", response.Code, response.Body.String())
	}

	var bundle spec.ReplayBundle
	if err := json.NewDecoder(response.Body).Decode(&bundle); err != nil {
		t.Fatal(err)
	}
	if bundle.SchemaVersion != "v1alpha1" {
		t.Fatalf("expected schema version v1alpha1, got %q", bundle.SchemaVersion)
	}
	if bundle.Taskpack.TaskpackID != "4e1fe00c-6303-453c-8cb6-2c34f84896e4" {
		t.Fatalf("unexpected taskpack id %q", bundle.Taskpack.TaskpackID)
	}
	if len(bundle.DriBindings) != 1 {
		t.Fatalf("expected one DRI binding, got %d", len(bundle.DriBindings))
	}
	if len(bundle.Artifacts) != 1 {
		t.Fatalf("expected one artifact, got %d", len(bundle.Artifacts))
	}
	if len(bundle.PromotionRecords) != 1 {
		t.Fatalf("expected one promotion record, got %d", len(bundle.PromotionRecords))
	}
}

func TestReplayBundleRejectsMalformedTaskpackID(t *testing.T) {
	router := newTestRouter()
	response := httptest.NewRecorder()

	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/replay/taskpacks/not-a-uuid", nil))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func newTestRouter() http.Handler {
	return NewRouter(RouterDeps{
		Config:      config.Config{UIOrigin: "http://localhost:3000"},
		Tasks:       fakeTasks{},
		DRI:         fakeDRI{},
		Artifacts:   fakeArtifacts{},
		Promotions:  fakePromotions{},
		Evaluator:   fakeEvaluator{},
		Institution: fakeInstitution{},
	})
}

type fakeTasks struct{}

func (fakeTasks) Status() map[string]any     { return map[string]any{"service": "tasks"} }
func (fakeTasks) Create(spec.Taskpack) error { return nil }
func (fakeTasks) Get(id string) (spec.Taskpack, error) {
	if id != "4e1fe00c-6303-453c-8cb6-2c34f84896e4" {
		return spec.Taskpack{}, storage.ErrNotFound
	}
	return spec.Taskpack{
		SchemaVersion: "v1alpha1",
		TaskpackID:    id,
		Title:         "Replayable task",
		Objective:     "Prove replay bundles are portable.",
		TaskType:      "analysis",
		Priority:      "medium",
		RequestedBy: spec.ActorRef{
			ActorID:     "52a203e8-0a20-4f0b-9b3d-78631e00d6ab",
			ActorType:   "human",
			DisplayName: "Quentin",
		},
		ContextBudget: spec.ContextBudget{
			MaxInputTokens:  12000,
			MaxOutputTokens: 2500,
			ContextStrategy: "artifact_refs_first",
		},
		Permissions: spec.Permissions{
			ApprovalMode: "ask",
		},
		Acceptance: []spec.AcceptanceCriteria{
			{
				CriterionID: "bundle-export",
				Description: "Replay bundle exports.",
				Required:    true,
			},
		},
		CreatedAt: "2026-04-24T10:00:00Z",
	}, nil
}
func (fakeTasks) List() ([]spec.Taskpack, error) { return nil, nil }

type fakeDRI struct{}

func (fakeDRI) Status() map[string]any       { return map[string]any{"service": "dri"} }
func (fakeDRI) Create(spec.DriBinding) error { return nil }
func (fakeDRI) Get(string) (spec.DriBinding, error) {
	return spec.DriBinding{}, storage.ErrNotFound
}
func (fakeDRI) List() ([]spec.DriBinding, error) {
	return []spec.DriBinding{
		{
			SchemaVersion: "v1alpha1",
			DriBindingID:  "19887415-bb68-438b-9599-0b07b5f13e97",
			TaskpackID:    "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
			Owner: spec.ActorRef{
				ActorID:     "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
				ActorType:   "agent",
				DisplayName: "owner",
			},
			Status:    "assigned",
			CreatedAt: "2026-04-24T10:01:00Z",
		},
		{
			SchemaVersion: "v1alpha1",
			DriBindingID:  "84c3fcf4-6592-4e8b-9847-3220864b6867",
			TaskpackID:    "4fd1c0f8-f6f8-4440-8314-37ea85b3da48",
			Owner: spec.ActorRef{
				ActorID:     "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
				ActorType:   "agent",
				DisplayName: "other-owner",
			},
			Status:    "assigned",
			CreatedAt: "2026-04-24T10:01:00Z",
		},
	}, nil
}

type fakeArtifacts struct{}

func (fakeArtifacts) Status() map[string]any     { return map[string]any{"service": "artifacts"} }
func (fakeArtifacts) Create(spec.Artifact) error { return nil }
func (fakeArtifacts) Get(string) (spec.Artifact, error) {
	return spec.Artifact{}, storage.ErrNotFound
}
func (fakeArtifacts) List() ([]spec.Artifact, error) {
	return []spec.Artifact{
		{
			SchemaVersion: "v1alpha1",
			ArtifactID:    "5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
			TaskpackID:    "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
			Kind:          "plan",
			Title:         "Replay plan",
			Producer: spec.ActorRef{
				ActorID:     "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
				ActorType:   "agent",
				DisplayName: "owner",
			},
			Storage: spec.ArtifactStorage{
				URI:      "s3://guild/replay-plan.md",
				MimeType: "text/markdown",
			},
			Version:   1,
			CreatedAt: "2026-04-24T10:02:00Z",
		},
	}, nil
}
func (fakeArtifacts) ListByTaskpack(string) ([]spec.Artifact, error) {
	return fakeArtifacts{}.List()
}

type fakePromotions struct{}

func (fakePromotions) Status() map[string]any            { return map[string]any{"service": "promotions"} }
func (fakePromotions) Create(spec.PromotionRecord) error { return nil }
func (fakePromotions) Get(string) (spec.PromotionRecord, error) {
	return spec.PromotionRecord{}, storage.ErrNotFound
}

type fakeEvaluator struct{}

func (fakeEvaluator) Status() map[string]any {
	return map[string]any{"service": "evaluator"}
}
func (fakeEvaluator) Enqueue(suite evaluator.ReplaySuite) (evaluator.EvaluationJob, error) {
	return evaluator.EvaluationJob{
		SchemaVersion:   "v1alpha1",
		EvaluationJobID: "2b63be41-d0a1-4d26-acb1-cab92cf3301e",
		Status:          "queued",
		Suite:           suite,
		CreatedAt:       "2026-04-24T10:00:00Z",
		UpdatedAt:       "2026-04-24T10:00:00Z",
		QueuedAt:        "2026-04-24T10:00:00Z",
	}, nil
}
func (fakeEvaluator) Get(id string) (evaluator.EvaluationJob, error) {
	if id != "2b63be41-d0a1-4d26-acb1-cab92cf3301e" {
		return evaluator.EvaluationJob{}, storage.ErrNotFound
	}
	return evaluator.EvaluationJob{
		SchemaVersion:   "v1alpha1",
		EvaluationJobID: id,
		Status:          "queued",
		CreatedAt:       "2026-04-24T10:00:00Z",
		UpdatedAt:       "2026-04-24T10:00:00Z",
		QueuedAt:        "2026-04-24T10:00:00Z",
	}, nil
}
func (fakeEvaluator) List() ([]evaluator.EvaluationJob, error) {
	job, _ := fakeEvaluator{}.Get("2b63be41-d0a1-4d26-acb1-cab92cf3301e")
	return []evaluator.EvaluationJob{job}, nil
}
func (fakeEvaluator) RunJob(id string) (evaluator.EvaluationJob, error) {
	job, err := fakeEvaluator{}.Get(id)
	if err != nil {
		return evaluator.EvaluationJob{}, err
	}
	job.Status = "succeeded"
	return job, nil
}

type fakeInstitution struct{}

func (fakeInstitution) Status() map[string]any                           { return map[string]any{"service": "institution"} }
func (fakeInstitution) CreatePolicy(spec.GovernancePolicy) error         { return nil }
func (fakeInstitution) ListPolicies() ([]spec.GovernancePolicy, error)   { return nil, nil }
func (fakeInstitution) CreateApproval(spec.ApprovalRequest) error        { return nil }
func (fakeInstitution) ListApprovals() ([]spec.ApprovalRequest, error)   { return nil, nil }
func (fakeInstitution) CreateGate(spec.PromotionGate) error              { return nil }
func (fakeInstitution) ListGates() ([]spec.PromotionGate, error)         { return nil, nil }
func (fakeInstitution) CreateCommonsEntry(spec.CommonsEntry) error       { return nil }
func (fakeInstitution) ListCommonsEntries() ([]spec.CommonsEntry, error) { return nil, nil }
func (fakePromotions) List() ([]spec.PromotionRecord, error) {
	return []spec.PromotionRecord{
		{
			SchemaVersion: "v1alpha1",
			PromotionID:   "b2ddb0dd-b29c-4a28-b1ba-e9a2f8ff23fb",
			InstitutionID: "7fb72c51-b13a-4c68-bf67-c042ac2fa10c",
			CandidateKind: "workflow_pattern",
			CandidateRef: spec.ArtifactRef{
				ArtifactID: "5d19ef0f-31b5-4709-a66b-da2d3bb4a731",
				Kind:       "plan",
				URI:        "s3://guild/replay-plan.md",
				Version:    1,
			},
			Decision:  "accepted",
			DecidedAt: "2026-04-24T10:03:00Z",
		},
		{
			SchemaVersion: "v1alpha1",
			PromotionID:   "88f97639-04dd-4886-8810-06978c5716f2",
			InstitutionID: "7fb72c51-b13a-4c68-bf67-c042ac2fa10c",
			CandidateKind: "workflow_pattern",
			CandidateRef: spec.ArtifactRef{
				ArtifactID: "a662a2ab-72cd-4030-bf33-72997d65ba09",
				Kind:       "plan",
				URI:        "s3://guild/other-plan.md",
				Version:    1,
			},
			Decision:  "accepted",
			DecidedAt: "2026-04-24T10:03:00Z",
		},
	}, nil
}
