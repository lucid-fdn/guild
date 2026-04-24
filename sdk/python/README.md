# Python SDK

The Python SDK is a small, dependency-free client for Python-based agent stacks.

Primary goals:

- talk to the Guild control-plane API
- make adapters easy to write for Python orchestrators
- keep the first client boring enough to vendor or fork

## Usage

```python
from guild_client import GuildAPIError, GuildClient

client = GuildClient("http://localhost:8080")

try:
    status = client.get_status()
    taskpacks = client.list_taskpacks()
except GuildAPIError as error:
    print(error.status, error.message)
```

## Current Surface

- `get_status()`
- `list_taskpacks()`, `get_taskpack(id)`, `create_taskpack(payload)`
- `list_dri_bindings()`, `get_dri_binding(id)`, `create_dri_binding(payload)`
- `list_artifacts()`, `list_artifacts_for_taskpack(taskpack_id)`, `get_artifact(id)`, `create_artifact(payload)`
- `list_promotion_records()`, `get_promotion_record(id)`, `create_promotion_record(payload)`
- `export_replay_bundle(taskpack_id)`

The public JSON Schemas and OpenAPI document remain the source of truth. This
handwritten client is intentionally minimal until the SDK generation path is
locked.

## Verification

```bash
python3 -m compileall -q sdk/python/src
PYTHONPATH=sdk/python/src python3 -m unittest discover -s sdk/python/tests
```
