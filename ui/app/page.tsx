import Link from "next/link";
import type { CSSProperties } from "react";
import { getGuildExperience, recordsForTask } from "./guild-data";

export default async function HomePage() {
  const experience = await getGuildExperience();
  const acceptedPromotions = experience.promotionRecords.filter((record) => record.decision === "accepted").length;
  const approvalsRequired = experience.taskpacks.filter((taskpack) => taskpack.permissions.approval_mode !== "none").length;
  const pendingApprovals = experience.approvalRequests.filter((request) => request.status === "pending").length;

  return (
    <main style={styles.shell}>
      <section style={styles.hero}>
        <div>
          <p style={styles.eyebrow}>Guild Experience Plane</p>
          <h1 style={styles.title}>The institution layer for AI teams.</h1>
          <p style={styles.subtitle}>
            Bring your own orchestrator. Guild gives every task one owner, every handoff a bounded packet,
            every output an artifact, and every promoted learning a replayable audit trail.
          </p>
        </div>
        <div style={styles.statusCard}>
          <span style={styles.statusDot} />
          <strong>{experience.status?.mode ?? "unknown"} mode</strong>
          <span style={styles.muted}>{experience.source === "api" ? "Live control plane" : "Offline demo data"}</span>
        </div>
      </section>

      <section style={styles.metricsGrid}>
        <Metric label="Taskpacks" value={experience.taskpacks.length} />
        <Metric label="DRI bindings" value={experience.driBindings.length} />
        <Metric label="Artifacts" value={experience.artifacts.length} />
        <Metric label="Commons entries" value={experience.commonsEntries.length} />
      </section>

      <section style={styles.grid}>
        <div style={styles.panelWide}>
          <div style={styles.panelHeader}>
            <div>
              <p style={styles.kicker}>Ownership</p>
              <h2 style={styles.h2}>DRI task board</h2>
            </div>
            <span style={styles.badge}>{approvalsRequired} approval-aware</span>
          </div>
          <div style={styles.taskList}>
            {experience.taskpacks.map((taskpack) => {
              const records = recordsForTask(experience, taskpack.taskpack_id);
              const owner = records.driBindings[0]?.owner;
              return (
                <Link key={taskpack.taskpack_id} href={`/tasks/${taskpack.taskpack_id}`} style={styles.taskCard}>
                  <div style={styles.taskCardTop}>
                    <span style={styles.priority}>{taskpack.priority}</span>
                    <span style={styles.muted}>{taskpack.task_type}</span>
                  </div>
                  <h3 style={styles.h3}>{taskpack.title}</h3>
                  <p style={styles.copy}>{taskpack.objective}</p>
                  <div style={styles.row}>
                    <Pill label="DRI" value={owner?.display_name ?? "unassigned"} />
                    <Pill label="Artifacts" value={String(records.artifacts.length)} />
                    <Pill label="Promotions" value={String(records.promotionRecords.length)} />
                  </div>
                </Link>
              );
            })}
          </div>
        </div>

        <div style={styles.panel}>
          <p style={styles.kicker}>Approval Inbox</p>
          <h2 style={styles.h2}>Human gates</h2>
          <div style={styles.stack}>
            {experience.approvalRequests.length === 0 ? (
              <p style={styles.copy}>No approval requests yet.</p>
            ) : experience.approvalRequests.map((request) => {
              const taskpack = experience.taskpacks.find((item) => item.taskpack_id === request.taskpack_id);
              const records = recordsForTask(experience, request.taskpack_id);
              const owner = records.driBindings[0]?.owner;
              return (
                <Link key={request.approval_id} href={`/tasks/${request.taskpack_id}`} style={styles.inboxItemLink}>
                  <div style={styles.taskCardTop}>
                    <strong>{request.status}</strong>
                    <span style={styles.badgeSmall}>{request.required_approvals} approval</span>
                  </div>
                  <span>{taskpack?.title ?? request.taskpack_id}</span>
                  <small>Owner: {owner?.display_name ?? "unassigned"}</small>
                  <small>{request.reason}</small>
                </Link>
              );
            })}
          </div>
        </div>

        <div style={styles.panel}>
          <p style={styles.kicker}>Commons</p>
          <h2 style={styles.h2}>Registry browser</h2>
          <p style={styles.copy}>{acceptedPromotions} accepted promotion records. {pendingApprovals} approval request pending.</p>
          <div style={styles.stack}>
            {experience.commonsEntries.map((entry) => (
              <div key={entry.commons_entry_id} style={styles.inboxItem}>
                <strong>{entry.title}</strong>
                <span>{entry.scope} / {entry.status}</span>
                <small>{entry.summary}</small>
              </div>
            ))}
          </div>
        </div>

        <div style={styles.panelWide}>
          <p style={styles.kicker}>Governance</p>
          <h2 style={styles.h2}>Policies and promotion gates</h2>
          <div style={styles.governanceGrid}>
            {experience.governancePolicies.map((policy) => (
              <div key={policy.policy_id} style={styles.inboxItem}>
                <strong>{policy.name}</strong>
                <span>{policy.description}</span>
                <small>{policy.rules.length} rule(s)</small>
              </div>
            ))}
            {experience.promotionGates.map((gate) => (
              <div key={gate.gate_id} style={styles.inboxItem}>
                <strong>{gate.name}</strong>
                <span>{gate.candidate_kinds.join(", ")}</span>
                <small>{gate.min_replay_runs} replay run(s), approval {gate.requires_approval ? "required" : "optional"}</small>
              </div>
            ))}
          </div>
        </div>
      </section>
    </main>
  );
}

