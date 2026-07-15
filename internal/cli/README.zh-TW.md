# internal/cli/ — Cobra 指令樹、具 schema 版本的 JSON 輸出、型別化的錯誤合約

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

只對外提供 Cobra 指令的*建構函式*（`NewRootCmd` 及其他相關函式），絕不提供套件層級的指令
實例，也絕不呼叫 `os.Exit`，因此整個指令樹可以完整測試（ADD §10.1——
[`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md)）。套件合約請見
[`doc.go`](doc.go)，其中也說明了為何 hook 子指令採用 kebab-case 命名。

指令樹（[`root.go`](root.go)）：`version`、`init`、`hook`、`evaluate`、`decision`、
`checkpoint`、`progress`、`state`、`pause`、`resume`、`scheduler`、`daemon`、`status`、
`doctor`、`gc`、`export`、`run`。真正的處理器由 [`internal/app/wiring/`](../app/wiring/README.md)
替換進來；若某個指令對應的服務尚未被組裝，則會回傳型別化的 `unavailable` 錯誤，而不會假裝
自己能正常運作。

輸出合約：

- 成功時的輸出是每個指令各自具備 schema 版本的 JSON——例如 `auspex.evaluate.v1`（[`evaluate.go`](evaluate.go)）、`auspex.checkpoint-create.v1`（[`checkpoint.go`](checkpoint.go)）、`auspex.daemon-install.v1`（[`daemon.go`](daemon.go)）。
- 每一條錯誤路徑都會輸出同一個共用的 `auspex.error.v1` 信封（envelope）（[`errors.go`](errors.go)：`SchemaVersionError`、`RenderErrorJSON`、`WithJSONErrorRendering`），包裹凍結的 `domain.Error` 欄位（`code`、`message`、`retryable`、`details`）；非 `domain.Error` 的值會被輸出為 `internal`／不可重試（non-retryable），而不是完全不輸出 JSON。`SilenceErrors` 讓 Cobra 不會再額外附加一行純文字錯誤訊息。

測試：[`golden_test.go`](golden_test.go) 會將完整的成功輸出結構與 `testdata/golden/*.golden.json`
中的固定資料（fixture）做結構性比對，因此任何欄位被悄悄新增／刪除／改名都會讓建置失敗；
[`errorcontract_test.go`](errorcontract_test.go) 則對每個指令的錯誤信封把關。

相關套件：業務邏輯位於 [`internal/orchestrator/`](../orchestrator/README.md) 之後，以及
[`internal/app/ports.go`](../app/ports.go) 中凍結的 port；受管理（managed）的 `run` 指令核心是
[`internal/managed/`](../managed/README.md)；daemon 相關指令則驅動
[`internal/daemon/`](../daemon/README.md)。
