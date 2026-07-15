# vscode/src/test/ — unit tests for the extension's pure logic layer

> 🌐 English | [繁體中文](README.zh-TW.md)

Tests use Node's built-in test runner (`node:test`) — no VS Code
download and no `@vscode/test-electron` harness (deliberately left
out of the MVP; see [`../../README.md`](../../README.md)).

- `paths.test.ts` — every OS branch of [`../paths.ts`](../paths.ts)
  with injected env/home.
- `sse.test.ts` — SSE parsing of the exact daemon stream shapes,
  chunk splits, CRLF, heartbeats, and the reconnect backoff schedule
  ([`../sse.ts`](../sse.ts)).
- `types.test.ts` — response/metadata parsing against fixtures copied
  field-for-field from the Go handlers ([`../types.ts`](../types.ts)).

## How they run

`npm test` compiles `src/` → `out/` (the `pretest` build), then
[`../../scripts/run-tests.js`](../../scripts/run-tests.js) enumerates
the compiled `out/test/*.test.js` files and hands `node --test` that
explicit file list, failing loudly if zero test files are discovered
— never a silent green on an empty run. CI's `vscode` job in
[`../../../.github/workflows/ci.yml`](../../../.github/workflows/ci.yml)
runs the same `npm test` step on Node pinned to exactly 22.11.0.

Not covered here: `extension.ts`/`tree.ts` (extension-host UI wiring,
exercised manually) and the live network paths in `client.ts`
(smoke-tested against a real daemon).
