from __future__ import annotations

import json
import threading
import unittest
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from typing import Any

from guild_client import GuildAPIError, GuildClient


class GuildClientTest(unittest.TestCase):
    def setUp(self) -> None:
        self.server = TestServer()
        self.server.start()
        self.client = GuildClient(self.server.base_url)

    def tearDown(self) -> None:
        self.server.stop()

    def test_strips_trailing_slash_and_lists_taskpacks(self) -> None:
        client = GuildClient(self.server.base_url + "/")

        taskpacks = client.list_taskpacks()

        self.assertEqual(taskpacks, [{"taskpack_id": "task-1"}])
        self.assertEqual(self.server.requests[-1]["path"], "/api/v1/taskpacks")

    def test_create_artifact_posts_json(self) -> None:
        payload = {"artifact_id": "artifact-1"}

        created = self.client.create_artifact(payload)

        self.assertEqual(created, {"artifact_id": "artifact-1", "created": True})
        self.assertEqual(self.server.requests[-1]["method"], "POST")
        self.assertEqual(self.server.requests[-1]["path"], "/api/v1/artifacts")
        self.assertEqual(self.server.requests[-1]["body"], payload)

    def test_exports_replay_bundle(self) -> None:
        bundle = self.client.export_replay_bundle("4e1fe00c-6303-453c-8cb6-2c34f84896e4")

        self.assertEqual(bundle["schema_version"], "v1alpha1")
        self.assertEqual(
            self.server.requests[-1]["path"],
            "/api/v1/replay/taskpacks/4e1fe00c-6303-453c-8cb6-2c34f84896e4",
        )

    def test_raises_api_error_with_server_message(self) -> None:
        with self.assertRaises(GuildAPIError) as raised:
            self.client.get_taskpack("missing")

        self.assertEqual(raised.exception.status, 404)
        self.assertEqual(raised.exception.message, "record not found")


class TestServer:
    def __init__(self) -> None:
        self.requests: list[dict[str, Any]] = []
        self.httpd = ThreadingHTTPServer(("127.0.0.1", 0), self._handler())
        self.thread = threading.Thread(target=self.httpd.serve_forever, daemon=True)

    @property
    def base_url(self) -> str:
        host, port = self.httpd.server_address
        return f"http://{host}:{port}"

    def start(self) -> None:
        self.thread.start()

    def stop(self) -> None:
        self.httpd.shutdown()
        self.thread.join(timeout=5)
        self.httpd.server_close()

    def _handler(self) -> type[BaseHTTPRequestHandler]:
        parent = self

        class Handler(BaseHTTPRequestHandler):
            def do_GET(self) -> None:
                parent.requests.append({"method": "GET", "path": self.path})
                if self.path == "/api/v1/taskpacks":
                    self._write_json(200, {"items": [{"taskpack_id": "task-1"}]})
                    return
                if self.path == "/api/v1/replay/taskpacks/4e1fe00c-6303-453c-8cb6-2c34f84896e4":
                    self._write_json(200, {"schema_version": "v1alpha1"})
                    return
                if self.path == "/api/v1/taskpacks/missing":
                    self._write_json(404, {"error": "record not found"})
                    return
                self._write_json(500, {"error": "unexpected path"})

            def do_POST(self) -> None:
                length = int(self.headers.get("Content-Length", "0"))
                body = json.loads(self.rfile.read(length).decode("utf-8"))
                parent.requests.append({"method": "POST", "path": self.path, "body": body})
                if self.path == "/api/v1/artifacts":
                    self._write_json(201, {**body, "created": True})
                    return
                self._write_json(500, {"error": "unexpected path"})

            def log_message(self, format: str, *args: object) -> None:
                return

            def _write_json(self, status: int, payload: dict[str, Any]) -> None:
                encoded = json.dumps(payload).encode("utf-8")
                self.send_response(status)
                self.send_header("Content-Type", "application/json")
                self.send_header("Content-Length", str(len(encoded)))
                self.end_headers()
                self.wfile.write(encoded)

        return Handler


if __name__ == "__main__":
    unittest.main()
