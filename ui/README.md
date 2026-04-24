# UI

Guild UI is the experience plane.

What exists today:

- Next.js App Router shell
- live/offline dashboard
- task detail route
- DRI ownership graph
- artifact graph
- artifact viewer with provenance and storage metadata
- replay timeline
- approval inbox with policy/scopes/owner context
- commons / promotion records panel
- production build and TypeScript checks in CI

Run it from the repository root:

```bash
make run-ui
```

Primary views:

- task list
- task detail
- DRI ownership view
- artifact detail
- replay timeline
- approvals inbox
- commons browser

Configuration:

- `GUILD_API_BASE_URL` points the UI at a Guild control plane.
- When the API is unavailable, the UI renders demo institution data so the landing experience still builds and explains the product.

Recommended stack:

- Next.js
- React
- TypeScript
- AG-UI-compatible event ingestion where useful
