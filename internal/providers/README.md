# internal/providers/ — provider adapters under the capability-based integration model

> 🌐 English | [繁體中文](README.zh-TW.md)

Home of per-provider adapter packages. Auspex talks to coding-agent
providers through narrow, capability-segregated ports declared in
[../app/ports.go](../app/ports.go) (ADD §9.10): `ProviderDetector`,
`ProviderCapabilityReader`, `HookNormalizer`, `ManagedRunner`,
`LiveObserver`, `TurnInterrupter`, `SessionResumer`, `QuotaReader`. What a
provider can do is declared by the frozen `domain.ProviderCapabilities`
struct ([../domain/capability.go](../domain/capability.go)), which matches
ADD §8.6 field for field. (ADD section citations refer to
`Auspex_ADD.md`, which now lives at
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md).)

The governing principle is capability-based integration (ADD §6.7): when a
provider lacks a capability, the dependent feature degrades explicitly and
the core keeps working — e.g. missing quota data is reported as unknown
with lower confidence, never substituted with zero (ADD §8.8 degradation
rules). No production implementation of the detector/capability-reader
ports is wired yet (see the `newDoctorCmd` note in internal/cli/root.go).

[claude/](claude/) is the only adapter today; it parses Claude Code's
status-line payloads. The other Claude Code surfaces live in sibling
directories: lifecycle-hook payload parsing in
[../hooks/claude/](../hooks/claude/) and normalization into the frozen
`pkg/protocol/v1.Event` envelope in
[../telemetry/claude/](../telemetry/claude/). A Codex adapter is milestone
M7/M8, tracked in issue
[#9](https://github.com/huaiche94/auspex/issues/9); until then the managed
runner (internal/managed) accepts only `claude`.

This directory holds no Go files of its own and no doc.go; each adapter
package carries its own package comment.
