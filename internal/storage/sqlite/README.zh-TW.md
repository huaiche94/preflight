# internal/storage/sqlite/ — SQLite runtime：連線、pragma、交易與 migration engine

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`sqlite` 套件是 Auspex 的 SQLite runtime。套件契約就是 `db.go` 檔案開頭的套件註解（沒有另外的
`doc.go`）。依照 Auspex_ADD.md §1.4 的技術堆疊決策（該 ADD 位於
[docs/design/Auspex_ADD.md](../../../docs/design/Auspex_ADD.md)），所使用的 driver 為
`modernc.org/sqlite`（純 Go 實作，不需要 CGO）。

分為兩部分：

- **`db.go` — 連線／pragma／交易引擎。** `Open(ctx, path)` 會回傳一個已設定好 journal／durability
  pragma 的 `*DB`，之後所有角色的儲存程式碼都仰賴這些設定。`DB` 實作了凍結的 `app.TxRunner` port
  （[../../app/ports.go](../../app/ports.go)）：`WithTx(ctx, fn)` 會在同一個交易中執行
  `fn`，遇到非 nil 錯誤時進行 rollback。`Querier` / `QuerierFromContext(ctx, db)` 讓 store
  程式碼可以在 `WithTx` callback 內或外執行相同的查詢。開啟一個空檔案會得到一個有效、設定正確、完全空白的資料庫 —
  `db.go` 不會建立任何 schema。
- **`migrate.go` — forward-only migration engine（ADD §12.5）。** `AllMigrations()`
  會載入透過 `go:embed` 內嵌自 [`migrations/`](migrations/README.md) 的每一個
  `NNNN_name.sql` 檔案；`LoadMigrationsFS` 刻意採取嚴格模式（檔名格式錯誤與重複版本一律視為硬性錯誤，絕不跳過）。
  `DB.Migrate` 會把待套用的 migration 視為對照 `schema_migrations` 中已套用資料列的**集合差集（set
  difference）** — 而不是「所有版本號大於 MAX(version) 的項目」 — 並在單一 `BEGIN IMMEDIATE`
  交易內套用，因此補登、編號有缺口的 migration 都能被正確套用，且並行的 migrator 不會發生 race（範圍／補登規則詳見
  [migrations/README.md](migrations/README.md)）。若資料庫已套用的最高版本比這個執行檔所認得的版本還新，
  `Migrate` 會回傳 `ErrSchemaNewerThanBinary` 且不套用任何內容；呼叫端必須 fail closed，退回唯讀診斷模式（ADD
  §12.5）。`CurrentVersion` 會回報目前已套用的最高版本，供診斷使用。

相鄰套件：所有需要持久化的內容都會經過這個套件 —
[`../../telemetry/claude`](../../telemetry/claude/) 的 EventStore、
[`../../progress`](../../progress/)、[`../../evaluation`](../../evaluation/)、
[`../../scheduler`](../../scheduler/)，以及 [`../../retention`](../../retention/)
全部都透過 `DB.WithTx` 寫入。
