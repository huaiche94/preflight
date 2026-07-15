# internal/daemon/ — M6 背景常駐程式（daemon）：生命週期、認證權杖、事件 broker、worker 迴圈

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`auspex daemon run` 背後的常駐行程（issue #7；ADD §23——
[`docs/design/Auspex_ADD.md`](../../docs/design/Auspex_ADD.md)）。其型態與權杖（token）儲存
方式為 [`docs/DECISION_LOG.md`](../../docs/DECISION_LOG.md) 中的決策 D-16：核心是一個與作業
系統無關的前景行程（foreground process），外圍可選擇性地包上一層 macOS LaunchAgent。

關鍵檔案：

- [`daemon.go`](daemon.go) —— `Daemon.Run` 依序組合：單例鎖（singleton lock，`daemon.lock`，透過 `internal/lock`）→ bearer token → 動態的 loopback 監聽器 → 執行期中繼資料（runtime metadata）→ serve + work；關機時則以相反順序執行。訊號（signal）處理留給 CLI 呼叫端負責；`Run` 只認得 context。
- [`token.go`](token.go) —— `GenerateToken` 產生一個 256 位元的十六進位權杖，寫入 `<data>/daemon.token`，權限為 `0600`，每次啟動都以 `O_TRUNC` 覆寫（即輪替），因此每次重啟都會讓先前發出的所有權杖失效（ADD §23.2、§27.5；D-16）。`VerifyToken` 以常數時間比較；預期權杖為空字串時一律不會比對成功。
- [`metadata.go`](metadata.go) —— 位於 `<runtime>/daemon.json` 的 `auspex.daemon.v1` 探索文件（discovery document）（PID、位址、權杖檔案路徑），以 `0600` 權限寫入，關機時最先被移除。
- [`broker.go`](broker.go) —— 純記憶體內（in-memory）的 pub/sub 廣播（fan-out），供應 SSE 串流。刻意只存在於記憶體中：broker 不保留任何歷史紀錄，因此沒有事件重播（event replay）——較晚訂閱的用戶端改為從 status／jobs 端點讀取目前狀態；反應較慢的訂閱者會被丟棄事件，而不會阻塞 worker。
- [`worker.go`](worker.go) —— 常駐的排程迴圈（ADD §23.6）：協調（reconcile）已過期的租約（lease）、認領（claim）到期的 wake job（[`../scheduler/`](../scheduler/README.md)）、執行 `pause.Wake` → `GracefulPauseService.Resume`（其中真正執行的是 `ValidateResume`，絕不繞過）、續約（renew）、完成／失敗處理，並透過 broker 發布 `pkg/protocol/v1` 事件。預設輪詢間隔為 5 秒。

HTTP 處理器本身位於 [`internal/httpapi/`](../httpapi/README.md)——本套件透過
`Config.NewHandler` 交給它一個剛產生好的權杖。launchd 安裝（`auspex daemon install`，附
KeepAlive 的 LaunchAgent plist）則實作在 [`internal/orchestrator/daemon.go`](../orchestrator/daemon.go)
與 [`internal/cli/daemon.go`](../cli/daemon.go)，而不是在本套件。本套件沒有 `doc.go`；每個
檔案各自帶有自己的合約註解。
