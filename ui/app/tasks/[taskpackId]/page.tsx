import Link from "next/link";
import type { CSSProperties } from "react";
import { getGuildExperience, getReplayBundle, recordsForTask } from "../../guild-data";

export default async function TaskDetailPage({ params }: { params: Promise<{ taskpackId: string }> }) {
  const { taskpackId } = await params;
  const experience = await getGuildExperience();
  const records = recordsForTask(experience, taskpackId);
  const replay = experience.source === "api" ? await getReplayBundle(taskpackId) : null;
  const taskpack = records.taskpack ?? replay?.taskpack;

  if (!taskpack) {
    return (
      <main style={styles.shell}>
        <Link href="/" style={styles.back}>Back to control room</Link>
        <section style={styles.panel}>
          <h1 style={styles.title}>Taskpack not found</h1>
          <p style={styles.copy}>No live or demo record exists for {taskpackId}.</p>
        </section>
      </main>
    );
  }

  const driBindings = replay?.dri_bindings ?? records.driBindings;
  const artifacts = replay?.artifacts ?? records.artifacts;
  const promotions = replay?.promotion_records ?? records.promotionRecords;
  const owner = driBindings[0]?.owner;

  return (
    <main style={styles.shell}>
      <Link href="/" style={styles.back}>Back to control room</Link>
      <section style={styles.hero}>
        <p style={styles.eyebrow}>Task Detail</p>
        <h1 style={styles.title}>{taskpack.title}</h1>
        <p style={styles.copy}>{taskpack.objective}</p>
        <div style={styles.row}>
          <Pill label="Priority" value={taskpack.priority} />
          <Pill label="Type" value={taskpack.task_type} />
          <Pill label="Approval" value={taskpack.permissions.approval_mode} />
          <Pill label="Context" value={`${taskpack.context_budget.max_input_tokens} in / ${taskpack.context_budget.max_output_tokens} out`} />
        </div>
      </section>

      <section style={styles.grid}>
        <div style={styles.panel}>
          <p style={styles.eyebrow}>DRI Graph</p>
          <h2 style={styles.h2}>One owner, many contributors</h2>
          <div style={styles.graph}>
            <Node label="Requester" value={taskpack.requested_by.display_name ?? taskpack.requested_by.actor_id} />
            <Edge />
            <Node label="DRI Owner" value={owner?.display_name ?? "unassigned"} strong />
            <Edge />
            <Node label="Reviewers" value={driBindings[0]?.reviewers?.map((actor) => actor.display_name ?? actor.actor_id).join(", ") ?? "none"} />
          </div>
        </div>

        <div style={styles.panel}>
          <p style={styles.eyebrow}>Acceptance</p>
          <h2 style={styles.h2}>Definition of done</h2>
          <div style={styles.stack}>
            {taskpack.acceptance_criteria.map((criterion) => (
              <div key={criterion.criterion_id} style={styles.item}>
                <strong>{criterion.criterion_id}</strong>
                <span>{criterion.description}</span>
                <small>{criterion.required ? "required" : "optional"}</small>
              </div>
            ))}
          </div>
        </div>

        <div style={styles.panelWide}>
          <p style={styles.eyebrow}>Artifact Graph</p>
          <h2 style={styles.h2}>Durable outputs, not chat history</h2>
          <div style={styles.artifactGrid}>
            <div style={styles.timeline}>
              {artifacts.map((artifact) => (
                <article key={artifact.artifact_id} style={styles.timelineItem}>
                  <span style={styles.dot} />
                  <div>
                    <div style={styles.rowBetween}>
                      <strong>{artifact.title}</strong>
                      <span style={styles.badge}>{artifact.kind}</span>
                    </div>
                    <p style={styles.copy}>{artifact.summary ?? "No summary captured."}</p>
                    <code style={styles.code}>{artifact.storage.uri}</code>
                  </div>
                </article>
              ))}
            </div>
            <div style={styles.viewer}>
              <p style={styles.eyebrow}>Artifact Viewer</p>
              {artifacts.length === 0 ? (
                <p style={styles.copy}>No artifacts have been published for this task.</p>
              ) : (
                artifacts.map((artifact) => (
                  <div key={artifact.artifact_id} style={styles.viewerCard}>
                    <div style={styles.rowBetween}>
                      <strong>{artifact.title}</strong>
                      <span style={styles.badge}>{artifact.version ? `v${artifact.version}` : "v1"}</span>
                    </div>
                    <dl style={styles.definitionList}>
                      <div>
                        <dt>ID</dt>
                        <dd>{artifact.artifact_id}</dd>
                      </div>
                      <div>
                        <dt>Producer</dt>
                        <dd>{artifact.producer.display_name ?? artifact.producer.actor_id}</dd>
                      </div>
                      <div>
                        <dt>MIME</dt>
                        <dd>{artifact.storage.mime_type}</dd>
                      </div>
                      <div>
                        <dt>Lineage</dt>
                        <dd>{artifact.lineage?.trace_id ?? taskpack.trace_id ?? "not captured"}</dd>
                      </div>
                    </dl>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        <div style={styles.panel}>
          <p style={styles.eyebrow}>Approval Inbox</p>
          <h2 style={styles.h2}>Pending gates</h2>
          <div style={styles.stack}>
            <div style={styles.item}>
              <strong>{taskpack.permissions.approval_mode}</strong>
              <span>{taskpack.permissions.allow_network ? "Network access requested" : "No network access requested"}</span>
              <small>Scopes: {taskpack.permissions.scopes?.join(", ") ?? "no explicit scopes"}</small>
            </div>
            <div style={styles.item}>
              <strong>External write</strong>
              <span>{taskpack.permissions.allow_external_write ? "Requires explicit approval" : "Blocked by policy"}</span>
              <small>Owner: {owner?.display_name ?? "unassigned"}</small>
            </div>
          </div>
        </div>

        <div style={styles.panel}>
          <p style={styles.eyebrow}>Replay Timeline</p>
          <h2 style={styles.h2}>What happened</h2>
          <div style={styles.stack}>
            <TimelineItem time={taskpack.created_at} title="Taskpack created" />
            {driBindings.map((binding) => (
              <TimelineItem key={binding.dri_binding_id} time={binding.created_at} title={`DRI assigned to ${binding.owner.display_name ?? binding.owner.actor_id}`} />
            ))}
            {artifacts.map((artifact) => (
              <TimelineItem key={artifact.artifact_id} time={artifact.created_at} title={`Artifact published: ${artifact.title}`} />
            ))}
            {promotions.map((record) => (
              <TimelineItem key={record.promotion_record_id} time={record.decided_at} title={`Promotion ${record.decision}: ${record.candidate_kind}`} />
            ))}
          </div>
        </div>

        <div style={styles.panel}>
          <p style={styles.eyebrow}>Commons</p>
          <h2 style={styles.h2}>Promoted learning</h2>
          <div style={styles.stack}>
            {promotions.length === 0 ? (
              <p style={styles.copy}>No promotion records yet.</p>
            ) : (
              promotions.map((record) => (
                <div key={record.promotion_record_id} style={styles.item}>
                  <strong>{record.decision}</strong>
                  <span>{record.candidate_kind}</span>
                  <small>{record.decision_reason ?? "No decision reason captured."}</small>
                </div>
              ))
            )}
          </div>
        </div>
      </section>
    </main>
  );
}

function Pill({ label, value }: { label: string; value: string }) {
  return (
    <span style={styles.pill}>
      {label}: <strong>{value}</strong>
    </span>
  );
}

function Node({ label, value, strong }: { label: string; value: string; strong?: boolean }) {
  return (
    <div style={{ ...styles.node, ...(strong ? styles.nodeStrong : {}) }}>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function Edge() {
  return <div style={styles.edge} />;
}

function TimelineItem({ time, title }: { time: string; title: string }) {
  return (
    <div style={styles.item}>
      <strong>{title}</strong>
      <small>{time}</small>
    </div>
  );
}

const styles = {
  shell: {
    minHeight: "100vh",
    padding: "34px clamp(20px, 5vw, 72px) 56px",
    background:
      "radial-gradient(circle at 85% 10%, rgba(217, 109, 43, 0.18), transparent 30%), linear-gradient(135deg, #f8f0df 0%, #e8f1ea 46%, #d6e7ef 100%)",
    color: "#17211b",
    fontFamily: "Georgia, 'Times New Roman', serif"
  },
  back: {
    display: "inline-flex",
    marginBottom: 28,
    color: "#17211b",
    textDecoration: "none",
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
  },
  hero: {
    maxWidth: 1180,
    margin: "0 auto 18px",
    padding: 28,
    borderRadius: 30,
    background: "rgba(255, 255, 255, 0.66)",
    border: "1px solid rgba(23, 33, 27, 0.14)"
  },
  eyebrow: {
    margin: "0 0 8px",
    color: "#7f5d19",
    fontFamily: "ui-monospace, SFMono-Regular, Menlo, monospace",
    fontSize: 12,
    letterSpacing: "0.12em",
    textTransform: "uppercase" as const
  },
  title: {
    margin: 0,
    maxWidth: 920,
    fontSize: "clamp(38px, 6vw, 72px)",
    lineHeight: 0.96,
    letterSpacing: "-0.05em"
  },
  copy: {
    color: "#4a574e",
    lineHeight: 1.55,
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
  },
  grid: {
    display: "grid",
    gridTemplateColumns: "repeat(2, minmax(0, 1fr))",
    gap: 14,
    maxWidth: 1180,
    margin: "0 auto"
  },
  panel: {
    padding: 24,
    borderRadius: 28,
    background: "rgba(255, 255, 255, 0.68)",
    border: "1px solid rgba(23, 33, 27, 0.14)"
  },
  panelWide: {
    gridColumn: "1 / -1",
    padding: 24,
    borderRadius: 28,
    background: "rgba(255, 255, 255, 0.74)",
    border: "1px solid rgba(23, 33, 27, 0.14)"
  },
  h2: {
    margin: 0,
    fontSize: 28,
    lineHeight: 1.05
  },
  row: {
    display: "flex",
    gap: 8,
    flexWrap: "wrap" as const,
    marginTop: 16
  },
  rowBetween: {
    display: "flex",
    justifyContent: "space-between",
    gap: 16,
    alignItems: "center"
  },
  pill: {
    padding: "8px 11px",
    borderRadius: 999,
    background: "#e9dfcb",
    color: "#26342b",
    fontFamily: "ui-sans-serif, system-ui, sans-serif",
    fontSize: 13
  },
  graph: {
    display: "grid",
    gap: 12,
    marginTop: 18,
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
  },
  node: {
    display: "grid",
    gap: 4,
    padding: 16,
    borderRadius: 18,
    background: "#fffaf0",
    border: "1px solid rgba(23, 33, 27, 0.1)"
  },
  nodeStrong: {
    background: "#17211b",
    color: "#fff7e8"
  },
  edge: {
    width: 2,
    height: 24,
    marginLeft: 22,
    background: "#c39b4e"
  },
  stack: {
    display: "grid",
    gap: 12,
    marginTop: 18,
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
  },
  item: {
    display: "grid",
    gap: 5,
    padding: 14,
    borderRadius: 16,
    background: "rgba(255, 250, 240, 0.78)",
    border: "1px solid rgba(23, 33, 27, 0.1)"
  },
  timeline: {
    display: "grid",
    gap: 14,
    marginTop: 18,
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
  },
  artifactGrid: {
    display: "grid",
    gridTemplateColumns: "minmax(0, 1fr) minmax(280px, 0.72fr)",
    gap: 18,
    alignItems: "start"
  },
  viewer: {
    display: "grid",
    gap: 12,
    marginTop: 18,
    padding: 16,
    borderRadius: 22,
    background: "#17211b",
    color: "#fff7e8",
    fontFamily: "ui-sans-serif, system-ui, sans-serif"
  },
  viewerCard: {
    display: "grid",
    gap: 12,
    padding: 14,
    borderRadius: 16,
    background: "rgba(255, 247, 232, 0.1)",
    border: "1px solid rgba(255, 247, 232, 0.16)"
  },
  definitionList: {
    display: "grid",
    gap: 9,
    margin: 0
  },
  timelineItem: {
    display: "grid",
    gridTemplateColumns: "18px minmax(0, 1fr)",
    gap: 14
  },
  dot: {
    width: 12,
    height: 12,
    marginTop: 5,
    borderRadius: 99,
    background: "#d96d2b",
    boxShadow: "0 0 0 7px rgba(217, 109, 43, 0.12)"
  },
  badge: {
    padding: "6px 9px",
    borderRadius: 999,
    background: "#17211b",
    color: "#fff7e8",
    fontSize: 12
  },
  code: {
    display: "block",
    maxWidth: "100%",
    overflow: "auto",
    padding: 10,
    borderRadius: 12,
    background: "#17211b",
    color: "#fff7e8"
  }
} satisfies Record<string, CSSProperties>;
