# Contributing

Thanks for helping build Guild.

## Development principles

- Keep the public spec surface small and stable.
- Preserve orchestrator-agnostic design.
- Prefer additive changes over breaking changes.
- Treat replayability and provenance as first-class concerns.

## Local development

1. Copy `.env.example` if needed.
2. Run `make dev-up`.
3. Start the server from `server/cmd/guildd`.
4. Validate schemas with `make lint-spec`.

## Commit guidance

- Keep schema changes and example changes together.
- When changing public spec objects, update docs and examples in the same change.
- Add tests or fixtures for state-machine changes when the server implementation grows.
