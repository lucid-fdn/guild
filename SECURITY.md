# Security Policy

## Supported Versions

Guild is pre-1.0 alpha software.
Security fixes are applied to the default branch until tagged releases begin.

## Reporting A Vulnerability

Please do not open a public issue for a suspected vulnerability.

Email the maintainers or use GitHub private vulnerability reporting if it is enabled for the repository.
Include:

- affected commit or version
- reproduction steps
- expected impact
- any logs, proof artifacts, or replay bundles that help us understand the issue

We aim to acknowledge reports within 72 hours.

## Security Model

Guild treats agent work as an auditable contract:

- mandates define allowed scope
- preflight checks gate risky actions
- approvals capture human consent
- proof artifacts and replay bundles preserve evidence

Do not store secrets in `agentdesk.yaml`, `.agentdesk/`, replay bundles, proof artifacts, or issue comments.
