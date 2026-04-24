import fs from "node:fs";
import path from "node:path";
import process from "node:process";
import Ajv2020 from "ajv/dist/2020.js";

const root = process.cwd();
const schema = JSON.parse(fs.readFileSync(path.join(root, "conformance/adapter-profile.schema.json"), "utf8"));
const profilesDir = path.join(root, "conformance/profiles");
const ajv = new Ajv2020({ allErrors: true, strict: true });
const validate = ajv.compile(schema);

const profileFiles = fs.readdirSync(profilesDir).filter((file) => file.endsWith(".json")).sort();
if (profileFiles.length === 0) {
  throw new Error("no adapter profiles found");
}

for (const file of profileFiles) {
  const absolute = path.join(profilesDir, file);
  const profile = JSON.parse(fs.readFileSync(absolute, "utf8"));
  if (!validate(profile)) {
    throw new Error(`${file}: ${ajv.errorsText(validate.errors)}`);
  }
  const badgePath = path.normalize(path.join(profilesDir, profile.badge.url));
  if (!badgePath.startsWith(path.join(root, "conformance"))) {
    throw new Error(`${file}: badge URL must stay inside conformance/`);
  }
  if (!fs.existsSync(badgePath)) {
    throw new Error(`${file}: missing badge ${profile.badge.url}`);
  }
  console.log(`ok ${path.relative(root, absolute)}`);
}
