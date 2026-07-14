# internal/scheduler/ — 具持久性（durable）的喚醒排程器：針對 wake_jobs 的租約認領／續約／完成／失敗

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`Store` 操作 `wake_jobs` 資料表（migration
`internal/storage/sqlite/migrations/0051_wake_jobs.sql`），精確實作了 ADD §12.4 所規範的
`BEGIN IMMEDIATE` 租約認領交易（[`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md)）：
在有多個併發認領者（claimer）的情況下，一次只會有一個 worker 取得恰好一個到期且尚未被租用的
job，且具持久性保證。套件合約請見 [`doc.go`](doc.go)。

Job 狀態是本套件自有的詞彙表——`scheduled` → `leased` → `done` | `dead`——與凍結的
`domain.PauseStatus` 列舉是不同的兩套概念。

關鍵檔案：

- [`lease.go`](lease.go) —— `NewStore`、`Schedule`、`Get`、`GetByPauseKind`、`Claim`、`Renew`、`Complete`、`Fail`（在還有剩餘嘗試次數時會退避（backoff）並重新設為 `scheduled`，用盡後才變為 `dead`）、`ReclaimExpired`；`DefaultLeaseDuration` 為 60 秒（ADD §20.7）。
- [`restart.go`](restart.go) —— 租約復原（lease recovery）：`Restart` 在行程啟動時，會把每一列 `leased` 狀態的資料釋放回 `scheduled`（因為一個全新的行程絕不可能是租約持有者），並在 `RestartReport` 中回報已逾期且可認領的 job；實際認領仍是 `Claim` 的職責。
- [`list.go`](list.go) —— 提供給 daemon 狀態介面使用的唯讀 `List`；不會取得租約，也不會影響認領順序。
- [`cancel.go`](cancel.go) —— 操作者主動取消（FR-163，issue #10）：沿用終態（terminal）的 `dead`，並將 `last_error` 設為 `CancelledByOperator`，而不是額外發明第五種狀態。

執行模式：`auspex scheduler run-once`（[`internal/orchestrator/pauselifecycle.go`](../orchestrator/pauselifecycle.go)）
只執行一次認領掃描，並刻意在 `Claim` 之後就停止；daemon 的常駐 worker
（[`internal/daemon/worker.go`](../daemon/README.md)）則會持續驅動完整的
claim → wake → resume → complete/fail 迴圈。payload 的意義歸屬於
[`internal/pause/`](../pause/README.md)，而非本套件。
