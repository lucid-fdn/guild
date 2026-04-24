package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/lucid-fdn/guild/pkg/spec"
	specvalidate "github.com/lucid-fdn/guild/pkg/spec/validate"
)

const usage = `Guild CLI

Usage:
  guild validate --kind taskpack --file spec/examples/taskpack.example.json
  guild validate --kind replay-bundle --file spec/examples/replay-bundle.example.json
  guild conformance --base-url http://localhost:8080
  guild replay-export --base-url http://localhost:8080 --taskpack-id <uuid> [--file replay.json]
  guild replay-suite --base-url http://localhost:8080 --suite examples/replay-suite.example.json
  guild eval-submit --base-url http://localhost:8080 --suite examples/replay-suite.example.json [--wait]
  guild agentdesk init
  guild agentdesk mandate create "Fix failing auth tests"
  guild agentdesk next
  guild agentdesk preflight --id <uuid> --action write --path src/auth/login.ts
  guild agentdesk context compile --id <uuid> --role coder
  guild agentdesk approval request --id <uuid> --reason "Need owner consent"
  guild agentdesk proof add --id <uuid> --kind test_report --path test-results.xml
  guild agentdesk handoff create --id <uuid> --to reviewer --summary "Ready for review"
  guild agentdesk verify --id <uuid>
  guild agentdesk close --id <uuid>
  guild agentdesk replay export --id <uuid> [--file replay.json]

Commands:
  validate      Validate one Guild spec object with strict decoding
  conformance   Run API conformance checks against a running Guild server
  replay-export Export a replay bundle for one taskpack
  replay-suite  Run a replay/evaluation suite and propose a promotion candidate
  eval-submit   Queue a replay/evaluation job through the control plane
  agentdesk     Local-first mandate, preflight, proof, and replay workflow for agents
`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return errors.New("command is required")
	}

	switch args[0] {
	case "validate":
		return runValidate(args[1:], stdout)
	case "conformance":
		return runConformance(args[1:], stdout)
	case "replay-export":
		return runReplayExport(args[1:], stdout)
	case "replay-suite":
		return runReplaySuite(args[1:], stdout)
	case "eval-submit":
		return runEvalSubmit(args[1:], stdout)
	case "agentdesk":
		return runAgentDesk(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return nil
	default:
		fmt.Fprint(stderr, usage)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runEvalSubmit(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("eval-submit", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	baseURL := fs.String("base-url", "http://localhost:8080", "Guild-compatible API base URL")
	suitePath := fs.String("suite", "", "path to replay suite JSON")
	wait := fs.Bool("wait", false, "run the queued job immediately and return the completed job")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *suitePath == "" {
		return errors.New("--suite is required")
	}
	data, err := os.ReadFile(*suitePath)
	if err != nil {
		return err
	}
	var suite replaySuiteFile
	if err := decodeStrict(data, &suite); err != nil {
		return err
	}
	if err := validateReplaySuiteFile(suite); err != nil {
		return err
	}
	runner := replaySuiteRunner{
		baseURL: strings.TrimRight(*baseURL, "/"),
		client:  http.Client{Timeout: 5 * time.Second},
		stdout:  stdout,
	}
	var job struct {
		EvaluationJobID string `json:"evaluation_job_id"`
		Status          string `json:"status"`
	}
	if err := runner.postJSONStatus("/api/v1/evaluation-jobs", suite, http.StatusAccepted, &job); err != nil {
		return err
	}
	if *wait {
		if err := runner.postJSONStatus("/api/v1/evaluation-jobs/"+job.EvaluationJobID+"/run", map[string]string{}, http.StatusOK, &job); err != nil {
			return err
		}
	}
	fmt.Fprintf(stdout, "eval-job-%s %s\n", job.Status, job.EvaluationJobID)
	return nil
}

type replaySuiteFile struct {
	SuiteID     string   `json:"suite_id"`
	Name        string   `json:"name"`
	TaskpackIDs []string `json:"taskpack_ids"`
	MetricName  string   `json:"metric_name"`
	Before      float64  `json:"before"`
	After       float64  `json:"after"`
	Direction   string   `json:"direction"`
}

func runReplaySuite(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("replay-suite", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	baseURL := fs.String("base-url", "http://localhost:8080", "Guild-compatible API base URL")
	suitePath := fs.String("suite", "", "path to replay suite JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *suitePath == "" {
		return errors.New("--suite is required")
	}

	data, err := os.ReadFile(*suitePath)
	if err != nil {
		return err
	}
	var suite replaySuiteFile
	if err := decodeStrict(data, &suite); err != nil {
		return err
	}
	if err := validateReplaySuiteFile(suite); err != nil {
		return err
	}

	runner := replaySuiteRunner{
		baseURL: strings.TrimRight(*baseURL, "/"),
		client:  http.Client{Timeout: 5 * time.Second},
		stdout:  stdout,
	}
	return runner.run(suite)
}

func validateReplaySuiteFile(suite replaySuiteFile) error {
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

type replaySuiteRunner struct {
	baseURL string
	client  http.Client
	stdout  io.Writer
}

func (r replaySuiteRunner) run(suite replaySuiteFile) error {
	bundles := make([]spec.ReplayBundle, 0, len(suite.TaskpackIDs))
	for _, taskpackID := range suite.TaskpackIDs {
		var bundle spec.ReplayBundle
		if err := r.getJSON("/api/v1/replay/taskpacks/"+taskpackID, &bundle); err != nil {
			return err
		}
		if err := specvalidate.ReplayBundle(bundle); err != nil {
			return fmt.Errorf("invalid replay bundle for %s: %w", taskpackID, err)
		}
		bundles = append(bundles, bundle)
	}
	root := bundles[0].Taskpack
	if root.InstitutionID == "" {
		return errors.New("root taskpack must include institution_id to open a promotion candidate")
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
		ArtifactID:    mustNewUUID(),
		TaskpackID:    root.TaskpackID,
		Kind:          "benchmark_result",
		Title:         "Replay suite benchmark: " + suite.Name,
		Summary:       fmt.Sprintf("%s moved from %.3f to %.3f across %d replay bundle(s).", suite.MetricName, suite.Before, suite.After, len(bundles)),
		Producer:      evaluator,
		Storage: spec.ArtifactStorage{
			URI:      "guild://replay-suites/" + suite.SuiteID + "/benchmark-result.json",
			MimeType: "application/json",
		},
		EvaluationState: &spec.EvaluationState{
			Status:         "passed",
			Score:          suite.After,
			BenchmarkSuite: suite.SuiteID,
		},
		Labels:    []string{"replay-suite", "benchmark-result"},
		Version:   1,
		CreatedAt: now,
	}
	if err := specvalidate.Artifact(benchmarkArtifact); err != nil {
		return err
	}
	if err := r.postJSON("/api/v1/artifacts", benchmarkArtifact, nil); err != nil {
		return err
	}

	candidateArtifact := spec.Artifact{
		SchemaVersion:     "v1alpha1",
		ArtifactID:        mustNewUUID(),
		TaskpackID:        root.TaskpackID,
		ParentArtifactIDs: []string{benchmarkArtifact.ArtifactID},
		Kind:              "skill_candidate",
		Title:             "Promotion candidate from replay suite: " + suite.Name,
		Summary:           "Candidate learning generated from replay benchmark evidence and awaiting human review.",
		Producer:          evaluator,
		Storage: spec.ArtifactStorage{
			URI:      "guild://replay-suites/" + suite.SuiteID + "/promotion-candidate.md",
			MimeType: "text/markdown",
		},
		EvaluationState: &spec.EvaluationState{
			Status:         "pending",
			Score:          suite.After,
			BenchmarkSuite: suite.SuiteID,
		},
		Labels:    []string{"replay-suite", "promotion-candidate"},
		Version:   1,
		CreatedAt: now,
	}
	if err := specvalidate.Artifact(candidateArtifact); err != nil {
		return err
	}
	if err := r.postJSON("/api/v1/artifacts", candidateArtifact, nil); err != nil {
		return err
	}

	record := spec.PromotionRecord{
		SchemaVersion:  "v1alpha1",
		PromotionID:    mustNewUUID(),
		InstitutionID:  root.InstitutionID,
		CandidateKind:  "skill",
		CandidateRef:   spec.ArtifactRef{ArtifactID: candidateArtifact.ArtifactID, Kind: candidateArtifact.Kind, URI: candidateArtifact.Storage.URI, Version: candidateArtifact.Version},
		SourceRunIDs:   suite.TaskpackIDs,
		BenchmarkSuite: suite.SuiteID,
		Metrics: []spec.MetricDelta{
			{Name: suite.MetricName, Before: suite.Before, After: suite.After, Direction: suite.Direction},
		},
		Decision:       "needs_human_review",
		DecisionReason: "Replay suite runner created this candidate; a human must approve promotion into the commons.",
		Deciders:       []spec.ActorRef{evaluator},
		DecidedAt:      now,
	}
	if err := specvalidate.PromotionRecord(record); err != nil {
		return err
	}
	if err := r.postJSON("/api/v1/promotion-records", record, nil); err != nil {
		return err
	}

	fmt.Fprintf(r.stdout, "replay-suite-ok %s benchmark_artifact=%s candidate_artifact=%s promotion_record=%s\n", suite.SuiteID, benchmarkArtifact.ArtifactID, candidateArtifact.ArtifactID, record.PromotionID)
	return nil
}

func (r replaySuiteRunner) getJSON(path string, dest any) error {
	response, err := r.client.Get(r.baseURL + path)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("GET %s expected 200, got %d: %s", path, response.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(response.Body).Decode(dest)
}

func (r replaySuiteRunner) postJSON(path string, payload any, dest any) error {
	return r.postJSONStatus(path, payload, http.StatusCreated, dest)
}

func (r replaySuiteRunner) postJSONStatus(path string, payload any, expectedStatus int, dest any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	response, err := r.client.Post(r.baseURL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != expectedStatus {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("POST %s expected %d, got %d: %s", path, expectedStatus, response.StatusCode, strings.TrimSpace(string(body)))
	}
	if dest != nil {
		return json.NewDecoder(response.Body).Decode(dest)
	}
	return nil
}

func mustNewUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func runValidate(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	kind := fs.String("kind", "", "object kind: taskpack, dri-binding, artifact, promotion-record")
	file := fs.String("file", "", "path to JSON file")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *kind == "" {
		return errors.New("--kind is required")
	}
	if *file == "" {
		return errors.New("--file is required")
	}

	data, err := os.ReadFile(*file)
	if err != nil {
		return err
	}
	if err := validateObject(*kind, data); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "ok %s %s\n", *kind, *file)
	return nil
}

func validateObject(kind string, data []byte) error {
	switch kind {
	case "taskpack":
		var payload spec.Taskpack
		if err := decodeStrict(data, &payload); err != nil {
			return err
		}
		return specvalidate.Taskpack(payload)
	case "dri-binding":
		var payload spec.DriBinding
		if err := decodeStrict(data, &payload); err != nil {
			return err
		}
		return specvalidate.DriBinding(payload)
	case "artifact":
		var payload spec.Artifact
		if err := decodeStrict(data, &payload); err != nil {
			return err
		}
		return specvalidate.Artifact(payload)
	case "promotion-record":
		var payload spec.PromotionRecord
		if err := decodeStrict(data, &payload); err != nil {
			return err
		}
		return specvalidate.PromotionRecord(payload)
	case "governance-policy":
		var payload spec.GovernancePolicy
		if err := decodeStrict(data, &payload); err != nil {
			return err
		}
		return specvalidate.GovernancePolicy(payload)
	case "approval-request":
		var payload spec.ApprovalRequest
		if err := decodeStrict(data, &payload); err != nil {
			return err
		}
		return specvalidate.ApprovalRequest(payload)
	case "promotion-gate":
		var payload spec.PromotionGate
		if err := decodeStrict(data, &payload); err != nil {
			return err
		}
		return specvalidate.PromotionGate(payload)
	case "commons-entry":
		var payload spec.CommonsEntry
		if err := decodeStrict(data, &payload); err != nil {
			return err
		}
		return specvalidate.CommonsEntry(payload)
	case "replay-bundle":
		var payload spec.ReplayBundle
		if err := decodeStrict(data, &payload); err != nil {
			return err
		}
		return specvalidate.ReplayBundle(payload)
	case "workspace-constitution":
		var payload spec.WorkspaceConstitution
		if err := decodeStrict(data, &payload); err != nil {
			return err
		}
		return specvalidate.WorkspaceConstitution(payload)
	case "context-pack":
		var payload spec.ContextPack
		if err := decodeStrict(data, &payload); err != nil {
			return err
		}
		return specvalidate.ContextPack(payload)
	case "preflight-decision":
		var payload spec.PreflightDecision
		if err := decodeStrict(data, &payload); err != nil {
			return err
		}
		return specvalidate.PreflightDecision(payload)
	default:
		return fmt.Errorf("unsupported kind %q", kind)
	}
}

func decodeStrict(data []byte, dest any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return err
	}
	var extra struct{}
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("file must contain a single JSON document")
	}
	return nil
}

func runConformance(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("conformance", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	baseURL := fs.String("base-url", "http://localhost:8080", "Guild-compatible API base URL")
	if err := fs.Parse(args); err != nil {
		return err
	}

	runner := conformanceRunner{
		baseURL: strings.TrimRight(*baseURL, "/"),
		client:  http.Client{Timeout: 5 * time.Second},
		stdout:  stdout,
	}
	return runner.run()
}

func runReplayExport(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("replay-export", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	baseURL := fs.String("base-url", "http://localhost:8080", "Guild-compatible API base URL")
	taskpackID := fs.String("taskpack-id", "", "taskpack UUID to export")
	file := fs.String("file", "", "optional output file; defaults to stdout")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *taskpackID == "" {
		return errors.New("--taskpack-id is required")
	}
	if !specvalidate.IsUUID(*taskpackID) {
		return errors.New("--taskpack-id must be a UUID")
	}

	client := http.Client{Timeout: 5 * time.Second}
	response, err := client.Get(strings.TrimRight(*baseURL, "/") + "/api/v1/replay/taskpacks/" + *taskpackID)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("expected 200, got %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}

	var bundle spec.ReplayBundle
	if err := json.NewDecoder(response.Body).Decode(&bundle); err != nil {
		return err
	}
	if bundle.SchemaVersion != "v1alpha1" {
		return fmt.Errorf("unsupported replay bundle schema_version %q", bundle.SchemaVersion)
	}
	if bundle.Taskpack.TaskpackID != *taskpackID {
		return fmt.Errorf("replay bundle taskpack_id mismatch: expected %q, got %q", *taskpackID, bundle.Taskpack.TaskpackID)
	}
	if err := specvalidate.ReplayBundle(bundle); err != nil {
		return fmt.Errorf("invalid replay bundle: %w", err)
	}

	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if *file == "" {
		_, err = stdout.Write(data)
		return err
	}
	return os.WriteFile(*file, data, 0o644)
}

type conformanceRunner struct {
	baseURL string
	client  http.Client
	stdout  io.Writer
}

func (r conformanceRunner) run() error {
	checks := []struct {
		name string
		fn   func() error
	}{
		{name: "health endpoint", fn: r.checkHealth},
		{name: "status endpoint", fn: r.checkStatus},
		{name: "taskpack listing", fn: r.checkTaskpacks},
		{name: "artifact creation", fn: r.checkArtifactCreate},
		{name: "orphan DRI rejection", fn: r.checkOrphanDRIRejected},
		{name: "unknown field rejection", fn: r.checkUnknownFieldRejected},
		{name: "malformed path id rejection", fn: r.checkMalformedPathIDRejected},
		{name: "unsupported item method rejection", fn: r.checkUnsupportedItemMethodRejected},
		{name: "replay bundle export", fn: r.checkReplayBundleExport},
	}

	for _, check := range checks {
		if err := check.fn(); err != nil {
			return fmt.Errorf("%s: %w", check.name, err)
		}
		fmt.Fprintf(r.stdout, "ok %s\n", check.name)
	}
	fmt.Fprintf(r.stdout, "conformance-ok %s\n", r.baseURL)
	return nil
}

func (r conformanceRunner) checkHealth() error {
	var payload map[string]any
	if err := r.getJSON("/healthz", &payload); err != nil {
		return err
	}
	if payload["status"] != "ok" {
		return fmt.Errorf("expected status ok, got %v", payload["status"])
	}
	return nil
}

func (r conformanceRunner) checkStatus() error {
	var payload map[string]any
	if err := r.getJSON("/api/v1/status", &payload); err != nil {
		return err
	}
	if payload["name"] != "guild" {
		return fmt.Errorf("expected name guild, got %v", payload["name"])
	}
	return nil
}

func (r conformanceRunner) checkTaskpacks() error {
	var payload struct {
		Items []spec.Taskpack `json:"items"`
	}
	if err := r.getJSON("/api/v1/taskpacks", &payload); err != nil {
		return err
	}
	if len(payload.Items) == 0 {
		return errors.New("expected at least one taskpack")
	}
	return specvalidate.Taskpack(payload.Items[0])
}

func (r conformanceRunner) checkArtifactCreate() error {
	artifact := spec.Artifact{
		SchemaVersion: "v1alpha1",
		ArtifactID:    "c4ce7f7b-6d4b-49c3-a6a4-632ce4317a9c",
		TaskpackID:    "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
		Kind:          "plan",
		Title:         "Conformance artifact",
		Producer: spec.ActorRef{
			ActorID:     "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
			ActorType:   "agent",
			DisplayName: "conformance-agent",
		},
		Storage: spec.ArtifactStorage{
			URI:      "s3://guild/conformance/artifact.md",
			MimeType: "text/markdown",
		},
		Version:   1,
		CreatedAt: "2026-04-24T12:00:00Z",
	}
	return r.postJSON("/api/v1/artifacts", artifact, http.StatusCreated, nil)
}

func (r conformanceRunner) checkOrphanDRIRejected() error {
	binding := spec.DriBinding{
		SchemaVersion: "v1alpha1",
		DriBindingID:  "9eb2d8f5-f756-402c-9872-6652f2418f53",
		TaskpackID:    "11111111-1111-1111-1111-111111111111",
		Owner: spec.ActorRef{
			ActorID:     "0a3657eb-2f37-4614-a8a7-9c6bd51714a8",
			ActorType:   "agent",
			DisplayName: "orphan-owner",
		},
		Status:    "assigned",
		CreatedAt: "2026-04-24T12:01:00Z",
	}
	return r.postJSON("/api/v1/dri-bindings", binding, http.StatusBadRequest, nil)
}

func (r conformanceRunner) checkUnknownFieldRejected() error {
	payload := map[string]any{
		"schema_version": "v1alpha1",
		"taskpack_id":    "d013e9c3-3fdc-4f72-a79f-3ca30d0fe111",
		"title":          "Invalid unknown field task",
		"objective":      "Prove unknown fields are rejected.",
		"task_type":      "analysis",
		"priority":       "medium",
		"requested_by": map[string]any{
			"actor_id":     "52a203e8-0a20-4f0b-9b3d-78631e00d6ab",
			"actor_type":   "human",
			"display_name": "Quentin",
		},
		"context_budget": map[string]any{
			"max_input_tokens":  4000,
			"max_output_tokens": 1500,
			"context_strategy":  "artifact_refs_first",
		},
		"permissions": map[string]any{
			"allow_network":        false,
			"allow_shell":          false,
			"allow_external_write": false,
			"approval_mode":        "ask",
		},
		"acceptance_criteria": []map[string]any{
			{
				"criterion_id": "unknown-field",
				"description":  "Reject unknown fields.",
				"required":     true,
			},
		},
		"created_at": "2026-04-24T12:02:00Z",
		"surprise":   true,
	}
	return r.postJSON("/api/v1/taskpacks", payload, http.StatusBadRequest, nil)
}

func (r conformanceRunner) checkMalformedPathIDRejected() error {
	return r.request("GET", "/api/v1/taskpacks/not-a-uuid", nil, http.StatusBadRequest, nil)
}

func (r conformanceRunner) checkUnsupportedItemMethodRejected() error {
	return r.request("POST", "/api/v1/taskpacks/4e1fe00c-6303-453c-8cb6-2c34f84896e4", nil, http.StatusMethodNotAllowed, nil)
}

func (r conformanceRunner) checkReplayBundleExport() error {
	var bundle spec.ReplayBundle
	if err := r.getJSON("/api/v1/replay/taskpacks/4e1fe00c-6303-453c-8cb6-2c34f84896e4", &bundle); err != nil {
		return err
	}
	if bundle.SchemaVersion != "v1alpha1" {
		return fmt.Errorf("expected schema version v1alpha1, got %q", bundle.SchemaVersion)
	}
	if bundle.Taskpack.TaskpackID != "4e1fe00c-6303-453c-8cb6-2c34f84896e4" {
		return fmt.Errorf("unexpected replay taskpack_id %q", bundle.Taskpack.TaskpackID)
	}
	if len(bundle.DriBindings) == 0 {
		return errors.New("expected at least one DRI binding in replay bundle")
	}
	if len(bundle.Artifacts) == 0 {
		return errors.New("expected at least one artifact in replay bundle")
	}
	return nil
}

func (r conformanceRunner) getJSON(path string, dest any) error {
	response, err := r.client.Get(r.baseURL + path)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200, got %d", response.StatusCode)
	}
	return json.NewDecoder(response.Body).Decode(dest)
}

func (r conformanceRunner) postJSON(path string, payload any, wantStatus int, dest any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return r.request("POST", path, data, wantStatus, dest)
}

func (r conformanceRunner) request(method, path string, data []byte, wantStatus int, dest any) error {
	var body io.Reader
	if data != nil {
		body = bytes.NewReader(data)
	}
	request, err := http.NewRequest(method, r.baseURL+path, body)
	if err != nil {
		return err
	}
	if data != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := r.client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != wantStatus {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return fmt.Errorf("expected %d, got %d: %s", wantStatus, response.StatusCode, strings.TrimSpace(string(body)))
	}
	if dest == nil {
		io.Copy(io.Discard, response.Body)
		return nil
	}
	return json.NewDecoder(response.Body).Decode(dest)
}
