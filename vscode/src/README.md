# vscode/src/ — TypeScript sources of the VS Code companion extension

> 🌐 English | [繁體中文](README.zh-TW.md)

What the extension does and how it connects to the daemon is
documented in [`../README.md`](../README.md); this is the file map.
Compiled by `tsc` to `out/` (`npm run build`,
[`../tsconfig.json`](../tsconfig.json)).

## Extension-host layer (imports `vscode`; exercised manually)

- `extension.ts` — activation, the status bar item, polling +
  SSE-driven refresh, and the command palette surface (FR-162/163/164).
- `tree.ts` — the Auspex activity-bar tree view. Renders only fields
  the daemon API actually serves; FR-162 sections the API does not
  expose yet render as explicit "not exposed by the daemon API yet"
  placeholders.

## Pure logic layer (no `vscode` import; unit-testable under plain Node)

- `client.ts` — daemon discovery (`daemon.json` metadata + bearer
  token file), authenticated HTTP against the loopback API, and the
  reconnecting SSE subscription. Node built-ins only.
- `paths.ts` — per-OS resolution of Auspex's user directories; a
  line-for-line TypeScript mirror of the Go daemon's
  `internal/paths/paths.go`, so extension and daemon agree on where
  `daemon.json` and `daemon.token` live.
- `sse.ts` — minimal Server-Sent Events parser for the daemon's
  `GET /v1/events/stream`.
- `types.ts` — TypeScript mirrors of the daemon's wire shapes
  (`internal/httpapi/httpapi.go` responses,
  `internal/daemon/metadata.go`, and `pkg/protocol/v1`'s `Event` —
  whose SSE payload uses PascalCase keys because the Go struct has no
  json tags). Every field exists in the Go handlers; nothing invented.

Unit tests for the pure layer live in [`test/`](test/README.md).
