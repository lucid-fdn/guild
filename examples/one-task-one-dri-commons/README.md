# One Task, One DRI, Commons

This is the canonical Guild example.

It demonstrates the full institution loop:

- one bounded `Taskpack`
- one accountable `DRI Binding`
- durable `Artifact` records
- recursive `Replay Bundle` export
- benchmark-backed `Promotion Record`
- human `Approval Request`
- `Promotion Gate`
- accepted `Commons Entry`

## Run It

From the repository root:

```bash
examples/one-task-one-dri-commons/run.sh
```

Expected output:

```text
simulation-ok http://127.0.0.1:<port>
```

The script delegates to the same simulation used by CI and `make release-check`.

## Why This Example Matters

Most multi-agent demos stop at collaboration.

This example shows the institutional layer:

- ownership is explicit
- outputs are artifacts
- replay is portable evidence
- learning must pass a gate
- humans approve promotion
- accepted knowledge enters the commons

That is the product wedge: Guild is not the orchestrator. Guild is the system of record for accountability and learning above orchestrators.

## Manual Exploration

Start the server:

```bash
make run-server
```

Then explore the records:

```bash
curl -fsS http://localhost:8080/api/v1/taskpacks
curl -fsS http://localhost:8080/api/v1/dri-bindings
curl -fsS http://localhost:8080/api/v1/artifacts
curl -fsS http://localhost:8080/api/v1/promotion-records
curl -fsS http://localhost:8080/api/v1/governance-policies
curl -fsS http://localhost:8080/api/v1/approval-requests
curl -fsS http://localhost:8080/api/v1/promotion-gates
curl -fsS http://localhost:8080/api/v1/commons-entries
```

Export the replay bundle:

```bash
go run ./cli/cmd/guild replay-export \
  --base-url http://localhost:8080 \
  --taskpack-id 4e1fe00c-6303-453c-8cb6-2c34f84896e4 \
  --file replay.json
```

