# internal/statecheckpoint/ — State Checkpoint（狀態檢查點）manifest：建構、封存、驗證、持久化

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

State Checkpoint（狀態檢查點）是任務進度樹（Progress Tree）在某個語意邊界上，具持久性、
可重播（replayable）的快照（Auspex_ADD.md §18.8 及附錄 B——ADD 現位於
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)；線上 schema 為
`auspex.state-checkpoint.v1`）。本套件負責 manifest 的 Go 結構型別、其確定性（deterministic）
JSON 序列化方式、完整性校驗碼，以及 `state_checkpoints` 儲存層——但不負責決定「何時」該建立
checkpoint，那屬於 [`../progress/`](../progress/) `CompleteNode` 協定的職責（Constitution
§6.3：每次節點完成都會在同一個原子操作中建立一份 State Checkpoint）。

關鍵組成：

- **`Manifest`**（`manifest.go`）—— 附錄 B 所定義的文件；`IntegritySHA256` 欄位宣告在最後，且
  不計入自身的摘要值（digest）之中。
- **`Build`**（`build.go`）—— 從呼叫端已持有的 `BuildInput` 快照組裝出一份尚未封存的
  manifest；刻意不匯入（import）`internal/progress`（依賴方向是反過來的）。
- **`Digest` ／ `Seal` ／ `Marshal` ／ `Verify`**（`serialize.go`）—— 摘要值是對「`IntegritySHA256`
  歸零後的標準 JSON 編碼」計算 SHA-256；欄位順序固定為結構體宣告順序，因此編碼結果可重現。
  `Verify` 一律會從 manifest 自身內容重新計算摘要值，再與儲存的 `integrity_sha256` 比對——
  儲存的值只會被檢查，絕不會被直接信任。
- **`Store`**（`store.go`）—— 對 `state_checkpoints` 進行 CRUD（`migrations/0023_state_checkpoints.sql`）；
  資料列會重複儲存 `manifest_json` 中可查詢的一個子集。
- **`Service`**（`service.go`）—— 實作凍結的 `app.StateCheckpointService` port：`Create`
  （獨立於任何節點完成之外、臨時建立的快照）、`LoadLatest`、`Snapshot`、`Verify`。
- **`Reconciler`**（`reconcile.go`）—— 啟動時的協調流程（ADD §18.9）：一個唯讀的完整性掃描，
  會重新計算每一列的摘要值、檢查 schema 版本，並交叉核對 manifest ID 與資料列欄位是否一致。
  依設計它不會修復任何東西——唯一的持久化寫入（`Store.Insert`）是單一原子 SQL 陳述式，因此
  不可能出現寫到一半的資料列狀態。
