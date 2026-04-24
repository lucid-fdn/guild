package spec

type Taskpack struct {
	SchemaVersion    string               `json:"schema_version"`
	TaskpackID       string               `json:"taskpack_id"`
	InstitutionID    string               `json:"institution_id,omitempty"`
	ParentTaskpackID string               `json:"parent_taskpack_id,omitempty"`
	Title            string               `json:"title"`
	Objective        string               `json:"objective"`
	ProblemStatement string               `json:"problem_statement,omitempty"`
	TaskType         string               `json:"task_type"`
	Priority         string               `json:"priority"`
	RequestedBy      ActorRef             `json:"requested_by"`
	RoleHint         string               `json:"role_hint,omitempty"`
	Inputs           []ArtifactRef        `json:"inputs,omitempty"`
	References       []string             `json:"references,omitempty"`
	Labels           []string             `json:"labels,omitempty"`
	ContextBudget    ContextBudget        `json:"context_budget"`
	Constraints      *TaskConstraints     `json:"constraints,omitempty"`
	Permissions      Permissions          `json:"permissions"`
	Acceptance       []AcceptanceCriteria `json:"acceptance_criteria"`
	DelegationPolicy *DelegationPolicy    `json:"delegation_policy,omitempty"`
	CreatedAt        string               `json:"created_at"`
	TraceID          string               `json:"trace_id,omitempty"`
}

type DriBinding struct {
	SchemaVersion    string            `json:"schema_version"`
	DriBindingID     string            `json:"dri_binding_id"`
	TaskpackID       string            `json:"taskpack_id"`
	Owner            ActorRef          `json:"owner"`
	Reviewers        []ActorRef        `json:"reviewers,omitempty"`
	Specialists      []ActorRef        `json:"specialists,omitempty"`
	Approvers        []ActorRef        `json:"approvers,omitempty"`
	EscalationPolicy *EscalationPolicy `json:"escalation_policy,omitempty"`
	ApprovalPolicy   *ApprovalPolicy   `json:"approval_policy,omitempty"`
	VisibilityPolicy *VisibilityPolicy `json:"visibility_policy,omitempty"`
	Status           string            `json:"status"`
	CreatedAt        string            `json:"created_at"`
}

type Artifact struct {
	SchemaVersion          string           `json:"schema_version"`
	ArtifactID             string           `json:"artifact_id"`
	TaskpackID             string           `json:"taskpack_id"`
	ParentArtifactIDs      []string         `json:"parent_artifact_ids,omitempty"`
	Kind                   string           `json:"kind"`
	Title                  string           `json:"title"`
	Summary                string           `json:"summary,omitempty"`
	Producer               ActorRef         `json:"producer"`
	Storage                ArtifactStorage  `json:"storage"`
	Lineage                *ArtifactLineage `json:"lineage,omitempty"`
	EvaluationState        *EvaluationState `json:"evaluation_state,omitempty"`
	Labels                 []string         `json:"labels,omitempty"`
	SecurityClassification string           `json:"security_classification,omitempty"`
	Version                int              `json:"version"`
	CreatedAt              string           `json:"created_at"`
}

type PromotionRecord struct {
	SchemaVersion  string        `json:"schema_version"`
	PromotionID    string        `json:"promotion_record_id"`
	InstitutionID  string        `json:"institution_id"`
	CandidateKind  string        `json:"candidate_kind"`
	CandidateRef   ArtifactRef   `json:"candidate_ref"`
	SourceRunIDs   []string      `json:"source_run_ids,omitempty"`
	BenchmarkSuite string        `json:"benchmark_suite,omitempty"`
	Metrics        []MetricDelta `json:"metrics,omitempty"`
	Decision       string        `json:"decision"`
	DecisionReason string        `json:"decision_reason,omitempty"`
	Deciders       []ActorRef    `json:"deciders,omitempty"`
	AcceptedScope  string        `json:"accepted_scope,omitempty"`
	DecidedAt      string        `json:"decided_at"`
}

type GovernancePolicy struct {
	SchemaVersion string       `json:"schema_version"`
	PolicyID      string       `json:"policy_id"`
	InstitutionID string       `json:"institution_id"`
	Name          string       `json:"name"`
	Description   string       `json:"description,omitempty"`
	AppliesTo     []string     `json:"applies_to"`
	Rules         []PolicyRule `json:"rules"`
	CreatedBy     ActorRef     `json:"created_by"`
	CreatedAt     string       `json:"created_at"`
}

type PolicyRule struct {
	RuleID       string   `json:"rule_id"`
	Effect       string   `json:"effect"`
	Condition    string   `json:"condition"`
	Requires     []string `json:"requires,omitempty"`
	MinApprovals int      `json:"min_approvals,omitempty"`
}

