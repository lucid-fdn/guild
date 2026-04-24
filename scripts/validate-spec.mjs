import { readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import Ajv2020 from "ajv/dist/2020.js";
import addFormats from "ajv-formats";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");

const pairs = [
  ["spec/taskpack.schema.json", "spec/examples/taskpack.example.json"],
  ["spec/dri-binding.schema.json", "spec/examples/dri-binding.example.json"],
  ["spec/artifact.schema.json", "spec/examples/artifact.example.json"],
  ["spec/promotion-record.schema.json", "spec/examples/promotion-record.example.json"],
  ["spec/governance-policy.schema.json", "spec/examples/governance-policy.example.json"],
  ["spec/approval-request.schema.json", "spec/examples/approval-request.example.json"],
  ["spec/promotion-gate.schema.json", "spec/examples/promotion-gate.example.json"],
  ["spec/commons-entry.schema.json", "spec/examples/commons-entry.example.json"],
  ["spec/governance-policy.schema.json", "server/internal/bootstrap/fixtures/governance-policy.json"],
  ["spec/approval-request.schema.json", "server/internal/bootstrap/fixtures/approval-request.json"],
  ["spec/promotion-gate.schema.json", "server/internal/bootstrap/fixtures/promotion-gate.json"],
  ["spec/commons-entry.schema.json", "server/internal/bootstrap/fixtures/commons-entry.json"],
  ["spec/replay-bundle.schema.json", "spec/examples/replay-bundle.example.json"],
  ["spec/workspace-constitution.schema.json", "spec/examples/workspace-constitution.example.json"],
  ["spec/context-pack.schema.json", "spec/examples/context-pack.example.json"],
  ["spec/preflight-decision.schema.json", "spec/examples/preflight-decision.example.json"],
  ["spec/taskpack.schema.json", "server/internal/bootstrap/fixtures/taskpack.json"],
  ["spec/dri-binding.schema.json", "server/internal/bootstrap/fixtures/dri-binding.json"],
  ["spec/artifact.schema.json", "server/internal/bootstrap/fixtures/artifact.json"],
  ["spec/promotion-record.schema.json", "server/internal/bootstrap/fixtures/promotion-record.json"],
];

async function readJSON(relativePath) {
  const payload = await readFile(path.join(root, relativePath), "utf8");
  return JSON.parse(payload);
}

function formatError(error) {
  const location = error.instancePath || "/";
  const detail = error.message ?? "failed validation";
  return `${location} ${detail}`;
}

const ajv = new Ajv2020({
  allErrors: true,
  strict: true,
});
addFormats(ajv);

for (const schemaPath of [
  "spec/common.schema.json",
  "spec/taskpack.schema.json",
  "spec/dri-binding.schema.json",
  "spec/artifact.schema.json",
  "spec/promotion-record.schema.json",
  "spec/governance-policy.schema.json",
  "spec/approval-request.schema.json",
  "spec/promotion-gate.schema.json",
  "spec/commons-entry.schema.json",
  "spec/workspace-constitution.schema.json",
  "spec/context-pack.schema.json",
  "spec/preflight-decision.schema.json",
]) {
  const schema = await readJSON(schemaPath);
  ajv.addSchema(schema, path.basename(schemaPath));
}

const validators = new Map();

for (const [schemaPath, dataPath] of pairs) {
  if (!validators.has(schemaPath)) {
    const schemaKey = path.basename(schemaPath);
    const existing = ajv.getSchema(schemaKey);
    validators.set(schemaPath, existing ?? ajv.compile(await readJSON(schemaPath)));
  }
  const data = await readJSON(dataPath);
  const validate = validators.get(schemaPath);

  if (!validate(data)) {
    const errors = validate.errors?.map(formatError).join("\n  - ") ?? "unknown validation error";
    throw new Error(`${dataPath} does not match ${schemaPath}:\n  - ${errors}`);
  }

  console.log(`ok ${dataPath}`);
}
