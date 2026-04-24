package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/lucid-fdn/guild/pkg/spec"
	specvalidate "github.com/lucid-fdn/guild/pkg/spec/validate"
	"gopkg.in/yaml.v3"
)

const agentDeskUsage = `Guild AgentDesk

Usage:
  guild agentdesk init
  guild agentdesk mandate create "Fix failing auth tests"
  guild agentdesk mandate show --id <uuid>
  guild agentdesk next [--source local|github] [--repo owner/repo] [--include-claimed]
  guild agentdesk claim --id <uuid> [--agent codex]
  guild agentdesk preflight --id <uuid> --action write --path src/auth/login.ts
  guild agentdesk context compile --id <uuid> --role coder [--budget 12000]
  guild agentdesk approval request --id <uuid> --reason "Need to edit auth policy"
  guild agentdesk approval resolve --approval-id <uuid> --decision approved
  guild agentdesk proof add --id <uuid> --kind test_report --path test-results.xml [--summary "..."]
  guild agentdesk handoff create --id <uuid> --to reviewer --summary "Ready for review"
  guild agentdesk verify --id <uuid>
  guild agentdesk close --id <uuid>
  guild agentdesk replay export --id <uuid> [--file replay.json]
`

func runAgentDesk(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, agentDeskUsage)
		return errors.New("agentdesk command is required")
	}
	switch args[0] {
	case "init":
		return runAgentDeskInit(args[1:], stdout)
	case "mandate":
		return runAgentDeskMandate(args[1:], stdout, stderr)
	case "next":
		return runAgentDeskNext(args[1:], stdout)
	case "claim":
		return runAgentDeskClaim(args[1:], stdout)
	case "preflight":
		return runAgentDeskPreflight(args[1:], stdout)
	case "context":
		return runAgentDeskContext(args[1:], stdout, stderr)
	case "approval":
		return runAgentDeskApproval(args[1:], stdout, stderr)
	case "proof":
		return runAgentDeskProof(args[1:], stdout, stderr)
	case "handoff":
		return runAgentDeskHandoff(args[1:], stdout, stderr)
	case "verify":
		return runAgentDeskVerify(args[1:], stdout)
	case "close":
		return runAgentDeskClose(args[1:], stdout)
	case "replay":
		return runAgentDeskReplay(args[1:], stdout, stderr)
	case "help", "-h", "--help":
		fmt.Fprint(stdout, agentDeskUsage)
		return nil
	default:
		fmt.Fprint(stderr, agentDeskUsage)
		return fmt.Errorf("unknown agentdesk command %q", args[0])
	}
}

func runAgentDeskInit(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("agentdesk init", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	workspace := fs.String("workspace", "", "workspace name; defaults to current directory")
	mission := fs.String("mission", "Ship reliable product changes with accountable AI agents.", "workspace mission")
	if err := fs.Parse(args); err != nil {
		return err
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	if *workspace == "" {
		*workspace = filepath.Base(root)
	}
	config := defaultAgentDeskConfig(*workspace, *mission)
	if _, err := os.Stat("agentdesk.yaml"); err == nil {
		return errors.New("agentdesk.yaml already exists")
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}
	if err := os.WriteFile("agentdesk.yaml", data, 0o644); err != nil {
		return err
	}
	for _, dir := range []string{".agentdesk/mandates", ".agentdesk/proof", ".agentdesk/replay", ".agentdesk/handoffs", ".agentdesk/approvals", ".agentdesk/claims", ".agentdesk/closed"} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	fmt.Fprintln(stdout, "agentdesk-init-ok agentdesk.yaml")
	return nil
}

func runAgentDeskMandate(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, agentDeskUsage)
		return errors.New("agentdesk mandate command is required")
	}
	switch args[0] {
	case "create":
		return runAgentDeskMandateCreate(args[1:], stdout)
	case "show":
		return runAgentDeskMandateShow(args[1:], stdout)
	default:
		return fmt.Errorf("unknown agentdesk mandate command %q", args[0])
	}
}

