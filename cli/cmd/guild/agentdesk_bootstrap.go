package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/lucid-fdn/guild/pkg/spec"
	specvalidate "github.com/lucid-fdn/guild/pkg/spec/validate"
	"gopkg.in/yaml.v3"
)

const defaultAgentDeskBootstrapVersion = "v0.1.0-alpha.3"

type agentDeskBootstrapReport struct {
	SchemaVersion string                         `json:"schema_version"`
	Ready         bool                           `json:"ready"`
	Files         []agentDeskBootstrapFileStatus `json:"files"`
	Labels        []agentDeskBootstrapLabel      `json:"labels"`
	NextSteps     []string                       `json:"next_steps"`
}

type agentDeskBootstrapFileStatus struct {
	Path   string `json:"path"`
	Status string `json:"status"`
}

type agentDeskBootstrapLabel struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type bootstrapGitHubLabel struct {
	Name        string
	Color       string
	Description string
}

func runAgentDeskBootstrap(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, agentDeskUsage)
		return errors.New("agentdesk bootstrap command is required")
	}
	switch args[0] {
	case "github":
		return runAgentDeskBootstrapGitHub(args[1:], stdout)
	default:
		return fmt.Errorf("unknown agentdesk bootstrap command %q", args[0])
	}
}

func runAgentDeskBootstrapGitHub(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("agentdesk bootstrap github", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	repo := fs.String("repo", envOr("GITHUB_REPOSITORY", ""), "GitHub owner/repo to configure")
	workspace := fs.String("workspace", "", "workspace name; defaults to current directory")
	mission := fs.String("mission", "Ship reliable product changes with accountable AI agents.", "workspace mission")
	version := fs.String("version", defaultAgentDeskBootstrapVersion, "Guild CLI version used by generated GitHub Actions workflow")
	force := fs.Bool("force", false, "overwrite existing AgentDesk files")
	skipLabels := fs.Bool("skip-labels", false, "skip creating GitHub labels through the API")
	if err := fs.Parse(args); err != nil {
		return err
	}
	*repo = strings.TrimSpace(*repo)
	if *repo == "" || !strings.Contains(*repo, "/") {
		return fmt.Errorf("--repo must be owner/repo, got %q", *repo)
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	if strings.TrimSpace(*workspace) == "" {
		*workspace = filepath.Base(root)
	}

	report := agentDeskBootstrapReport{
		SchemaVersion: "v1alpha1",
		Ready:         true,
		NextSteps: []string{
			"Create or label a GitHub issue with agent:ready.",
			"Run `GITHUB_TOKEN=$(gh auth token) guild agentdesk next --source github --repo " + *repo + "`.",
			"Run `guild agentdesk claim --id <mandate-id> --agent <agent-name>`.",
			"Attach proof, run `guild agentdesk verify --id <mandate-id>`, and open a PR.",
		},
	}
	addFile := func(path, status string) {
		report.Files = append(report.Files, agentDeskBootstrapFileStatus{Path: path, Status: status})
	}
	addLabel := func(name, status string) {
		report.Labels = append(report.Labels, agentDeskBootstrapLabel{Name: name, Status: status})
		if strings.HasPrefix(status, "failed") {
			report.Ready = false
		}
	}

	config := defaultAgentDeskConfig(*workspace, *mission)
	config.TaskSources = []spec.TaskSource{
		{Type: "local", Path: ".agentdesk/mandates"},
		{Type: "github_issues", Repo: *repo, Query: defaultGitHubIssueQuery},
	}
	if err := specvalidate.WorkspaceConstitution(config); err != nil {
		return err
	}
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	status, err := writeBootstrapFile("agentdesk.yaml", data, *force)
	if err != nil {
		return err
	}
	addFile("agentdesk.yaml", status)

	for _, dir := range []string{".agentdesk/mandates", ".agentdesk/proof", ".agentdesk/replay", ".agentdesk/handoffs", ".agentdesk/approvals", ".agentdesk/claims", ".agentdesk/closed"} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		addFile(dir, "ready")
	}

	for path, content := range map[string]string{
		".github/ISSUE_TEMPLATE/agent-ready.yml":    agentReadyIssueTemplate,
		".github/ISSUE_TEMPLATE/config.yml":         agentReadyIssueTemplateConfig,
		".github/workflows/agent-work-contract.yml": agentWorkContractWorkflow(*version),
		".github/workflows/agentdesk-doctor.yml":    agentDeskDoctorWorkflow(*version),
	} {
		status, err := writeBootstrapFile(path, []byte(content), *force)
		if err != nil {
			return err
		}
		addFile(path, status)
	}

	if *skipLabels {
		for _, label := range bootstrapGitHubLabels() {
			addLabel(label.Name, "skipped")
		}
	} else if os.Getenv("GITHUB_TOKEN") == "" {
		for _, label := range bootstrapGitHubLabels() {
			addLabel(label.Name, "skipped: GITHUB_TOKEN is not set")
		}
	} else {
		for _, label := range bootstrapGitHubLabels() {
			status, err := ensureGitHubLabel(*repo, label)
			if err != nil {
				addLabel(label.Name, "failed: "+err.Error())
				continue
			}
			addLabel(label.Name, status)
		}
	}

	return writeJSON(stdout, report)
}

func writeBootstrapFile(path string, content []byte, force bool) (string, error) {
	if _, err := os.Stat(path); err == nil && !force {
		return "exists", nil
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	if len(content) == 0 || content[len(content)-1] != '\n' {
		content = append(content, '\n')
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		return "", err
	}
	if force {
		return "written", nil
	}
	return "created", nil
}

func bootstrapGitHubLabels() []bootstrapGitHubLabel {
	return []bootstrapGitHubLabel{
		{Name: "agent:ready", Color: "0e8a16", Description: "Ready for an autonomous agent to claim through Guild AgentDesk."},
		{Name: "priority:p1", Color: "b60205", Description: "High priority agent mandate."},
		{Name: "priority:p2", Color: "d93f0b", Description: "Medium priority agent mandate."},
		{Name: "priority:p3", Color: "fbca04", Description: "Low priority agent mandate."},
	}
}

func ensureGitHubLabel(repo string, label bootstrapGitHubLabel) (string, error) {
	apiURL := strings.TrimRight(envOr("GITHUB_API_URL", "https://api.github.com"), "/")
	payload, err := json.Marshal(map[string]string{
		"name":        label.Name,
		"color":       label.Color,
		"description": label.Description,
	})
	if err != nil {
		return "", err
	}
	request, err := http.NewRequest(http.MethodPost, apiURL+"/repos/"+repo+"/labels", bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	setGitHubHeaders(request)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case http.StatusCreated:
		return "created", nil
	case http.StatusUnprocessableEntity:
		return "exists", nil
	default:
		body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
		return "", fmt.Errorf("GitHub label create failed with %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
}

const agentReadyIssueTemplate = `name: Agent-ready mandate
description: Create a task that an autonomous agent can claim through Guild AgentDesk.
title: "[agent] "
labels:
  - agent:ready
body:
  - type: markdown
    attributes:
      value: |
        Use this template when the task is ready for an agent to claim.
        A good mandate has a clear objective, bounded scope, acceptance criteria, and proof expectations.
  - type: textarea
    id: objective
    attributes:
      label: Objective
      description: What should the agent accomplish?
      placeholder: "Update the MCP setup docs so a new agent can connect in under 90 seconds."
    validations:
      required: true
  - type: textarea
    id: scope
    attributes:
      label: Allowed scope
      description: Which files or directories may the agent read/write?
      placeholder: "docs/**, src/**"
    validations:
      required: true
  - type: textarea
    id: acceptance
    attributes:
      label: Acceptance criteria
      description: What proof is required before the mandate is done?
      value: |
        - Tests pass or failure is explained.
        - Every modified file is listed.
        - A proof artifact is attached.
        - A reviewer handoff is created.
    validations:
      required: true
  - type: dropdown
    id: priority
    attributes:
      label: Priority
      options:
        - priority:p1
        - priority:p2
        - priority:p3
    validations:
      required: true
  - type: textarea
    id: notes
    attributes:
      label: Notes for the agent
      description: Add constraints, links, risks, or reviewer preferences.
`

const agentReadyIssueTemplateConfig = `blank_issues_enabled: true
`

func agentWorkContractWorkflow(version string) string {
	return fmt.Sprintf(`name: Agent Work Contract

on:
  pull_request:
    paths:
      - ".agentdesk/**"
      - "agentdesk.yaml"
      - ".github/workflows/agent-work-contract.yml"
  workflow_dispatch:
    inputs:
      mandate_id:
        description: AgentDesk mandate/taskpack UUID to verify
        required: true
        type: string
      replay_file:
        description: Replay bundle path to generate and reference
        required: false
        type: string
        default: .agentdesk/replay/replay.json

permissions:
  contents: read
  issues: write
  pull-requests: write

jobs:
  verify:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23.0"
          cache: false
      - name: Install Guild
        run: go install github.com/lucid-fdn/guild/cli/cmd/guild@%s
      - name: Resolve mandate
        id: mandate
        env:
          INPUT_MANDATE_ID: ${{ inputs.mandate_id }}
        run: |
          if [[ -n "${INPUT_MANDATE_ID}" ]]; then
            echo "mandate_id=${INPUT_MANDATE_ID}" >> "${GITHUB_OUTPUT}"
            exit 0
          fi
          mandate_file="$(find .agentdesk/mandates -name '*.json' -type f | sort | head -n 1 || true)"
          if [[ -z "${mandate_file}" ]]; then
            echo "No AgentDesk mandate found. Commit .agentdesk/mandates/<id>.json or pass mandate_id." >&2
            exit 1
          fi
          mandate_id="$(node -e 'const fs=require("fs"); const p=process.argv[1]; process.stdout.write(JSON.parse(fs.readFileSync(p, "utf8")).taskpack_id)' "${mandate_file}")"
          echo "mandate_id=${mandate_id}" >> "${GITHUB_OUTPUT}"
      - name: Export replay bundle
        env:
          MANDATE_ID: ${{ steps.mandate.outputs.mandate_id }}
          REPLAY_FILE: ${{ inputs.replay_file }}
        run: |
          if [[ -z "${REPLAY_FILE}" ]]; then
            REPLAY_FILE=".agentdesk/replay/replay.json"
          fi
          mkdir -p "$(dirname "${REPLAY_FILE}")"
          guild agentdesk replay export --id "${MANDATE_ID}" --file "${REPLAY_FILE}"
      - name: Verify agent work contract
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          MANDATE_ID: ${{ steps.mandate.outputs.mandate_id }}
          REPLAY_FILE: ${{ inputs.replay_file }}
        run: |
          if [[ -z "${REPLAY_FILE}" ]]; then
            REPLAY_FILE=".agentdesk/replay/replay.json"
          fi
          guild agentdesk verify \
            --id "${MANDATE_ID}" \
            --github-report \
            --replay-file "${REPLAY_FILE}"
`, version)
}

func agentDeskDoctorWorkflow(version string) string {
	return fmt.Sprintf(`name: AgentDesk Doctor

on:
  workflow_dispatch:
  pull_request:
    paths:
      - "agentdesk.yaml"
      - ".github/workflows/agentdesk-doctor.yml"

permissions:
  contents: read

jobs:
  doctor:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.23.0"
          cache: false
      - name: Install Guild
        run: go install github.com/lucid-fdn/guild/cli/cmd/guild@%s
      - name: Run AgentDesk doctor
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          GITHUB_REPOSITORY: ${{ github.repository }}
        run: guild agentdesk doctor --repo "${GITHUB_REPOSITORY}"
`, version)
}
