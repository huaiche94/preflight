# internal/storage/sqlite/migrations/ — 內嵌、forward-only 的 schema migration，依角色劃分編號範圍

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

直接放在這裡的每一個 `NNNN_name.sql` 檔案都會被內嵌進執行檔中（在
[../migrate.go](../migrate.go) 裡以 `go:embed` 處理），並由 `DB.Migrate`
（[../README.md](../README.md)）套用。檔名格式錯誤與重複版本一律是載入時的錯誤，絕不會被跳過。Migration
一律 forward-only — 沒有 down migration（Auspex_ADD.md §12.5；該 ADD 位於
[docs/design/Auspex_ADD.md](../../../../docs/design/Auspex_ADD.md)）。

## 編號範圍（依照 [CONTRACT_FREEZE.md](../../../../docs/implementation/vertical-slice/CONTRACT_FREEZE.md)「Migration ranges」— 依角色分配，而非依時間順序）

| Range | Owner |
|---|---|
| 0000–0009 | foundation |
| 0010–0019 | claude-provider |
| 0020–0029 | checkpoint Part A（progress/state） |
| 0030–0039 | checkpoint Part B（repository） |
| 0040–0049 | predictor |
| 0050–0059 | runtime Part A（pause/scheduler） |
| 0060–0069 | retention/gc（cross-cutting；由 [ADR-046](../../../../docs/adr/0046-tiered-telemetry-retention.md) 指派） |

runtime Part B 沒有分配範圍；除非 contract-integrator 指派，否則它不會新增任何 schema。

## 集合差集套用與編號有缺口的補登（backfill）

由於版本號是依範圍分配的，一個 migration 可能會在對應範圍已經被實際資料庫套用「之後」才進到 git（issue #22：
`0045` 是在 `0050–0052` 上線之後才進來的）。#22 的修法是：`Migrate` 會把待處理的工作計算為對照所有已套用
`schema_migrations` 資料列的**集合差集** — 而不是「所有版本號大於 MAX(version)
的項目」，因為後者會永遠悄悄跳過這類補登。如同 [../migrate.go](../migrate.go) 中
`Migration.Version` 上所記載的：一個補登 migration 的 SQL，是在既有資料庫上、於較大編號的 migration
已經執行完之後才執行，因此它不能依賴與自身範圍以上之範圍的相對順序（只要是針對自己所屬 domain
資料表的新增性語句，自然就能滿足這個條件）。fail-closed 的 `ErrSchemaNewerThanBinary` 檢查只以*最大*已套用版本為依據
— 一個低於執行檔自身最大版本的已套用版本，代表這是一個此執行檔尚未認識、但已經套用過的補登，會被正確地忽略。單次
`Migrate` 呼叫會在單一 `BEGIN IMMEDIATE` 交易中套用所有待處理的 migration。