func runAgentDeskMandateCreate(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("agentdesk mandate create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	objective := fs.String("objective", "", "mandate objective; defaults to title")
	priority := fs.String("priority", "medium", "priority: low, medium, high, critical")
	role := fs.String("role", "builder", "role hint")
	writable := fs.String("writable", "", "comma-separated writable scope override")
	if err := fs.Parse(reorderInterspersedFlags(args, map[string]bool{
		"objective": true,
		"priority":  true,
		"role":      true,
		"writable":  true,
	})); err != nil {
		return err
	}
	title := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if title == "" {
		return errors.New("mandate title is required")
	}
	if *objective == "" {
		*objective = title
	}
	config, err := loadAgentDeskConfig()
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	scopes := append([]string{}, config.Scope.Writable...)
	if strings.TrimSpace(*writable) != "" {
		scopes = splitCSV(*writable)
	}
	mandate := spec.Taskpack{
		SchemaVersion: "v1alpha1",
		TaskpackID:    mustNewUUID(),
		Title:         title,
		Objective:     *objective,
		TaskType:      "implementation",
		Priority:      *priority,
		RequestedBy: spec.ActorRef{
			ActorID:     mustNewUUID(),
			ActorType:   "human",
			DisplayName: "agentdesk-cli",
		},
		RoleHint: "builder",
		Labels:   []string{"agentdesk", "open"},
		ContextBudget: spec.ContextBudget{
			MaxInputTokens:  max(config.Defaults.ContextBudgetTokens, 256),
			MaxOutputTokens: 1024,
			ContextStrategy: "artifact_refs_first",
		},
		Permissions: spec.Permissions{
			AllowNetwork:       false,
			AllowShell:         true,
			AllowExternalWrite: false,
			ApprovalMode:       "ask",
			Scopes:             scopes,
		},
		Acceptance: defaultAcceptance(config.SuccessCriteria),
		CreatedAt:  now,
	}
	if *role != "" {
		mandate.RoleHint = normalizeRoleHint(*role)
	}
	if err := specvalidate.Taskpack(mandate); err != nil {
		return err
	}
	if err := writeAgentDeskJSON(mandatePath(mandate.TaskpackID), mandate); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "mandate-created %s\n", mandate.TaskpackID)
	return nil
}

