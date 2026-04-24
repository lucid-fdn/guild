# Evaluator Worker

The evaluator worker runs:

- replay jobs
- benchmark suites
- candidate promotion checks

It must stay off the hot path for task execution.

Evaluation workloads should be queue-driven, reproducible, and easy to rerun.

What exists now:

- `guild replay-suite` exports recursive replay bundles from a running control plane.
- The runner creates a `benchmark_result` artifact.
- The runner creates a `skill_candidate` artifact linked to the benchmark result.
- The runner opens a `Promotion Record` with `decision=needs_human_review`.
- `POST /api/v1/evaluation-jobs` queues a durable evaluation job.
- `GET /api/v1/evaluation-jobs` lists queued/running/succeeded/failed jobs.
- `POST /api/v1/evaluation-jobs/{id}/run` runs a job immediately for deterministic tests and local workflows.
- `guild eval-submit --wait` queues and runs a job through the control plane.
- the server starts an in-process worker by default and processes queued jobs off the request path.

Run locally:

```bash
go run ./cli/cmd/guild replay-suite \
  --base-url http://localhost:8080 \
  --suite examples/replay-suite.example.json
```

Queue through the worker path:

```bash
go run ./cli/cmd/guild eval-submit \
  --base-url http://localhost:8080 \
  --suite examples/replay-suite.example.json \
  --wait
```

Worker configuration:

- `GUILD_WORKER_ENABLED=true`
- `GUILD_WORKER_INTERVAL=1s`

Redis/NATS-backed queues are still the next scale hardening pass. The current worker is durable through the configured Guild store and keeps evaluation off the task execution hot path.
