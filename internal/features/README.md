# internal/features/ — privacy-bounded feature extraction and the task classifier

> 🌐 English | [繁體中文](README.zh-TW.md)

This package derives prediction-input signals from prompts, repositories, sessions, and Progress
Trees (ADD §14.2), and classifies each turn onto the fixed ADD §14.3 task taxonomy. It is the
layer [`internal/predictor`](../predictor/README.md) sits above.

Privacy boundary (Constitution §7 rule 2, the frozen privacy contract): raw prompt text enters
this package through exactly one function, `ExtractPromptFeatures` (`prompt.go`), and never
leaves it. No exported type carries raw prompt text, a substring, or any reversible encoding —
only derived counts, flags, scores, and a SHA-256 digest. Adding a field capable of holding raw
prompt text is a contract violation, not a feature.

Key pieces:

- `PromptFeatures`: size signals, the ADD §14.7 tokenizer-free `ApproxTokens` estimate (always
  ConfidenceLow), structure counts (paths, list items, acceptance criteria), verb booleans
  (fix/implement/refactor/investigate/migrate — vocabulary widened per issue #42), and domain
  indicators (tests, schema/API, security, performance, docs, open-ended, cross-layer,
  repository-wide).
- `ClassifyTask` (`classifier.go`): deterministic, fixed-precedence mapping of derived features
  onto the 16 `TaskClass` values (`taskclass.go`). Action verbs win verb collisions (#49);
  insufficient signal returns `TaskClassUnknown` with ConfidenceUnavailable — never a guess.
- `RepositoryFeatures` / `SessionFeatures` / `ProgressFeatures` (`dto.go`): derived-only DTOs
  with pointer quantiles where nil means unknown, never a substituted zero.

Open issues touching this package (both open at time of writing): #50 — the persisted
prompt-feature payload schema duplicates its key literals between the hook-side writer
(`internal/hooks/claude`, `internal/telemetry/claude`) and the read-back decoder
(`internal/evaluation/datasource_sql.go`), with no extraction-version tag; #51 —
`ExtractPromptFeatures` makes several full passes over the prompt (line split, lowercase copy,
word set, field scan, rune loop) with per-call allocations on the blocking hook path.

See `doc.go` for the package contract; ADD sections cited above live in
[Auspex_ADD.md](../../docs/design/Auspex_ADD.md).