func runAgentDeskMandateShow(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("agentdesk mandate show", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.String("id", "", "mandate/taskpack UUID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	mandate, err := loadMandate(*id)
	if err != nil {
		return err
	}
	return writeJSON(stdout, mandate)
}

func runAgentDeskNext(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("agentdesk next", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	source := fs.String("source", "local", "task source: local or github")
	repo := fs.String("repo", "", "GitHub owner/repo override")
	query := fs.String("query", "", "GitHub issue search query override")
	includeClaimed := fs.Bool("include-claimed", false, "include mandates with active claim locks")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *source == "github" {
		if err := syncGitHubIssueMandates(*repo, *query); err != nil {
			return err
		}
	} else if *source != "local" {
		return fmt.Errorf("--source must be local or github, got %q", *source)
	}
	mandates, err := loadOpenMandates(*includeClaimed)
	if err != nil {
		return err
	}
	if len(mandates) == 0 {
		return errors.New("no open mandates found")
	}
	sort.Slice(mandates, func(i, j int) bool {
		if priorityRank(mandates[i].Priority) == priorityRank(mandates[j].Priority) {
			return mandates[i].CreatedAt < mandates[j].CreatedAt
		}
		return priorityRank(mandates[i].Priority) > priorityRank(mandates[j].Priority)
	})
	return writeJSON(stdout, mandates[0])
}

func runAgentDeskClaim(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("agentdesk claim", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.String("id", "", "mandate/taskpack UUID")
	agent := fs.String("agent", envOr("AGENT_NAME", "agentdesk-agent"), "claiming agent name")
	ttlMinutes := fs.Int("ttl-minutes", 120, "claim lease TTL in minutes")
	force := fs.Bool("force", false, "replace an existing active claim")
	if err := fs.Parse(args); err != nil {
		return err
	}
	mandate, err := loadMandate(*id)
	if err != nil {
		return err
	}
	if *ttlMinutes < 1 {
		return errors.New("--ttl-minutes must be >= 1")
	}
	existing, active, err := loadClaim(mandate.TaskpackID)
	if err != nil {
		return err
	}
	if active && !*force {
		return fmt.Errorf("mandate already claimed by %s until %s", existing.Agent, existing.ExpiresAt)
	}
	now := time.Now().UTC()
	claim := agentDeskClaim{
		SchemaVersion: "v1alpha1",
		MandateID:     mandate.TaskpackID,
		Agent:         *agent,
		ClaimedAt:     now.Format(time.RFC3339),
		ExpiresAt:     now.Add(time.Duration(*ttlMinutes) * time.Minute).Format(time.RFC3339),
	}
	if err := writeAgentDeskJSON(claimPath(mandate.TaskpackID), claim); err != nil {
		return err
	}
	return writeJSON(stdout, claim)
}

func runAgentDeskPreflight(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("agentdesk preflight", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.String("id", "", "mandate/taskpack UUID")
	action := fs.String("action", "", "action: read, write, run, network, secret, git_push, dependency_install, prod_access")
	pathValue := fs.String("path", "", "path to check")
	command := fs.String("command", "", "command to check")
	if err := fs.Parse(args); err != nil {
		return err
	}
	config, err := loadAgentDeskConfig()
	if err != nil {
		return err
	}
	mandate, err := loadMandate(*id)
	if err != nil {
		return err
	}
	decision := evaluatePreflight(config, mandate, *action, *pathValue, *command)
	if err := specvalidate.PreflightDecision(decision); err != nil {
		return err
	}
	return writeJSON(stdout, decision)
}

func runAgentDeskContext(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, agentDeskUsage)
		return errors.New("agentdesk context command is required")
	}
	if args[0] != "compile" {
		return fmt.Errorf("unknown agentdesk context command %q", args[0])
	}
	fs := flag.NewFlagSet("agentdesk context compile", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.String("id", "", "mandate/taskpack UUID")
	role := fs.String("role", "coder", "agent role")
	budget := fs.Int("budget", 0, "token budget override")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	config, err := loadAgentDeskConfig()
	if err != nil {
		return err
	}
	mandate, err := loadMandate(*id)
	if err != nil {
		return err
	}
	if *budget == 0 {
		*budget = config.Defaults.ContextBudgetTokens
	}
	pack := spec.ContextPack{
		SchemaVersion: "v1alpha1",
		MandateID:     mandate.TaskpackID,
		Role:          *role,
		BudgetTokens:  *budget,
		MustRead:      refsToPaths(mandate.References),
		MayRead:       uniqueStrings(append(config.Scope.Writable, mandate.Permissions.Scopes...)),
		MayWrite:      uniqueStrings(mandate.Permissions.Scopes),
		Forbidden:     config.Scope.Forbidden,
		Summary:       mandate.Objective,
		ProofRequired: proofKindsFromAcceptance(mandate.Acceptance),
		OmittedReasons: []string{
			"Full transcript omitted; agentdesk emits bounded context packs from mandates, scope, and artifacts.",
		},
	}
	if err := specvalidate.ContextPack(pack); err != nil {
		return err
	}
	return writeJSON(stdout, pack)
}

func runAgentDeskApproval(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, agentDeskUsage)
		return errors.New("agentdesk approval command is required")
	}
	switch args[0] {
	case "request":
		return runAgentDeskApprovalRequest(args[1:], stdout)
	case "resolve":
		return runAgentDeskApprovalResolve(args[1:], stdout)
	default:
		return fmt.Errorf("unknown agentdesk approval command %q", args[0])
	}
}

func runAgentDeskApprovalRequest(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("agentdesk approval request", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.String("id", "", "mandate/taskpack UUID")
	reason := fs.String("reason", "", "approval reason")
	required := fs.Int("required", 1, "required approvals")
	if err := fs.Parse(args); err != nil {
		return err
	}
	mandate, err := loadMandate(*id)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*reason) == "" {
		return errors.New("--reason is required")
	}
	request := spec.ApprovalRequest{
		SchemaVersion: "v1alpha1",
		ApprovalID:    mustNewUUID(),
		TaskpackID:    mandate.TaskpackID,
		RequestedBy: spec.ActorRef{
			ActorID:      mustNewUUID(),
			ActorType:    "agent",
			DisplayName:  "agentdesk-cli",
			Orchestrator: "agentdesk",
		},
		Reason:            *reason,
		RequiredApprovals: *required,
		Status:            "pending",
		CreatedAt:         time.Now().UTC().Format(time.RFC3339),
	}
	if err := specvalidate.ApprovalRequest(request); err != nil {
		return err
	}
	if err := writeAgentDeskJSON(approvalPath(mandate.TaskpackID, request.ApprovalID), request); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "approval-requested %s\n", request.ApprovalID)
	return nil
}

func runAgentDeskApprovalResolve(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("agentdesk approval resolve", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	approvalID := fs.String("approval-id", "", "approval UUID")
	decision := fs.String("decision", "", "approved or rejected")
	reason := fs.String("reason", "", "decision reason")
	actor := fs.String("actor", "human-reviewer", "approver display name")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !specvalidate.IsUUID(*approvalID) {
		return errors.New("--approval-id must be a UUID")
	}
	if *decision != "approved" && *decision != "rejected" {
		return errors.New("--decision must be approved or rejected")
	}
	request, path, err := findApproval(*approvalID)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	request.Approvals = append(request.Approvals, spec.ApprovalDecision{
		Actor: spec.ActorRef{
			ActorID:     mustNewUUID(),
			ActorType:   "human",
			DisplayName: *actor,
		},
		Decision:  *decision,
		Reason:    *reason,
		DecidedAt: now,
	})
	if *decision == "rejected" {
		request.Status = "rejected"
		request.DecidedAt = now
	} else if len(request.Approvals) >= request.RequiredApprovals {
		request.Status = "approved"
		request.DecidedAt = now
	}
	if err := specvalidate.ApprovalRequest(request); err != nil {
		return err
	}
	if err := writeAgentDeskJSON(path, request); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "approval-%s %s\n", request.Status, request.ApprovalID)
	return nil
}

func runAgentDeskProof(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, agentDeskUsage)
		return errors.New("agentdesk proof command is required")
	}
	if args[0] != "add" {
		return fmt.Errorf("unknown agentdesk proof command %q", args[0])
	}
	fs := flag.NewFlagSet("agentdesk proof add", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.String("id", "", "mandate/taskpack UUID")
	kind := fs.String("kind", "custom", "artifact kind")
	pathValue := fs.String("path", "", "proof file path")
	summary := fs.String("summary", "", "proof summary")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	mandate, err := loadMandate(*id)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*pathValue) == "" {
		return errors.New("--path is required")
	}
	artifact, err := buildProofArtifact(mandate, *kind, *pathValue, *summary)
	if err != nil {
		return err
	}
	if err := specvalidate.Artifact(artifact); err != nil {
		return err
	}
	if err := writeAgentDeskJSON(proofPath(mandate.TaskpackID, artifact.ArtifactID), artifact); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "proof-added %s\n", artifact.ArtifactID)
	return nil
}

