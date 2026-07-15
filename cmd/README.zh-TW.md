# cmd/ — Go 執行檔的進入點

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

標準的 Go `cmd/` 佈局：每個執行檔各自一個子目錄。Auspex 目前僅發行一個
binary：

- [`auspex/`](auspex/README.md) — `auspex` CLI（`go build
  ./cmd/auspex`）。一個薄薄的 `main` 加上組裝根（composition root）；它所
  組裝的每個服務實作皆位於 [`../internal/`](../internal/) 之下，而它回傳的
  凍結指令樹則由 `internal/app/wiring` 建構。

`cmd/` 底下不包含任何商業邏輯（依
[`Auspex_ADD.md`](../docs/design/Auspex_ADD.md) §10.1）——若這裡的變更超出
組裝（wiring）、路徑解析、或 DTO 形狀轉換的範疇，就應該改放到某個
`internal/` 套件中。
