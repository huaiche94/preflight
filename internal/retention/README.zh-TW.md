# internal/retention/ — ADR-046 分層 telemetry retention：hot window → rollup → gzip archive → verified delete

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`retention` 套件實作了 [ADR-046](../../docs/adr/0046-tiered-telemetry-retention.md)
的三層 retention 機制，也是 `auspex gc`（輸出格式為有 schema 版本的 `auspex.gc.v1`）背後的引擎。套件契約就是
`engine.go` 檔案開頭的套件註解（沒有另外的 `doc.go`）。

三個層級：

1. **Hot raw window**（`policy.go`）— 比這個時間窗還新的原始資料列絕不會被動到。預設值
   `DefaultRetentionDays = 90`；`Policy.Cutoff` 採用嚴格的「早於（older than）」判斷，因此剛好落在
   cutoff 那一刻的資料列會被保留。
2. **Rollup 摘要資料表**（`rollup.go`，migration `0060_retention.sql`，屬於 ADR-046
   指派的 0060–0069 範圍 — 詳見
   [../storage/sqlite/migrations/README.md](../storage/sqlite/migrations/README.md)）—
   在原始資料列離開 hot tier 之前，`usage_rollups_daily` 與 `calibration_samples`
   （M13 calibration 所需的 prediction-vs-actual 配對）會在同一筆刪除交易中寫入。
3. **先寫入 Gzip JSONL 封存，再刪除 — fail-closed**（`archive.go`）— 過期的資料列會以每列一個
   JSON 物件、保留完整欄位的方式，寫入
   `<data-dir>/archive/<table>/<YYYY-MM>/…jsonl.gz`，並採用
   [`../repocheckpoint`](../repocheckpoint/) `atomicwrite.go` 中的 temp-file →
   fsync → rename 紀律；在任何刪除動作執行之前，都會重新開啟並重新驗證（列數 + SHA-256）。

進入點：`Engine.Run(ctx, RunRequest) (RunResult, error)`。整個流程有嚴格的順序 — select、
archive、verify，唯有全部完成之後，才會在單一 `app.TxRunner.WithTx` 交易中刪除所有類別的資料及
rollup，並檢查受影響的列數；任何前面步驟失敗，都會讓所有原始資料列維持不變，並在
`retention_runs` 中記錄一筆失敗的執行紀錄。dry run 只會執行 selection 這一步。依賴項目為凍結的
`domain.Clock`／`domain.IDGenerator` port 加上 `*sqlite.DB`
（[../storage/sqlite/README.md](../storage/sqlite/README.md)），因此測試具備決定性。

Retention 屬於跨領域（cross-cutting）關注點（它會跨越每個角色的資料表進行封存與刪除），不屬於任何
vertical-slice 角色所擁有，也不新增任何凍結 port — gc 是 CLI 背後的內部維運事務。
