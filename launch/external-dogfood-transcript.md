# External Dogfood Transcript

This is the clean-repo proof path for `v0.1.0-alpha.2`.

## Canonical Links

- Release: https://github.com/lucid-fdn/guild/releases/tag/v0.1.0-alpha.2
- External repo: https://github.com/lucid-fdn/guild-agentdesk-dogfood
- Issue: https://github.com/lucid-fdn/guild-agentdesk-dogfood/issues/1
- PR: https://github.com/lucid-fdn/guild-agentdesk-dogfood/pull/2
- Agent Work Contract comment: https://github.com/lucid-fdn/guild-agentdesk-dogfood/pull/2#issuecomment-4318664382
- CI run: https://github.com/lucid-fdn/guild-agentdesk-dogfood/actions/runs/24928040619
- Mandate ID: `951ffef3-36e0-5816-9c58-ecfcbca04f15`

## Captured Flow

1. Created a clean external repo: `lucid-fdn/guild-agentdesk-dogfood`.
2. Installed Guild from the public tag:

```bash
go install github.com/lucid-fdn/guild/cli/cmd/guild@v0.1.0-alpha.2
```

3. Bootstrapped GitHub issue intake and CI:

```bash
GITHUB_TOKEN="$(gh auth token)" \
guild agentdesk bootstrap github --repo lucid-fdn/guild-agentdesk-dogfood
```

4. Created issue #1 labeled `agent:ready`.
5. Ingested the issue:

```bash
GITHUB_TOKEN="$(gh auth token)" \
guild agentdesk next --source github --repo lucid-fdn/guild-agentdesk-dogfood
```

6. Claimed the mandate:

```bash
guild agentdesk claim \
  --id 951ffef3-36e0-5816-9c58-ecfcbca04f15 \
  --agent codex-dogfood
```

7. Checked bounded context and preflight:

```bash
guild agentdesk context compile \
  --id 951ffef3-36e0-5816-9c58-ecfcbca04f15 \
  --role coder

guild agentdesk preflight \
  --id 951ffef3-36e0-5816-9c58-ecfcbca04f15 \
  --action write \
  --path docs/dogfood-proof.md
```

8. Attached proof and handoff:

```bash
guild agentdesk proof add \
  --id 951ffef3-36e0-5816-9c58-ecfcbca04f15 \
  --kind test_report \
  --path test-results.xml

guild agentdesk proof add \
  --id 951ffef3-36e0-5816-9c58-ecfcbca04f15 \
  --kind changed_files \
  --path changed-files.json

guild agentdesk handoff create \
  --id 951ffef3-36e0-5816-9c58-ecfcbca04f15 \
  --to reviewer \
  --summary "Ready for review: public alpha dogfood loop completed from clean external repo."
```

9. Verified and exported replay:

```bash
guild agentdesk verify --id 951ffef3-36e0-5816-9c58-ecfcbca04f15
guild agentdesk replay export \
  --id 951ffef3-36e0-5816-9c58-ecfcbca04f15 \
  --file .agentdesk/replay/951ffef3-36e0-5816-9c58-ecfcbca04f15.json
```

10. Opened PR #2.
11. Generated Agent Work Contract workflow installed `v0.1.0-alpha.2`, verified the run, and posted a passing PR comment.
12. Merged PR #2, closing issue #1.

## Result

The public alpha path worked from a clean external repo:

- issue to mandate
- mandate to claim
- claim to proof
- proof to replay
- replay to PR check
- PR check to GitHub comment
