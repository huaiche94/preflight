# A04 — Git 觀察與儲存庫檢查點

> 🌐 [English](04-repository-checkpoint.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。本文件屬歷史封存，非現行實作指引。

## 模型

較便宜的程式碼模型即可實作其中大部分內容；路徑／race／安全性審查請使用 Fable。

## ADD 負責範圍

§19、§§27–29 中 Git／安全性相關的面向、Appendix D，以及 day-one 流程所需的 M2 子集。

## 專屬路徑

```text
internal/gitx/**
internal/repocheckpoint/**
internal/redact/**
schemas/repository-checkpoint.schema.json
testdata/repositories/**
testdata/checkpoints/repository/**
internal/storage/sqlite/migrations/0030-0039_*.sql
docs/implementation/day1/A04.md
```

## 任務

在不變更目前作用中分支的情況下，擷取並驗證儲存庫證據。提供安全的檢查點基本操作（primitive）給 A03／A06／A07 使用。

## P0 產出物

1. 儲存庫／worktree 解析器。
2. `git status --porcelain=v2 -z` 解析器。
3. 快照指紋（fingerprint）：
   - 儲存庫身分；
   - worktree 路徑；
   - 分支／HEAD；
   - index／worktree 狀態；
   - 變更的路徑與 numstat；
   - 未追蹤（untracked）政策中繼資料。
4. Repository Checkpoint 的建立與驗證。
5. 依 ADD 規定，產生二進位安全（binary-safe）的 patch 或清單（manifest）參照。
6. 具備大小／路徑／機密資料過濾條件的安全未追蹤檔案封存政策。
7. 從暫存到最終產出物的原子性寫入與清理。
8. 若擷取過程中 Git 狀態發生變化，偵測競爭（race）情形。
9. 還原（restore）**dry-run（試跑）**；實際還原屬於加分項目。

## 安全需求

- 拒絕路徑穿越（path traversal）與符號連結逃逸（symlink escape）；
- 絕不包含 `.git` 內部檔案或已設定排除的路徑；
- 預設遮蔽／省略可能的機密資料；
- 絕不執行 shell 字串；使用 argv 形式的處理程序呼叫；
- 限制產出物大小與檔案數量上限；
- 在規劃還原之前先驗證檢查碼。

## 必要測試

已追蹤／已暫存（staged）／未暫存／未追蹤、重新命名／刪除、二進位檔案、在平台允許的情況下路徑含空白／換行、巢狀 worktree、並行變更、暫存清理、路徑穿越、超出大小限制，以及排除疑似機密資料的檔案。

## 邊界

不要直接更新 Progress Tree。透過 A00 的 ports 回傳 `RepositoryCheckpoint` 與證據參照。
