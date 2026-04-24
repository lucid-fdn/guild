package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/guild-labs/guild/pkg/spec"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStore(databaseURL string) (*PostgresStore, error) {
	if strings.TrimSpace(databaseURL) == "" {
		return nil, fmt.Errorf("postgres storage requires GUILD_DATABASE_URL")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &PostgresStore{db: db}, nil
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) RunMigrations(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if _, err := s.db.Exec(string(data)); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
	}
	return nil
}

func (s *PostgresStore) Put(collection, id string, value any) error {
	if err := validateKey(collection, id); err != nil {
		return err
	}
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	switch collection {
	case "institutions":
		var institution spec.Institution
		if err := json.Unmarshal(payload, &institution); err != nil {
			return err
		}
		_, err = s.db.Exec(`
			insert into institutions (institution_id, schema_version, name, payload, created_at)
			values ($1, $2, $3, $4::jsonb, $5)
			on conflict (institution_id) do update set
				schema_version = excluded.schema_version,
				name = excluded.name,
				payload = excluded.payload,
				created_at = excluded.created_at`,
			institution.InstitutionID, institution.SchemaVersion, institution.Name, string(payload), institution.CreatedAt)
	case "taskpacks":
		var taskpack spec.Taskpack
		if err := json.Unmarshal(payload, &taskpack); err != nil {
			return err
		}
		_, err = s.db.Exec(`
			insert into taskpacks (
				taskpack_id, institution_id, parent_taskpack_id, schema_version, title, objective,
				task_type, priority, requested_by, role_hint, context_budget, permissions,
				acceptance_criteria, payload, created_at
			)
			values ($1, nullif($2, '')::uuid, nullif($3, '')::uuid, $4, $5, $6, $7, $8, $9::jsonb, nullif($10, ''), $11::jsonb, $12::jsonb, $13::jsonb, $14::jsonb, $15)
			on conflict (taskpack_id) do update set
				institution_id = excluded.institution_id,
				parent_taskpack_id = excluded.parent_taskpack_id,
				schema_version = excluded.schema_version,
				title = excluded.title,
				objective = excluded.objective,
				task_type = excluded.task_type,
				priority = excluded.priority,
				requested_by = excluded.requested_by,
				role_hint = excluded.role_hint,
				context_budget = excluded.context_budget,
				permissions = excluded.permissions,
				acceptance_criteria = excluded.acceptance_criteria,
				payload = excluded.payload,
				created_at = excluded.created_at`,
			taskpack.TaskpackID, taskpack.InstitutionID, taskpack.ParentTaskpackID, taskpack.SchemaVersion, taskpack.Title, taskpack.Objective,
			taskpack.TaskType, taskpack.Priority, mustJSON(taskpack.RequestedBy), taskpack.RoleHint, mustJSON(taskpack.ContextBudget), mustJSON(taskpack.Permissions),
			mustJSON(taskpack.Acceptance), string(payload), taskpack.CreatedAt)
	case "dri-bindings":
		var binding spec.DriBinding
		if err := json.Unmarshal(payload, &binding); err != nil {
			return err
		}
		_, err = s.db.Exec(`
			insert into dri_bindings (
				dri_binding_id, taskpack_id, owner, reviewers, specialists, approvers,
				escalation_policy, approval_policy, visibility_policy, status, payload, created_at
			)
			values ($1, $2, $3::jsonb, $4::jsonb, $5::jsonb, $6::jsonb, $7::jsonb, $8::jsonb, $9::jsonb, $10, $11::jsonb, $12)
			on conflict (dri_binding_id) do update set
				taskpack_id = excluded.taskpack_id,
				owner = excluded.owner,
				reviewers = excluded.reviewers,
				specialists = excluded.specialists,
				approvers = excluded.approvers,
				escalation_policy = excluded.escalation_policy,
				approval_policy = excluded.approval_policy,
				visibility_policy = excluded.visibility_policy,
				status = excluded.status,
				payload = excluded.payload,
				created_at = excluded.created_at`,
			binding.DriBindingID, binding.TaskpackID, mustJSON(binding.Owner), mustJSON(binding.Reviewers), mustJSON(binding.Specialists), mustJSON(binding.Approvers),
			nullableJSON(binding.EscalationPolicy), nullableJSON(binding.ApprovalPolicy), nullableJSON(binding.VisibilityPolicy), binding.Status, string(payload), binding.CreatedAt)
	case "artifacts":
		var artifact spec.Artifact
		if err := json.Unmarshal(payload, &artifact); err != nil {
			return err
		}
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		if _, err = tx.Exec(`
			insert into artifacts (artifact_id, taskpack_id, kind, title, producer, storage_uri, mime_type, version, payload, created_at)
			values ($1, $2, $3, $4, $5::jsonb, $6, $7, $8, $9::jsonb, $10)
			on conflict (artifact_id) do update set
				taskpack_id = excluded.taskpack_id,
				kind = excluded.kind,
				title = excluded.title,
				producer = excluded.producer,
				storage_uri = excluded.storage_uri,
				mime_type = excluded.mime_type,
				version = excluded.version,
				payload = excluded.payload,
				created_at = excluded.created_at`,
			artifact.ArtifactID, artifact.TaskpackID, artifact.Kind, artifact.Title, mustJSON(artifact.Producer), artifact.Storage.URI, artifact.Storage.MimeType, artifact.Version, string(payload), artifact.CreatedAt); err != nil {
			return err
		}
		if _, err = tx.Exec(`delete from artifact_edges where child_artifact_id = $1`, artifact.ArtifactID); err != nil {
			return err
		}
		for _, parentID := range artifact.ParentArtifactIDs {
			if _, err = tx.Exec(`insert into artifact_edges (parent_artifact_id, child_artifact_id) values ($1, $2) on conflict do nothing`, parentID, artifact.ArtifactID); err != nil {
				return err
			}
		}
		return tx.Commit()
	case "promotion-records":
		var record spec.PromotionRecord
		if err := json.Unmarshal(payload, &record); err != nil {
			return err
		}
		_, err = s.db.Exec(`
			insert into promotion_records (
				promotion_record_id, institution_id, candidate_kind, candidate_ref, decision,
				decision_reason, accepted_scope, payload, decided_at
			)
			values ($1, $2, $3, $4::jsonb, $5, nullif($6, ''), nullif($7, ''), $8::jsonb, $9)
			on conflict (promotion_record_id) do update set
				institution_id = excluded.institution_id,
				candidate_kind = excluded.candidate_kind,
				candidate_ref = excluded.candidate_ref,
				decision = excluded.decision,
				decision_reason = excluded.decision_reason,
				accepted_scope = excluded.accepted_scope,
				payload = excluded.payload,
				decided_at = excluded.decided_at`,
			record.PromotionID, record.InstitutionID, record.CandidateKind, mustJSON(record.CandidateRef), record.Decision, record.DecisionReason, record.AcceptedScope, string(payload), record.DecidedAt)
	case "evaluation-jobs":
		_, err = s.db.Exec(`
			insert into evaluation_jobs (evaluation_job_id, status, payload, created_at, updated_at)
			values ($1, coalesce($2, 'queued'), $3::jsonb, coalesce(($3::jsonb ->> 'created_at')::timestamptz, now()), coalesce(($3::jsonb ->> 'updated_at')::timestamptz, now()))
			on conflict (evaluation_job_id) do update set
				status = excluded.status,
				payload = excluded.payload,
				updated_at = excluded.updated_at`,
			id, jsonStringField(payload, "status"), string(payload))
	case "governance-policies":
		_, err = s.db.Exec(`insert into governance_policies (policy_id, institution_id, payload, created_at) values ($1, ($2::jsonb ->> 'institution_id')::uuid, $2::jsonb, ($2::jsonb ->> 'created_at')::timestamptz) on conflict (policy_id) do update set institution_id = excluded.institution_id, payload = excluded.payload`, id, string(payload))
	case "approval-requests":
		_, err = s.db.Exec(`insert into approval_requests (approval_id, taskpack_id, status, payload, created_at, updated_at) values ($1, ($2::jsonb ->> 'taskpack_id')::uuid, $2::jsonb ->> 'status', $2::jsonb, ($2::jsonb ->> 'created_at')::timestamptz, now()) on conflict (approval_id) do update set status = excluded.status, payload = excluded.payload, updated_at = now()`, id, string(payload))
	case "promotion-gates":
		_, err = s.db.Exec(`insert into promotion_gates (gate_id, institution_id, payload, created_at) values ($1, ($2::jsonb ->> 'institution_id')::uuid, $2::jsonb, ($2::jsonb ->> 'created_at')::timestamptz) on conflict (gate_id) do update set institution_id = excluded.institution_id, payload = excluded.payload`, id, string(payload))
	case "commons-entries":
		_, err = s.db.Exec(`insert into commons_entries (commons_entry_id, institution_id, promotion_record_id, status, payload, created_at) values ($1, ($2::jsonb ->> 'institution_id')::uuid, ($2::jsonb ->> 'promotion_record_id')::uuid, $2::jsonb ->> 'status', $2::jsonb, ($2::jsonb ->> 'created_at')::timestamptz) on conflict (commons_entry_id) do update set status = excluded.status, payload = excluded.payload`, id, string(payload))
	}
	return err
}

