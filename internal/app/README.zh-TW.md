# internal/app/ — 凍結的跨元件服務 port（介面 + DTO）

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

唯一一個原始碼檔案 [`ports.go`](ports.go)：所有元件溝通所依循的凍結合約（ADD §9.9、§9.10——
ADD 文件位於 [`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md)）。介面刻意保持狹窄；
套件層級的註解明文禁止將任何介面擴大成一個「上帝介面」（God interface）。

`ports.go` 定義的內容：

- `TxRunner` ／ `TxFunc` —— 唯一一種凍結的儲存層交易（storage-transaction）回呼（callback）形狀。
- `EvaluationService` —— evaluate／decide／authorize，以及 `PolicyAction`（`RUN` … `BLOCK`）與 `Evaluation` ／ `DecisionResult` ／ `Authorization` 這些 DTO。
- 預測管線各階段的 port（ADR-041）：`ScopeEstimator`、`TokenForecaster`、`QuotaForecaster`、`RiskCombiner`——每個階段皆可獨立替換。
- `ProgressTreeService`、`StateCheckpointService`、`RepositoryCheckpointService`、`GracefulPauseService`（Observe／RequestPause／ReachSafePoint／EnterSleep／Resume／Cancel）。
- 依能力（capability）切分的 provider port：`ProviderDetector`、`ProviderCapabilityReader`、`HookNormalizer`、`ManagedRunner`、`LiveObserver`。

所有權：`internal/app/ports.go` 是凍結的跨切面檔案，僅由 `contract-integrator` 擁有
（[CONSTITUTION.md §4.3](../../CONSTITUTION.md)）；其他角色一律不得編輯。

相關套件：DTO 是以 [`internal/domain/`](../domain/README.md) 的型別建構而成。具體實作分散於
各自擁有的套件中（例如 `GracefulPauseService` 由 [`internal/pause/`](../pause/README.md) 的
`Service` 實作、`EvaluationService` 由 `internal/evaluation` 實作），再由
[`./wiring/`](wiring/README.md) 組裝成單一容器。[`internal/orchestrator/`](../orchestrator/README.md)
依序呼叫這些 port，但不需要知道背後的具體實作為何。本套件沒有 `doc.go`；套件註解位於
`ports.go` 檔案開頭。
