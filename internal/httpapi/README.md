# internal/httpapi/ — the daemon's authenticated loopback HTTP/JSON + SSE surface

> 🌐 English | [繁體中文](README.zh-TW.md)

One source file, [`httpapi.go`](httpapi.go); its package comment is the contract (there is
no separate `doc.go`). Implements the M6 daemon's API (issue #7; ADD §23.2–§23.5, NFR-022 —
[`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md)).

`NewHandler(deps, bearerToken)` mounts:

- `GET /v1/health`, `/v1/version`, `/v1/capabilities`, `/v1/status`, `/v1/scheduler/jobs` — the read surface.
- `POST /v1/scheduler/jobs/{id}/cancel` — the single mutation (FR-163, issue #10: cancel a scheduled resume).
- `GET /v1/events/stream` — SSE live event stream with `: ping` heartbeats every 15s.

Security posture (`guard` middleware, ADD §23.2/§27.5): every endpoint requires
`Authorization: Bearer <token>` compared in constant time; the `Host` header must be
loopback (DNS-rebinding defense); request bodies are capped at 1 MiB; CORS is disabled by
omission; errors render as the ADD §23.5 envelope with typed `domain.Error` codes.

Dependencies are deliberately narrow interfaces: `JobLister` / `JobCanceller` (slices of
[`internal/scheduler/`](../scheduler/README.md)'s `Store` — the canceller is separate so
read-only compositions never gain mutation), `EventSource`
([`internal/daemon/`](../daemon/README.md)'s `Broker`), and `domain.Clock`. The bearer token
is minted per restart by the daemon; this package never reads the token file.

Deferred by design: the remaining ADD §23.4 pause-mutation endpoints (`POST /v1/pauses`,
`:cancel`, `:resume`) — `auspex pause|resume` already covers manual mutation locally.
