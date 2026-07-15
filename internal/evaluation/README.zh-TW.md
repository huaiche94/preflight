# internal/evaluation/ — EvaluateTurn 管線：執行、持久化、決策、授權

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`Service`（`service.go`）實作了凍結的 `app.EvaluationService` 合約
（`internal/app/ports.go`，ADD §9.9）——這是每一個 provider hook handler 與 runtime
orchestrator 呼叫的 evaluate/decide/authorize 介面，而非直接呼叫具體的 predictor／policy
型別。

主要進入點：

- `EvaluateTurn` 執行完整的 ADR-041 鏈（`pipeline.go`）：透過四個 `app.*` 管線 port
  （由 [`internal/predictor/{scope,token,quota,risk}`](../predictor/README.md) 實作）
  走過 Scope → Token → Quota → Risk，再交給 [`internal/policy`](../policy/README.md) 的
  `Decider`。它會在單一 transaction 內持久化一筆 `feature_vectors`（migration 0040）、
  一筆 `predictions`（0041），以及一筆 `policy_decisions`（0043）——有 prediction
  卻沒有對應 decision，屬於不完整的 evaluation，不是合法狀態。
- `Decide` 是讀回（read-back），而非重新計算：`app.DecideRequest` 只帶有一個
  EvaluationID，因此它回傳的是 `EvaluateTurn` 早已存好的那筆 `policy_decisions` 資料列。
- `IssueAuthorization` / `ConsumeAuthorization`（migration 0044）：一次性授權，具備恰好
  一次（exactly-once）的消耗、綁定時鐘的到期時間，以及在儲存層以原子方式強制執行的
  prompt／session 綁定（`authorization_test.go` 涵蓋 replay／綁定強化，predictor-10）。
- `ForecastCard` / `LatestForecastCard` / `StatusLineText`（`forecastcard.go`）：
  issue #14 的呈現層（presenter），讀回持久化的資料列供 UserPromptSubmit hook、
  statusline 與 `auspex evaluate` 使用。在某個校準波次真正持久化一個值之前，
  `Probability` 結構上為 nil（JSON null）；對外呈現時，未校準的輸出會被標示為一種
  估計值（estimate）（Constitution principle #2）。
- `DataSource`（`datasource.go`）現在是凍結的 `app.FeatureDataSource` port（ADR-044）
  的別名；`SQLDataSource`（`datasource_sql.go`）在 SQLite 之上實作它，把
  [`internal/features`](../features/README.md) 形狀的 DTO 提供給管線各階段。

獨立的 runway 推估（[`internal/predictor/runway`](../predictor/runway/README.md)）絕不會
在這裡重新計算——DataSource 會提供最近一次已計算好的結果，讓 policy 的 runway gate
得以被評估。成本範圍來自 [`internal/pricing`](../pricing/README.md)。原始 prompt 文字
絕不會進入本套件——只會有它的 hash（privacy 合約）。程式碼中引用的 ADD 章節見
[Auspex_ADD.md](../../docs/design/Auspex_ADD.md)。套件合約詳見 `doc.go`。
