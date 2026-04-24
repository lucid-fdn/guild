package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/lucid-fdn/guild/pkg/spec"
	specvalidate "github.com/lucid-fdn/guild/pkg/spec/validate"
	"github.com/lucid-fdn/guild/server/internal/config"
	"github.com/lucid-fdn/guild/server/internal/evaluator"
	"github.com/lucid-fdn/guild/server/internal/storage"
)

type TaskpackService interface {
	Status() map[string]any
	Create(spec.Taskpack) error
	Get(string) (spec.Taskpack, error)
	List() ([]spec.Taskpack, error)
}

type DriService interface {
	Status() map[string]any
	Create(spec.DriBinding) error
	Get(string) (spec.DriBinding, error)
	List() ([]spec.DriBinding, error)
}

type ArtifactService interface {
	Status() map[string]any
	Create(spec.Artifact) error
	Get(string) (spec.Artifact, error)
	List() ([]spec.Artifact, error)
	ListByTaskpack(string) ([]spec.Artifact, error)
}

type PromotionService interface {
	Status() map[string]any
	Create(spec.PromotionRecord) error
	Get(string) (spec.PromotionRecord, error)
	List() ([]spec.PromotionRecord, error)
}

type EvaluatorService interface {
	Status() map[string]any
	Enqueue(evaluator.ReplaySuite) (evaluator.EvaluationJob, error)
	Get(string) (evaluator.EvaluationJob, error)
	List() ([]evaluator.EvaluationJob, error)
	RunJob(string) (evaluator.EvaluationJob, error)
}

type InstitutionService interface {
	Status() map[string]any
	CreatePolicy(spec.GovernancePolicy) error
	ListPolicies() ([]spec.GovernancePolicy, error)
	CreateApproval(spec.ApprovalRequest) error
	ListApprovals() ([]spec.ApprovalRequest, error)
	CreateGate(spec.PromotionGate) error
	ListGates() ([]spec.PromotionGate, error)
	CreateCommonsEntry(spec.CommonsEntry) error
	ListCommonsEntries() ([]spec.CommonsEntry, error)
}

type RouterDeps struct {
	Config      config.Config
	Tasks       TaskpackService
	DRI         DriService
	Artifacts   ArtifactService
	Promotions  PromotionService
	Evaluator   EvaluatorService
	Institution InstitutionService
}

