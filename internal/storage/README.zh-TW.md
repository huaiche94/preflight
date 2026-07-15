# internal/storage/ — 儲存轉接器（目前僅有 SQLite）

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

本目錄是 Auspex_ADD.md §7.6 所固定之依賴方向中的 storage-adapter 分支（該 ADD 位於
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)）：adapter 位於整個堆疊的最底層，在
domain service 與 interface 之下。這一層沒有任何 Go 程式碼 — 它只是為具體的儲存後端命名空間。

它唯一的子套件是 [`sqlite/`](sqlite/README.md)，也就是
[CONTRACT_FREEZE.md](../../docs/implementation/vertical-slice/CONTRACT_FREEZE.md)
import-path 對照表中凍結的「SQLite runtime」import path：連線設定與持久性 pragma、每個涉及儲存的操作都要經過的
`app.TxRunner` 交易邊界，以及包含內嵌 [`sqlite/migrations/`](sqlite/migrations/README.md)
檔案的 forward-only migration engine。

更上層不會直接引入 `database/sql` 的細節；它們要嘛依賴 [`../app`](../app/) 中凍結的 port（例如
`TxRunner.WithTx`），要嘛（對於角色擁有的 store，例如 [`../progress`](../progress/)、
[`../telemetry/claude`](../telemetry/claude/)）依賴 `sqlite.DB` 的交易／querier 介接點。