type ApprovalRequest struct {
	SchemaVersion     string             `json:"schema_version"`
	ApprovalID        string             `json:"approval_id"`
	TaskpackID        string             `json:"taskpack_id"`
	PolicyID          string             `json:"policy_id,omitempty"`
	RequestedBy       ActorRef           `json:"requested_by"`
	Reason            string             `json:"reason"`
	RequiredApprovals int                `json:"required_approvals"`
	Approvals         []ApprovalDecision `json:"approvals,omitempty"`
	Status            string             `json:"status"`
	CreatedAt         string             `json:"created_at"`
	DecidedAt         string             `json:"decided_at,omitempty"`
}

type ApprovalDecision struct {
	Actor     ActorRef `json:"actor"`
	Decision  string   `json:"decision"`
	Reason    string   `json:"reason,omitempty"`
	DecidedAt string   `json:"decided_at"`
}

type PromotionGate struct {
	SchemaVersion    string        `json:"schema_version"`
	GateID           string        `json:"gate_id"`
	InstitutionID    string        `json:"institution_id"`
	Name             string        `json:"name"`
	CandidateKinds   []string      `json:"candidate_kinds"`
	RequiredMetrics  []MetricDelta `json:"required_metrics"`
	MinReplayRuns    int           `json:"min_replay_runs"`
	RequiresApproval bool          `json:"requires_approval"`
	CreatedAt        string        `json:"created_at"`
}

type CommonsEntry struct {
	SchemaVersion     string      `json:"schema_version"`
	CommonsEntryID    string      `json:"commons_entry_id"`
	InstitutionID     string      `json:"institution_id"`
	PromotionRecordID string      `json:"promotion_record_id"`
	Title             string      `json:"title"`
	Summary           string      `json:"summary"`
	ArtifactRef       ArtifactRef `json:"artifact_ref"`
	Scope             string      `json:"scope"`
	Status            string      `json:"status"`
	CreatedAt         string      `json:"created_at"`
}

type ReplayBundle struct {
	SchemaVersion    string            `json:"schema_version"`
	RootTaskpackID   string            `json:"root_taskpack_id,omitempty"`
	Taskpack         Taskpack          `json:"taskpack"`
	Taskpacks        []Taskpack        `json:"taskpacks,omitempty"`
	DriBindings      []DriBinding      `json:"dri_bindings"`
	Artifacts        []Artifact        `json:"artifacts"`
	PromotionRecords []PromotionRecord `json:"promotion_records"`
}

type WorkspaceConstitution struct {
	SchemaVersion   string               `json:"schema_version" yaml:"version"`
	Workspace       string               `json:"workspace" yaml:"workspace"`
	Mission         string               `json:"mission" yaml:"mission"`
	Defaults        WorkspaceDefaults    `json:"defaults" yaml:"defaults"`
	TaskSources     []TaskSource         `json:"task_sources,omitempty" yaml:"task_sources,omitempty"`
	Scope           WorkspaceScope       `json:"scope" yaml:"scope"`
	ApprovalRules   []WorkspaceRule      `json:"approval_rules,omitempty" yaml:"approval_rules,omitempty"`
	SuccessCriteria []string             `json:"success_criteria,omitempty" yaml:"success_criteria,omitempty"`
	Escalation      *WorkspaceEscalation `json:"escalation,omitempty" yaml:"escalation,omitempty"`
}

type WorkspaceDefaults struct {
	MaxRuntimeMinutes   int     `json:"max_runtime_minutes" yaml:"max_runtime_minutes"`
	MaxCostUSD          float64 `json:"max_cost_usd" yaml:"max_cost_usd"`
	ContextBudgetTokens int     `json:"context_budget_tokens" yaml:"context_budget_tokens"`
}

type TaskSource struct {
	Type  string `json:"type" yaml:"type"`
	Repo  string `json:"repo,omitempty" yaml:"repo,omitempty"`
	Query string `json:"query,omitempty" yaml:"query,omitempty"`
	Path  string `json:"path,omitempty" yaml:"path,omitempty"`
}

type WorkspaceScope struct {
	Writable  []string `json:"writable,omitempty" yaml:"writable,omitempty"`
	Forbidden []string `json:"forbidden,omitempty" yaml:"forbidden,omitempty"`
}

type WorkspaceRule struct {
	When    string `json:"when" yaml:"when"`
	Require string `json:"require" yaml:"require"`
}

type WorkspaceEscalation struct {
	DefaultOwner string              `json:"default_owner,omitempty" yaml:"default_owner,omitempty"`
	Channels     []EscalationChannel `json:"channels,omitempty" yaml:"channels,omitempty"`
}

type EscalationChannel struct {
	Type string `json:"type" yaml:"type"`
}

