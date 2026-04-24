package evaluator

import (
	"crypto/rand"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/lucid-fdn/guild/pkg/spec"
	specvalidate "github.com/lucid-fdn/guild/pkg/spec/validate"
	"github.com/lucid-fdn/guild/server/internal/storage"
)

type TaskpackService interface {
	Get(string) (spec.Taskpack, error)
	List() ([]spec.Taskpack, error)
}

type DriService interface {
	List() ([]spec.DriBinding, error)
}

type ArtifactService interface {
	Create(spec.Artifact) error
	List() ([]spec.Artifact, error)
}

type PromotionService interface {
	Create(spec.PromotionRecord) error
	List() ([]spec.PromotionRecord, error)
}

type ReplaySuite struct {
	SuiteID     string   `json:"suite_id"`
	Name        string   `json:"name"`
	TaskpackIDs []string `json:"taskpack_ids"`
	MetricName  string   `json:"metric_name"`
	Before      float64  `json:"before"`
	After       float64  `json:"after"`
	Direction   string   `json:"direction"`
}

type EvaluationJob struct {
	SchemaVersion       string      `json:"schema_version"`
	EvaluationJobID     string      `json:"evaluation_job_id"`
	Status              string      `json:"status"`
	Suite               ReplaySuite `json:"suite"`
	Attempts            int         `json:"attempts"`
	MaxAttempts         int         `json:"max_attempts"`
	Error               string      `json:"error,omitempty"`
	BenchmarkArtifactID string      `json:"benchmark_artifact_id,omitempty"`
	CandidateArtifactID string      `json:"candidate_artifact_id,omitempty"`
	PromotionRecordID   string      `json:"promotion_record_id,omitempty"`
	QueuedAt            string      `json:"queued_at"`
	StartedAt           string      `json:"started_at,omitempty"`
	CompletedAt         string      `json:"completed_at,omitempty"`
	CreatedAt           string      `json:"created_at"`
	UpdatedAt           string      `json:"updated_at"`
}

type Service struct {
	store      storage.Store
	tasks      TaskpackService
	dri        DriService
	artifacts  ArtifactService
	promotions PromotionService
}

func NewService(store storage.Store, tasks TaskpackService, dri DriService, artifacts ArtifactService, promotions PromotionService) *Service {
	return &Service{store: store, tasks: tasks, dri: dri, artifacts: artifacts, promotions: promotions}
}

func (s *Service) Enqueue(suite ReplaySuite) (EvaluationJob, error) {
	if err := ValidateReplaySuite(suite); err != nil {
		return EvaluationJob{}, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	job := EvaluationJob{
		SchemaVersion:   "v1alpha1",
		EvaluationJobID: newUUID(),
		Status:          "queued",
		Suite:           suite,
		MaxAttempts:     3,
		QueuedAt:        now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	if err := s.store.Put("evaluation-jobs", job.EvaluationJobID, job); err != nil {
		return EvaluationJob{}, err
	}
	return job, nil
}

func (s *Service) Get(id string) (EvaluationJob, error) {
	var job EvaluationJob
	err := s.store.Get("evaluation-jobs", id, &job)
	return job, err
}

func (s *Service) List() ([]EvaluationJob, error) {
	var jobs []EvaluationJob
	err := s.store.List("evaluation-jobs", &jobs)
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].CreatedAt < jobs[j].CreatedAt
	})
	return jobs, err
}

func (s *Service) RunJob(id string) (EvaluationJob, error) {
	job, err := s.Get(id)
	if err != nil {
		return EvaluationJob{}, err
	}
	if job.Status == "succeeded" {
		return job, nil
	}
	if job.Attempts >= job.MaxAttempts {
		job.Status = "failed"
		job.Error = "max attempts exceeded"
		return s.save(job)
	}

	now := time.Now().UTC().Format(time.RFC3339)
	job.Status = "running"
	job.Attempts++
	job.StartedAt = now
	job.UpdatedAt = now
	if err := s.store.Put("evaluation-jobs", job.EvaluationJobID, job); err != nil {
		return EvaluationJob{}, err
	}

	result, runErr := s.runSuite(job)
	if runErr != nil {
		job.Status = "failed"
		if job.Attempts < job.MaxAttempts {
			job.Status = "queued"
		}
		job.Error = runErr.Error()
		job.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
		_, _ = s.save(job)
		return job, runErr
	}
	return result, nil
}

