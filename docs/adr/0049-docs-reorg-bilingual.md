# ADR-049 — Documentation reorganization: design docs under `docs/design/`, per-folder READMEs, Traditional Chinese translations

> 🌐 English | [繁體中文](0049-docs-reorg-bilingual.zh-TW.md)

Status: Accepted
Date: 2026-07-14
Owner: lead
Approved by: repository owner, 2026-07-14 (documentation reorganization request)

## Context

The repository root carried three large design documents
(`Auspex_ADD.md`, `Auspex_Predictor_Design_Supplement.md`,
`Auspex_Parallel_Execution_Plan.md`) alongside the nine
community/process files GitHub and agent tooling expect at root
(`README.md`, `AGENTS.md`, `CHANGELOG.md`, `CODE_OF_CONDUCT.md`,
`CONSTITUTION.md`, `CONTRIBUTING.md`, `GOVERNANCE.md`, `SECURITY.md`,
`SUPPORT.md`). A first-time viewer landing on the repo had no way to
tell which of the twelve root markdown files was the entry point, and
most folders had no introduction at all. The repository owner requested
(2026-07-14): a README rewritten for first-time viewers, root markdown
organized into `docs/`, a `README.md` introduction in every folder, and
a Traditional Chinese version of every markdown document.

Changing where `Auspex_ADD.md` lives requires editing the Constitution's
path references (§1, §2, §8), which per Constitution §3 requires an ADR.

## Decision

1. **`docs/design/` is the new home of the three design documents.**
   Filenames are unchanged, so section citations by document name
   (`Auspex_ADD.md §31`) remain unambiguous and greppable.
2. **Living documents cite the new path.** `CONSTITUTION.md`,
   `CONTRIBUTING.md`, `GOVERNANCE.md`, `SECURITY.md`, `SUPPORT.md`,
   `AGENTS.md`, `README.md`, and `agents/*.md` now reference
   `docs/design/Auspex_ADD.md` (and siblings).
3. **Historical records are NOT rewritten** — accepted ADRs
   (immutable per Constitution §3.3), `docs/archive/**`,
   `docs/implementation/**` progress logs, Go source comments, JSON
   schema description strings, and checksummed test fixtures
   (`testdata/**`) keep their original citations. The document names
   still resolve by grep; only hyperlinks in living documents needed to
   stay valid.
4. **Every folder gets a `README.md` introduction**, except fixture
   directories whose contents are enumerated or checksummed by tests
   (`testdata/*` leaf dirs, `internal/cli/testdata/`,
   `internal/managed/testdata/`) — those are documented by their nearest
   ancestor README instead, so adding files cannot break tests.
5. **Bilingual documentation policy.** Every documentation markdown
   file gets a sibling Traditional Chinese translation named
   `<name>.zh-TW.md`, cross-linked from the top of both files.
   - **The document's original authorship language is the normative
     text.** For every documentation file that is English — which is
     all of them except the two named below — the English document is
     normative and the `.zh-TW.md` sibling is a non-normative reading
     aid: where they diverge, the English document wins and the
     translation is the bug.
   - **Two documents are authored in Traditional Chinese and are
     normative as written:** `docs/design/Auspex_ADD.md` (the
     architecture authority — its prose body is Traditional Chinese
     with English section labels and code) and `docs/DECISION_LOG.md`.
     They get no `.zh-TW.md` sibling (they would be duplicates that
     drift). If an English translation is ever added for either, that
     translation is the non-normative side.
   - Code blocks, JSON payloads, command names, file paths, schema
     version strings, and identifiers are never translated.
   - Markdown files that are test inputs (for example
     `testdata/checkpoints/state/add-section-18-*.md`) are fixtures,
     not documentation, and are not translated.

## Consequences

- Root markdown count drops from 12 to 9; everything left at root is
  either a GitHub community convention file, agent-tooling convention
  (`AGENTS.md`), or the process authority (`CONSTITUTION.md`).
- Constitution §1's source-of-truth table now points at
  `docs/design/Auspex_ADD.md`; the authority *content* is unchanged —
  this ADR changes locations and adds a translation policy, no
  architectural or process rule.
- Anyone adding a new documentation file after this ADR is expected to
  add its `.zh-TW.md` sibling in the same change; `docs/repository_inventory.md`
  records the convention.
- Translation freshness is best-effort: a PR that edits an English
  document should update its translation, but a stale translation is
  a documentation bug, never an authority question.
