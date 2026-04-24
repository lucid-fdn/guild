package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateReplayBundleCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := run([]string{
		"validate",
		"--kind", "replay-bundle",
		"--file", "../../../spec/examples/replay-bundle.example.json",
	}, &stdout, &stderr)

	if err != nil {
		t.Fatalf("expected replay bundle to validate, got %v: %s", err, stderr.String())
	}
	if !strings.Contains(stdout.String(), "ok replay-bundle") {
		t.Fatalf("expected ok output, got %q", stdout.String())
	}
}

func TestValidateRejectsUnknownFields(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "invalid-taskpack.json")
	if err := os.WriteFile(file, []byte(`{
		"schema_version": "v1alpha1",
		"taskpack_id": "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
		"title": "Invalid task",
		"objective": "Prove strict decoding.",
		"task_type": "analysis",
		"priority": "medium",
		"requested_by": {
			"actor_id": "52a203e8-0a20-4f0b-9b3d-78631e00d6ab",
			"actor_type": "human",
			"display_name": "Quentin"
		},
		"context_budget": {
			"max_input_tokens": 4000,
			"max_output_tokens": 1500,
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
				"criterion_id": "strict-decode",
				"description": "Reject unknown fields.",
				"required": true
			}
		],
		"created_at": "2026-04-24T12:00:00Z",
		"unexpected": true
	}`), 0o644); err != nil {
		t.Fatal(err)
	}

	err := run([]string{"validate", "--kind", "taskpack", "--file", file}, ioDiscard(), ioDiscard())

	if err == nil {
		t.Fatal("expected validation error")
	}
	if !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}

func TestReplayExportWritesValidatedBundle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/replay/taskpacks/4e1fe00c-6303-453c-8cb6-2c34f84896e4" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "../../../spec/examples/replay-bundle.example.json")
	}))
	defer server.Close()

	outputPath := filepath.Join(t.TempDir(), "replay.json")
	err := run([]string{
		"replay-export",
		"--base-url", server.URL,
		"--taskpack-id", "4e1fe00c-6303-453c-8cb6-2c34f84896e4",
		"--file", outputPath,
	}, ioDiscard(), ioDiscard())

	if err != nil {
		t.Fatalf("expected replay export to succeed, got %v", err)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["schema_version"] != "v1alpha1" {
		t.Fatalf("expected schema_version v1alpha1, got %v", payload["schema_version"])
	}
}

func TestReplayExportRejectsTaskpackMismatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "../../../spec/examples/replay-bundle.example.json")
	}))
	defer server.Close()

	err := run([]string{
		"replay-export",
		"--base-url", server.URL,
		"--taskpack-id", "11111111-1111-1111-1111-111111111111",
	}, ioDiscard(), ioDiscard())

	if err == nil {
		t.Fatal("expected taskpack mismatch error")
	}
	if !strings.Contains(err.Error(), "taskpack_id mismatch") {
		t.Fatalf("expected mismatch error, got %v", err)
	}
}

func TestConformanceCommandRunsAgainstCompatibleServer(t *testing.T) {
	server := newConformanceServer(t)
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{"conformance", "--base-url", server.URL}, &stdout, ioDiscard())

	if err != nil {
		t.Fatalf("expected conformance to pass, got %v", err)
	}
	if !strings.Contains(stdout.String(), "conformance-ok") {
		t.Fatalf("expected conformance-ok output, got %q", stdout.String())
	}
}

func TestReplaySuiteCreatesBenchmarkCandidateAndPromotion(t *testing.T) {
	var artifactPosts int
	var promotionPosts int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/replay/taskpacks/4e1fe00c-6303-453c-8cb6-2c34f84896e4":
			w.Header().Set("Content-Type", "application/json")
			http.ServeFile(w, r, "../../../spec/examples/replay-bundle.example.json")
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/artifacts":
			var artifact map[string]any
			if err := json.NewDecoder(r.Body).Decode(&artifact); err != nil {
				t.Fatal(err)
			}
			if artifact["kind"] != "benchmark_result" && artifact["kind"] != "skill_candidate" {
				t.Fatalf("unexpected artifact kind %v", artifact["kind"])
			}
			artifactPosts++
			w.WriteHeader(http.StatusCreated)
			writeTestJSON(t, w, artifact)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/promotion-records":
			var record map[string]any
			if err := json.NewDecoder(r.Body).Decode(&record); err != nil {
				t.Fatal(err)
			}
			if record["decision"] != "needs_human_review" {
				t.Fatalf("expected needs_human_review, got %v", record["decision"])
			}
			promotionPosts++
			w.WriteHeader(http.StatusCreated)
			writeTestJSON(t, w, record)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{
		"replay-suite",
		"--base-url", server.URL,
		"--suite", "../../../examples/replay-suite.example.json",
	}, &stdout, ioDiscard())

	if err != nil {
		t.Fatalf("expected replay suite to succeed, got %v", err)
	}
	if artifactPosts != 2 {
		t.Fatalf("expected 2 artifact posts, got %d", artifactPosts)
	}
	if promotionPosts != 1 {
		t.Fatalf("expected 1 promotion post, got %d", promotionPosts)
	}
	if !strings.Contains(stdout.String(), "replay-suite-ok") {
		t.Fatalf("expected replay-suite-ok output, got %q", stdout.String())
	}
}