type ContextPack struct {
	SchemaVersion  string   `json:"schema_version"`
	MandateID      string   `json:"mandate_id"`
	Role           string   `json:"role"`
	BudgetTokens   int      `json:"budget_tokens"`
	MustRead       []string `json:"must_read,omitempty"`
	MayRead        []string `json:"may_read,omitempty"`
	MayWrite       []string `json:"may_write,omitempty"`
	Forbidden      []string `json:"forbidden,omitempty"`
	Summary        string   `json:"summary"`
	ProofRequired  []string `json:"proof_required,omitempty"`
	OmittedReasons []string `json:"omitted_reasons,omitempty"`
}

type PreflightDecision struct {
	SchemaVersion    string   `json:"schema_version"`
	MandateID        string   `json:"mandate_id"`
	Action           string   `json:"action"`
	Path             string   `json:"path,omitempty"`
	Command          string   `json:"command,omitempty"`
	Decision         string   `json:"decision"`
	Reason           string   `json:"reason"`
	ApprovalRequired bool     `json:"approval_required"`
	MatchedRules     []string `json:"matched_rules,omitempty"`
}

type Institution struct {
	SchemaVersion string `json:"schema_version"`
	InstitutionID string `json:"institution_id"`
	Name          string `json:"name"`
	CreatedAt     string `json:"created_at"`
}

type ActorRef struct {
	ActorID      string `json:"actor_id"`
	ActorType    string `json:"actor_type"`
	DisplayName  string `json:"display_name"`
	Orchestrator string `json:"orchestrator,omitempty"`
	Model        string `json:"model,omitempty"`
	Endpoint     string `json:"endpoint,omitempty"`
}

type ArtifactRef struct {
	ArtifactID string `json:"artifact_id"`
	Kind       string `json:"kind"`
	URI        string `json:"uri"`
	Version    int    `json:"version,omitempty"`
}

type ContextBudget struct {
	MaxInputTokens     int    `json:"max_input_tokens"`
	MaxOutputTokens    int    `json:"max_output_tokens"`
	ContextStrategy    string `json:"context_strategy"`
	MaxArtifactsInline int    `json:"max_artifacts_inline,omitempty"`
}

type TaskConstraints struct {
	MustNot        []string `json:"must_not,omitempty"`
	MustInclude    []string `json:"must_include,omitempty"`
	DeadlineAt     string   `json:"deadline_at,omitempty"`
	CostCeilingUSD float64  `json:"cost_ceiling_usd,omitempty"`
	ToolAllowlist  []string `json:"tool_allowlist,omitempty"`
}

type Permissions struct {
	AllowNetwork       bool     `json:"allow_network"`
	AllowShell         bool     `json:"allow_shell"`
	AllowExternalWrite bool     `json:"allow_external_write"`
	ApprovalMode       string   `json:"approval_mode"`
	Scopes             []string `json:"scopes,omitempty"`
}

type AcceptanceCriteria struct {
	CriterionID      string `json:"criterion_id"`
	Description      string `json:"description"`
	Required         bool   `json:"required"`
	VerificationHint string `json:"verification_hint,omitempty"`
}

type DelegationPolicy struct {
	AllowSubtasks       bool `json:"allow_subtasks"`
	MaxDepth            int  `json:"max_depth"`
	MaxParallelSubtasks int  `json:"max_parallel_subtasks"`
}

type EscalationPolicy struct {
	Mode                 string     `json:"mode"`
	EscalateAfterSeconds int        `json:"escalate_after_seconds,omitempty"`
	Targets              []ActorRef `json:"targets,omitempty"`
}

type ApprovalPolicy struct {
	RequiredFor  []string `json:"required_for,omitempty"`
	MinApprovals int      `json:"min_approvals,omitempty"`
}

type VisibilityPolicy struct {
	Mode    string     `json:"mode"`
	Readers []ActorRef `json:"readers,omitempty"`
}

type ArtifactStorage struct {
	URI       string `json:"uri"`
	MimeType  string `json:"mime_type"`
	SHA256    string `json:"sha256,omitempty"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
}

type ArtifactLineage struct {
	TraceID           string   `json:"trace_id,omitempty"`
	RunID             string   `json:"run_id,omitempty"`
	SourceTaskpackIDs []string `json:"source_taskpack_ids,omitempty"`
}

type EvaluationState struct {
	Status         string  `json:"status,omitempty"`
	Score          float64 `json:"score,omitempty"`
	BenchmarkSuite string  `json:"benchmark_suite,omitempty"`
}

type MetricDelta struct {
	Name      string  `json:"name"`
	Before    float64 `json:"before"`
	After     float64 `json:"after"`
	Direction string  `json:"direction"`
	Unit      string  `json:"unit,omitempty"`
}
