# Deploy

Local development stack includes:

- Postgres
- Redis
- NATS JetStream
- MinIO
- OpenTelemetry Collector

Production deployment should start as a single backend process plus managed dependencies.
Do not split into microservices before clear scale evidence exists.

Current runtime support:

- `GUILD_STORAGE_DRIVER=postgres` enables Postgres metadata storage.
- `GUILD_DATABASE_URL` points to the managed Postgres database.
- `GUILD_MIGRATIONS_DIR` is executed at startup.
- `GUILD_OBJECT_DIR` enables the local object backend for artifact metadata mirroring.

Redis and NATS are intentionally still optional/later. They should be added for async event fanout and evaluator queues once the single-process control plane has real load.