function Metric({ label, value }: { label: string; value: number }) {
  return (
    <div style={styles.metric}>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function Pill({ label, value }: { label: string; value: string }) {
  return (
    <span style={styles.pill}>
      {label}: <strong>{value}</strong>
    </span>
  );
}

const styles = {
  shell: {
    minHeight: "100vh",
    padding: "56px clamp(20px, 5vw, 72px)",
    background:
      "radial-gradient(circle at 20% 10%, rgba(255, 202, 92, 0.28), transparent 32%), linear-gradient(135deg, #f8f0df 0%, #e8f1ea 46%, #d6e7ef 100%)",
    color: "#17211b",
    fontFamily: "Georgia, 'Times New Roman', serif"
  },
  hero: {
    display: "grid",
    gridTemplateColumns: "minmax(0, 1fr) minmax(220px, 320px)",
    gap: 28,
    alignItems: "end",
    maxWidth: 1180,
    margin: "0 auto 28px"
  },
  eyebrow: {
    margin: 0,
    color: "#6b4d16",
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
    fontSize: 12,
    letterSpacing: "0.16em",
    textTransform: "uppercase" as const
  },
  title: {
    margin: "10px 0 14px",
    maxWidth: 820,
    fontSize: "clamp(46px, 8vw, 94px)",
    lineHeight: 0.92,
    letterSpacing: "-0.06em"
  },
  subtitle: {
    margin: 0,
    maxWidth: 760,
    color: "#35443a",
    fontSize: 20,
    lineHeight: 1.5,
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
  },
  statusCard: {
    display: "grid",
    gap: 8,
    padding: 22,
    border: "1px solid rgba(23, 33, 27, 0.18)",
    borderRadius: 24,
    background: "rgba(255, 255, 255, 0.58)",
    boxShadow: "0 24px 80px rgba(55, 63, 48, 0.12)",
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
  },
  statusDot: {
    width: 12,
    height: 12,
    borderRadius: 99,
    background: "#28764c",
    boxShadow: "0 0 0 8px rgba(40, 118, 76, 0.12)"
  },
  metricsGrid: {
    display: "grid",
    gridTemplateColumns: "repeat(4, minmax(0, 1fr))",
    gap: 14,
    maxWidth: 1180,
    margin: "0 auto 14px"
  },
  metric: {
    display: "grid",
    gap: 8,
    padding: 20,
    borderRadius: 22,
    background: "rgba(23, 33, 27, 0.86)",
    color: "#fff7e8",
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
  },
  grid: {
    display: "grid",
    gridTemplateColumns: "minmax(0, 1.45fr) minmax(280px, 0.75fr)",
    gap: 14,
    maxWidth: 1180,
    margin: "0 auto"
  },
  panelWide: {
    gridRow: "span 2",
    padding: 24,
    borderRadius: 28,
    background: "rgba(255, 255, 255, 0.74)",
    border: "1px solid rgba(23, 33, 27, 0.14)"
  },
  panel: {
    padding: 24,
    borderRadius: 28,
    background: "rgba(255, 255, 255, 0.64)",
    border: "1px solid rgba(23, 33, 27, 0.14)"
  },
  panelHeader: {
    display: "flex",
    justifyContent: "space-between",
    gap: 16,
    alignItems: "start"
  },
  kicker: {
    margin: "0 0 6px",
    color: "#7f5d19",
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
    fontSize: 12,
    letterSpacing: "0.12em",
    textTransform: "uppercase" as const
  },
  h2: {
    margin: 0,
    fontSize: 30,
    lineHeight: 1.05
  },
  h3: {
    margin: "10px 0 8px",
    fontSize: 24,
    lineHeight: 1.12
  },
  taskList: {
    display: "grid",
    gap: 14,
    marginTop: 20
  },
  taskCard: {
    display: "block",
    padding: 20,
    borderRadius: 22,
    background: "#fffaf0",
    border: "1px solid rgba(23, 33, 27, 0.12)",
    color: "inherit",
    textDecoration: "none",
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
  },
  taskCardTop: {
    display: "flex",
    justifyContent: "space-between",
    gap: 12
  },
  priority: {
    padding: "5px 10px",
    borderRadius: 999,
    background: "#d96d2b",
    color: "white",
    fontSize: 12,
    textTransform: "uppercase" as const,
    letterSpacing: "0.08em"
  },
  copy: {
    color: "#4a574e",
    lineHeight: 1.5
  },
  row: {
    display: "flex",
    gap: 8,
    flexWrap: "wrap" as const
  },
  pill: {
    padding: "7px 10px",
    borderRadius: 999,
    background: "#e9dfcb",
    color: "#26342b",
    fontSize: 13
  },
  badge: {
    padding: "8px 11px",
    borderRadius: 999,
    background: "#17211b",
    color: "#fff7e8",
    fontFamily: "ui-sans-serif, system-ui, sans-serif",
    fontSize: 13
  },
  stack: {
    display: "grid",
    gap: 12,
    marginTop: 18,
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
  },
  inboxItem: {
    display: "grid",
    gap: 5,
    padding: 14,
    borderRadius: 16,
    background: "rgba(255, 250, 240, 0.78)",
    border: "1px solid rgba(23, 33, 27, 0.1)"
  },
  inboxItemLink: {
    display: "grid",
    gap: 7,
    padding: 14,
    borderRadius: 16,
    background: "rgba(255, 250, 240, 0.78)",
    border: "1px solid rgba(23, 33, 27, 0.1)",
    color: "inherit",
    textDecoration: "none"
  },
  badgeSmall: {
    padding: "4px 8px",
    borderRadius: 999,
    background: "#e9dfcb",
    color: "#26342b",
    fontSize: 12
  },
  muted: {
    color: "#657269",
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
  },
  governanceGrid: {
    display: "grid",
    gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
    gap: 12,
    marginTop: 18,
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
  }
} satisfies Record<string, CSSProperties>;