func TestEvalSubmitQueuesAndRunsEvaluationJob(t *testing.T) {
	var runCalled bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/evaluation-jobs":
			w.WriteHeader(http.StatusAccepted)
			writeTestJSON(t, w, map[string]any{
				"evaluation_job_id": "2b63be41-d0a1-4d26-acb1-cab92cf3301e",
				"status":            "queued",
			})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/evaluation-jobs/2b63be41-d0a1-4d26-acb1-cab92cf3301e/run":
			runCalled = true
			writeTestJSON(t, w, map[string]any{
				"evaluation_job_id": "2b63be41-d0a1-4d26-acb1-cab92cf3301e",
				"status":            "succeeded",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	var stdout bytes.Buffer
	err := run([]string{
		"eval-submit",
		"--base-url", server.URL,
		"--suite", "../../../examples/replay-suite.example.json",
		"--wait",
	}, &stdout, ioDiscard())

	if err != nil {
		t.Fatalf("expected eval submit to succeed, got %v", err)
	}
	if !runCalled {
		t.Fatal("expected run endpoint to be called")
	}
	if !strings.Contains(stdout.String(), "eval-job-succeeded") {
		t.Fatalf("expected eval-job-succeeded output, got %q", stdout.String())
	}
}

func TestAgentDeskLocalWorkflow(t *testing.T) {
	dir := t.TempDir()
	restore := chdir(t, dir)
	defer restore()

	var stdout bytes.Buffer
	if err := run([]string{"agentdesk", "init", "--workspace", "demo"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "agentdesk.yaml")); err != nil {
		t.Fatal(err)
	}

	stdout.Reset()
	if err := run([]string{"agentdesk", "mandate", "create", "Fix failing auth tests", "--writable", "src/auth/**,tests/auth/**", "--priority", "high"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("mandate create failed: %v", err)
	}
	mandateID := strings.TrimSpace(strings.TrimPrefix(stdout.String(), "mandate-created "))
	if mandateID == "" {
		t.Fatalf("expected mandate id, got %q", stdout.String())
	}

	stdout.Reset()
	if err := run([]string{"agentdesk", "next"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("next failed: %v", err)
	}
	var mandate map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &mandate); err != nil {
		t.Fatal(err)
	}
	if mandate["taskpack_id"] != mandateID {
		t.Fatalf("expected next mandate %s, got %v", mandateID, mandate["taskpack_id"])
	}

	stdout.Reset()
	if err := run([]string{"agentdesk", "preflight", "--id", mandateID, "--action", "write", "--path", "src/auth/login.ts"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("preflight allow failed: %v", err)
	}
	var allowDecision map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &allowDecision); err != nil {
		t.Fatal(err)
	}
	if allowDecision["decision"] != "allow" {
		t.Fatalf("expected allow decision, got %v", allowDecision["decision"])
	}

	stdout.Reset()
	if err := run([]string{"agentdesk", "preflight", "--id", mandateID, "--action", "write", "--path", ".env"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("preflight approval failed: %v", err)
	}
	var approvalDecision map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &approvalDecision); err != nil {
		t.Fatal(err)
	}
	if approvalDecision["decision"] != "needs_approval" {
		t.Fatalf("expected needs_approval, got %v", approvalDecision["decision"])
	}

	stdout.Reset()
	if err := run([]string{"agentdesk", "context", "compile", "--id", mandateID, "--role", "coder"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("context compile failed: %v", err)
	}
	var contextPack map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &contextPack); err != nil {
		t.Fatal(err)
	}
	if contextPack["mandate_id"] != mandateID {
		t.Fatalf("expected context for %s, got %v", mandateID, contextPack["mandate_id"])
	}

	proofFile := filepath.Join(dir, "test-results.xml")
	if err := os.WriteFile(proofFile, []byte("<testsuite failures=\"0\"></testsuite>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if err := run([]string{"agentdesk", "proof", "add", "--id", mandateID, "--kind", "test_report", "--path", proofFile}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("proof add failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "proof-added") {
		t.Fatalf("expected proof-added, got %q", stdout.String())
	}

	changedFiles := filepath.Join(dir, "changed-files.json")
	if err := os.WriteFile(changedFiles, []byte(`["src/auth/login.ts"]`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	if err := run([]string{"agentdesk", "proof", "add", "--id", mandateID, "--kind", "changed_files", "--path", changedFiles}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("changed files proof add failed: %v", err)
	}

	stdout.Reset()
	if err := run([]string{"agentdesk", "approval", "request", "--id", mandateID, "--reason", "Need reviewer consent for auth fixture update"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("approval request failed: %v", err)
	}
	approvalID := strings.TrimSpace(strings.TrimPrefix(stdout.String(), "approval-requested "))
	stdout.Reset()
	if err := run([]string{"agentdesk", "verify", "--id", mandateID}, &stdout, ioDiscard()); err == nil {
		t.Fatal("expected verify to fail with pending approval and missing handoff")
	}
	stdout.Reset()
	if err := run([]string{"agentdesk", "approval", "resolve", "--approval-id", approvalID, "--decision", "approved"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("approval resolve failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "approval-approved") {
		t.Fatalf("expected approval-approved, got %q", stdout.String())
	}

	stdout.Reset()
	if err := run([]string{"agentdesk", "handoff", "create", "--id", mandateID, "--to", "reviewer", "--summary", "Auth tests are green and ready for review."}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("handoff create failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "handoff-created") {
		t.Fatalf("expected handoff-created, got %q", stdout.String())
	}

	stdout.Reset()
	if err := run([]string{"agentdesk", "verify", "--id", mandateID}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("verify failed: %v output=%s", err, stdout.String())
	}
	var verify map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &verify); err != nil {
		t.Fatal(err)
	}
	if verify["ready"] != true {
		t.Fatalf("expected verify ready, got %v", verify["ready"])
	}

	stdout.Reset()
	if err := run([]string{"agentdesk", "close", "--id", mandateID}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("close failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "mandate-closed") {
		t.Fatalf("expected mandate-closed, got %q", stdout.String())
	}

	stdout.Reset()
	if err := run([]string{"agentdesk", "replay", "export", "--id", mandateID}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("replay export failed: %v", err)
	}
	var replay map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &replay); err != nil {
		t.Fatal(err)
	}
	if replay["root_taskpack_id"] != mandateID {
		t.Fatalf("expected replay root %s, got %v", mandateID, replay["root_taskpack_id"])
	}
}

func TestAgentDeskCloseRequiresProof(t *testing.T) {
	dir := t.TempDir()
	restore := chdir(t, dir)
	defer restore()

	if err := run([]string{"agentdesk", "init"}, ioDiscard(), ioDiscard()); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	var stdout bytes.Buffer
	if err := run([]string{"agentdesk", "mandate", "create", "Write docs"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("mandate create failed: %v", err)
	}
	mandateID := strings.TrimSpace(strings.TrimPrefix(stdout.String(), "mandate-created "))
	err := run([]string{"agentdesk", "close", "--id", mandateID}, ioDiscard(), ioDiscard())
	if err == nil {
		t.Fatal("expected close to require proof")
	}
	if !strings.Contains(err.Error(), "without at least one proof") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgentDeskClaimSkipsClaimedMandates(t *testing.T) {
	dir := t.TempDir()
	restore := chdir(t, dir)
	defer restore()

	if err := run([]string{"agentdesk", "init"}, ioDiscard(), ioDiscard()); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	var stdout bytes.Buffer
	if err := run([]string{"agentdesk", "mandate", "create", "First task", "--priority", "high"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("first create failed: %v", err)
	}
	firstID := strings.TrimSpace(strings.TrimPrefix(stdout.String(), "mandate-created "))
	stdout.Reset()
	if err := run([]string{"agentdesk", "mandate", "create", "Second task", "--priority", "medium"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("second create failed: %v", err)
	}
	secondID := strings.TrimSpace(strings.TrimPrefix(stdout.String(), "mandate-created "))
	stdout.Reset()
	if err := run([]string{"agentdesk", "claim", "--id", firstID, "--agent", "codex", "--ttl-minutes", "30"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("claim failed: %v", err)
	}
	var claim map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &claim); err != nil {
		t.Fatal(err)
	}
	if claim["agent"] != "codex" {
		t.Fatalf("expected codex claim, got %v", claim["agent"])
	}
	err := run([]string{"agentdesk", "claim", "--id", firstID, "--agent", "other"}, ioDiscard(), ioDiscard())
	if err == nil {
		t.Fatal("expected duplicate claim to fail")
	}
	stdout.Reset()
	if err := run([]string{"agentdesk", "next"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("next failed: %v", err)
	}
	var mandate map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &mandate); err != nil {
		t.Fatal(err)
	}
	if mandate["taskpack_id"] != secondID {
		t.Fatalf("expected unclaimed second mandate %s, got %v", secondID, mandate["taskpack_id"])
	}
}

func TestAgentDeskNextFromGitHubIssuesCreatesMandate(t *testing.T) {
	dir := t.TempDir()
	restore := chdir(t, dir)
	defer restore()

	var requestedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestedPath = r.URL.Path
		if r.URL.Path != "/search/issues" {
			http.NotFound(w, r)
			return
		}
		if !strings.Contains(r.URL.Query().Get("q"), "repo:lucid-fdn/app") {
			t.Fatalf("expected repo query, got %q", r.URL.Query().Get("q"))
		}
		writeTestJSON(t, w, map[string]any{
			"items": []map[string]any{
				{
					"number":     184,
					"title":      "Fix failing auth tests",
					"body":       "Auth tests fail in CI after fixture drift.",
					"html_url":   "https://github.com/lucid-fdn/app/issues/184",
					"state":      "open",
					"created_at": "2026-04-24T12:00:00Z",
					"user": map[string]any{
						"login": "quentin",
					},
					"labels": []map[string]any{
						{"name": "agent:ready"},
						{"name": "priority:p1"},
						{"name": "scope:auth"},
					},
				},
			},
		})
	}))
	defer server.Close()

	t.Setenv("GITHUB_API_URL", server.URL)
	if err := run([]string{"agentdesk", "init", "--workspace", "demo"}, ioDiscard(), ioDiscard()); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	var stdout bytes.Buffer
	if err := run([]string{"agentdesk", "next", "--source", "github", "--repo", "lucid-fdn/app", "--query", "label:agent:ready state:open"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("github next failed: %v", err)
	}
	if requestedPath != "/search/issues" {
		t.Fatalf("expected GitHub search request, got %q", requestedPath)
	}
	var mandate map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &mandate); err != nil {
		t.Fatal(err)
	}
	if mandate["title"] != "Fix failing auth tests" {
		t.Fatalf("unexpected title %v", mandate["title"])
	}
	if mandate["priority"] != "high" {
		t.Fatalf("expected priority high, got %v", mandate["priority"])
	}
	permissions := mandate["permissions"].(map[string]any)
	scopes := permissions["scopes"].([]any)
	if !containsAny(scopes, "auth/**") {
		t.Fatalf("expected scope auth/**, got %v", scopes)
	}
}

func TestAgentDeskVerifyPublishesGitHubReport(t *testing.T) {
	dir := t.TempDir()
	restore := chdir(t, dir)
	defer restore()

	var commentBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/repos/lucid-fdn/app/issues/12/comments" {
			http.NotFound(w, r)
			return
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatal(err)
		}
		commentBody = payload["body"]
		w.WriteHeader(http.StatusCreated)
		writeTestJSON(t, w, map[string]any{"id": 1})
	}))
	defer server.Close()

	t.Setenv("GITHUB_API_URL", server.URL)
	t.Setenv("GITHUB_TOKEN", "test-token")
	t.Setenv("GITHUB_REPOSITORY", "lucid-fdn/app")
	t.Setenv("GITHUB_PR_NUMBER", "12")
	summaryPath := filepath.Join(dir, "summary.md")
	t.Setenv("GITHUB_STEP_SUMMARY", summaryPath)

	if err := run([]string{"agentdesk", "init", "--workspace", "demo"}, ioDiscard(), ioDiscard()); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	var stdout bytes.Buffer
	if err := run([]string{"agentdesk", "mandate", "create", "Fix failing auth tests", "--writable", "src/auth/**,tests/auth/**"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("mandate create failed: %v", err)
	}
	mandateID := strings.TrimSpace(strings.TrimPrefix(stdout.String(), "mandate-created "))
	testReport := filepath.Join(dir, "test-results.xml")
	changedFiles := filepath.Join(dir, "changed-files.json")
	if err := os.WriteFile(testReport, []byte("<testsuite failures=\"0\"></testsuite>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(changedFiles, []byte(`["src/auth/login.ts"]`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"agentdesk", "proof", "add", "--id", mandateID, "--kind", "test_report", "--path", testReport}, ioDiscard(), ioDiscard()); err != nil {
		t.Fatalf("proof add failed: %v", err)
	}
	if err := run([]string{"agentdesk", "proof", "add", "--id", mandateID, "--kind", "changed_files", "--path", changedFiles}, ioDiscard(), ioDiscard()); err != nil {
		t.Fatalf("changed files proof failed: %v", err)
	}
	if err := run([]string{"agentdesk", "handoff", "create", "--id", mandateID, "--to", "reviewer", "--summary", "Ready for review."}, ioDiscard(), ioDiscard()); err != nil {
		t.Fatalf("handoff failed: %v", err)
	}
	stdout.Reset()
	if err := run([]string{"agentdesk", "verify", "--id", mandateID, "--github-report", "--replay-file", ".agentdesk/replay/replay.json"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("verify failed: %v output=%s", err, stdout.String())
	}
	summary, err := os.ReadFile(summaryPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, haystack := range []string{string(summary), commentBody} {
		if !strings.Contains(haystack, "Agent Work Contract: passed") {
			t.Fatalf("expected passed report, got %q", haystack)
		}
		if !strings.Contains(haystack, "Proof: test_report") {
			t.Fatalf("expected proof summary, got %q", haystack)
		}
		if !strings.Contains(haystack, "Approvals: resolved") {
			t.Fatalf("expected approvals resolved, got %q", haystack)
		}
		if !strings.Contains(haystack, "Replay: attached") {
			t.Fatalf("expected replay attached, got %q", haystack)
		}
	}
}

func TestAgentDeskDoctorReportsReady(t *testing.T) {
	dir := t.TempDir()
	restore := chdir(t, dir)
	defer restore()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/lucid-fdn/guild/labels" {
			http.NotFound(w, r)
			return
		}
		writeTestJSON(t, w, []map[string]string{{"name": "agent:ready"}})
	}))
	defer server.Close()
	t.Setenv("GITHUB_API_URL", server.URL)
	t.Setenv("GITHUB_TOKEN", "test-token")

	if err := run([]string{"agentdesk", "init", "--workspace", "demo"}, ioDiscard(), ioDiscard()); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	var stdout bytes.Buffer
	if err := run([]string{"agentdesk", "mandate", "create", "Doctor task"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("mandate create failed: %v", err)
	}
	mandateID := strings.TrimSpace(strings.TrimPrefix(stdout.String(), "mandate-created "))
	testReport := filepath.Join(dir, "test-results.xml")
	changedFiles := filepath.Join(dir, "changed-files.json")
	if err := os.WriteFile(testReport, []byte("<testsuite failures=\"0\"></testsuite>\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(changedFiles, []byte("[\"docs/demo.md\"]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := run([]string{"agentdesk", "proof", "add", "--id", mandateID, "--kind", "test_report", "--path", testReport}, ioDiscard(), ioDiscard()); err != nil {
		t.Fatalf("test proof failed: %v", err)
	}
	if err := run([]string{"agentdesk", "proof", "add", "--id", mandateID, "--kind", "changed_files", "--path", changedFiles}, ioDiscard(), ioDiscard()); err != nil {
		t.Fatalf("changed files proof failed: %v", err)
	}
	if err := run([]string{"agentdesk", "handoff", "create", "--id", mandateID, "--to", "reviewer", "--summary", "Ready."}, ioDiscard(), ioDiscard()); err != nil {
		t.Fatalf("handoff failed: %v", err)
	}
	stdout.Reset()
	if err := run([]string{"agentdesk", "doctor", "--id", mandateID, "--repo", "lucid-fdn/guild"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("doctor failed: %v output=%s", err, stdout.String())
	}
	var report agentDeskDoctorReport
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if !report.Ready {
		t.Fatalf("expected ready doctor report, got %#v", report)
	}
	if !doctorHasCheck(report, "github_labels", "pass") {
		t.Fatalf("expected github_labels pass, got %#v", report.Checks)
	}
	if !doctorHasCheck(report, "proof_readiness", "pass") {
		t.Fatalf("expected proof_readiness pass, got %#v", report.Checks)
	}
}

func TestMCPServerToolsAndLocalFlow(t *testing.T) {
	dir := t.TempDir()
	restore := chdir(t, dir)
	defer restore()

	if err := run([]string{"agentdesk", "init", "--workspace", "mcp"}, ioDiscard(), ioDiscard()); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	var stdout bytes.Buffer
	if err := run([]string{"agentdesk", "mandate", "create", "MCP task"}, &stdout, ioDiscard()); err != nil {
		t.Fatalf("mandate create failed: %v", err)
	}

	list := handleMCPRequest(mcpRequest{JSONRPC: "2.0", ID: float64(1), Method: "tools/list"})
	data, err := json.Marshal(list.Result)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "guild_claim_mandate") {
		t.Fatalf("expected claim tool, got %s", string(data))
	}

	next := handleMCPRequest(mcpRequest{
		JSONRPC: "2.0",
		ID:      float64(2),
		Method:  "tools/call",
		Params: map[string]any{
			"name":      "guild_get_next_mandate",
			"arguments": map[string]any{},
		},
	})
	result := next.Result.(mcpToolResult)
	if result.IsError {
		t.Fatalf("expected get_next success, got %#v", result)
	}
	var mandate map[string]any
	if err := json.Unmarshal([]byte(result.Content[0].Text), &mandate); err != nil {
		t.Fatal(err)
	}
	mandateID := mandate["taskpack_id"].(string)

	claim := handleMCPRequest(mcpRequest{
		JSONRPC: "2.0",
		ID:      float64(3),
		Method:  "tools/call",
		Params: map[string]any{
			"name": "guild_claim_mandate",
			"arguments": map[string]any{
				"taskpack_id": mandateID,
				"agent":       "mcp-test",
				"ttl_minutes": float64(10),
			},
		},
	})
	claimResult := claim.Result.(mcpToolResult)
	if claimResult.IsError {
		t.Fatalf("expected claim success, got %#v", claimResult)
	}
	if !strings.Contains(claimResult.Content[0].Text, "mcp-test") {
		t.Fatalf("expected mcp-test claim, got %s", claimResult.Content[0].Text)
	}
}

func doctorHasCheck(report agentDeskDoctorReport, name, status string) bool {
	for _, check := range report.Checks {
		if check.Name == name && check.Status == status {
			return true
		}
	}
	return false
}

func newConformanceServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(t, w, map[string]any{"status": "ok", "service": "guildd"})
	})
	mux.HandleFunc("/api/v1/status", func(w http.ResponseWriter, r *http.Request) {
		writeTestJSON(t, w, map[string]any{"name": "guild"})
	})
	mux.HandleFunc("/api/v1/taskpacks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeTestFixtureList(t, w, "../../../spec/examples/taskpack.example.json")
		case http.MethodPost:
			http.Error(w, `{"error":"bad request"}`, http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/artifacts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	})
	mux.HandleFunc("/api/v1/dri-bindings", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		http.Error(w, `{"error":"taskpack_id does not exist"}`, http.StatusBadRequest)
	})
	mux.HandleFunc("/api/v1/taskpacks/not-a-uuid", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"taskpack_id must be a UUID"}`, http.StatusBadRequest)
	})
	mux.HandleFunc("/api/v1/taskpacks/4e1fe00c-6303-453c-8cb6-2c34f84896e4", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/api/v1/replay/taskpacks/4e1fe00c-6303-453c-8cb6-2c34f84896e4", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "../../../spec/examples/replay-bundle.example.json")
	})
	return httptest.NewServer(mux)
}

func writeTestFixtureList(t *testing.T, w http.ResponseWriter, fixturePath string) {
	t.Helper()
	data, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"items":[`))
	_, _ = w.Write(data)
	_, _ = w.Write([]byte(`]}`))
}

func writeTestJSON(t *testing.T, w http.ResponseWriter, payload any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Fatal(err)
	}
}

func ioDiscard() *bytes.Buffer {
	return &bytes.Buffer{}
}

func chdir(t *testing.T, dir string) func() {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	return func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatal(err)
		}
	}
}

func containsAny(values []any, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
