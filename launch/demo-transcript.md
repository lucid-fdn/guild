# Demo Transcript

This is the recorded launch proof path.

## Canonical Links

- Demo PR: https://github.com/lucid-fdn/guild/pull/9
- Agent Work Contract comment: https://github.com/lucid-fdn/guild/pull/9#issuecomment-4316590055
- Dogfood issue: https://github.com/lucid-fdn/guild/issues/3
- Mandate ID: `620d256b-fbe8-5e36-a773-85fe6ea936da`

## Capture Sequence

1. Human creates a GitHub issue labeled `agent:ready`.
2. Agent runs `guild agentdesk next --source github --repo lucid-fdn/guild`.
3. Guild creates a local mandate from the issue.
4. Agent runs `guild agentdesk claim --id 620d256b-fbe8-5e36-a773-85fe6ea936da --agent codex`.
5. Agent attaches proof artifacts: `test_report`, `changed_files`, and `handoff_summary`.
6. CI runs `guild agentdesk verify --github-report --id 620d256b-fbe8-5e36-a773-85fe6ea936da`.
7. GitHub receives the Agent Work Contract report.
8. Replay is available from `.agentdesk/replay/620d256b-fbe8-5e36-a773-85fe6ea936da.json`.

## Screenshot/GIF Beats

- Issue with `agent:ready` label.
- Terminal showing `agentdesk next` output.
- Terminal showing claim JSON.
- Terminal showing `agentdesk doctor --id ...` ready state.
- PR checks showing Agent Work Contract passed.
- PR comment with mandate, proof, approvals, and replay.

## Launch Caption

Every agent run starts with a mandate and ends with proof.

This PR was created through Guild AgentDesk: GitHub issue in, mandate claim, proof artifacts, CI verification, PR comment, and replay bundle out.
