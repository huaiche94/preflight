# cmd/auspex/ — `auspex` CLI binary（薄 main ＋ 組裝根）

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

依 [`Auspex_ADD.md`](../../docs/design/Auspex_ADD.md) §10.1（於 `main.go`
的 package comment 中引用），本套件只負責組裝（wiring）與行程結束——這裡
以及 Cobra 指令處理器中都不存放任何商業邏輯。凍結的指令樹本身（`evaluate`、
`decision`、`checkpoint`、`pause`/`resume`/`scheduler`、`status`、
`doctor`、`hook claude ...`）由
[`internal/app/wiring`](../../internal/app/wiring/) 建構；本套件負責把
真正的服務實作組裝進該容器中。

## 檔案

- `main.go` — 進入點。`main` 呼叫 `run()`，並把其回傳碼交給 `os.Exit`，
  因此延遲清理（deferred cleanup，例如關閉 DB）一定會在結束前執行完畢。
  此外也保留了 `newRootCmd`，一個僅顯示版本號的最小後備指令，由
  `main_test.go` 直接測試。
- `wire.go` — 組裝根：在作業系統的使用者資料目錄下開啟並 migrate SQLite
  資料庫，為每個凍結的 `app.*` 服務（progress tree、state/repository
  checkpoint、evaluation pipeline、graceful pause、daemon、retention）
  建構一個真正的實作，並組裝成 `internal/app/wiring.Services`；
  `wiring.New(services).RootCmd()` 會回傳完整組裝好的 Cobra 指令樹。
- `adapters.go` — 純粹的介面橋接膠合層：DTO 形狀轉換，以及供套件內部
  接縫（例如 `pause.SessionContextResolver`）使用的唯讀 SQL adapter，
  再加上針對尚未存在能力（managed turn interrupt）、有明確文件說明的
  fail-closed stub。
- `main_test.go` — 測試僅顯示版本號的後備指令。

## 關聯

- [`../../internal/`](../../internal/) — 所有服務實作。
- [`../../pkg/protocol/v1/`](../../pkg/protocol/v1/README.md) — 組裝完成
  的服務所使用的凍結 wire protocol。
