# vscode/scripts/ — build/test helper scripts for the extension

> 🌐 English | [繁體中文](README.zh-TW.md)

One script today:

- `run-tests.js` — deterministic launcher for the compiled unit
  tests; `npm test` runs it (see `scripts.test` in
  [`../package.json`](../package.json), after `pretest` builds
  `src/` → `out/`). It enumerates `out/test/*.test.js` itself,
  **fails loudly if zero test files are found** (zero discovered
  tests must never look green), and hands `node --test` an explicit
  file list.

The script exists because `node --test`'s positional-path semantics
differ across Node versions (its header comment records the CI
regression this replaced): Node 20 scans a directory argument, Node
22 treats it as a module and dies with `ERR_MODULE_NOT_FOUND`, and on
Node ≥ 21 an unmatched glob runs zero tests and exits 0. An explicit
file list behaves identically everywhere.

CI runs the same `npm test` entrypoint in the `vscode` job of
[`../../.github/workflows/ci.yml`](../../.github/workflows/ci.yml)
(Node pinned to exactly 22.11.0, per the repository's exact-version
pinning policy). The tests themselves live in
[`../src/test/`](../src/test/README.md).