func runAgentDeskHandoff(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, agentDeskUsage)
		return errors.New("agentdesk handoff command is required")
	}
	if args[0] != "create" {
		return fmt.Errorf("unknown agentdesk handoff command %q", args[0])
	}
	fs := flag.NewFlagSet("agentdesk handoff create", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.String("id", "", "mandate/taskpack UUID")
	to := fs.String("to", "", "target agent or role")
	summary := fs.String("summary", "", "handoff summary")
	file := fs.String("file", "", "handoff summary file")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	mandate, err := loadMandate(*id)
	if err != nil {
		return err
	}
	if strings.TrimSpace(*to) == "" {
		return errors.New("--to is required")
	}
	body := strings.TrimSpace(*summary)
	if *file != "" {
		data, err := os.ReadFile(*file)
		if err != nil {
			return err
		}
		body = strings.TrimSpace(string(data))
	}
	if body == "" {
		return errors.New("--summary or --file is required")
	}
	artifactID := mustNewUUID()
	handoffFile := filepath.Join(".agentdesk", "handoffs", mandate.TaskpackID, artifactID+".md")
	content := fmt.Sprintf("# Handoff\n\nMandate: %s\nTo: %s\nCreated: %s\n\n%s\n", mandate.TaskpackID, *to, time.Now().UTC().Format(time.RFC3339), body)
	if err := os.MkdirAll(filepath.Dir(handoffFile), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(handoffFile, []byte(content), 0o644); err != nil {
		return err
	}
	artifact, err := buildProofArtifact(mandate, "handoff_summary", handoffFile, "Handoff to "+*to+": "+body)
	if err != nil {
		return err
	}
	artifact.ArtifactID = artifactID
	if err := specvalidate.Artifact(artifact); err != nil {
		return err
	}
	if err := writeAgentDeskJSON(proofPath(mandate.TaskpackID, artifact.ArtifactID), artifact); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "handoff-created %s\n", artifact.ArtifactID)
	return nil
}

func runAgentDeskVerify(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("agentdesk verify", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.String("id", "", "mandate/taskpack UUID")
	githubReport := fs.Bool("github-report", false, "write a GitHub Actions summary and PR comment when environment variables are available")
	replayFile := fs.String("replay-file", "", "replay bundle path or URL to reference in CI output")
	if err := fs.Parse(args); err != nil {
		return err
	}
	mandate, err := loadMandate(*id)
	if err != nil {
		return err
	}
	report, err := verifyMandate(mandate)
	if err != nil {
		return err
	}
	if *githubReport {
		if err := publishGitHubAgentWorkReport(mandate, report, *replayFile); err != nil {
			return err
		}
	}
	if err := writeJSON(stdout, report); err != nil {
		return err
	}
	if !report.Ready {
		return errors.New("mandate is not ready")
	}
	return nil
}

func runAgentDeskClose(args []string, stdout io.Writer) error {
	fs := flag.NewFlagSet("agentdesk close", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.String("id", "", "mandate/taskpack UUID")
	if err := fs.Parse(args); err != nil {
		return err
	}
	mandate, err := loadMandate(*id)
	if err != nil {
		return err
	}
	artifacts, err := loadProofArtifacts(mandate.TaskpackID)
	if err != nil {
		return err
	}
	if len(artifacts) == 0 {
		return errors.New("cannot close mandate without at least one proof artifact")
	}
	approvals, err := loadApprovals(mandate.TaskpackID)
	if err != nil {
		return err
	}
	if countPendingApprovals(approvals) > 0 {
		return errors.New("cannot close mandate with pending approvals")
	}
	record := map[string]any{
		"schema_version": "v1alpha1",
		"mandate_id":     mandate.TaskpackID,
		"closed_at":      time.Now().UTC().Format(time.RFC3339),
		"proof_count":    len(artifacts),
	}
	if err := writeAgentDeskJSON(filepath.Join(".agentdesk", "closed", mandate.TaskpackID+".json"), record); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "mandate-closed %s proof_count=%d\n", mandate.TaskpackID, len(artifacts))
	return nil
}

func runAgentDeskReplay(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, agentDeskUsage)
		return errors.New("agentdesk replay command is required")
	}
	if args[0] != "export" {
		return fmt.Errorf("unknown agentdesk replay command %q", args[0])
	}
	fs := flag.NewFlagSet("agentdesk replay export", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	id := fs.String("id", "", "mandate/taskpack UUID")
	file := fs.String("file", "", "optional output file")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	mandate, err := loadMandate(*id)
	if err != nil {
		return err
	}
	artifacts, err := loadProofArtifacts(mandate.TaskpackID)
	if err != nil {
		return err
	}
	bundle := spec.ReplayBundle{
		SchemaVersion:    "v1alpha1",
		RootTaskpackID:   mandate.TaskpackID,
		Taskpack:         mandate,
		Taskpacks:        []spec.Taskpack{mandate},
		DriBindings:      []spec.DriBinding{},
		Artifacts:        artifacts,
		PromotionRecords: []spec.PromotionRecord{},
	}
	if err := specvalidate.ReplayBundle(bundle); err != nil {
		return err
	}
	if *file == "" {
		return writeJSON(stdout, bundle)
	}
	return writeAgentDeskJSON(*file, bundle)
}

