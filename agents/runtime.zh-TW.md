# Runtime

> 🌐 [English](runtime.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

整併了原本兩個獨立的 agent packet——**Graceful Pause, Safe Points, and
Durable Scheduler** 與 **Application Orchestration, CLI, and Local API**
——成為單一 bounded context：涵蓋一切驅動系統即時運作的部分，
相對於它所儲存或預測的內容。Part A 與 Part B 應維持為各自獨立的內部
子元件；Part B 建立在 Part A 的 ports 之上，且預期會在 Part A 的
狀態機與 migration 都已存在後才開始，即便兩者同屬這一個角色。

## 模型

Part A 使用 Fable（pause/resume 是一個與正確性、狀態機正確運作
高度相關的關鍵邊界），Part B 中 authorization/pause orchestration 的
審查也使用 Fable；Part B 其餘部分使用較便宜的 coding model 即已足夠。

## ADD 負責章節

Part A：§20、§§15/17/28/29 中與 pause 相關的部分、附錄 C、ADR-031 至 ADR-040。
Part B：§13 pipeline orchestration、§§23–24、§28 中的維運（operational）子集、附錄 F。

## 專屬路徑

```text
# Part A — Graceful Pause, Safe Points, Durable Scheduler
internal/pause/**
internal/scheduler/**
schemas/pause.schema.json
testdata/pause-scenarios/**
internal/storage/sqlite/migrations/0050-0059_*.sql

# Part B — Application Orchestration, CLI, Local API
internal/orchestrator/**
internal/cli/**
internal/httpapi/**
internal/daemon/**
internal/app/wiring/**
internal/testutil/fakes/** (coordinate with the qa role)

docs/implementation/vertical-slice/runtime.md
```

不得編輯 `cmd/auspex/main.go`；root wiring 由 contract-integrator 與
foundation 角色負責整合。請在本角色擁有的路徑下新增指令 constructor。
除非 contract-integrator 明確指派範圍，否則 Part B 不得新增 schema。

---

## Part A — Graceful Pause、Safe Points 與 Durable Scheduler

### 任務目標

實作與 provider 無關（provider-neutral）的 pause/resume 狀態機，
以及持久化的 wake 排程。僅得依賴凍結後的 ports 來存取 predictor、
progress/state checkpoint、repository checkpoint、provider 中斷／恢復、
quota 讀取、clock，以及 lease。

### 必要狀態路徑

```text
observing
→ pause_requested
→ quiescing
→ safe_point_reached
→ persisting
→ interrupting
→ sleeping
→ wake_due
→ validating
→ resuming
→ resumed
```

需納入 ADD 中的終止／衝突／取消／失敗狀態。

### P0 交付項目

1. 狀態轉換驗證器。
2. 具備 debounce/hysteresis 狀態的 `Observe` 處理。
3. `RequestPause` 的冪等性。
4. 針對 turn/section 邊界觀測值的 safe-point coordinator 介面與實作。
5. Persist 階段的 orchestration：
   - Progress Tree snapshot；
   - State Checkpoint；
   - Repository Checkpoint；
   - Pause Record；
   - Wake Job。
6. 具備 claim/renew/complete/fail/retry 的持久化 scheduler lease。
7. 重新啟動時，復原逾期／已租用（leased）的工作。
8. Resume 驗證：
   - quota 安全；
   - repository 指紋（fingerprint）相容；
   - session/provider 能力有效；
   - authorization/consent 有效。
9. 重複 wake 的 exactly-once 行為。
10. Cancel 會阻止未來的 resume。
11. Provider interrupter/resumer 的 fake 契約測試。

### 首日現實考量

由於資料不足，經校準的自動 pause（calibrated auto-pause）可能無法使用。
需同時支援：

- 經校準的觸發條件：連續觀測值 `P_hit_10m >= threshold`；
- 明確標示為未經校準的緊急政策，並使用不同的原因代碼。

先實作持久化 wake 與 fake resumer。實際的 managed Claude resume
屬於延伸目標（stretch），且不得因此削弱狀態機測試的嚴謹度。

### 必要測試

- 兩次符合條件的觀測值會觸發請求；
- 單一次的尖峰（spike）不會觸發；
- safe point 會在中斷前先完成 checkpoint 的持久化；
- 每個階段之後發生 crash，都能正確 resume／一致化（reconcile）；
- 重新啟動可復原 wake job；
- 不安全的 quota 會重新排程；
- repo 有重疊時應阻擋；
- 不相關的 repo 變更則依設定的政策處理；
- 重複的 worker 只會產生一次 resume；
- 過期的 lease 會被回收；
- cancel 在與 wake 的競態中勝出；
- provider 中斷失敗時，狀態仍應保持可復原。

---

## Part B — Application Orchestration、CLI 與 Local API

### 任務目標

將凍結後的 ports 接線（wire）進一個以 in-process 為優先的應用程式，
並透過穩定的 CLI/JSON 契約，對外提供 day-one 流程。HTTP daemon 的
優先順序低於一個可運作的 CLI。

### P0 指令

```text
auspex version
auspex init
auspex hook claude statusline
auspex hook claude user-prompt-submit
auspex hook claude stop
auspex hook claude stop-failure
auspex evaluate
auspex decision allow
auspex decision deny
auspex checkpoint create
auspex progress show
auspex state show
auspex pause request
auspex pause cancel
auspex resume
auspex scheduler run-once
auspex status
auspex doctor
```

### 管線行為

1. 接收 provider 正規化後的輸入，或 CLI 輸入。
2. 解析 repository/worktree/session。
3. 載入目前的 Progress Tree 與用量觀測值。
4. 對輕量的 Git 狀態做 snapshot。
5. 透過 predictor 角色進行評估。
6. 套用政策。
7. 若為 allow：產生與 provider 相容的回應。
8. 若為 block/checkpoint：持久化 evaluation，並回傳穩定的 decision ID／指示。
9. `checkpoint create` 會依照凍結後的交易／orchestration 契約，
   先呼叫 checkpoint 角色的 Part A（state），再呼叫其 Part B（repository）。
10. `decision allow` 會發放一次性 authorization。
11. 重新送出的 prompt，在被允許前必須先恰好消耗一次 authorization。
12. Stop/StopFailure 會完成結果標記（outcome labeling）。

### JSON 與錯誤

- 具穩定 schema 版本的輸出；
- 具型別的錯誤代碼、訊息、是否可重試（retryable）、詳細資訊；
- log/錯誤中不得出現原始 prompt；
- machine mode 絕不對 stdout 輸出裝飾性文字；
- 當 Auspex 失敗時，hook fallback 仍需維持語法上合法。

### HTTP 延伸目標

只有在 CLI 的 E2E 測試通過後，才實作具驗證機制的 loopback endpoint。
在核心迴圈穩定之前，不實作 SSE。

### 測試

- CLI 的 golden test；
- 無 TTY 時的行為；
- 格式錯誤的 stdin；
- 高風險的 block 與 allow-once 流程；
- 第二次 authorization 的 replay 應被拒絕；
- checkpoint 失敗時不得發放 authorization；
- provider hook 永遠會收到合法的回應；
- 行程（process）結束代碼（exit code）；
- 使用同一個 SQLite 檔案的 in-process 重新啟動。
