create table if not exists institutions (
  institution_id uuid primary key,
  schema_version text not null default 'v1alpha1',
  name text not null,
  payload jsonb not null default '{}'::jsonb,
  created_at timestamptz not null default now()
);

create table if not exists taskpacks (
  taskpack_id uuid primary key,
  institution_id uuid null references institutions (institution_id),
  parent_taskpack_id uuid null references taskpacks (taskpack_id),
  schema_version text not null,
  title text not null,
  objective text not null,
  task_type text not null,
  priority text not null,
  requested_by jsonb not null,
  role_hint text null,
  context_budget jsonb not null,
  permissions jsonb not null,
  acceptance_criteria jsonb not null,
  payload jsonb not null,
  created_at timestamptz not null
);

create table if not exists dri_bindings (
  dri_binding_id uuid primary key,
  taskpack_id uuid not null references taskpacks (taskpack_id) on delete cascade,
  owner jsonb not null,
  reviewers jsonb not null default '[]'::jsonb,
  specialists jsonb not null default '[]'::jsonb,
  approvers jsonb not null default '[]'::jsonb,
  escalation_policy jsonb null,
  approval_policy jsonb null,
  visibility_policy jsonb null,
  status text not null,
  payload jsonb not null,
  created_at timestamptz not null
);

create table if not exists artifacts (
  artifact_id uuid primary key,
  taskpack_id uuid not null references taskpacks (taskpack_id) on delete cascade,
  kind text not null,
  title text not null,
  producer jsonb not null,
  storage_uri text not null,
  mime_type text not null,
  version integer not null,
  payload jsonb not null,
  created_at timestamptz not null
);

create table if not exists artifact_edges (
  parent_artifact_id uuid not null references artifacts (artifact_id) on delete cascade,
  child_artifact_id uuid not null references artifacts (artifact_id) on delete cascade,
  primary key (parent_artifact_id, child_artifact_id)
);

create table if not exists promotion_records (
  promotion_record_id uuid primary key,
  institution_id uuid not null references institutions (institution_id),
  candidate_kind text not null,
  candidate_ref jsonb not null,
  decision text not null,
  decision_reason text null,
  accepted_scope text null,
  payload jsonb not null,
  decided_at timestamptz not null
);

create index if not exists idx_taskpacks_institution_created_at
  on taskpacks (institution_id, created_at desc);

create index if not exists idx_dri_bindings_taskpack_id
  on dri_bindings (taskpack_id);

create index if not exists idx_artifacts_taskpack_id
  on artifacts (taskpack_id);

create index if not exists idx_promotion_records_institution_decided_at
  on promotion_records (institution_id, decided_at desc);