func defaultAgentDeskConfig(workspace, mission string) spec.WorkspaceConstitution {
	return spec.WorkspaceConstitution{
		SchemaVersion: "v1alpha1",
		Workspace:     workspace,
		Mission:       mission,
		Defaults: spec.WorkspaceDefaults{
			MaxRuntimeMinutes:   45,
			MaxCostUSD:          5,
			ContextBudgetTokens: 12000,
		},
		TaskSources: []spec.TaskSource{
			{Type: "local", Path: ".agentdesk/mandates"},
		},
		Scope: spec.WorkspaceScope{
			Writable:  []string{"src/**", "tests/**", "docs/**"},
			Forbidden: []string{".env", "infra/prod/**", "billing/**"},
		},
		ApprovalRules: []spec.WorkspaceRule{
			{When: "touches_forbidden_path", Require: "human"},
			{When: "runs_destructive_command", Require: "human"},
			{When: "pushes_to_main", Require: "human"},
		},
		SuccessCriteria: []string{
			"Tests pass or failure is explained.",
			"Every modified file is listed.",
			"A proof artifact is attached.",
			"A reviewer handoff is created.",
		},
		Escalation: &spec.WorkspaceEscalation{
			DefaultOwner: "@owner",
			Channels: []spec.EscalationChannel{
				{Type: "cli_prompt"},
			},
		},
	}
}

func loadAgentDeskConfig() (spec.WorkspaceConstitution, error) {
	root, err := findAgentDeskRoot()
	if err != nil {
		return spec.WorkspaceConstitution{}, err
	}
	data, err := os.ReadFile(filepath.Join(root, "agentdesk.yaml"))
	if err != nil {
		return spec.WorkspaceConstitution{}, err
	}
	var config spec.WorkspaceConstitution
	decoder := yaml.NewDecoder(strings.NewReader(string(data)))
	decoder.KnownFields(true)
	if err := decoder.Decode(&config); err != nil {
		return spec.WorkspaceConstitution{}, err
	}
	if err := specvalidate.WorkspaceConstitution(config); err != nil {
		return spec.WorkspaceConstitution{}, err
	}
	return config, nil
}

func findAgentDeskRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "agentdesk.yaml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("agentdesk.yaml not found; run `guild agentdesk init`")
		}
		dir = parent
	}
}

func loadMandate(id string) (spec.Taskpack, error) {
	if !specvalidate.IsUUID(id) {
		return spec.Taskpack{}, errors.New("--id must be a UUID")
	}
	root, err := findAgentDeskRoot()
	if err != nil {
		return spec.Taskpack{}, err
	}
	var mandate spec.Taskpack
	if err := readAgentDeskJSON(filepath.Join(root, mandatePath(id)), &mandate); err != nil {
		return spec.Taskpack{}, err
	}
	if err := specvalidate.Taskpack(mandate); err != nil {
		return spec.Taskpack{}, err
	}
	return mandate, nil
}

func loadOpenMandates(includeClaimed bool) ([]spec.Taskpack, error) {
	root, err := findAgentDeskRoot()
	if err != nil {
		return nil, err
	}
	files, err := filepath.Glob(filepath.Join(root, ".agentdesk", "mandates", "*.json"))
	if err != nil {
		return nil, err
	}
	mandates := make([]spec.Taskpack, 0, len(files))
	for _, file := range files {
		var mandate spec.Taskpack
		if err := readAgentDeskJSON(file, &mandate); err != nil {
			return nil, err
		}
		if _, err := os.Stat(filepath.Join(root, ".agentdesk", "closed", mandate.TaskpackID+".json")); err == nil {
			continue
		}
		if !includeClaimed {
			if _, active, err := loadClaim(mandate.TaskpackID); err != nil {
				return nil, err
			} else if active {
				continue
			}
		}
		if err := specvalidate.Taskpack(mandate); err != nil {
			return nil, err
		}
		mandates = append(mandates, mandate)
	}
	return mandates, nil
}

type agentDeskClaim struct {
	SchemaVersion string `json:"schema_version"`
	MandateID     string `json:"mandate_id"`
	Agent         string `json:"agent"`
	ClaimedAt     string `json:"claimed_at"`
	ExpiresAt     string `json:"expires_at"`
}

