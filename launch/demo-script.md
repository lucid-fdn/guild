# 90-Second Demo Script

## 0:00 - Problem

"Agents already get tasks from humans, but the handoff is mushy. A GitHub issue says what to do. It rarely says what the agent may touch, when it must stop, or what proof makes the work complete."

## 0:10 - Issue To Mandate

"Guild turns an issue labeled `agent:ready` into a mandate an agent can consume."

Command:

```bash
guild agentdesk next --source github --repo lucid-fdn/guild
```

## 0:25 - Claim

"The agent claims the mandate, so two agents do not silently work the same task."

Command:

```bash
guild agentdesk claim --id <mandate-id> --agent codex
```

## 0:35 - Guardrails

"Before editing, the agent checks scope and compiles bounded context instead of inhaling the whole repo."

Command:

```bash
guild agentdesk context compile --id <mandate-id> --role coder
guild agentdesk preflight --id <mandate-id> --action write --path docs/demo.md
```

## 0:50 - Proof

"The run ends with proof: tests, changed files, and a handoff summary. Not vibes. Not transcript archaeology."

Command:

```bash
guild agentdesk proof add --id <mandate-id> --kind test_report --path test-results.xml
guild agentdesk proof add --id <mandate-id> --kind changed_files --path changed-files.json
guild agentdesk handoff create --id <mandate-id> --to reviewer --summary "Ready for review"
```

## 1:05 - Verify

"CI verifies the Agent Work Contract and posts the report back to the PR."

Command:

```bash
guild agentdesk verify --id <mandate-id> --github-report
```

## 1:20 - Replay

"Anyone can export the replay bundle later and inspect mandate, claim, approvals, proof, and handoff."

Command:

```bash
guild agentdesk replay export --id <mandate-id>
```

## 1:30 - Close

"Guild is not another orchestrator. It is the work contract agents consume before, during, and after work."

## Real Proof

- Demo PR: https://github.com/lucid-fdn/guild/pull/9
- Agent Work Contract comment: https://github.com/lucid-fdn/guild/pull/9#issuecomment-4316590055
