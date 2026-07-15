# internal/redact/ — secret detection for checkpoint artifacts

> 🌐 English | [繁體中文](README.zh-TW.md)

Pattern-based secret detection applied before working-tree content reaches durable checkpoint artifacts
(Auspex_ADD.md §19.5 "secret scan" and §27.8 detector lists — the ADD now lives at
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)). See `doc.go` for the package contract,
including an explicit statement of what these detectors do and do not catch.

Two detector classes:

- **Filename patterns** (`filename.go`, `MatchesSecretFilename`) — the §27.8 name list (`.env`, `.env.*`,
  `*.pem`, `*.key`, `*.pfx`, `*.p12`, `id_rsa`, `id_ed25519`, `credentials.json`, `auth.json`,
  `secrets.*`), matched against the base filename before any content is read.
- **Content detectors** (`patterns.go`, `Detectors`) — fixed regular expressions for bearer tokens, PEM
  private-key headers, GitHub/OpenAI/Anthropic API key shapes, Azure storage account keys, JWT-shaped
  tokens, and password/connection-string patterns. `ScanPath`/`ScanContent` (`scan.go`) return `Finding`s;
  content is scanned up to `MaxContentScanBytes` (1 MiB) per file, and likely-binary content is skipped
  (`IsLikelyBinary`).

Consumers, both in [`../repocheckpoint/`](../repocheckpoint/): the untracked-archive path skips matching
files and records a skip-ledger entry (`archive.go`), and the patch path redacts detected spans in place on
`+`/`-` line bodies of staged/unstaged patches (`patchredact.go`).

Accepted residual risk: [ADR-042](../../docs/adr/0042-patch-redaction-residual-surface.md) deliberately
excludes patch file-path header lines and binary-diff header/payload lines from redaction — rewriting them
would break `git apply` and destroy the checkpoint's evidentiary value; the boundary is pinned by tests in
`../repocheckpoint/patchredact_internal_test.go`.

This is a documented set of shape-based detectors, not an exhaustive scanner — one layer of
defense-in-depth alongside qa's independent leakage scan (`doc.go`).