func loadClaim(mandateID string) (agentDeskClaim, bool, error) {
	root, err := findAgentDeskRoot()
	if err != nil {
		return agentDeskClaim{}, false, err
	}
	var claim agentDeskClaim
	path := filepath.Join(root, claimPath(mandateID))
	if err := readAgentDeskJSON(path, &claim); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return agentDeskClaim{}, false, nil
		}
		return agentDeskClaim{}, false, err
	}
	if claim.SchemaVersion != "v1alpha1" {
		return agentDeskClaim{}, false, fmt.Errorf("claim %s has unsupported schema_version %q", mandateID, claim.SchemaVersion)
	}
	if !specvalidate.IsUUID(claim.MandateID) {
		return agentDeskClaim{}, false, fmt.Errorf("claim %s has invalid mandate_id", mandateID)
	}
	expiresAt, err := time.Parse(time.RFC3339, claim.ExpiresAt)
	if err != nil {
		return agentDeskClaim{}, false, fmt.Errorf("claim %s has invalid expires_at", mandateID)
	}
	return claim, time.Now().UTC().Before(expiresAt), nil
}

func loadProofArtifacts(mandateID string) ([]spec.Artifact, error) {
	root, err := findAgentDeskRoot()
	if err != nil {
		return nil, err
	}
	files, err := filepath.Glob(filepath.Join(root, ".agentdesk", "proof", mandateID, "*.json"))
	if err != nil {
		return nil, err
	}
	artifacts := make([]spec.Artifact, 0, len(files))
	for _, file := range files {
		var artifact spec.Artifact
		if err := readAgentDeskJSON(file, &artifact); err != nil {
			return nil, err
		}
		if err := specvalidate.Artifact(artifact); err != nil {
			return nil, err
		}
		artifacts = append(artifacts, artifact)
	}
	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].CreatedAt < artifacts[j].CreatedAt
	})
	return artifacts, nil
}

func loadApprovals(mandateID string) ([]spec.ApprovalRequest, error) {
	root, err := findAgentDeskRoot()
	if err != nil {
		return nil, err
	}
	files, err := filepath.Glob(filepath.Join(root, ".agentdesk", "approvals", mandateID, "*.json"))
	if err != nil {
		return nil, err
	}
	approvals := make([]spec.ApprovalRequest, 0, len(files))
	for _, file := range files {
		var approval spec.ApprovalRequest
		if err := readAgentDeskJSON(file, &approval); err != nil {
			return nil, err
		}
		if err := specvalidate.ApprovalRequest(approval); err != nil {
			return nil, err
		}
		approvals = append(approvals, approval)
	}
	sort.Slice(approvals, func(i, j int) bool {
		return approvals[i].CreatedAt < approvals[j].CreatedAt
	})
	return approvals, nil
}

func findApproval(approvalID string) (spec.ApprovalRequest, string, error) {
	root, err := findAgentDeskRoot()
	if err != nil {
		return spec.ApprovalRequest{}, "", err
	}
	files, err := filepath.Glob(filepath.Join(root, ".agentdesk", "approvals", "*", approvalID+".json"))
	if err != nil {
		return spec.ApprovalRequest{}, "", err
	}
	if len(files) == 0 {
		return spec.ApprovalRequest{}, "", fmt.Errorf("approval %q not found", approvalID)
	}
	var approval spec.ApprovalRequest
	if err := readAgentDeskJSON(files[0], &approval); err != nil {
		return spec.ApprovalRequest{}, "", err
	}
	if err := specvalidate.ApprovalRequest(approval); err != nil {
		return spec.ApprovalRequest{}, "", err
	}
	return approval, files[0], nil
}

type agentDeskVerifyReport struct {
	SchemaVersion        string   `json:"schema_version"`
	MandateID            string   `json:"mandate_id"`
	Ready                bool     `json:"ready"`
	ProofCount           int      `json:"proof_count"`
	RequiredProofKinds   []string `json:"required_proof_kinds"`
	PresentProofKinds    []string `json:"present_proof_kinds"`
	MissingProofKinds    []string `json:"missing_proof_kinds,omitempty"`
	PendingApprovalCount int      `json:"pending_approval_count"`
	OpenIssues           []string `json:"open_issues,omitempty"`
}

func verifyMandate(mandate spec.Taskpack) (agentDeskVerifyReport, error) {
	artifacts, err := loadProofArtifacts(mandate.TaskpackID)
	if err != nil {
		return agentDeskVerifyReport{}, err
	}
	approvals, err := loadApprovals(mandate.TaskpackID)
	if err != nil {
		return agentDeskVerifyReport{}, err
	}
	required := proofKindsFromAcceptance(mandate.Acceptance)
	present := artifactKinds(artifacts)
	missing := missingStrings(required, present)
	pending := countPendingApprovals(approvals)
	issues := []string{}
	if len(artifacts) == 0 {
		issues = append(issues, "at least one proof artifact is required")
	}
	if pending > 0 {
		issues = append(issues, "pending approvals must be resolved")
	}
	for _, kind := range missing {
		issues = append(issues, "missing proof kind: "+kind)
	}
	return agentDeskVerifyReport{
		SchemaVersion:        "v1alpha1",
		MandateID:            mandate.TaskpackID,
		Ready:                len(issues) == 0,
		ProofCount:           len(artifacts),
		RequiredProofKinds:   required,
		PresentProofKinds:    present,
		MissingProofKinds:    missing,
		PendingApprovalCount: pending,
		OpenIssues:           issues,
	}, nil
}

