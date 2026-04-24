# 90-Second Demo Script

## 0:00 - Problem

"Most multi-agent demos show agents talking. Guild shows agents becoming a team."

## 0:10 - Task Ownership

"We start with a Taskpack: one bounded unit of work. Guild assigns exactly one DRI, so accountability is explicit."

Command:

```bash
make run-server
```

## 0:25 - Artifacts

"The DRI does not leave a transcript as the source of truth. It publishes durable artifacts with provenance."

Open:

```text
GET /api/v1/artifacts
```

## 0:40 - Replay

"Now we export the run as a replay bundle. This includes the task tree, DRI bindings, artifacts, and promotion evidence."

Command:

```bash
go run ./cli/cmd/guild replay-export \
  --base-url http://localhost:8080 \
  --taskpack-id 4e1fe00c-6303-453c-8cb6-2c34f84896e4
```

## 0:55 - Evaluation

"A replay suite creates benchmark evidence and a skill candidate. This is learning off the hot path."

Command:

```bash
go run ./cli/cmd/guild eval-submit \
  --base-url http://localhost:8080 \
  --suite examples/replay-suite.example.json \
  --wait
```

## 1:10 - Governance

"The candidate cannot enter the commons automatically. A policy, promotion gate, and approval request make institutional learning governed."

Open:

```text
GET /api/v1/governance-policies
GET /api/v1/promotion-gates
GET /api/v1/approval-requests
```

## 1:25 - Commons

"Once approved, the learning becomes a commons entry: reusable institutional memory."

Open:

```text
GET /api/v1/commons-entries
```

## 1:30 - Close

"Guild is not another orchestrator. It is the institutional layer for AI teams."