func (s *PostgresStore) Get(collection, id string, dest any) error {
	if err := validateKey(collection, id); err != nil {
		return err
	}
	table, err := tableForCollection(collection)
	if err != nil {
		return err
	}
	idColumn := idColumnForCollection(collection)
	var payload []byte
	err = s.db.QueryRow(fmt.Sprintf(`select payload from %s where %s = $1`, table, idColumn), id).Scan(&payload)
	if err == sql.ErrNoRows {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, dest)
}

func (s *PostgresStore) List(collection string, dest any) error {
	table, err := tableForCollection(collection)
	if err != nil {
		return err
	}
	idColumn := idColumnForCollection(collection)
	rows, err := s.db.Query(fmt.Sprintf(`select payload from %s order by %s`, table, idColumn))
	if err != nil {
		return err
	}
	defer rows.Close()
	items := make([]json.RawMessage, 0)
	for rows.Next() {
		var payload []byte
		if err := rows.Scan(&payload); err != nil {
			return err
		}
		items = append(items, payload)
	}
	if err := rows.Err(); err != nil {
		return err
	}
	payload, err := json.Marshal(items)
	if err != nil {
		return err
	}
	return json.Unmarshal(payload, dest)
}

func tableForCollection(collection string) (string, error) {
	switch collection {
	case "institutions":
		return "institutions", nil
	case "taskpacks":
		return "taskpacks", nil
	case "dri-bindings":
		return "dri_bindings", nil
	case "artifacts":
		return "artifacts", nil
	case "promotion-records":
		return "promotion_records", nil
	case "evaluation-jobs":
		return "evaluation_jobs", nil
	case "governance-policies":
		return "governance_policies", nil
	case "approval-requests":
		return "approval_requests", nil
	case "promotion-gates":
		return "promotion_gates", nil
	case "commons-entries":
		return "commons_entries", nil
	default:
		return "", fmt.Errorf("unknown collection %q", collection)
	}
}

func idColumnForCollection(collection string) string {
	switch collection {
	case "institutions":
		return "institution_id"
	case "taskpacks":
		return "taskpack_id"
	case "dri-bindings":
		return "dri_binding_id"
	case "artifacts":
		return "artifact_id"
	case "promotion-records":
		return "promotion_record_id"
	case "evaluation-jobs":
		return "evaluation_job_id"
	case "governance-policies":
		return "policy_id"
	case "approval-requests":
		return "approval_id"
	case "promotion-gates":
		return "gate_id"
	case "commons-entries":
		return "commons_entry_id"
	default:
		return "id"
	}
}

func jsonStringField(data []byte, key string) string {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}
	value, _ := payload[key].(string)
	return value
}

func mustJSON(value any) string {
	data, _ := json.Marshal(value)
	return string(data)
}

func nullableJSON(value any) any {
	if value == nil {
		return nil
	}
	return mustJSON(value)
}