func (s *Service) RunNext() (EvaluationJob, bool, error) {
	jobs, err := s.List()
	if err != nil {
		return EvaluationJob{}, false, err
	}
	for _, job := range jobs {
		if job.Status == "queued" {
			result, err := s.RunJob(job.EvaluationJobID)
			return result, true, err
		}
	}
	return EvaluationJob{}, false, nil
}

func (s *Service) Status() map[string]any {
	jobs, _ := s.List()
	counts := map[string]int{}
	for _, job := range jobs {
		counts[job.Status]++
	}
	return map[string]any{
		"service": "evaluator",
		"state":   "ready",
		"count":   len(jobs),
		"jobs":    counts,
	}
}

func (s *Service) runSuite(job EvaluationJob) (EvaluationJob, error) {
	bundles := make([]spec.ReplayBundle, 0, len(job.Suite.TaskpackIDs))
	for _, taskpackID := range job.Suite.TaskpackIDs {
		bundle, err := s.buildReplayBundle(taskpackID)
		if err != nil {
			return job, err
		}
		if err := specvalidate.ReplayBundle(bundle); err != nil {
			return job, fmt.Errorf("invalid replay bundle for %s: %w", taskpackID, err)
		}
		bundles = append(bundles, bundle)
	}
	root := bundles[0].Taskpack
	if root.InstitutionID == "" {
		return job, errors.New("root taskpack must include institution_id")
	}
	now := time.Now().UTC().Format(time.RFC3339)
	evaluator := spec.ActorRef{
		ActorID:      "7a9b2679-bc52-47cb-9218-422fce18bd80",
		ActorType:    "agent",
		DisplayName:  "guild-evaluator",
		Orchestrator: "guild",
	}
	benchmarkArtifact := spec.Artifact{
		SchemaVersion: "v1alpha1",
		ArtifactID:    newUUID(),
		TaskpackID:    root.TaskpackID,
		Kind:          "benchmark_result",
		Title:         "Replay suite benchmark: " + job.Suite.Name,
		Summary:       fmt.Sprintf("%s moved from %.3f to %.3f across %d replay bundle(s).", job.Suite.MetricName, job.Suite.Before, job.Suite.After, len(bundles)),
		Producer:      evaluator,
		Storage: spec.ArtifactStorage{
			URI:      "guild://replay-suites/" + job.Suite.SuiteID + "/jobs/" + job.EvaluationJobID + "/benchmark-result.json",
			MimeType: "application/json",
		},
		EvaluationState: &spec.EvaluationState{
			Status:         "passed",
			Score:          job.Suite.After,
			BenchmarkSuite: job.Suite.SuiteID,
		},
		Labels:    []string{"replay-suite", "benchmark-result", "evaluation-job"},
		Version:   1,
		CreatedAt: now,
	}
	if err := s.artifacts.Create(benchmarkArtifact); err != nil {
		return job, err
	}
	candidateArtifact := spec.Artifact{
		SchemaVersion:     "v1alpha1",
		ArtifactID:        newUUID(),
		TaskpackID:        root.TaskpackID,
		ParentArtifactIDs: []string{benchmarkArtifact.ArtifactID},
		Kind:              "skill_candidate",
		Title:             "Promotion candidate from replay suite: " + job.Suite.Name,
		Summary:           "Candidate learning generated from replay benchmark evidence and awaiting human review.",
		Producer:          evaluator,
		Storage: spec.ArtifactStorage{
			URI:      "guild://replay-suites/" + job.Suite.SuiteID + "/jobs/" + job.EvaluationJobID + "/promotion-candidate.md",
			MimeType: "text/markdown",
		},
		EvaluationState: &spec.EvaluationState{
			Status:         "pending",
			Score:          job.Suite.After,
			BenchmarkSuite: job.Suite.SuiteID,
		},
		Labels:    []string{"replay-suite", "promotion-candidate", "evaluation-job"},
		Version:   1,
		CreatedAt: now,
	}
	if err := s.artifacts.Create(candidateArtifact); err != nil {
		return job, err
	}
	record := spec.PromotionRecord{
		SchemaVersion:  "v1alpha1",
		PromotionID:    newUUID(),
		InstitutionID:  root.InstitutionID,
		CandidateKind:  "skill",
		CandidateRef:   spec.ArtifactRef{ArtifactID: candidateArtifact.ArtifactID, Kind: candidateArtifact.Kind, URI: candidateArtifact.Storage.URI, Version: candidateArtifact.Version},
		SourceRunIDs:   job.Suite.TaskpackIDs,
		BenchmarkSuite: job.Suite.SuiteID,
		Metrics: []spec.MetricDelta{
			{Name: job.Suite.MetricName, Before: job.Suite.Before, After: job.Suite.After, Direction: job.Suite.Direction},
		},
		Decision:       "needs_human_review",
		DecisionReason: "Evaluation job created this candidate; a human must approve promotion into the commons.",
		Deciders:       []spec.ActorRef{evaluator},
		DecidedAt:      now,
	}
	if err := s.promotions.Create(record); err != nil {
		return job, err
	}
	job.Status = "succeeded"
	job.Error = ""
	job.BenchmarkArtifactID = benchmarkArtifact.ArtifactID
	job.CandidateArtifactID = candidateArtifact.ArtifactID
	job.PromotionRecordID = record.PromotionID
	job.CompletedAt = now
	job.UpdatedAt = now
	return s.save(job)
}