func NewRouter(deps RouterDeps) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "ok",
			"service": "guildd",
		})
	})

	mux.HandleFunc("/api/v1/status", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"name":      "guild",
			"version":   "0.1.0",
			"mode":      "bootstrap",
			"ui_origin": deps.Config.UIOrigin,
			"services": []map[string]any{
				deps.Tasks.Status(),
				deps.DRI.Status(),
				deps.Artifacts.Status(),
				deps.Promotions.Status(),
				deps.Evaluator.Status(),
				deps.Institution.Status(),
			},
		})
	})

	mux.HandleFunc("/api/v1/taskpacks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			items, err := deps.Tasks.List()
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
		case http.MethodPost:
			var payload spec.Taskpack
			if err := decodeJSON(r, &payload); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			if err := deps.Tasks.Create(payload); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			writeJSON(w, http.StatusCreated, payload)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/dri-bindings", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			items, err := deps.DRI.List()
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
		case http.MethodPost:
			var payload spec.DriBinding
			if err := decodeJSON(r, &payload); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			if err := deps.DRI.Create(payload); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			writeJSON(w, http.StatusCreated, payload)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/artifacts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			items, err := deps.Artifacts.List()
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
		case http.MethodPost:
			var payload spec.Artifact
			if err := decodeJSON(r, &payload); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			if err := deps.Artifacts.Create(payload); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			writeJSON(w, http.StatusCreated, payload)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/promotion-records", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			items, err := deps.Promotions.List()
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
		case http.MethodPost:
			var payload spec.PromotionRecord
			if err := decodeJSON(r, &payload); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			if err := deps.Promotions.Create(payload); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			writeJSON(w, http.StatusCreated, payload)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/evaluation-jobs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			items, err := deps.Evaluator.List()
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
		case http.MethodPost:
			var payload evaluator.ReplaySuite
			if err := decodeJSON(r, &payload); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			job, err := deps.Evaluator.Enqueue(payload)
			if err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			writeJSON(w, http.StatusAccepted, job)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/governance-policies", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			items, err := deps.Institution.ListPolicies()
			writeCollectionOrError(w, items, err)
		case http.MethodPost:
			var payload spec.GovernancePolicy
			if decodeOrCreate(w, r, &payload, deps.Institution.CreatePolicy) {
				writeJSON(w, http.StatusCreated, payload)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/approval-requests", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			items, err := deps.Institution.ListApprovals()
			writeCollectionOrError(w, items, err)
		case http.MethodPost:
			var payload spec.ApprovalRequest
			if decodeOrCreate(w, r, &payload, deps.Institution.CreateApproval) {
				writeJSON(w, http.StatusCreated, payload)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/promotion-gates", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			items, err := deps.Institution.ListGates()
			writeCollectionOrError(w, items, err)
		case http.MethodPost:
			var payload spec.PromotionGate
			if decodeOrCreate(w, r, &payload, deps.Institution.CreateGate) {
				writeJSON(w, http.StatusCreated, payload)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/commons-entries", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			items, err := deps.Institution.ListCommonsEntries()
			writeCollectionOrError(w, items, err)
		case http.MethodPost:
			var payload spec.CommonsEntry
			if decodeOrCreate(w, r, &payload, deps.Institution.CreateCommonsEntry) {
				writeJSON(w, http.StatusCreated, payload)
			}
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/replay/taskpacks/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		taskpackID := strings.TrimPrefix(r.URL.Path, "/api/v1/replay/taskpacks/")
		if !validatePathUUID(w, "taskpack_id", taskpackID) {
			return
		}
		bundle, err := buildReplayBundle(deps, taskpackID)
		if err != nil {
			handleLookupError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, bundle)
	})

	mux.HandleFunc("/api/v1/taskpacks/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/taskpacks/")
		if suffix == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if strings.HasSuffix(suffix, "/artifacts") {
			taskpackID := strings.TrimSuffix(suffix, "/artifacts")
			if !validatePathUUID(w, "taskpack_id", taskpackID) {
				return
			}
			items, err := deps.Artifacts.ListByTaskpack(taskpackID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, err)
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"items": items})
			return
		}

		if !validatePathUUID(w, "taskpack_id", suffix) {
			return
		}
		item, err := deps.Tasks.Get(suffix)
		if err != nil {
			handleLookupError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	})

	mux.HandleFunc("/api/v1/dri-bindings/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/dri-bindings/")
		if !validatePathUUID(w, "dri_binding_id", id) {
			return
		}
		item, err := deps.DRI.Get(id)
		if err != nil {
			handleLookupError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	})

	mux.HandleFunc("/api/v1/artifacts/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/artifacts/")
		if !validatePathUUID(w, "artifact_id", id) {
			return
		}
		item, err := deps.Artifacts.Get(id)
		if err != nil {
			handleLookupError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	})

	mux.HandleFunc("/api/v1/promotion-records/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		id := strings.TrimPrefix(r.URL.Path, "/api/v1/promotion-records/")
		if !validatePathUUID(w, "promotion_record_id", id) {
			return
		}
		item, err := deps.Promotions.Get(id)
		if err != nil {
			handleLookupError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	})

	mux.HandleFunc("/api/v1/evaluation-jobs/", func(w http.ResponseWriter, r *http.Request) {
		suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/evaluation-jobs/")
		if strings.HasSuffix(suffix, "/run") {
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			id := strings.TrimSuffix(suffix, "/run")
			if !validatePathUUID(w, "evaluation_job_id", id) {
				return
			}
			job, err := deps.Evaluator.RunJob(id)
			if err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
			writeJSON(w, http.StatusOK, job)
			return
		}
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if !validatePathUUID(w, "evaluation_job_id", suffix) {
			return
		}
		item, err := deps.Evaluator.Get(suffix)
		if err != nil {
			handleLookupError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, item)
	})

	return withCORS(mux, deps.Config.UIOrigin)
}