func artifactKinds(artifacts []spec.Artifact) []string {
	kinds := make([]string, 0, len(artifacts))
	for _, artifact := range artifacts {
		kinds = append(kinds, artifact.Kind)
	}
	return orderProofKinds(uniqueStrings(kinds))
}

func missingStrings(required, present []string) []string {
	seen := map[string]struct{}{}
	for _, value := range present {
		seen[value] = struct{}{}
	}
	missing := []string{}
	for _, value := range required {
		if _, ok := seen[value]; !ok {
			missing = append(missing, value)
		}
	}
	return missing
}

func countPendingApprovals(approvals []spec.ApprovalRequest) int {
	count := 0
	for _, approval := range approvals {
		if approval.Status == "pending" {
			count++
		}
	}
	return count
}

func evaluatePreflight(config spec.WorkspaceConstitution, mandate spec.Taskpack, action, pathValue, command string) spec.PreflightDecision {
	decision := spec.PreflightDecision{
		SchemaVersion: "v1alpha1",
		MandateID:     mandate.TaskpackID,
		Action:        action,
		Path:          pathValue,
		Command:       command,
		Decision:      "allow",
		Reason:        "Action is within mandate and workspace policy.",
		MatchedRules:  []string{"default.allow"},
	}
	if action == "" {
		decision.Action = inferPreflightAction(pathValue, command)
	}
	if pathValue != "" {
		normalized := filepath.ToSlash(filepath.Clean(pathValue))
		if matchesAny(normalized, config.Scope.Forbidden) {
			return approvalDecision(mandate.TaskpackID, decision.Action, pathValue, command, "Path matches forbidden workspace scope.", "scope.forbidden")
		}
		if decision.Action == "write" && !matchesAny(normalized, append(config.Scope.Writable, mandate.Permissions.Scopes...)) {
			return approvalDecision(mandate.TaskpackID, decision.Action, pathValue, command, "Path is outside writable workspace and mandate scope.", "scope.write_outside_allowlist")
		}
	}
	if command != "" {
		lower := strings.ToLower(strings.TrimSpace(command))
		if isDestructiveCommand(lower) {
			return approvalDecision(mandate.TaskpackID, "run", pathValue, command, "Command looks destructive and requires approval.", "approval.runs_destructive_command")
		}
		if strings.HasPrefix(lower, "git push") {
			return approvalDecision(mandate.TaskpackID, "git_push", pathValue, command, "Pushing requires approval.", "approval.pushes_to_main")
		}
		if !mandate.Permissions.AllowShell {
			return approvalDecision(mandate.TaskpackID, "run", pathValue, command, "Mandate does not allow shell commands.", "permissions.allow_shell")
		}
	}
	if decision.Action == "network" && !mandate.Permissions.AllowNetwork {
		return approvalDecision(mandate.TaskpackID, "network", pathValue, command, "Mandate does not allow network access.", "permissions.allow_network")
	}
	return decision
}

func approvalDecision(mandateID, action, pathValue, command, reason, rule string) spec.PreflightDecision {
	return spec.PreflightDecision{
		SchemaVersion:    "v1alpha1",
		MandateID:        mandateID,
		Action:           action,
		Path:             pathValue,
		Command:          command,
		Decision:         "needs_approval",
		Reason:           reason,
		ApprovalRequired: true,
		MatchedRules:     []string{rule},
	}
}

func inferPreflightAction(pathValue, command string) string {
	if command != "" {
		return "run"
	}
	if pathValue != "" {
		return "read"
	}
	return "custom"
}

func isDestructiveCommand(command string) bool {
	dangerous := []string{
		"rm -rf",
		"rm -fr",
		"git reset --hard",
		"git clean -fd",
		"chmod -r",
		"chown -r",
		"drop database",
		"terraform apply",
		"kubectl delete",
	}
	for _, item := range dangerous {
		if strings.Contains(command, item) {
			return true
		}
	}
	return false
}

