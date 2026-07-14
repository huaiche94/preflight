# internal/progress/ — 進度樹（Progress Tree）：任務狀態的權威持久化來源

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

實作進度樹（Progress Tree）領域服務（Constitution §6；Auspex_ADD.md §18——ADD 現位於
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)）。進度樹是任務狀態的權威持久化
來源：對話情境（conversation context）以及 agent 自己宣稱的「已完成」，都絕不是真實依據
（source of truth）。

關鍵組成：

- **Stores** —— `TaskStore`（`tasks`）、`NodeStore`（`progress_nodes`）、`EdgeStore`（`progress_edges`）、
  `ArtifactStore`（`artifacts` 證據列及其 `validation_status`）。`NodeStore` 在持久化狀態變更之前，
  一定會先呼叫 `ValidateTransition`。
- **狀態機**（`statemachine.go`）—— 固定的 `domain.ProgressNodeStatus` 列舉（`pending`、`ready`、
  `in_progress`、`checkpointing`、`paused`、`completed`、`failed`、`skipped`、`blocked`；
  `internal/domain/status.go`，Constitution §6.4），搭配凍結的轉換表；不允許任何臨時（ad hoc）狀態。
- **`CompleteNode`**（`complete_node.go`）—— 以證據為門檻（evidence-gated）的原子完成協定：先對照
  `node_completions` 帳本（ledger）進行冪等性檢查（`idempotency.go`；相同鍵值加相同 payload 會重播
  先前結果，相同鍵值但不同 payload 視為衝突，絕不會被靜默合併——Constitution §6.6），接著將證據
  暫存（staging）為內容定址（content-addressed）的副本（`stager.go`），透過
  [`../artifacts/`](../artifacts/) 的驗證器（validator）進行驗證，最後在單一 SQLite 交易中完成節點
  狀態轉換、提交 artifact 資料列，並透過 [`../statecheckpoint/`](../statecheckpoint/) 封存並插入一份
  State Checkpoint（狀態檢查點）manifest（Constitution §6.3）。完成與否從不採信 agent 自身的宣稱：
  證據必須真實存在並通過驗證器（§6.2）。事件只有在交易提交（commit）之後才會發布。
- **`Reconciler`**（`reconcile.go`）—— 針對「已暫存 artifact 與資料庫不一致」這段當機時間窗
  （crash window）所做的啟動時協調（ADD §18.9）；這是一個唯讀掃描，只會揭露孤兒（orphaned）
  暫存證據，而不會修改任何狀態。
- **`Service`**（`service.go`）—— 凍結的 `app.ProgressTreeService` port 的具體實作
  （`internal/app/ports.go`），組合以上所有組成；只負責 DTO 轉譯，不含任何新邏輯。

當機復原能力是透過階段層級的當機注入（crash injection）驗證的（`complete_node_crash_test.go`
會在每個具名 `Phase` 之後中止協定，並斷言協調（reconciliation）結果依然成立——Constitution §6.5）。