func (s *Service) buildReplayBundle(taskpackID string) (spec.ReplayBundle, error) {
	taskpack, err := s.tasks.Get(taskpackID)
	if err != nil {
		return spec.ReplayBundle{}, err
	}
	allTaskpacks, err := s.tasks.List()
	if err != nil {
		return spec.ReplayBundle{}, err
	}
	taskpacks := recursiveTaskpacks(taskpack, allTaskpacks)
	taskpackIDs := make(map[string]struct{}, len(taskpacks))
	for _, item := range taskpacks {
		taskpackIDs[item.TaskpackID] = struct{}{}
	}
	allArtifacts, err := s.artifacts.List()
	if err != nil {
		return spec.ReplayBundle{}, err
	}
	artifacts := make([]spec.Artifact, 0)
	artifactIDs := map[string]struct{}{}
	for _, artifact := range allArtifacts {
		if _, ok := taskpackIDs[artifact.TaskpackID]; ok {
			artifacts = append(artifacts, artifact)
			artifactIDs[artifact.ArtifactID] = struct{}{}
		}
	}
	allBindings, err := s.dri.List()
	if err != nil {
		return spec.ReplayBundle{}, err
	}
	bindings := make([]spec.DriBinding, 0)
	for _, binding := range allBindings {
		if _, ok := taskpackIDs[binding.TaskpackID]; ok {
			bindings = append(bindings, binding)
		}
	}
	allPromotions, err := s.promotions.List()
	if err != nil {
		return spec.ReplayBundle{}, err
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

func (s *Service) save(job EvaluationJob) (EvaluationJob, error) {
	job.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	err := s.store.Put("evaluation-jobs", job.EvaluationJobID, job)
	return job, err
}

func ValidateReplaySuite(suite ReplaySuite) error {
	if strings.TrimSpace(suite.SuiteID) == "" {
		return errors.New("suite_id is required")
	}
	if strings.TrimSpace(suite.Name) == "" {
		return errors.New("name is required")
	}
	if len(suite.TaskpackIDs) == 0 {
		return errors.New("taskpack_ids must not be empty")
	}
	for i, taskpackID := range suite.TaskpackIDs {
		if !specvalidate.IsUUID(taskpackID) {
			return fmt.Errorf("taskpack_ids[%d] must be a UUID", i)
		}
	}
	if strings.TrimSpace(suite.MetricName) == "" {
		return errors.New("metric_name is required")
	}
	if suite.Direction != "higher_is_better" && suite.Direction != "lower_is_better" {
		return errors.New("direction must be higher_is_better or lower_is_better")
	}
	return nil
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

func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
