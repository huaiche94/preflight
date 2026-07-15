# internal/pause/ — 優雅暫停（Graceful Pause）／安全點（Safe Points）狀態機及其組合服務

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

實作 ADD §20 所定義的 Graceful Pause 生命週期（[`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md)）。
套件合約請見 [`doc.go`](doc.go)，其中包含十二個凍結的 `domain.PauseStatus` 線上字串，如何
對應到設計文件中以文字描述的狀態名稱。

關鍵檔案：

- [`statemachine.go`](statemachine.go) —— 針對凍結列舉的純粹 `(current state, event) → next state` 轉換表；`Validate`、`Apply`、`IsTerminal`、`ValidEvents`。不含任何 I/O。
- [`observe.go`](observe.go) —— 針對 runway 觀測值的防手震（debounce）／遲滯（hysteresis）處理；[`requestpause.go`](requestpause.go) —— 具冪等性（idempotent）的 `RequestPause`，以及作為參考實作的記憶體內 `MemStore`。
- [`safepoint.go`](safepoint.go) —— 安全邊界偵測，以及 `PersistThenInterrupt` 的順序保證；[`persistphase.go`](persistphase.go) —— 五階段的持久化協調器（persist orchestrator）；[`interrupt.go`](interrupt.go) —— `InterruptAndSleep`。
- [`lifecycle.go`](lifecycle.go) —— 手動觸發的 `Cancel` ／ `Resume`；[`wake.go`](wake.go) —— 由 scheduler 驅動的 `Wake`，強制在 pause 層級保證 exactly-once（恰好一次）轉換，即使發生 split-brain 導致兩個租約持有者同時存在，也不會讓同一個 pause 被推進兩次。
- [`resumevalidation.go`](resumevalidation.go) —— `ValidateResume` 檢查清單（配額安全性、repository 相容性、session 能力、授權），以及當配額仍不安全時重新排程 wake job 的邏輯。
- [`service.go`](service.go) —— `Service`，即具體的 `app.GracefulPauseService`（[`../app/ports.go`](../app/ports.go)），組合了以上所有元件。
- [`sqlitestore.go`](sqlitestore.go) —— 建構於 `pause_records` 資料表之上、具持久性的 `PauseStore`（migration 0050）。
- [`contextstore.go`](contextstore.go) —— pause 情境（`QuotaBaseline`、`GitHeadBaseline`、`WorktreeID`、`PausedWorkPaths`）持久化於既有 `pause_records.metadata_json` 欄位中的 `"context"` 鍵之下，並以保留合併（merge-preserving）的 read-modify-write 方式寫入，確保同層其他鍵值不會遺失；這是必要設計，因為發出 pause 請求的行程，絕不會是之後負責 resume 的 daemon 行程（決策 D-16，[`docs/DECISION_LOG.md`](../../docs/DECISION_LOG.md)）。

處於休眠（sleeping）狀態的 pause，是透過 [`internal/scheduler/`](../scheduler/README.md) 所
擁有的持久化 wake job 來喚醒；在無人值守（unattended）情況下驅動 wake → resume 的常駐迴圈，是
[`internal/daemon/`](../daemon/README.md) 的 worker。CLI 介面則透過
[`internal/orchestrator/pauselifecycle.go`](../orchestrator/pauselifecycle.go)。