func buildReplayBundle(deps RouterDeps, taskpackID string) (spec.ReplayBundle, error) {
	taskpack, err := deps.Tasks.Get(taskpackID)
	if err != nil {
		return spec.ReplayBundle{}, err
	}

	allTaskpacks, err := deps.Tasks.List()
	if err != nil {
		return spec.ReplayBundle{}, err
	}
	taskpacks := recursiveTaskpacks(taskpack, allTaskpacks)
	taskpackIDs := make(map[string]struct{}, len(taskpacks))
	for _, item := range taskpacks {
		taskpackIDs[item.TaskpackID] = struct{}{}
	}

	allArtifacts, err := deps.Artifacts.List()
	if err != nil {
		return spec.ReplayBundle{}, err
	}
	artifacts := make([]spec.Artifact, 0)
	for _, artifact := range allArtifacts {
		if _, ok := taskpackIDs[artifact.TaskpackID]; ok {
			artifacts = append(artifacts, artifact)
		}
	}
	allBindings, err := deps.DRI.List()
	if err != nil {
		return spec.ReplayBundle{}, err
	}
	allPromotions, err := deps.Promotions.List()
	if err != nil {
		return spec.ReplayBundle{}, err
	}

	artifactIDs := make(map[string]struct{}, len(artifacts))
	for _, artifact := range artifacts {
		artifactIDs[artifact.ArtifactID] = struct{}{}
	}

	bindings := make([]spec.DriBinding, 0)
	for _, binding := range allBindings {
		if _, ok := taskpackIDs[binding.TaskpackID]; ok {
			bindings = append(bindings, binding)
		}
	}

	promotions := make([]spec.PromotionRecord, 0)
	for _, promotion := range allPromotions {
		if _, ok := artifactIDs[promotion.CandidateRef.ArtifactID]; ok {
			promotions = append(promotions, promotion)
		}
	}

	return spec.ReplayBundle{
		SchemaVersion:    "v1alpha1",
		RootTaskpackID:   taskpack.TaskpackID,
		Taskpack:         taskpack,
		Taskpacks:        taskpacks,
		DriBindings:      bindings,
		Artifacts:        artifacts,
		PromotionRecords: promotions,
	}, nil
}

func recursiveTaskpacks(root spec.Taskpack, all []spec.Taskpack) []spec.Taskpack {
	childrenByParent := make(map[string][]spec.Taskpack)
	for _, taskpack := range all {
		if taskpack.ParentTaskpackID != "" {
			childrenByParent[taskpack.ParentTaskpackID] = append(childrenByParent[taskpack.ParentTaskpackID], taskpack)
		}
	}
	result := []spec.Taskpack{root}
	seen := map[string]struct{}{root.TaskpackID: {}}
	var walk func(string)
	walk = func(parentID string) {
		for _, child := range childrenByParent[parentID] {
			if _, ok := seen[child.TaskpackID]; ok {
				continue
			}
			seen[child.TaskpackID] = struct{}{}
			result = append(result, child)
			walk(child.TaskpackID)
		}
	}
	walk(root.TaskpackID)
	return result
}

func withCORS(next http.Handler, origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeCollectionOrError(w http.ResponseWriter, items any, err error) {
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func decodeOrCreate[T any](w http.ResponseWriter, r *http.Request, payload *T, create func(T) error) bool {
	if err := decodeJSON(r, payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return false
	}
	if err := create(*payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return false
	}
	return true
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

func handleLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeError(w, http.StatusInternalServerError, err)
}

func validatePathUUID(w http.ResponseWriter, name, value string) bool {
	if specvalidate.IsUUID(value) {
		return true
	}
	writeError(w, http.StatusBadRequest, errors.New(name+" must be a UUID"))
	return false
}

func decodeJSON(r *http.Request, dest any) error {
	defer r.Body.Close()
	decoder := json.NewDecoder(io.LimitReader(r.Body, 2<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return err
	}
	var extra struct{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request body must contain a single JSON document")
	}
	return nil
}
