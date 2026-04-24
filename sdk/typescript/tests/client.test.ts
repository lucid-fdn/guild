import assert from "node:assert/strict";
import { afterEach, test } from "node:test";
import { GuildAPIError, GuildClient } from "../src/index";

type FetchCall = {
  url: string;
  init: RequestInit | undefined;
};

const calls: FetchCall[] = [];
const originalFetch = globalThis.fetch;

afterEach(() => {
  calls.length = 0;
  globalThis.fetch = originalFetch;
});

test("normalizes trailing slashes and lists taskpacks", async () => {
  mockFetch((url, init) => {
    calls.push({ url: url.toString(), init });
    return jsonResponse(200, { items: [{ taskpack_id: "task-1" }] });
  });

  const client = new GuildClient("http://guild.test/");
  const taskpacks = await client.listTaskpacks();

  assert.deepEqual(taskpacks, [{ taskpack_id: "task-1" }]);
  assert.equal(calls[0]?.url, "http://guild.test/api/v1/taskpacks");
  assert.equal(calls[0]?.init?.method, "GET");
});

test("posts JSON when creating artifacts", async () => {
  mockFetch((url, init) => {
    calls.push({ url: url.toString(), init });
    return jsonResponse(201, { artifact_id: "artifact-1", created: true });
  });

  const client = new GuildClient("http://guild.test");
  const created = await client.createArtifact({ artifact_id: "artifact-1" } as never);

  assert.deepEqual(created, { artifact_id: "artifact-1", created: true });
  assert.equal(calls[0]?.url, "http://guild.test/api/v1/artifacts");
  assert.equal(calls[0]?.init?.method, "POST");
  assert.deepEqual(JSON.parse(calls[0]?.init?.body as string), { artifact_id: "artifact-1" });
  assert.deepEqual(calls[0]?.init?.headers, { "Content-Type": "application/json" });
});

test("exports replay bundles", async () => {
  mockFetch((url, init) => {
    calls.push({ url: url.toString(), init });
    return jsonResponse(200, { schema_version: "v1alpha1" });
  });

  const client = new GuildClient("http://guild.test");
  const bundle = await client.exportReplayBundle("4e1fe00c-6303-453c-8cb6-2c34f84896e4");

  assert.equal(bundle.schema_version, "v1alpha1");
  assert.equal(calls[0]?.url, "http://guild.test/api/v1/replay/taskpacks/4e1fe00c-6303-453c-8cb6-2c34f84896e4");
});

test("throws structured API errors with server messages", async () => {
  mockFetch(() => jsonResponse(404, { error: "record not found" }));

  const client = new GuildClient("http://guild.test");
  await assert.rejects(client.getTaskpack("missing"), (error) => {
    assert.ok(error instanceof GuildAPIError);
    assert.equal(error.status, 404);
    assert.equal(error.message, "record not found");
    return true;
  });
});

function mockFetch(handler: (url: URL, init: RequestInit | undefined) => Response): void {
  globalThis.fetch = async (input, init) => handler(new URL(input.toString()), init);
}

function jsonResponse(status: number, payload: unknown): Response {
  return new Response(JSON.stringify(payload), {
    status,
    headers: {
      "Content-Type": "application/json",
    },
  });
}
