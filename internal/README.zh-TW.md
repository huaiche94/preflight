# internal/ — Auspex 的私有 Go 套件：凍結契約、角色擁有的服務與轉接器

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`internal/` 底下的所有內容都是 `github.com/huaiche94/auspex` 模組的私有程式碼。跨角色共用且已凍結的介面位於 [`domain`](domain/)（實體）與 [`app`](app/)（port），詳見 [docs/implementation/vertical-slice/CONTRACT_FREEZE.md](../docs/implementation/vertical-slice/CONTRACT_FREEZE.md) 的 import-path 對照表；其餘套件則是在這些契約之下、由各角色擁有的實作。

## Layering rule

Auspex_ADD.md §7.6「Core dependency direction」（該 ADD 位於
[docs/design/Auspex_ADD.md](../docs/design/Auspex_ADD.md)）固定了依賴方向：

```text
cli/http/hooks
    ↓
application/orchestrator
    ↓
domain services + interfaces
    ↓
adapters: sqlite/git/providers/filesystem
```

domain 層不得引入 provider、SQLite、Cobra 或 VS Code 專屬的型別。

## Package index

若套件有 `doc.go`，該檔案就是該套件的契約 — 修改套件前請先閱讀它。

| Package | Role |
|---|---|
| [app](app/) | 凍結的跨元件 port（ADD §9.9、§9.10）：精簡的服務介面、request/response DTO，以及 `TxRunner` 交易邊界。`app/wiring` 會將實際實作組裝成可執行的 App。 |
| [artifacts](artifacts/) | Checkpoint Part A 的 artifact 驗證層：具體檢查項目（checksum、file-exists、heading、fence balance），在 `progress` 記錄之前，把宣稱的證據轉為經驗證的證據（「Completed means evidenced」，Constitution §6.2）。 |
| [buildinfo](buildinfo/) | `auspex version` 背後最精簡的 build/version 中繼資料；在真正的發行封裝出現之前，先以固定的開發用佔位值代替。詳見 [buildinfo/README.md](buildinfo/README.md)。 |
| [cli](cli/) | `auspex` CLI 的 Cobra command-tree 建構函式（`NewRootCmd` 及其相關函式）；採用建構函式而非套件層級單例，讓整個命令樹可在不呼叫 `os.Exit` 的情況下進行測試。詳見其 `doc.go`。 |
| [clock](clock/) | `domain.Clock` 的真實 wall-clock 實作；production 程式碼依賴介面本身，讓測試可以替換成 fake。詳見 [clock/README.md](clock/README.md)。 |
| [config](config/) | 依照 ADD §26.1 的優先序鏈載入分層 YAML 設定，包含 `schema_version` envelope，以及 unknown-field 的 warn/strict 驗證。詳見 [config/README.md](config/README.md)。 |
| [daemon](daemon/) | M6 daemon 生命週期：singleton lock、每次重啟產生的 bearer token、動態 loopback listener、runtime metadata 檔案、常駐 worker loop，以及記憶體內的 SSE event broker。 |
| [domain](domain/) | 凍結的 domain 實體：以 UUIDv7 為底的 opaque ID 型別、狀態 enum、凍結的 `domain.Error` 結構、forecast/usage/capability 型別，以及 `Clock`／`IDGenerator`／`ProcessRunner` port。 |
| [evaluation](evaluation/) | 實作 `app.EvaluationService`：針對單一 turn 執行 ADR-041 的 predictor pipeline，在同一筆交易中寫入 feature-vector／prediction／decision 資料列，並強制 exactly-once 授權。詳見其 `doc.go`。 |
| [features](features/) | 從 prompt、repository、session 與 Progress Tree（ADD §14.2）衍生出 prediction 的輸入訊號；這裡是隱私邊界，原始 prompt 文字只會進來、絕不會流出。詳見其 `doc.go`。 |
| [gitx](gitx/) | 用於 repository checkpoint 的 Git plumbing：repository／worktree 解析與 `git status --porcelain=v2 -z` 解析，一律透過 `domain.ProcessRunner` 以 argv 形式執行（絕不使用 shell 字串）。 |
| [hooks](hooks/) | Provider 生命週期 hook payload 解析。`hooks/claude` 負責解析 Claude Code hook 的 stdin payload（UserPromptSubmit、Stop、StopFailure），並編碼出與 provider 相容的 stdout 回應。 |
| [httpapi](httpapi/) | daemon 的已驗證 loopback HTTP/JSON + SSE 介面（ADD §23.2–23.5）：health／version／capabilities／status／jobs、即時 event stream，以及 scheduler-job 的 cancel endpoint。 |
| [idgen](idgen/) | `domain.IDGenerator` 的 UUIDv7 實作；每一個 Auspex 擁有的實體 ID 都是 UUIDv7，且絕不會被拿來解析其含義。詳見 [idgen/README.md](idgen/README.md)。 |
| [integrationtest](integrationtest/) | 由 qa 擁有的跨角色 integration 與 end-to-end 測試：高風險 fixture 流程、同一個 DB 上重新啟動、隱私／洩漏掃描、scheduler 雙 worker 競爭等等。 |
| [lock](lock/) | 單機、PID 檔案風格的 advisory file lock，用來確保對 runtime directory 的獨占所有權；刻意不做成分散式或網路鎖。詳見 [lock/README.md](lock/README.md)。 |
| [managed](managed/) | `auspex run`（ADD §8.1）背後的 managed one-shot runner：pre-prompt gate、provider subprocess 的生命週期（`claude -p … stream-json`），以及 terminal-outcome 歸因。詳見其 `doc.go`。 |
| [orchestrator](orchestrator/) | 跨凍結的 `app` port，排序 day-one 的 evaluate/decide pipeline（ADD §13）；本身不實作任何 prediction／policy／checkpoint 邏輯。詳見其 `doc.go`。 |
| [paths](paths/) | 以可注入的 environment，依作業系統解析全域的 config／data／cache／runtime 目錄。詳見 [paths/README.md](paths/README.md)。 |
| [pause](pause/) | 針對 Graceful Pause 狀態機（ADD §20）、涵蓋十二個凍結的 `domain.PauseStatus` 值的純狀態轉換驗證器；不涉及任何 I/O。詳見其 `doc.go`。 |
| [policy](policy/) | predictor pipeline 的終端階段：依照 ADD §17.3 固定的優先順序，把 risk score 與 runway forecast 合併成凍結的 `app.PolicyAction`；絕不會把未經校準的分數標示為機率。詳見其 `doc.go`。 |
| [predictor](predictor/) | 建立在 `features` 之上、具決定性且可解釋的 prediction 基礎元件（ADD §15–§16）：risk score 與 quantile 估計值，day one 階段絕不提供經校準的機率。詳見其 `doc.go`。 |
| [pricing](pricing/) | 本地、手動維護的 per-model 價格表（ADR-043），把 token forecast 換算成估計的美元成本範圍；絕不在執行期抓取，且一律標示為估計值。 |
| [progress](progress/) | Progress Tree domain service（Constitution §6、ADD §18）：node／edge／artifact／task 儲存區、node 狀態機，以及 stage/verify/commit 的 CompleteNode 協定。 |
| [providers](providers/) | Provider 原生 payload 解析。`providers/claude` 會把 Claude Code 的 status-line JSON 解析成中繼、可容忍 unknown-field 的 struct。 |
| [redact](redact/) | 以內容為基礎的機密偵測（ADD §19.5、§27.8）：檔名 pattern 加上 untracked-file archive 政策所使用的固定內容偵測器；文件中已註明並非窮舉。詳見其 `doc.go`。 |
| [repocheckpoint](repocheckpoint/) | Repository Checkpoint 的 create／verify／restore（ADD §19）：patch 加上 untracked archive、採用原子寫入；在擷取期間絕不會更動目前的 active branch 或 working tree。 |
| [retention](retention/) | `auspex gc` 背後的 ADR-046 分層 telemetry retention（hot window → rollup → gzip archive → verified delete）。詳見 [retention/README.md](retention/README.md)。 |
| [scheduler](scheduler/) | 建立在 `wake_jobs` 資料表之上、具持久性的 wake scheduler lease（ADD §12.4）：claim／renew／complete／fail／retry，在並行 worker 下仍能保持正確。詳見其 `doc.go`。 |
| [statecheckpoint](statecheckpoint/) | State Checkpoint manifest（ADD §18.8）：manifest 的 Go 結構、具決定性的 JSON 序列化，以及完整性 checksum。 |
| [storage](storage/) | 儲存轉接器。`storage/sqlite` 是 SQLite runtime（pragma、交易）以及 forward-only migration engine。詳見 [storage/README.md](storage/README.md)。 |
| [telemetry](telemetry/) | Provider telemetry 正規化，並持久化寫入凍結的 `pkg/protocol/v1.Event` envelope；`telemetry/claude` 是唯一處理 Claude payload 的路徑。詳見 [telemetry/README.md](telemetry/README.md)。 |
| [testutil](testutil/) | 測試輔助工具。`testutil/fakes` 為每個凍結的 `app` port 提供手寫的 fake，具備編譯期 interface assertion，且未設定的方法會 fail-loud。詳見 `testutil/fakes/doc.go`。 |
