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
