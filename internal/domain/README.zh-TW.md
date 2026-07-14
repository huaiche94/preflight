# internal/domain/ — 凍結的領域模型（domain model）：實體、ID、列舉與錯誤結構

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

僅包含純型別（pure types）——沒有 I/O、沒有服務邏輯、沒有第三方套件匯入。其他每一層
（ports、orchestrator、CLI、daemon）都使用這些型別溝通。

關鍵檔案：

- [`ids.go`](ids.go) — 以 `string` 為底層型別的不透明（opaque）ID 型別（`RepositoryID`、`SessionID`、`TurnID`、`EvaluationID`、`TaskID`、`PauseID`、`WakeJobID`……）；產生時使用 UUIDv7，且從不解析其內容意義。
- [`status.go`](status.go) — 凍結的列舉線上字串（wire strings）（`TurnStatus`、`PauseStatus` 及其他相關型別），由 [`status_test.go`](status_test.go) 驗證。
- [`errors.go`](errors.go) — `domain.Error`，具備型別化的 `ErrorCode`（`validation`、`not_found`、`conflict`、`unauthorized`、`integrity`、`unavailable`、`internal`）以及 `Retryable`；這是每個指令與 API 端點都會輸出的錯誤結構。
- [`usage.go`](usage.go) — `UsageObservation`／配額（quota）／情境（context）觀測值，欄位皆為指標（pointer）：`nil` 代表未知，絕不會以零值替代。
- [`forecast.go`](forecast.go) — `ReasonCode`，預測與政策決策所引用的封閉詞彙表（ADD §16.4）。
- [`capability.go`](capability.go) — `ProviderCapabilities`；`false` 代表「確認不存在」，而非「尚未檢查」。
- [`clock.go`](clock.go) — `Clock` 與 `IDGenerator` 介面。
- [`checkpoint.go`](checkpoint.go)、[`measurement.go`](measurement.go)、[`failure.go`](failure.go)、[`artifact.go`](artifact.go) — checkpoint、量測來源（measurement-source）、失敗類別（failure-class）與證據／產出物（evidence/artifact）型別。

所有權：`internal/domain/**` 是共用的跨切面（cross-cutting）路徑，僅由 `contract-integrator`
角色擁有——任何其他角色都不得編輯（[CONSTITUTION.md §4.3](../../CONSTITUTION.md)）。合約層級
（contract-level）的變更需經過 ADR 流程（Constitution §3，[`docs/adr/`](../../docs/adr)）。凍結的結構彙整於
[`docs/implementation/vertical-slice/CONTRACT_FREEZE.md`](../../docs/implementation/vertical-slice/CONTRACT_FREEZE.md)。

相關套件：[`internal/app/ports.go`](../app/ports.go) 中凍結的服務 port 會以這些型別建構其
DTO。程式碼中所有「ADD」章節參照皆指向
[`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md)。本套件沒有 `doc.go`；
合約說明分散於各檔案的註解中。
