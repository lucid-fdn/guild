# Starter Issues

## Good First Issue: Add an OpenAI Agents SDK adapter

Build the second real orchestrator adapter beside `@guild/adapter-langgraph`.

Acceptance criteria:

- Converts an OpenAI Agents SDK run into a Taskpack, DRI Binding, and Artifact.
- Includes local tests.
- Adds an adapter conformance profile.

## Good First Issue: Add managed object storage backend

Implement an S3-compatible artifact metadata/payload backend behind the current object-store interface.

Acceptance criteria:

- Supports MinIO locally.
- Keeps local object store as default.
- Documents required environment variables.

## Good First Issue: Add commons detail UI

Create a `/commons` route that lists commons entries and links them to promotion records and artifacts.

Acceptance criteria:

- Works with live API and offline demo data.
- Shows scope, status, artifact ref, and promotion record.

## Good First Issue: Add Redis/NATS evaluator queue backend

Add distributed evaluator queue support without replacing the current durable in-process worker.

Acceptance criteria:

- Worker leasing prevents duplicate execution.
- Failed jobs are retryable.
- Dead-letter state is visible through the API.

## Good First Issue: Add promotion gate evaluator

Make promotion gates executable instead of descriptive.

Acceptance criteria:

- Evaluates metric thresholds against promotion records.
- Emits a clear pass/fail result.
- Blocks commons entry creation when gates fail.
