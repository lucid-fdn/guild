create table if not exists evaluation_jobs (
  evaluation_job_id uuid primary key,
  status text not null,
  payload jsonb not null,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

create index if not exists idx_evaluation_jobs_status_updated_at
  on evaluation_jobs (status, updated_at);
