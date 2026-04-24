alter table institutions
  add column if not exists schema_version text not null default 'v1alpha1';

alter table institutions
  add column if not exists payload jsonb not null default '{}'::jsonb;
