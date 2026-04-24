import { access, readdir, readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const ignoredDirs = new Set([".git", ".next", "node_modules"]);
const markdownFiles = await collectMarkdownFiles(root);
const failures = [];

for (const file of markdownFiles) {
  const body = await readFile(file, "utf8");
  const links = body.matchAll(/!?\[[^\]]*]\(([^)\s]+)(?:\s+"[^"]*")?\)/g);

  for (const match of links) {
    const target = match[1];
    if (shouldSkip(target)) {
      continue;
    }

    const [targetPath] = target.split("#");
    const decodedPath = decodeURIComponent(targetPath);
    const resolved = path.resolve(path.dirname(file), decodedPath);

    if (!isInsideRoot(resolved)) {
      failures.push(`${relative(file)} links outside repo: ${target}`);
      continue;
    }

    try {
      await access(resolved);
    } catch {
      failures.push(`${relative(file)} links to missing path: ${target}`);
    }
  }
}

if (failures.length > 0) {
  throw new Error(`Markdown link check failed:\n- ${failures.join("\n- ")}`);
}

console.log(`ok ${markdownFiles.length} markdown files`);

async function collectMarkdownFiles(dir) {
  const entries = await readdir(dir, { withFileTypes: true });
  const files = [];

  for (const entry of entries) {
    if (ignoredDirs.has(entry.name)) {
      continue;
    }
    const fullPath = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      files.push(...(await collectMarkdownFiles(fullPath)));
      continue;
    }
    if (entry.isFile() && entry.name.endsWith(".md")) {
      files.push(fullPath);
    }
  }

  return files;
}

function shouldSkip(target) {
  return (
    target.startsWith("#") ||
    target.startsWith("http://") ||
    target.startsWith("https://") ||
    target.startsWith("mailto:")
  );
}

function isInsideRoot(targetPath) {
  const relativePath = path.relative(root, targetPath);
  return relativePath === "" || (!relativePath.startsWith("..") && !path.isAbsolute(relativePath));
}

function relative(file) {
  return path.relative(root, file);
}
