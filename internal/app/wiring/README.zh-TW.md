# internal/app/wiring/ — 凍結 port 的行程內組合根（composition root）

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

唯一一個原始碼檔案 [`wiring.go`](wiring.go)。它只負責一件事：把每個凍結服務介面
（[`../ports.go`](../ports.go)）各挑一個實作，收集成一個經過驗證的容器，供執行環境其餘部分
使用，且完全不需要知道拿到的是哪個具體實作（ADD §13——
[`docs/design/Auspex_ADD.md`](../../../docs/design/Auspex_ADD.md)）。

關鍵型別與進入點：

- `Services` —— 五個必填的介面欄位（`Evaluation`、`ProgressTree`、`StateCheckpoint`、`GracefulPause`、`RepositoryCheckpoint`），外加選填的支援套件（bundle）：`Hooks`（hook 處理器／事件持久化）、`Diagnostics`（`auspex doctor` 的檢查項目——省略代表每項檢查都回報「已略過」而非錯誤），以及 `pause` ／ `resume` ／ `scheduler run-once` 所需的 pause 生命週期支援。
- `New(Services) (*App, error)` —— 拒絕接受只填了一部分的結構；真實實作、`internal/testutil/fakes` 的替身（doubles），以及未來的組合實作都是同樣有效的值。
- `App.RootCmd()` —— 取用 [`internal/cli/`](../../cli/README.md) 的樁（stub）指令樹，並針對已設定的選填套件替換成真正的處理器（`replaceSubcommand`），藉此建構出 Cobra 指令樹。

明確排除的目標：在 `cmd/auspex/main.go` 中的最上層組裝並非本套件的職責——將此容器組裝進
最終執行檔的工作，由 `contract-integrator` ／ `foundation` 角色負責。

相關套件：使用 [`internal/cli/`](../../cli/README.md) 的指令建構函式，以及
[`internal/orchestrator/`](../../orchestrator/README.md) 的依賴套件（deps bundle）；替換測試
（`evaluate_swap_test.go`、`gc_swap_test.go`、`restart_test.go`）證明了真實實作可以在不改變函式
簽章的情況下取代假實作。本套件沒有 `doc.go`；套件註解位於 `wiring.go` 檔案開頭。
