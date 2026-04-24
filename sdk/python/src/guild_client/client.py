from __future__ import annotations

import json
import urllib.error
import urllib.request
from dataclasses import dataclass
from typing import Any


class GuildAPIError(RuntimeError):
    def __init__(self, status: int, message: str):
        self.status = status
        self.message = message
        super().__init__(f"Guild API request failed with {status}: {message}")


@dataclass
class GuildClient:
    base_url: str

    def __post_init__(self) -> None:
        self.base_url = self.base_url.rstrip("/")

    def get_status(self) -> dict[str, Any]:
        return self._get_json("/api/v1/status")

    def list_taskpacks(self) -> list[dict[str, Any]]:
        return self._get_collection("/api/v1/taskpacks")

    def get_taskpack(self, taskpack_id: str) -> dict[str, Any]:
        return self._get_json(f"/api/v1/taskpacks/{taskpack_id}")

    def create_taskpack(self, taskpack: dict[str, Any]) -> dict[str, Any]:
        return self._post_json("/api/v1/taskpacks", taskpack)

    def list_dri_bindings(self) -> list[dict[str, Any]]:
        return self._get_collection("/api/v1/dri-bindings")

    def get_dri_binding(self, dri_binding_id: str) -> dict[str, Any]:
        return self._get_json(f"/api/v1/dri-bindings/{dri_binding_id}")

    def create_dri_binding(self, binding: dict[str, Any]) -> dict[str, Any]:
        return self._post_json("/api/v1/dri-bindings", binding)

    def list_artifacts(self) -> list[dict[str, Any]]:
        return self._get_collection("/api/v1/artifacts")

    def list_artifacts_for_taskpack(self, taskpack_id: str) -> list[dict[str, Any]]:
        return self._get_collection(f"/api/v1/taskpacks/{taskpack_id}/artifacts")

    def get_artifact(self, artifact_id: str) -> dict[str, Any]:
        return self._get_json(f"/api/v1/artifacts/{artifact_id}")

    def create_artifact(self, artifact: dict[str, Any]) -> dict[str, Any]:
        return self._post_json("/api/v1/artifacts", artifact)

    def list_promotion_records(self) -> list[dict[str, Any]]:
        return self._get_collection("/api/v1/promotion-records")

    def get_promotion_record(self, promotion_record_id: str) -> dict[str, Any]:
        return self._get_json(f"/api/v1/promotion-records/{promotion_record_id}")

    def create_promotion_record(self, record: dict[str, Any]) -> dict[str, Any]:
        return self._post_json("/api/v1/promotion-records", record)

    def export_replay_bundle(self, taskpack_id: str) -> dict[str, Any]:
        return self._get_json(f"/api/v1/replay/taskpacks/{taskpack_id}")

    def _get_collection(self, path: str) -> list[dict[str, Any]]:
        payload = self._get_json(path)
        return payload.get("items", [])

    def _get_json(self, path: str) -> dict[str, Any]:
        request = urllib.request.Request(f"{self.base_url}{path}", method="GET")
        return self._send(request)

    def _post_json(self, path: str, payload: dict[str, Any]) -> dict[str, Any]:
        request = urllib.request.Request(
            f"{self.base_url}{path}",
            data=json.dumps(payload).encode("utf-8"),
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        return self._send(request)

    def _send(self, request: urllib.request.Request) -> dict[str, Any]:
        try:
            with urllib.request.urlopen(request, timeout=30) as response:
                return json.loads(response.read().decode("utf-8"))
        except urllib.error.HTTPError as error:
            message = getattr(error, "reason", error.msg)
            try:
                payload = json.loads(error.read().decode("utf-8"))
                message = payload.get("error", message)
            except (json.JSONDecodeError, UnicodeDecodeError):
                pass
            raise GuildAPIError(error.code, str(message)) from error