func buildProofArtifact(mandate spec.Taskpack, kind, pathValue, summary string) (spec.Artifact, error) {
	absolute, err := filepath.Abs(pathValue)
	if err != nil {
		return spec.Artifact{}, err
	}
	info, err := os.Stat(absolute)
	if err != nil {
		return spec.Artifact{}, err
	}
	digest, err := sha256File(absolute)
	if err != nil {
		return spec.Artifact{}, err
	}
	if summary == "" {
		summary = "Proof artifact published from local agentdesk workflow."
	}
	uri := (&url.URL{Scheme: "file", Path: absolute}).String()
	return spec.Artifact{
		SchemaVersion: "v1alpha1",
		ArtifactID:    mustNewUUID(),
		TaskpackID:    mandate.TaskpackID,
		Kind:          kind,
		Title:         filepath.Base(pathValue),
		Summary:       summary,
		Producer: spec.ActorRef{
			ActorID:      mustNewUUID(),
			ActorType:    "agent",
			DisplayName:  "agentdesk-cli",
			Orchestrator: "agentdesk",
		},
		Storage: spec.ArtifactStorage{
			URI:       uri,
			MimeType:  mimeTypeForPath(pathValue),
			SHA256:    digest,
			SizeBytes: info.Size(),
		},
		Labels:    []string{"agentdesk", "proof"},
		Version:   1,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func sha256File(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func mimeTypeForPath(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	case ".md":
		return "text/markdown"
	case ".txt", ".log":
		return "text/plain"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}

func defaultAcceptance(criteria []string) []spec.AcceptanceCriteria {
	if len(criteria) == 0 {
		criteria = []string{"A proof artifact is attached."}
	}
	items := make([]spec.AcceptanceCriteria, 0, len(criteria))
	for i, criterion := range criteria {
		items = append(items, spec.AcceptanceCriteria{
			CriterionID: fmt.Sprintf("criterion_%02d", i+1),
			Description: criterion,
			Required:    true,
		})
	}
	return items
}

func proofKindsFromAcceptance(criteria []spec.AcceptanceCriteria) []string {
	required := []string{"test_report", "changed_files", "handoff_summary"}
	for _, criterion := range criteria {
		lower := strings.ToLower(criterion.Description)
		if strings.Contains(lower, "screenshot") {
			required = append(required, "screenshot")
		}
		if strings.Contains(lower, "benchmark") {
			required = append(required, "benchmark_result")
		}
	}
	return orderProofKinds(uniqueStrings(required))
}

func orderProofKinds(kinds []string) []string {
	rank := map[string]int{
		"test_report":     0,
		"changed_files":   1,
		"handoff_summary": 2,
	}
	sort.SliceStable(kinds, func(i, j int) bool {
		left, leftOK := rank[kinds[i]]
		right, rightOK := rank[kinds[j]]
		switch {
		case leftOK && rightOK:
			return left < right
		case leftOK:
			return true
		case rightOK:
			return false
		default:
			return kinds[i] < kinds[j]
		}
	})
	return kinds
}

func refsToPaths(refs []string) []string {
	paths := make([]string, 0, len(refs))
	for _, ref := range refs {
		parsed, err := url.Parse(ref)
		if err == nil && parsed.Scheme == "file" {
			paths = append(paths, parsed.Path)
			continue
		}
		paths = append(paths, ref)
	}
	return paths
}

func mandatePath(id string) string {
	return filepath.Join(".agentdesk", "mandates", id+".json")
}

func proofPath(mandateID, artifactID string) string {
	return filepath.Join(".agentdesk", "proof", mandateID, artifactID+".json")
}

func claimPath(mandateID string) string {
	return filepath.Join(".agentdesk", "claims", mandateID+".json")
}

func approvalPath(mandateID, approvalID string) string {
	return filepath.Join(".agentdesk", "approvals", mandateID, approvalID+".json")
}

func readAgentDeskJSON(path string, dest any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return decodeStrict(data, dest)
}

func writeAgentDeskJSON(path string, payload any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0o644)
}

func writeJSON(stdout io.Writer, payload any) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = stdout.Write(data)
	return err
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			items = append(items, trimmed)
		}
	}
	return items
}

func reorderInterspersedFlags(args []string, flagsWithValues map[string]bool) []string {
	flags := []string{}
	positionals := []string{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if !strings.HasPrefix(arg, "-") || arg == "-" {
			positionals = append(positionals, arg)
			continue
		}
		flags = append(flags, arg)
		name := strings.TrimLeft(strings.SplitN(arg, "=", 2)[0], "-")
		if strings.Contains(arg, "=") || !flagsWithValues[name] || i+1 >= len(args) {
			continue
		}
		i++
		flags = append(flags, args[i])
	}
	return append(flags, positionals...)
}

func uniqueStrings(values []string) []string {
	seen := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func priorityRank(priority string) int {
	switch priority {
	case "critical":
		return 4
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

func normalizeRoleHint(role string) string {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "coder", "implementer", "developer":
		return "builder"
	case "critic":
		return "skeptic"
	default:
		return role
	}
}

func matchesAny(pathValue string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchesPattern(pathValue, pattern) {
			return true
		}
	}
	return false
}

func matchesPattern(pathValue, pattern string) bool {
	normalized := filepath.ToSlash(strings.TrimPrefix(filepath.Clean(pattern), "./"))
	pathValue = filepath.ToSlash(strings.TrimPrefix(filepath.Clean(pathValue), "./"))
	if strings.HasSuffix(normalized, "/**") {
		prefix := strings.TrimSuffix(normalized, "/**")
		return pathValue == prefix || strings.HasPrefix(pathValue, prefix+"/")
	}
	if strings.HasSuffix(normalized, "/*") {
		prefix := strings.TrimSuffix(normalized, "/*")
		if !strings.HasPrefix(pathValue, prefix+"/") {
			return false
		}
		return !strings.Contains(strings.TrimPrefix(pathValue, prefix+"/"), "/")
	}
	if matched, err := filepath.Match(normalized, pathValue); err == nil && matched {
		return true
	}
	return pathValue == normalized
}
