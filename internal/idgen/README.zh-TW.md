# internal/idgen/ — domain.IDGenerator 的 UUIDv7 實作

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`idgen` 套件提供 [`../domain`](../domain/) 的 `IDGenerator` 介面之真實實作。套件契約就是
`idgen.go` 檔案開頭的套件註解（沒有另外的 `doc.go`）。

依照 [CONTRACT_FREEZE.md](../../docs/implementation/vertical-slice/CONTRACT_FREEZE.md)
的 ID 規則，所有 Auspex 擁有的實體 ID（`internal/domain/ids.go` 中的 opaque 字串型別）在產生時都是
UUIDv7，由這裡產生，且呼叫端絕不會解析其含義。

- `UUIDv7` — 以 `github.com/google/uuid` 為基礎的無狀態實作；可安全並行使用。
- `New() domain.IDGenerator` — production wiring 所使用的建構函式。
- `NewID()` 會回傳小寫、以連字號分隔的 UUIDv7 字串。UUIDv7 具時間排序性，讓
  [`../storage/sqlite`](../storage/sqlite/README.md) 的 primary-key／index locality
  維持在合理範圍。

Production 程式碼依賴的是 `domain.IDGenerator`，而非直接依賴這個套件，這樣測試就能注入具決定性的 ID。
