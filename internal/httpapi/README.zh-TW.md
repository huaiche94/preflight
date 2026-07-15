# internal/httpapi/ — daemon 具身分驗證的 loopback HTTP/JSON + SSE 介面

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

唯一一個原始碼檔案 [`httpapi.go`](httpapi.go)；其套件註解即為合約本身（沒有另外的
`doc.go`）。實作 M6 daemon 的 API（issue #7；ADD §23.2–§23.5、NFR-022——
[`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md)）。

`NewHandler(deps, bearerToken)` 掛載的端點：

- `GET /v1/health`、`/v1/version`、`/v1/capabilities`、`/v1/status`、`/v1/scheduler/jobs` —— 唯讀介面。
- `POST /v1/scheduler/jobs/{id}/cancel` —— 唯一的異動（mutation）端點（FR-163，issue #10：取消一個已排程的 resume）。
- `GET /v1/events/stream` —— SSE 即時事件串流，每 15 秒發送一次 `: ping` 心跳。

安全態勢（`guard` middleware，ADD §23.2／§27.5）：每個端點都需要
`Authorization: Bearer <token>`，並以常數時間比較；`Host` 標頭必須是 loopback（用以防禦
DNS-rebinding）；請求主體大小上限為 1 MiB；CORS 因未實作而預設停用；錯誤一律以 ADD §23.5 的
信封格式輸出，附上型別化的 `domain.Error` 代碼。

依賴項刻意採用狹窄的介面：`JobLister` ／ `JobCanceller`（皆為
[`internal/scheduler/`](../scheduler/README.md) `Store` 的切面（slice）——canceller 特意分開，
讓唯讀組合永遠不會意外取得異動能力）、`EventSource`（[`internal/daemon/`](../daemon/README.md)
的 `Broker`），以及 `domain.Clock`。bearer token 是由 daemon 在每次重啟時產生；本套件從不讀取
權杖檔案本身。

刻意延後實作：ADD §23.4 中其餘的 pause 異動端點（`POST /v1/pauses`、`:cancel`、`:resume`）——
本機端已可透過 `auspex pause|resume` 進行手動異動，故暫不需要。
