# internal/providers/ — 能力導向整合模型下的 provider adapter

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

Per-provider adapter 套件的根目錄。Auspex 透過 [../app/ports.go](../app/ports.go)
（ADD §9.10）中宣告的窄化、依能力切分的 port 與 coding-agent provider 溝通：
`ProviderDetector`、`ProviderCapabilityReader`、`HookNormalizer`、
`ManagedRunner`、`LiveObserver`、`TurnInterrupter`、`SessionResumer`、
`QuotaReader`。一個 provider 具備哪些能力，由凍結（frozen）的
`domain.ProviderCapabilities` struct
（[../domain/capability.go](../domain/capability.go)）逐欄位對應 ADD §8.6 而定。
（ADD 章節引用皆指 `Auspex_ADD.md`，該文件現位於
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)。）

指導原則是能力導向整合（capability-based integration，ADD §6.7）：當某個
provider 缺少某項能力時，依賴該能力的功能會明確降級，而核心功能仍可繼續運作
——例如缺少的配額資料會回報為「未知」並附帶較低的信心水準，絕不會以零值替代
（ADD §8.8 降級規則）。目前尚未接上任何 detector/capability-reader port 的
正式（production）實作（詳見 internal/cli/root.go 中 `newDoctorCmd` 的註記）。

[claude/](claude/) 是目前唯一的 adapter，負責解析 Claude Code 的 status-line
payload。Claude Code 的其他介面位於相鄰目錄：lifecycle-hook payload 的解析在
[../hooks/claude/](../hooks/claude/)，正規化為凍結的
`pkg/protocol/v1.Event` envelope 則在
[../telemetry/claude/](../telemetry/claude/)。Codex adapter 屬於里程碑
M7/M8，追蹤於 issue
[#9](https://github.com/huaiche94/auspex/issues/9)；在那之前，managed
runner（internal/managed）僅接受 `claude`。

此目錄本身沒有任何 Go 檔案，也沒有 doc.go；每個 adapter 套件都各自帶有自己的
package comment。
