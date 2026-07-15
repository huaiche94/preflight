# A07 — 應用程式編排、CLI、本機 API 與垂直切片接線

> 🌐 [English](07-runtime-cli-api.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。本文件屬歷史封存，非現行實作指引。

## 模型

使用較便宜的程式碼模型即可；授權／暫停編排審查請使用 Fable。

## ADD 負責範圍

§13 管線編排、§§23–24、§28 的維運子集、Appendix F。

## 專屬路徑

```text
internal/orchestrator/**
internal/cli/**
internal/httpapi/**
internal/daemon/**
internal/app/wiring/**
internal/testutil/fakes/** (coordinate with A08)
docs/implementation/day1/A07.md
```

不要編輯 `cmd/preflight/main.go`；根層級的接線（wiring）由 A00／A01 整合。請在自己擁有的路徑下新增指令建構子。

## 任務

將凍結的 ports 接線到以行程內（in-process）優先的應用程式中，並透過穩定的 CLI／JSON 合約公開 day-one 流程。HTTP daemon 的優先順序次於能運作的 CLI。

## P0 指令

```text
preflight version
preflight init
preflight hook claude statusline
preflight hook claude user-prompt-submit
preflight hook claude stop
preflight hook claude stop-failure
preflight evaluate
preflight decision allow
preflight decision deny
preflight checkpoint create
preflight progress show
preflight state show
preflight pause request
preflight pause cancel
preflight resume
preflight scheduler run-once
preflight status
preflight doctor
```

## 管線行為

1. 接收供應商正規化或 CLI 輸入。
2. 解析儲存庫／worktree／session。
3. 載入目前的 Progress Tree 與用量觀測值。
4. 為輕量的 Git 狀態建立快照。
5. 透過 A05 進行評估。
6. 套用政策。
7. 若為 allow：產生與供應商相容的回應。
8. 若為 block／checkpoint：持久化評估結果，並回傳穩定的決策 ID／指示。
9. `checkpoint create` 依照凍結的交易／編排合約，依序呼叫 A03 再呼叫 A04。
10. `decision allow` 會發放一次性授權。
11. 重新送出的提示詞在被允許之前，會恰好消費一次授權。
12. Stop／StopFailure 會完成結果標記（outcome labeling）。

## JSON 與錯誤

- 穩定且具 schema 版本的輸出；
- 型別化的錯誤代碼、訊息、是否可重試、詳細資訊；
- 記錄／錯誤中不含原始提示詞；
- 機器模式絕不對 stdout 輸出裝飾性文字；
- 當 Preflight 失敗時，hook 的後備方案（fallback）仍維持語法正確。

## HTTP 加分項目

只有在 CLI 的 E2E 測試通過後，才實作具身分驗證的 loopback 端點。核心迴圈穩定之前不導入 SSE。

## 測試

- CLI golden 測試；
- 無 TTY 行為；
- 格式錯誤的 stdin；
- 高風險阻擋與一次性允許流程；
- 第二次授權重播（replay）遭拒；
- 檢查點失敗時不發放授權；
- 供應商 hook 永遠收到有效回應；
- 處理程序結束代碼（exit codes）；
- 使用同一個 SQLite 檔案的行程內重啟。
