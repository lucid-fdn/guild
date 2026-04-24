import { readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const failures = [];

await validateBootstrapFixtures();
await validatePublicExampleReferences();
await validateReplayBundleExample();

if (failures.length > 0) {
  throw new Error(`Fixture reference check failed:\n- ${failures.join("\n- ")}`);
}

console.log("ok fixture references");

async function validateBootstrapFixtures() {
  const fixtureDir = "server/internal/bootstrap/fixtures";

  const institution = await readJSON(`${fixtureDir}/institution.json`);
  const taskpack = await readJSON(`${fixtureDir}/taskpack.json`);
  const driBinding = await readJSON(`${fixtureDir}/dri-binding.json`);
  const artifact = await readJSON(`${fixtureDir}/artifact.json`);
  const promotionRecord = await readJSON(`${fixtureDir}/promotion-record.json`);
  const policy = await readJSON(`${fixtureDir}/governance-policy.json`);
  const approval = await readJSON(`${fixtureDir}/approval-request.json`);
  const gate = await readJSON(`${fixtureDir}/promotion-gate.json`);
  const commons = await readJSON(`${fixtureDir}/commons-entry.json`);

  requireEqual("bootstrap taskpack.institution_id", taskpack.institution_id, institution.institution_id);
  requireEqual("bootstrap dri-binding.taskpack_id", driBinding.taskpack_id, taskpack.taskpack_id);
  requireEqual("bootstrap artifact.taskpack_id", artifact.taskpack_id, taskpack.taskpack_id);
  requireEqual("bootstrap promotion-record.institution_id", promotionRecord.institution_id, institution.institution_id);
  requireEqual("bootstrap policy.institution_id", policy.institution_id, institution.institution_id);
  requireEqual("bootstrap approval.taskpack_id", approval.taskpack_id, taskpack.taskpack_id);
  requireEqual("bootstrap approval.policy_id", approval.policy_id, policy.policy_id);
  requireEqual("bootstrap gate.institution_id", gate.institution_id, institution.institution_id);
  requireEqual("bootstrap commons.institution_id", commons.institution_id, institution.institution_id);
  requireEqual("bootstrap commons.promotion_record_id", commons.promotion_record_id, promotionRecord.promotion_record_id);
  requireArtifactRef("bootstrap promotion-record.candidate_ref", promotionRecord.candidate_ref, artifact);
  requireArtifactRef("bootstrap commons.artifact_ref", commons.artifact_ref, artifact);
  requireNoSelfParent("bootstrap artifact", artifact);
}

async function validatePublicExampleReferences() {
  const exampleDir = "spec/examples";

  const taskpack = await readJSON(`${exampleDir}/taskpack.example.json`);
  const driBinding = await readJSON(`${exampleDir}/dri-binding.example.json`);
  const artifact = await readJSON(`${exampleDir}/artifact.example.json`);
  const policy = await readJSON(`${exampleDir}/governance-policy.example.json`);
  const approval = await readJSON(`${exampleDir}/approval-request.example.json`);
  const gate = await readJSON(`${exampleDir}/promotion-gate.example.json`);
  const commons = await readJSON(`${exampleDir}/commons-entry.example.json`);
  const promotion = await readJSON(`${exampleDir}/promotion-record.example.json`);

  requireEqual("example dri-binding.taskpack_id", driBinding.taskpack_id, taskpack.taskpack_id);
  requireEqual("example artifact.taskpack_id", artifact.taskpack_id, taskpack.taskpack_id);
  requireEqual("example approval.taskpack_id", approval.taskpack_id, taskpack.taskpack_id);
  requireEqual("example approval.policy_id", approval.policy_id, policy.policy_id);
  requireEqual("example gate.institution_id", gate.institution_id, taskpack.institution_id);
  requireEqual("example commons.institution_id", commons.institution_id, taskpack.institution_id);
  requireEqual("example commons.promotion_record_id", commons.promotion_record_id, promotion.promotion_record_id);
  requireArtifactRef("example commons.artifact_ref", commons.artifact_ref, artifact);
  requireNoSelfParent("example artifact", artifact);

  for (const [index, sourceTaskpackID] of (artifact.lineage?.source_taskpack_ids ?? []).entries()) {
    requireEqual(`example artifact.lineage.source_taskpack_ids[${index}]`, sourceTaskpackID, taskpack.taskpack_id);
  }
}

async function validateReplayBundleExample() {
  const replayBundle = await readJSON("spec/examples/replay-bundle.example.json");
  const artifactByID = new Map(replayBundle.artifacts.map((artifact) => [artifact.artifact_id, artifact]));
  const taskpackIDs = new Set((replayBundle.taskpacks ?? [replayBundle.taskpack]).map((taskpack) => taskpack.taskpack_id));

  requireEqual("replay-bundle.root_taskpack_id", replayBundle.root_taskpack_id, replayBundle.taskpack.taskpack_id);
  if (!taskpackIDs.has(replayBundle.taskpack.taskpack_id)) {
    failures.push("replay-bundle.taskpacks must include taskpack.taskpack_id");
  }

  for (const [index, binding] of replayBundle.dri_bindings.entries()) {
    requireInSet(`replay-bundle.dri_bindings[${index}].taskpack_id`, binding.taskpack_id, taskpackIDs);
  }
  for (const [index, artifact] of replayBundle.artifacts.entries()) {
    requireInSet(`replay-bundle.artifacts[${index}].taskpack_id`, artifact.taskpack_id, taskpackIDs);
    requireNoSelfParent(`replay-bundle.artifacts[${index}]`, artifact);
  }
  for (const [index, record] of replayBundle.promotion_records.entries()) {
    const artifact = artifactByID.get(record.candidate_ref?.artifact_id);
    if (!artifact) {
      failures.push(`replay-bundle.promotion_records[${index}].candidate_ref.artifact_id must exist in artifacts`);
      continue;
    }
    requireArtifactRef(`replay-bundle.promotion_records[${index}].candidate_ref`, record.candidate_ref, artifact);
  }
}

function requireInSet(name, actual, expectedSet) {
  if (!expectedSet.has(actual)) {
    failures.push(`${name}: expected one of ${JSON.stringify([...expectedSet])}, got ${JSON.stringify(actual)}`);
  }
}

async function readJSON(relativePath) {
  const body = await readFile(path.join(root, relativePath), "utf8");
  return JSON.parse(body);
}

function requireEqual(name, actual, expected) {
  if (actual !== expected) {
    failures.push(`${name}: expected ${JSON.stringify(expected)}, got ${JSON.stringify(actual)}`);
  }
}

function requireArtifactRef(name, ref, artifact) {
  requireEqual(`${name}.artifact_id`, ref?.artifact_id, artifact.artifact_id);
  requireEqual(`${name}.kind`, ref?.kind, artifact.kind);
  requireEqual(`${name}.uri`, ref?.uri, artifact.storage?.uri);
  requireEqual(`${name}.version`, ref?.version, artifact.version);
}

function requireNoSelfParent(name, artifact) {
  for (const [index, parentID] of (artifact.parent_artifact_ids ?? []).entries()) {
    if (parentID === artifact.artifact_id) {
      failures.push(`${name}.parent_artifact_ids[${index}] must not reference artifact_id`);
    }
  }
}
