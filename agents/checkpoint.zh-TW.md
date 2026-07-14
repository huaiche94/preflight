# Checkpoint

> 🌐 [English](checkpoint.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

整併了原本兩個獨立的 agent packet——**Progress
Tree / State Checkpointing** 與 **Repository Checkpoint**——成為單一
bounded context。兩個半邊由同一個角色擁有，因為它們在實務上總是一起
被使用（runtime 角色在 pause 的 persist 階段，會在同一個邏輯步驟中
同時寫入一個 State Checkpoint 與一個 Repository Checkpoint），但兩者
在內部仍是各自獨立的子元件：在本角色的路徑範圍內，Part A 與 Part B
的實作、migration、測試都應保持分離。

## 模型

兩個部分都使用 Fable——本角色同時擁有產品核心完整性的邊界（Part A）
與 Git 安全性的邊界（Part B）。

## ADD 負責章節

Part A：§18、附錄 A/B、§29.5 的 State Checkpoint 情境、ADR-027 至 ADR-030 及 ADR-039（作為限制條件）。
Part B：§19、§§27–29 中與 Git／安全相關的部分、附錄 D、day-one 流程所需的 M2 子集。

## 專屬路徑

```text
# Part A — Progress Tree / State Checkpointing
internal/progress/**
internal/statecheckpoint/**
internal/artifacts/**
schemas/progress-tree.schema.json
schemas/state-checkpoint.schema.json
testdata/progress-trees/**
testdata/checkpoints/state/**
internal/storage/sqlite/migrations/0020-0029_*.sql

# Part B — Repository Checkpoint
internal/gitx/**
internal/repocheckpoint/**
internal/redact/**
schemas/repository-checkpoint.schema.json
testdata/repositories/**
testdata/checkpoints/repository/**
internal/storage/sqlite/migrations/0030-0039_*.sql

docs/implementation/vertical-slice/checkpoint.md
```

---

## Part A — Progress Tree 與 State Checkpointing

### 任務目標

讓 Progress Tree 成為權威的（canonical）持久任務狀態，並強制規定：
沒有經過驗證的 artifact 證據，節點（node）就不能被標記為完成。

### 交付項目

1. Task/node/edge/artifact 儲存層（store）。
2. 具備明確合法轉換（valid transitions）的節點狀態機。
3. Artifact 驗證器（validators）：
   - 檔案存在；
   - checksum 相符；
   - Markdown 標題（heading）存在；
   - Markdown code fence 成對平衡；
   - 可選的自訂驗證器介面。
4. `CompleteNode` 原子協定（atomic protocol）：
   - 暫存／驗證 artifact 證據；
   - 更新節點；
   - 建立 State Checkpoint；
   - 適用情況下在同一個資料庫交易（transaction）內提交；
   - 提交後發布正規化事件（normalized events）。
5. State Checkpoint manifest 的序列化與 checksum。
6. 針對「已暫存 artifact 但資料庫尚未提交」這類 crash 時間窗，實作啟動時的一致化（reconciliation）程序。
7. 完成動作的 idempotency key，以及重複 provider 事件的處理。
8. Snapshot／load-latest／verify 等 API。

### 必須拒絕的情況

- 「agent 宣稱已完成」但沒有 artifact；
- artifact 缺失或內容已變更；
- 已完成的子節點違反相依性政策（dependency policy）；
- 重複完成、且證據互相衝突；
- 不合法的狀態轉換；
- checkpoint manifest 引用了尚未提交（uncommitted）的資料列。

### 必要測試

- 合法的 Markdown 段落可以完成並產生 checkpoint；
- 缺少標題或 code fence 不平衡時應被拒絕；
- 在完成流程的每個階段注入 crash，並驗證一致化程序；
- 連續 100 個節點會產生 100 個可驗證的 checkpoint；
- 相同的 idempotency key 回傳相同結果；
- 衝突的 idempotency payload 應被拒絕；
- 併發完成的競態（race）情境。

---

## Part B — Repository Checkpoint

### 任務目標

在不變更目前作用中分支（active branch）的前提下，擷取並驗證 repository 證據。
為 runtime 角色與 contract-integrator 的最終審查，提供一個安全的 checkpoint 基本元件（primitive）。

### P0 交付項目

1. Repository/worktree 解析器（resolver）。
2. `git status --porcelain=v2 -z` 解析器。
3. Snapshot 指紋（fingerprint）：
   - repository 身分；
   - worktree 路徑；
   - branch/HEAD；
   - index/worktree 狀態；
   - 變更路徑與 numstat；
   - untracked 政策中繼資料。
4. Repository Checkpoint 的建立與驗證。
5. 依 ADD 規範產生二進位安全（binary-safe）的 patch，或提供 manifest 參照。
6. 安全的 untracked 封存政策，具備大小／路徑／機密內容過濾器。
7. Artifact 的原子式暫存檔轉正式檔寫入與清理。
8. 擷取（capture）過程中若 Git 狀態發生變化，需能偵測競態。
9. Restore **dry-run**；實際 restore 為延伸目標（stretch）。

### 安全需求

- 拒絕 path traversal 與 symlink escape；
- 絕不包含 `.git` 內部檔案或設定中排除的路徑；
- 預設遮蔽／省略疑似機密內容；
- 絕不執行 shell 字串；一律使用 argv 形式呼叫程序；
- 限制 artifact 大小與檔案數量上限；
- 在規劃 restore 前先驗證 checksum。

### 必要測試

已追蹤／已 staged／未 staged／未追蹤（untracked）、重新命名／刪除、二進位檔案、
平台允許時路徑中含空格／換行、巢狀 worktree、併發變更、暫存檔清理、
path traversal、超出大小上限，以及排除疑似機密內容的檔案。

---

## 跨部分邊界

Part A 透過凍結後的 ports 儲存對 Part B checkpoint 的參照；它不會直接
深入 Part B 的 Git plumbing，而 Part B 也不會直接更新 Progress Tree。
即使兩邊由同一個角色擁有，這條內部邊界仍應維持為真正的介面分界，
讓兩個半邊保持可獨立測試，也讓未來若真的需要拆回兩個角色時，
成本能維持低廉。
