# Auspex 垂直切片並行實作計畫

> 🌐 [English](Auspex_Parallel_Execution_Plan.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

## 1. 決策

**不要**逐章為 ADD 指派一個 agent。第 1–35 章混合了產品限制、領域合約（domain contract）、provider adapter、執行期（runtime）行為、測試與交付面向的考量。若採取「一章一個 agent」的切分方式，將產生重複的型別、不相容的 schema，以及嚴重的合併衝突。

改採一個 contract/integration 角色，加上六個 bounded-context 角色。每個角色都在專屬的 Git worktree 中工作，並各自擁有互不重疊的路徑集合。角色定義存放於 `agents/` 之下，每個角色各對應一個語意化命名的檔案（慣例說明見 `agents/README.md`）；七個角色中有兩個（`checkpoint`、`runtime`）各自涵蓋了原本兩個較窄的 bounded context，之所以合併，是因為這兩半在實務上總是被一起使用——內部仍保留的 Part A／Part B 切分方式，請參見 `agents/checkpoint.md` 與 `agents/runtime.md`。

第一天（day-one）的目標，是為 Claude Code 打造一個真正的垂直切片：

```text
Claude status-line / hook event
        ↓
normalized telemetry persisted in SQLite
        ↓
UserPromptSubmit auspex evaluation
        ↓
explainable risk decision
        ↓
ALLOW or BLOCK
        ↓
checkpoint + one-time authorization
        ↓
turn completion/failure outcome persisted
        ↓
pause state and durable wake job can be created/recovered
```

第一天的成功標準**並非**「所有 ADD 章節都已實作完成」，而是在保留長期架構的前提下，證明核心迴圈（core loop）可行。

## 2. 第一天的範圍邊界

### 必須合併

1. 儲存庫（repository）能在目標作業系統上初始化並建置成功。
2. 共用的 domain／event／store 合約一次凍結完成。
3. Claude status-line 與 hook 的 payload 可從 fixture 解析。
4. Quota／context／turn 遙測資料以冪等（idempotent）方式儲存。
5. 確定性（deterministic）的 scope／risk predictor 在冷啟動（cold start）期間回傳 P50/P90 估計值、原因代碼（reason code）、信心值（confidence），並標示 `calibrated=false`。
6. `UserPromptSubmit` 能在主要 turn 開始前允許（allow）或封鎖（block）。
7. 高風險流程能建立 State Checkpoint 與 Repository Checkpoint 的證據。
8. 若無持久化的（durable）artifact 證據，Progress node 的完成狀態會被拒絕。
9. 一次性授權（one-time authorization）可防止 prompt 重放（replay）。
10. Pause 狀態機與 durable wake job 在測試中能撐過行程（process）重啟。
11. CLI 對外暴露此垂直切片。
12. 一個端對端（end-to-end）fixture 測試涵蓋整個流程。

### 明確延後事項

- 第 21 章的完整 Codex 整合。
- 第 25 章的 VS Code 擴充套件。
- ML 訓練、個人化（personalization）、ONNX，以及 Python 研究用執行環境（research runtime）。
- 完全校準（calibrated）的機率宣稱。
- 超出基本 CI 骨架之外的公開套件管理器（package-manager）發佈自動化。
- 作業系統從關機／休眠狀態喚醒。
- 多 provider 路由（routing）。
- 外部 provider 外掛協定（plugin protocol）。
- 若 provider 生命週期支援未經 fixture 測試證實，則不做生產等級（production-grade）、自動喚起休眠中 Claude 行程的功能。

## 3. 模型配置

在正確性與狀態機推理占主導地位之處使用 Fable：

- `contract-integrator`：合約凍結與整合。
- `predictor`：predictor／policy。
- `runtime` Part A：pause／scheduler。
- 最終架構與競態條件（race-condition）審查。

在可行的情況下，對確定性的實作工作使用較便宜的程式碼模型：

- foundation／config／SQLite；
- Claude JSON 解析與 hook fixture；
- Git snapshot／checkpoint 的底層串接（plumbing）；
- CLI／API 串接（`runtime` Part B）；
- CI 與測試框架（test harness）。

若每個 worker 都必須使用 Fable，不要把完整的 161 KB ADD 給每個 worker。只給每個 worker：

1. 這份共通計畫；
2. `contract-integrator` 產出之後的 `CONTRACT_FREEZE.md`；
3. 它被指派的 ADD 章節；
4. 它在 `agents/` 下的角色檔案。

## 4. 並行拓撲

```text
                         ┌─────────────────────────────┐
                         │ contract-integrator          │
                         │ freeze types / ports / IDs   │
                         └──────────────┬──────────────┘
                                        │
                                        ▼
                              foundation
                              Go module / config / SQLite
                                        │
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
             claude-provider        checkpoint           predictor
             Telemetry/Hooks    State CP + Repo CP     Risk/Policy/Auth
                    │                   │                   │
                    └───────────────────┴───────────────────┘
                                        │
                                        ▼
                                    runtime
                        Pause+Scheduler, then CLI/API/Orchestration
                                        │
                                        ▼
                                       qa
                              Security/CI/E2E
                                        │
                                        ▼
                              contract-integrator
                              (final integration)
```

`runtime` 可以立即針對已凍結的介面與兩個部分各自的 fake 實作開始寫程式碼（它不需要等待 `checkpoint` 或 `predictor` 的具體實作完成）——但它的 Part B（CLI／API／orchestration）確實依賴於自身 Part A（pause／scheduler）進展到足以串接的程度，因為這兩部分現在同屬一個角色的序列。

## 5. 第 1–35 章的擁有權對照表

| ADD 章節 | 主要擁有者 | 第一天處理方式 |
|---|---|---|
| 1–8 | `contract-integrator` | 視為限制條件閱讀；產出合約凍結與範圍護欄（scope guardrail）。 |
| 9 | `contract-integrator` | 凍結 domain 值、ID、狀態（status）、service port、provider 能力型別。 |
| 10 | `foundation`（由 `contract-integrator` 治理） | 初始化精確的套件（package）配置；不建立投機性（speculative）套件。 |
| 11 | `contract-integrator` + `claude-provider` | `contract-integrator` 凍結 envelope；`claude-provider` 實作 Claude 正規化（normalization）／匯入（ingestion）。 |
| 12 | `foundation` + 功能擁有者 | `foundation` 擁有 DB 引擎／核心 migration；各角色擁有其被分配到的 migration 範圍。 |
| 13 | `runtime`（Part B） + `predictor` | `predictor` 負責評估；`runtime` 協調（orchestrate）授權與結果（outcome）生命週期。 |
| 14–17 | `predictor` | Scope、token／quota／runway 基準線、risk、policy。 |
| 18 | `checkpoint`（Part A） | Progress Tree 與 State Checkpointing。 |
| 19 | `checkpoint`（Part B） | Repository Checkpoint 與復原（recovery）。 |
| 20 | `runtime`（Part A） | Graceful Pause 與 durable wake job。 |
| 21 | 延後 | 短期內不做 Codex 生產環境 adapter，如有需要僅提供 fixture／介面。 |
| 22 | `claude-provider` | Claude plugin／hook／status-line／原生事件（native event）正規化。 |
| 23–24 | `runtime`（Part B） | 先做行程內（in-process）版本；提供輕量的 daemon／API／CLI 介面（surface）。 |
| 25 | 延後 | 短期內不實作 VS Code。 |
| 26 | `foundation` | 僅提供垂直切片所需的 config 模型與預設值。 |
| 27 | `contract-integrator` + `qa` | 隱私／安全限制與相關測試。 |
| 28 | `runtime`（Part B） + `qa` | 型別化錯誤（typed error）、日誌（logging）、recovery、doctor 基準線。 |
| 29 | 每個角色 + `qa` | 單元測試由各角色自行擁有；`qa` 擁有跨套件（cross-package）／E2E 測試。 |
| 30 | `foundation` + `qa` | 基本 OSS 檔案與 CI；完整的發佈矩陣（release matrix）延後。 |
| 31–32 | `contract-integrator` | 里程碑驗收與最終 DoD 關卡（gate）。 |
| 33 | `contract-integrator` | ADR 合規性；只有 `contract-integrator` 可以編輯已核准（accepted）的 ADR。 |
| 34 | 全員 | 執行與 durable-progress 合約。 |
| 35 | 依附錄拆分 | `claude-provider` 擁有 Claude 範本（template）；`checkpoint` 擁有 A/B/D；`runtime` 擁有 C/F；`qa` 負責測試／參考驗證。 |

## 6. 合約凍結關卡（Contract-freeze gate）

任何功能角色都不應自創競爭性的 domain 型別。`contract-integrator` 必須先提交 `docs/implementation/vertical-slice/CONTRACT_FREEZE.md`，以及以下項目的可編譯骨架（skeleton）：

- UUIDv7 風格的 ID 別名（alias）或包裝型別（wrapper）；
- `Session`、`Turn`、`Task`、`ProgressNode`、`ArtifactReference`；
- `StateCheckpoint`、`RepositoryCheckpoint`、`PauseRecord`、`WakeJob`；
- `UsageObservation`、`QuotaObservation`、`ContextObservation`；
- `Evaluation`、`PredictionResult`、`PolicyDecision`、`Authorization`；
- 事件 envelope 與事件型別常數（event-type constant）；
- 失敗分類（failure class）與型別化錯誤代碼；
- ADD §9.9 中的 service 介面；
- ADD §9.10 中的 provider 能力與 hook 正規化型別；
- `Clock` 與 `IDGenerator` 介面；
- SQLite 交易邊界（transaction boundary）慣例；
- JSON／YAML 欄位名稱與 schema 版本字串。

### 第一天不可變更的規則

1. 預設不持久化（persist）原始 prompt。
2. 未校準（uncalibrated）的分數絕不能標示為機率。
3. State Checkpoint 與 Repository Checkpoint 是不同的實體（entity）。
4. Progress Tree 是任務狀態的權威來源（canonical state）。
5. Node 完成需要 durable 的 artifact 證據。
6. Pause 的完整保證只適用於受管理（managed）的執行；原生 hook（native-hook）的行為是降級（degraded）且明確標示的。
7. 所有持久化寫入皆透過穩定的 event／operation ID 達成冪等（idempotent）。
8. Clock、行程執行（process execution）與 ID 產生皆可注入（injectable）。
9. Provider 的線路層 payload（wire payload）絕不能外洩至 domain／儲存層資料列（row）。
10. 運作性觀測（operational observation）失敗可以 fail open；狀態完整性（state-integrity）失敗必須 fail closed。

## 7. 共用檔案與相依性政策

### 僅由 `contract-integrator` 擁有的檔案

```text
Auspex_ADD.md
AGENTS.md
internal/domain/**
internal/app/ports.go
pkg/protocol/v1/**
docs/adr/**
docs/implementation/vertical-slice/CONTRACT_FREEZE.md
```

### 僅由 `foundation` 擁有的檔案

```text
go.mod
go.sum
cmd/auspex/main.go
internal/config/**
internal/paths/**
internal/buildinfo/**
internal/storage/sqlite/db.go
internal/storage/sqlite/migrate.go
internal/storage/sqlite/migrations/0000-0009_*.sql
Makefile
Taskfile.yml
.golangci.yml
```

沒有其他角色可編輯 `go.mod` 或 `go.sum`。相依性需求應寫入該角色自己的進度檔案，交由 `foundation`／`contract-integrator` 套用。

### Migration 分配

```text
0000–0009  foundation           core/session/config
0010–0019  claude-provider      telemetry/provider events
0020–0029  checkpoint (Part A)  progress/state checkpoints
0030–0039  checkpoint (Part B)  repository checkpoints
0040–0049  predictor            evaluations/predictions/authorizations
0050–0059  runtime (Part A)     pause/wake jobs
```

除非 `contract-integrator` 明確指派範圍，否則 `runtime` Part B 不新增 schema。

### 單元測試擁有權

每個功能角色在自己的套件（package）下撰寫單元測試。`qa` 只擁有：

- `internal/integrationtest/**`；
- `testdata/e2e/**`；
- 跨元件（cross-component）的 race／restart／security 測試；
- CI workflow。

## 8. Git worktree 設定

建議的分支（branch）：

```bash
git worktree add ../auspex-contract-integrator -b vertical-slice/contract-integrator
git worktree add ../auspex-foundation           -b vertical-slice/foundation
git worktree add ../auspex-claude-provider      -b vertical-slice/claude-provider
git worktree add ../auspex-checkpoint           -b vertical-slice/checkpoint
git worktree add ../auspex-predictor            -b vertical-slice/predictor
git worktree add ../auspex-runtime              -b vertical-slice/runtime
git worktree add ../auspex-qa                   -b vertical-slice/qa
```

`contract-integrator` 必須先落地（land）合約 commit。所有其他分支都必須先 rebase 到那個確切的 commit 上，才能開始撰寫正式（production）程式碼。

## 9. Durable 協調產物（Coordination Artifacts）

每個角色都恰好擁有一份進度產物（progress artifact）：

```text
docs/implementation/vertical-slice/contract-integrator.md
docs/implementation/vertical-slice/foundation.md
docs/implementation/vertical-slice/claude-provider.md
docs/implementation/vertical-slice/checkpoint.md
docs/implementation/vertical-slice/predictor.md
docs/implementation/vertical-slice/runtime.md
docs/implementation/vertical-slice/qa.md
```

每完成一個邏輯節點（logical node）之後，該角色都必須寫入：

```yaml
node: predictor-03
status: completed
artifacts:
  - internal/predictor/heuristic/predictor.go
  - internal/predictor/heuristic/predictor_test.go
validation:
  - go test ./internal/predictor/...
commit: <sha>
next_action: predictor-04 implement policy reason codes
assumptions: []
blockers: []
```

僅存在於對話中的進度不算數。

## 10. 合併順序

即使實作是並行進行，也要依照此順序合併：

1. `contract-integrator` 的合約凍結。
2. `foundation` 的核心 SQLite。
3. `claude-provider`、`checkpoint`、`predictor`，測試通過後可以任意順序合併。
4. `runtime`（先 Part A 的 pause／scheduler，再 Part B 的 CLI／API／orchestration）。
5. `qa` 的 CI／E2E／security 測試。
6. `contract-integrator` 的最終整合（reconciliation）與架構審查。

`contract-integrator` 應該 cherry-pick 或合併**整個已審查過的 commit**，而不是手動複製產生的程式碼。

## 11. 第一天最終展示（Demo）

最終分支應展示：

```bash
auspex version
auspex init

# Feed a Claude status-line fixture.
auspex hook claude statusline < testdata/provider-events/claude/statusline-high-usage.json

# Evaluate a prompt without persisting raw prompt text.
auspex evaluate \
  --provider claude \
  --prompt-file testdata/e2e/high-risk-prompt.txt \
  --json

# Simulate UserPromptSubmit; high risk returns provider-compatible block output.
auspex hook claude user-prompt-submit \
  < testdata/provider-events/claude/user-prompt-submit-high-risk.json

# Create both state and repository evidence.
auspex checkpoint create --evaluation <id> --json

# Issue a one-time allow decision and consume it once.
auspex decision allow --evaluation <id> --json

# Persist a normal turn completion or rate-limit failure fixture.
auspex hook claude stop < testdata/provider-events/claude/stop.json
auspex hook claude stop-failure < testdata/provider-events/claude/stop-failure-rate-limit.json

# Exercise pause/wake durability without depending on wall-clock sleep.
auspex pause request --session <id> --reason runway --json
auspex scheduler run-once --at <timestamp> --json
auspex status --json
```
</content>

## 12. 整合進度落後時的刪減順序

依此順序刪減功能：

1. HTTP daemon 傳輸層；保留行程內（in-process）CLI。
2. SSE／即時儀表板（live dashboard）。
3. 真正的 provider 行程自動恢復（auto-resume）；保留 durable wake 狀態與 fake resumer 的合約測試。
4. Repository 還原（restore）；保留建立／驗證（create／verify）功能。
5. 與既有使用者指令（user command）整合的原生 status-line 組合；改為記載手動設定方式。
6. 完整的 OSS 發佈矩陣。

絕不可刪減：

- prompt 隱私預設值；
- 授權重放保護（authorization replay protection）；
- artifact 證據要求；
- checkpoint 的原子性（atomicity）；
- calibrated／uncalibrated 的區分；
- pause 狀態的持久性（durability）；
- provider payload fixture 與合約測試。

## 13. 最終 Fable 審查提示詞（Review Prompt）

```text
Review the merged Auspex day-one vertical slice against Auspex_ADD.md.

Focus only on:
1. domain/schema contradictions;
2. raw-prompt or secret persistence;
3. authorization replay/staleness;
4. State Checkpoint atomicity and artifact evidence;
5. Repository Checkpoint race/path traversal handling;
6. quota/risk values incorrectly presented as calibrated probabilities;
7. pause/resume state-machine races, duplicate wake, stale lease, and repository conflict;
8. Claude hook output compatibility and unknown-field tolerance;
9. missing restart and idempotency tests.

Do not redesign the project or add future abstractions. Produce a severity-ranked
file-level report, then fix only P0/P1 findings and run the complete test suite.
```

---

# Agent 角色

每個角色定義的完整內容，都以獨立檔案的形式存放在
`agents/` 之下，因此單一檔案即可交給獨立的 agent／worktree，
而不需要這份計畫的其餘部分（參見上方 §3 與 `agents/README.md`）。

**`agents/*.md` 是角色內容的單一事實來源（single source of truth）。**
本節僅作為索引——角色細節請在該處編輯，而非在此處，以避免兩邊產生落差。

| 角色 | 檔案 | 模型 | 任務 |
|---|---|---|---|
| contract-integrator | [`agents/contract-integrator.md`](../../agents/contract-integrator.md) | Fable | 凍結編譯期／持久化合約；最終整合已審查過的分支。 |
| foundation | [`agents/foundation.md`](../../agents/foundation.md) | 較便宜的模型；migration／recovery 審查時用 Fable | 可建置的 Go 應用程式基礎架構，以及其他所有套件都依賴的 SQLite 執行環境（runtime）。 |
| claude-provider | [`agents/claude-provider.md`](../../agents/claude-provider.md) | hook 語意（semantics）用 Fable；parser／fixture 用較便宜的模型 | 以 fixture 為基礎，將 Claude Code 的 hook／status-line 正規化為已凍結的 Auspex 事件。 |
| checkpoint | [`agents/checkpoint.md`](../../agents/checkpoint.md) | Fable | Progress Tree + State Checkpointing（Part A）**以及** Repository Checkpoint（Part B）；沒有經過驗證的 artifact 證據就不算完成，且 checkpoint 不會變更目前使用中的分支。 |
| predictor | [`agents/predictor.md`](../../agents/predictor.md) | Fable | 確定性、可解釋、對冷啟動安全（cold-start-safe）的 predictor／policy／authorization 迴圈。 |
| runtime | [`agents/runtime.md`](../../agents/runtime.md) | Part A（pause／scheduler）用 Fable；Part B（CLI／API）大部分用較便宜的模型 | Graceful Pause + durable wake 排程（Part A）**以及**串接垂直切片的 CLI／API／orchestration（Part B）。 |
| qa | [`agents/qa.md`](../../agents/qa.md) | fixture／CI 用較便宜的模型；最終對抗式（adversarial）審查用 Fable | 提供客觀證據，證明此垂直切片安全、可重啟、冪等，且與 provider 相容。 |

`contract-integrator` 在產出 `docs/implementation/vertical-slice/CONTRACT_FREEZE.md` 時所填寫的合約凍結範本，存放於
[`agents/CONTRACT_FREEZE_TEMPLATE.md`](../../agents/CONTRACT_FREEZE_TEMPLATE.md)。

先前編號的九角色結構（`A00`–`A08`）已封存於
`docs/archive/agent-packets-v1/`，僅供歷史參考。
