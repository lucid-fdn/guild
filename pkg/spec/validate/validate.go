package validate

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/lucid-fdn/guild/pkg/spec"
)

const schemaVersion = "v1alpha1"

var (
	uuidPattern      = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	labelPattern     = regexp.MustCompile(`^[a-z0-9][a-z0-9._/-]{0,63}$`)
	criterionPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{1,31}$`)
	sha256Pattern    = regexp.MustCompile(`^[a-f0-9]{64}$`)
)

func Taskpack(t spec.Taskpack) error {
	switch {
	case t.SchemaVersion != schemaVersion:
		return fmt.Errorf("schema_version must be %q", schemaVersion)
	case !IsUUID(t.TaskpackID):
		return errors.New("taskpack_id must be a UUID")
	case t.InstitutionID != "" && !IsUUID(t.InstitutionID):
		return errors.New("institution_id must be a UUID")
	case t.ParentTaskpackID != "" && !IsUUID(t.ParentTaskpackID):
		return errors.New("parent_taskpack_id must be a UUID")
	case strings.TrimSpace(t.Title) == "":
		return errors.New("title is required")
	case strings.TrimSpace(t.Objective) == "":
		return errors.New("objective is required")
	case !inSet(t.TaskType, "analysis", "implementation", "review", "research", "triage", "evaluation", "operations", "custom"):
		return errors.New("task_type is invalid")
	case !inSet(t.Priority, "low", "medium", "high", "critical"):
		return errors.New("priority is invalid")
	case t.RoleHint != "" && !inSet(t.RoleHint, "dri", "explorer", "builder", "skeptic", "reviewer", "specialist", "approver"):
		return errors.New("role_hint is invalid")
	case t.ContextBudget.MaxInputTokens < 256:
		return errors.New("context_budget.max_input_tokens must be >= 256")
	case t.ContextBudget.MaxOutputTokens < 128:
		return errors.New("context_budget.max_output_tokens must be >= 128")
	case !inSet(t.ContextBudget.ContextStrategy, "artifact_refs_first", "summary_plus_refs", "full_bundle"):
		return errors.New("context_budget.context_strategy is invalid")
	case t.ContextBudget.MaxArtifactsInline < 0:
		return errors.New("context_budget.max_artifacts_inline must be >= 0")
	case !inSet(t.Permissions.ApprovalMode, "none", "ask", "required"):
		return errors.New("permissions.approval_mode is invalid")
	case len(t.Acceptance) == 0:
		return errors.New("acceptance_criteria must not be empty")
	case !isRFC3339(t.CreatedAt):
		return errors.New("created_at must be RFC3339 date-time")
	default:
		if err := errActorRef("requested_by", t.RequestedBy); err != nil {
			return err
		}
		return validateTaskpackCollections(t)
	}
}

func DriBinding(d spec.DriBinding) error {
	switch {
	case d.SchemaVersion != schemaVersion:
		return fmt.Errorf("schema_version must be %q", schemaVersion)
	case !IsUUID(d.DriBindingID):
		return errors.New("dri_binding_id must be a UUID")
	case !IsUUID(d.TaskpackID):
		return errors.New("taskpack_id must be a UUID")
	case !inSet(d.Status, "assigned", "accepted", "in_progress", "blocked", "completed", "canceled"):
		return errors.New("status is invalid")
	case !isRFC3339(d.CreatedAt):
		return errors.New("created_at must be RFC3339 date-time")
	default:
		if err := errActorRef("owner", d.Owner); err != nil {
			return err
		}
		return validateDriCollections(d)
	}
}

func Artifact(a spec.Artifact) error {
	switch {
	case a.SchemaVersion != schemaVersion:
		return fmt.Errorf("schema_version must be %q", schemaVersion)
	case !IsUUID(a.ArtifactID):
		return errors.New("artifact_id must be a UUID")
	case !IsUUID(a.TaskpackID):
		return errors.New("taskpack_id must be a UUID")
	case !inSet(a.Kind, "report", "code_patch", "review", "plan", "dataset", "decision_log", "benchmark_result", "skill_candidate", "test_report", "diff", "changed_files", "screenshot", "log_excerpt", "security_review", "handoff_summary", "human_approval", "custom"):
		return errors.New("kind is invalid")
	case strings.TrimSpace(a.Title) == "":
		return errors.New("title is required")
	case !isURI(a.Storage.URI):
		return errors.New("storage.uri must be a URI")
	case strings.TrimSpace(a.Storage.MimeType) == "":
		return errors.New("storage.mime_type is required")
	case a.Storage.SHA256 != "" && !sha256Pattern.MatchString(a.Storage.SHA256):
		return errors.New("storage.sha256 must be lowercase hex sha256")
	case a.Storage.SizeBytes < 0:
		return errors.New("storage.size_bytes must be >= 0")
	case a.SecurityClassification != "" && !inSet(a.SecurityClassification, "public", "internal", "restricted"):
		return errors.New("security_classification is invalid")
	case a.Version <= 0:
		return errors.New("version must be >= 1")
	case !isRFC3339(a.CreatedAt):
		return errors.New("created_at must be RFC3339 date-time")
	default:
		if err := errActorRef("producer", a.Producer); err != nil {
			return err
		}
		return validateArtifactCollections(a)
	}
}

func PromotionRecord(p spec.PromotionRecord) error {
	switch {
	case p.SchemaVersion != schemaVersion:
		return fmt.Errorf("schema_version must be %q", schemaVersion)
	case !IsUUID(p.PromotionID):
		return errors.New("promotion_record_id must be a UUID")
	case !IsUUID(p.InstitutionID):
		return errors.New("institution_id must be a UUID")
	case !inSet(p.CandidateKind, "skill", "rule", "prompt_pattern", "workflow_pattern", "review_heuristic"):
		return errors.New("candidate_kind is invalid")
	case !inSet(p.Decision, "accepted", "rejected", "needs_human_review", "deferred"):
		return errors.New("decision is invalid")
	case p.AcceptedScope != "" && !inSet(p.AcceptedScope, "institution", "team", "project", "role_specific"):
		return errors.New("accepted_scope is invalid")
	case !isRFC3339(p.DecidedAt):
		return errors.New("decided_at must be RFC3339 date-time")
	default:
		if err := errArtifactRef("candidate_ref", p.CandidateRef); err != nil {
			return err
		}
		for i, runID := range p.SourceRunIDs {
			if !IsUUID(runID) {
				return fmt.Errorf("source_run_ids[%d] must be a UUID", i)
			}
		}
		for i, metric := range p.Metrics {
			if strings.TrimSpace(metric.Name) == "" {
				return fmt.Errorf("metrics[%d].name is required", i)
			}
			if !inSet(metric.Direction, "higher_is_better", "lower_is_better") {
				return fmt.Errorf("metrics[%d].direction is invalid", i)
			}
		}
		for i, decider := range p.Deciders {
			if err := errActorRef(fmt.Sprintf("deciders[%d]", i), decider); err != nil {
				return err
			}
		}
		return nil
	}
}

func GovernancePolicy(policy spec.GovernancePolicy) error {
	switch {
	case policy.SchemaVersion != schemaVersion:
		return fmt.Errorf("schema_version must be %q", schemaVersion)
	case !IsUUID(policy.PolicyID):
		return errors.New("policy_id must be a UUID")
	case !IsUUID(policy.InstitutionID):
		return errors.New("institution_id must be a UUID")
	case strings.TrimSpace(policy.Name) == "":
		return errors.New("name is required")
	case len(policy.AppliesTo) == 0:
		return errors.New("applies_to must not be empty")
	case len(policy.Rules) == 0:
		return errors.New("rules must not be empty")
	case !isRFC3339(policy.CreatedAt):
		return errors.New("created_at must be RFC3339 date-time")
	default:
		if err := errActorRef("created_by", policy.CreatedBy); err != nil {
			return err
		}
		for i, target := range policy.AppliesTo {
			if !inSet(target, "taskpack", "artifact", "promotion_record", "commons_entry") {
				return fmt.Errorf("applies_to[%d] is invalid", i)
			}
		}
		for i, rule := range policy.Rules {
			if !criterionPattern.MatchString(rule.RuleID) {
				return fmt.Errorf("rules[%d].rule_id is invalid", i)
			}
			if !inSet(rule.Effect, "allow", "deny", "require_approval") {
				return fmt.Errorf("rules[%d].effect is invalid", i)
			}
			if strings.TrimSpace(rule.Condition) == "" {
				return fmt.Errorf("rules[%d].condition is required", i)
			}
			if rule.MinApprovals < 0 {
				return fmt.Errorf("rules[%d].min_approvals must be >= 0", i)
			}
		}
		return nil
	}
}

func ApprovalRequest(request spec.ApprovalRequest) error {
	switch {
	case request.SchemaVersion != schemaVersion:
		return fmt.Errorf("schema_version must be %q", schemaVersion)
	case !IsUUID(request.ApprovalID):
		return errors.New("approval_id must be a UUID")
	case !IsUUID(request.TaskpackID):
		return errors.New("taskpack_id must be a UUID")
	case request.PolicyID != "" && !IsUUID(request.PolicyID):
		return errors.New("policy_id must be a UUID")
	case strings.TrimSpace(request.Reason) == "":
		return errors.New("reason is required")
	case request.RequiredApprovals < 1:
		return errors.New("required_approvals must be >= 1")
	case !inSet(request.Status, "pending", "approved", "rejected", "canceled"):
		return errors.New("status is invalid")
	case !isRFC3339(request.CreatedAt):
		return errors.New("created_at must be RFC3339 date-time")
	case request.DecidedAt != "" && !isRFC3339(request.DecidedAt):
		return errors.New("decided_at must be RFC3339 date-time")
	default:
		if err := errActorRef("requested_by", request.RequestedBy); err != nil {
			return err
		}
		for i, approval := range request.Approvals {
			if err := errActorRef(fmt.Sprintf("approvals[%d].actor", i), approval.Actor); err != nil {
				return err
			}
			if !inSet(approval.Decision, "approved", "rejected") {
				return fmt.Errorf("approvals[%d].decision is invalid", i)
			}
			if !isRFC3339(approval.DecidedAt) {
				return fmt.Errorf("approvals[%d].decided_at must be RFC3339 date-time", i)
			}
		}
		return nil
	}
}

func PromotionGate(gate spec.PromotionGate) error {
	switch {
	case gate.SchemaVersion != schemaVersion:
		return fmt.Errorf("schema_version must be %q", schemaVersion)
	case !IsUUID(gate.GateID):
		return errors.New("gate_id must be a UUID")
	case !IsUUID(gate.InstitutionID):
		return errors.New("institution_id must be a UUID")
	case strings.TrimSpace(gate.Name) == "":
		return errors.New("name is required")
	case len(gate.CandidateKinds) == 0:
		return errors.New("candidate_kinds must not be empty")
	case gate.MinReplayRuns < 1:
		return errors.New("min_replay_runs must be >= 1")
	case !isRFC3339(gate.CreatedAt):
		return errors.New("created_at must be RFC3339 date-time")
	default:
		for i, kind := range gate.CandidateKinds {
			if !inSet(kind, "skill", "rule", "prompt_pattern", "workflow_pattern", "review_heuristic") {
				return fmt.Errorf("candidate_kinds[%d] is invalid", i)
			}
		}
		for i, metric := range gate.RequiredMetrics {
			if strings.TrimSpace(metric.Name) == "" {
				return fmt.Errorf("required_metrics[%d].name is required", i)
			}
			if !inSet(metric.Direction, "higher_is_better", "lower_is_better") {
				return fmt.Errorf("required_metrics[%d].direction is invalid", i)
			}
		}
		return nil
	}
}

func CommonsEntry(entry spec.CommonsEntry) error {
	switch {
	case entry.SchemaVersion != schemaVersion:
		return fmt.Errorf("schema_version must be %q", schemaVersion)
	case !IsUUID(entry.CommonsEntryID):
		return errors.New("commons_entry_id must be a UUID")
	case !IsUUID(entry.InstitutionID):
		return errors.New("institution_id must be a UUID")
	case !IsUUID(entry.PromotionRecordID):
		return errors.New("promotion_record_id must be a UUID")
	case strings.TrimSpace(entry.Title) == "":
		return errors.New("title is required")
	case strings.TrimSpace(entry.Summary) == "":
		return errors.New("summary is required")
	case !inSet(entry.Scope, "institution", "team", "project", "role_specific"):
		return errors.New("scope is invalid")
	case !inSet(entry.Status, "active", "deprecated", "archived"):
		return errors.New("status is invalid")
	case !isRFC3339(entry.CreatedAt):
		return errors.New("created_at must be RFC3339 date-time")
	default:
		return errArtifactRef("artifact_ref", entry.ArtifactRef)
	}
}

func ReplayBundle(bundle spec.ReplayBundle) error {
	if bundle.SchemaVersion != schemaVersion {
		return fmt.Errorf("schema_version must be %q", schemaVersion)
	}
	if err := Taskpack(bundle.Taskpack); err != nil {
		return fmt.Errorf("taskpack: %w", err)
	}
	if bundle.RootTaskpackID != "" && bundle.RootTaskpackID != bundle.Taskpack.TaskpackID {
		return errors.New("root_taskpack_id must match taskpack.taskpack_id")
	}
	taskpackIDs := map[string]struct{}{bundle.Taskpack.TaskpackID: {}}
	if len(bundle.Taskpacks) > 0 {
		foundRoot := false
		for i, taskpack := range bundle.Taskpacks {
			if err := Taskpack(taskpack); err != nil {
				return fmt.Errorf("taskpacks[%d]: %w", i, err)
			}
			if _, exists := taskpackIDs[taskpack.TaskpackID]; exists && taskpack.TaskpackID != bundle.Taskpack.TaskpackID {
				return fmt.Errorf("taskpacks[%d].taskpack_id is duplicated", i)
			}
			taskpackIDs[taskpack.TaskpackID] = struct{}{}
			if taskpack.TaskpackID == bundle.Taskpack.TaskpackID {
				foundRoot = true
			}
			if taskpack.ParentTaskpackID != "" {
				if _, ok := taskpackIDs[taskpack.ParentTaskpackID]; !ok && taskpack.ParentTaskpackID != bundle.Taskpack.TaskpackID {
					// Parent order is not guaranteed; validate after collecting IDs below.
				}
			}
		}
		if !foundRoot {
			return errors.New("taskpacks must include the root taskpack")
		}
		for i, taskpack := range bundle.Taskpacks {
			if taskpack.ParentTaskpackID != "" {
				if _, ok := taskpackIDs[taskpack.ParentTaskpackID]; !ok {
					return fmt.Errorf("taskpacks[%d].parent_taskpack_id must exist in taskpacks", i)
				}
			}
		}
	}
	for i, binding := range bundle.DriBindings {
		if err := DriBinding(binding); err != nil {
			return fmt.Errorf("dri_bindings[%d]: %w", i, err)
		}
		if _, ok := taskpackIDs[binding.TaskpackID]; !ok {
			return fmt.Errorf("dri_bindings[%d].taskpack_id must exist in replay taskpacks", i)
		}
	}
	for i, artifact := range bundle.Artifacts {
		if err := Artifact(artifact); err != nil {
			return fmt.Errorf("artifacts[%d]: %w", i, err)
		}
		if _, ok := taskpackIDs[artifact.TaskpackID]; !ok {
			return fmt.Errorf("artifacts[%d].taskpack_id must exist in replay taskpacks", i)
		}
	}

	artifactIDs := make(map[string]struct{}, len(bundle.Artifacts))
	for _, artifact := range bundle.Artifacts {
		artifactIDs[artifact.ArtifactID] = struct{}{}
	}
	for i, record := range bundle.PromotionRecords {
		if err := PromotionRecord(record); err != nil {
			return fmt.Errorf("promotion_records[%d]: %w", i, err)
		}
		if _, ok := artifactIDs[record.CandidateRef.ArtifactID]; !ok {
			return fmt.Errorf("promotion_records[%d].candidate_ref.artifact_id must exist in artifacts", i)
		}
	}
	return nil
}

func WorkspaceConstitution(config spec.WorkspaceConstitution) error {
	switch {
	case config.SchemaVersion != schemaVersion:
		return fmt.Errorf("version must be %q", schemaVersion)
	case strings.TrimSpace(config.Workspace) == "":
		return errors.New("workspace is required")
	case strings.TrimSpace(config.Mission) == "":
		return errors.New("mission is required")
	case config.Defaults.MaxRuntimeMinutes < 1:
		return errors.New("defaults.max_runtime_minutes must be >= 1")
	case config.Defaults.MaxCostUSD < 0:
		return errors.New("defaults.max_cost_usd must be >= 0")
	case config.Defaults.ContextBudgetTokens < 256:
		return errors.New("defaults.context_budget_tokens must be >= 256")
	default:
		for i, source := range config.TaskSources {
			if !inSet(source.Type, "local", "github_issues", "ci_failure", "linear", "jira", "slack", "paperclip", "openfang", "openclaw", "custom") {
				return fmt.Errorf("task_sources[%d].type is invalid", i)
			}
			if source.Type == "local" && strings.TrimSpace(source.Path) == "" {
				return fmt.Errorf("task_sources[%d].path is required for local sources", i)
			}
			if source.Type == "github_issues" && strings.TrimSpace(source.Repo) == "" {
				return fmt.Errorf("task_sources[%d].repo is required for github_issues sources", i)
			}
		}
		for i, rule := range config.ApprovalRules {
			if !inSet(rule.When, "touches_forbidden_path", "runs_destructive_command", "changes_auth_or_payments", "pushes_to_main", "external_api_call", "secret_access", "dependency_install", "production_access", "custom") {
				return fmt.Errorf("approval_rules[%d].when is invalid", i)
			}
			if !inSet(rule.Require, "human", "reviewer", "owner", "deny") {
				return fmt.Errorf("approval_rules[%d].require is invalid", i)
			}
		}
		return nil
	}
}

func ContextPack(pack spec.ContextPack) error {
	switch {
	case pack.SchemaVersion != schemaVersion:
		return fmt.Errorf("schema_version must be %q", schemaVersion)
	case !IsUUID(pack.MandateID):
		return errors.New("mandate_id must be a UUID")
	case strings.TrimSpace(pack.Role) == "":
		return errors.New("role is required")
	case pack.BudgetTokens < 256:
		return errors.New("budget_tokens must be >= 256")
	case strings.TrimSpace(pack.Summary) == "":
		return errors.New("summary is required")
	default:
		return nil
	}
}

func PreflightDecision(decision spec.PreflightDecision) error {
	switch {
	case decision.SchemaVersion != schemaVersion:
		return fmt.Errorf("schema_version must be %q", schemaVersion)
	case !IsUUID(decision.MandateID):
		return errors.New("mandate_id must be a UUID")
	case !inSet(decision.Action, "read", "write", "run", "network", "secret", "git_push", "dependency_install", "prod_access", "custom"):
		return errors.New("action is invalid")
	case !inSet(decision.Decision, "allow", "deny", "needs_approval", "needs_handoff"):
		return errors.New("decision is invalid")
	case strings.TrimSpace(decision.Reason) == "":
		return errors.New("reason is required")
	default:
		return nil
	}
}

func IsUUID(value string) bool {
	return uuidPattern.MatchString(value)
}

func validateTaskpackCollections(t spec.Taskpack) error {
	for i, input := range t.Inputs {
		if err := errArtifactRef(fmt.Sprintf("inputs[%d]", i), input); err != nil {
			return err
		}
	}
	for i, ref := range t.References {
		if !isURI(ref) {
			return fmt.Errorf("references[%d] must be a URI", i)
		}
	}
	if err := validateLabels(t.Labels); err != nil {
		return err
	}
	for i, criterion := range t.Acceptance {
		if !criterionPattern.MatchString(criterion.CriterionID) {
			return fmt.Errorf("acceptance_criteria[%d].criterion_id is invalid", i)
		}
		if strings.TrimSpace(criterion.Description) == "" {
			return fmt.Errorf("acceptance_criteria[%d].description is required", i)
		}
	}
	if t.Constraints != nil {
		if t.Constraints.DeadlineAt != "" && !isRFC3339(t.Constraints.DeadlineAt) {
			return errors.New("constraints.deadline_at must be RFC3339 date-time")
		}
		if t.Constraints.CostCeilingUSD < 0 {
			return errors.New("constraints.cost_ceiling_usd must be >= 0")
		}
	}
	if t.DelegationPolicy != nil {
		if t.DelegationPolicy.MaxDepth < 0 {
			return errors.New("delegation_policy.max_depth must be >= 0")
		}
		if t.DelegationPolicy.MaxParallelSubtasks < 1 {
			return errors.New("delegation_policy.max_parallel_subtasks must be >= 1")
		}
	}
	return nil
}

func validateDriCollections(d spec.DriBinding) error {
	for i, reviewer := range d.Reviewers {
		if err := errActorRef(fmt.Sprintf("reviewers[%d]", i), reviewer); err != nil {
			return err
		}
	}
	for i, specialist := range d.Specialists {
		if err := errActorRef(fmt.Sprintf("specialists[%d]", i), specialist); err != nil {
			return err
		}
	}
	for i, approver := range d.Approvers {
		if err := errActorRef(fmt.Sprintf("approvers[%d]", i), approver); err != nil {
			return err
		}
	}
	if d.EscalationPolicy != nil {
		if !inSet(d.EscalationPolicy.Mode, "manual", "time_based", "policy_based") {
			return errors.New("escalation_policy.mode is invalid")
		}
		if d.EscalationPolicy.EscalateAfterSeconds < 0 {
			return errors.New("escalation_policy.escalate_after_seconds must be >= 0")
		}
		for i, target := range d.EscalationPolicy.Targets {
			if err := errActorRef(fmt.Sprintf("escalation_policy.targets[%d]", i), target); err != nil {
				return err
			}
		}
	}
	if d.ApprovalPolicy != nil {
		if d.ApprovalPolicy.MinApprovals < 0 {
			return errors.New("approval_policy.min_approvals must be >= 0")
		}
		for i, requiredFor := range d.ApprovalPolicy.RequiredFor {
			if !inSet(requiredFor, "external_write", "network", "prod_access", "cost_overrun", "publication") {
				return fmt.Errorf("approval_policy.required_for[%d] is invalid", i)
			}
		}
	}
	if d.VisibilityPolicy != nil {
		if !inSet(d.VisibilityPolicy.Mode, "team", "project", "institution", "restricted") {
			return errors.New("visibility_policy.mode is invalid")
		}
		for i, reader := range d.VisibilityPolicy.Readers {
			if err := errActorRef(fmt.Sprintf("visibility_policy.readers[%d]", i), reader); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateArtifactCollections(a spec.Artifact) error {
	for i, parentID := range a.ParentArtifactIDs {
		if !IsUUID(parentID) {
			return fmt.Errorf("parent_artifact_ids[%d] must be a UUID", i)
		}
	}
	if err := validateLabels(a.Labels); err != nil {
		return err
	}
	if a.Lineage != nil {
		if a.Lineage.RunID != "" && !IsUUID(a.Lineage.RunID) {
			return errors.New("lineage.run_id must be a UUID")
		}
		for i, taskpackID := range a.Lineage.SourceTaskpackIDs {
			if !IsUUID(taskpackID) {
				return fmt.Errorf("lineage.source_taskpack_ids[%d] must be a UUID", i)
			}
		}
	}
	if a.EvaluationState != nil && a.EvaluationState.Status != "" && !inSet(a.EvaluationState.Status, "not_evaluated", "pending", "passed", "failed") {
		return errors.New("evaluation_state.status is invalid")
	}
	return nil
}

func validateLabels(labels []string) error {
	seen := make(map[string]struct{}, len(labels))
	for i, label := range labels {
		if !labelPattern.MatchString(label) {
			return fmt.Errorf("labels[%d] is invalid", i)
		}
		if _, ok := seen[label]; ok {
			return fmt.Errorf("labels[%d] is duplicated", i)
		}
		seen[label] = struct{}{}
	}
	return nil
}

func errActorRef(path string, actor spec.ActorRef) error {
	switch {
	case !IsUUID(actor.ActorID):
		return fmt.Errorf("%s.actor_id must be a UUID", path)
	case !inSet(actor.ActorType, "agent", "human", "service"):
		return fmt.Errorf("%s.actor_type is invalid", path)
	case strings.TrimSpace(actor.DisplayName) == "":
		return fmt.Errorf("%s.display_name is required", path)
	case actor.Endpoint != "" && !isURI(actor.Endpoint):
		return fmt.Errorf("%s.endpoint must be a URI", path)
	default:
		return nil
	}
}

func errArtifactRef(path string, ref spec.ArtifactRef) error {
	switch {
	case !IsUUID(ref.ArtifactID):
		return fmt.Errorf("%s.artifact_id must be a UUID", path)
	case strings.TrimSpace(ref.Kind) == "":
		return fmt.Errorf("%s.kind is required", path)
	case !isURI(ref.URI):
		return fmt.Errorf("%s.uri must be a URI", path)
	case ref.Version < 0:
		return fmt.Errorf("%s.version must be >= 1 when provided", path)
	default:
		return nil
	}
}

func isRFC3339(value string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	_, err := time.Parse(time.RFC3339, value)
	return err == nil
}

func isURI(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	return err == nil && parsed.Scheme != ""
}

func inSet(value string, allowed ...string) bool {
	for _, item := range allowed {
		if value == item {
			return true
		}
	}
	return false
}
