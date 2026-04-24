# Server

Guild server is the bootstrap control plane for v1.

What exists today:

- file-backed storage for local development
- Postgres storage behind the same service interfaces
- runtime migration execution for Postgres
- local object-storage backend that mirrors artifact metadata
- durable evaluation jobs
- in-process evaluator worker for queued replay suites
- seeded institution, taskpack, DRI binding, artifact, and promotion record fixtures
- `GET/POST` APIs for the four public spec objects
- runtime validation aligned with the public schemas
- referential integrity checks before records are persisted
- Go unit tests and local smoke coverage

Run it from the repository root:

```bash
make run-server
```

Use Postgres storage:

```bash
GUILD_STORAGE_DRIVER=postgres \
GUILD_DATABASE_URL='postgres://guild:guild@localhost:5432/guild?sslmode=disable' \
make run-server
```

Runtime storage environment:

- `GUILD_STORAGE_DRIVER=file|postgres`
- `GUILD_DATA_DIR=./data` for file metadata
- `GUILD_DATABASE_URL=...` for Postgres metadata
- `GUILD_MIGRATIONS_DIR=server/migrations`
- `GUILD_OBJECT_DIR=./data/objects` for local artifact metadata objects
- `GUILD_WORKER_ENABLED=true`
- `GUILD_WORKER_INTERVAL=1s`

Verify it:

```bash
make test-go
make smoke
```

Recommended internal module layout:

```text
server/
  cmd/
    guildd/
  internal/
    app/
    auth/
    tasks/
    dri/
    workflow/
    policy/
    approvals/
    artifacts/
    runs/
    commons/
    telemetry/
    storage/
  pkg/
    api/
    events/
    schemas/
```

Responsibilities:

- validate incoming spec objects
- persist metadata
- run metadata migrations
- mirror artifact metadata into object storage
- expose control-plane APIs
- emit events
- enforce policy hooks
- coordinate replay and promotion workflows
- run queued evaluation jobs off the request path
