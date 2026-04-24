create table if not exists governance_policies (
  policy_id uuid primary key,
  institution_id uuid not null references institutions (institution_id),
  payload jsonb not null,
  created_at timestamptz not null
);

create table if not exists approval_requests (
  approval_id uuid primary key,
  taskpack_id uuid not null references taskpacks (taskpack_id) on delete cascade,
  status text not null,
  payload jsonb not null,
  created_at timestamptz not null,
  updated_at timestamptz not null default now()
);

create table if not exists promotion_gates (
  gate_id uuid primary key,
  institution_id uuid not null references institutions (institution_id),
  payload jsonb not null,
  created_at timestamptz not null
);

create table if not exists commons_entries (
  commons_entry_id uuid primary key,
  institution_id uuid not null references institutions (institution_id),
  promotion_record_id uuid not null references promotion_records (promotion_record_id),
  status text not null,
  payload jsonb not null,
  created_at timestamptz not null
);

create index if not exists idx_approval_requests_status_created_at
  on approval_requests (status, created_at desc);

create index if not exists idx_commons_entries_institution_status
  on commons_entries (institution_id, status);
