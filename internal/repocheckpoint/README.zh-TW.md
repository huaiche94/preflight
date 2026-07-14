# internal/repocheckpoint/ — Repository Checkpoint（儲存庫檢查點）：工作樹證據擷取、驗證、還原

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

在 pause 或高風險 turn 發生之前，精確擷取工作樹（working-tree）證據，且絕不會異動（mutate）
repository（Auspex_ADD.md §19——ADD 現位於 [docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)；
線上 schema 為 `auspex.repository-checkpoint.v1`，附錄 D）。Constitution §7 規則 6：repository
checkpoint 必須是原子性的，且絕不會靜默地提交（commit）目前作用中的分支——擷取過程中的每一個
Git 操作，都是透過 [`../gitx/`](../gitx/) 執行的唯讀底層（plumbing）操作。

關鍵組成：

- **`Capture`**（`capture.go`）—— 寫出一個 checkpoint artifact 目錄：`staged.patch.gz` ／
  `unstaged.patch.gz`（來自 `gitx.Client.DiffPatch`）、`untracked.zip`（`archive.go`），以及
  `manifest.json`。
- **安全政策**（`security.go`）—— 每一個被歸檔的路徑都會通過 `validateUntrackedPath`：拒絕
  路徑穿越（path-traversal）、符號連結（symlink），以及 `.git` 內部檔案，並設有大小上限
  （單檔 5 MiB、總計 100 MiB、10,000 個檔案）；被略過的候選項目會被記錄於 `SkipReason` 帳本中，
  作為可復原性警告的依據。
- **機密內容遮蔽（redaction）** —— 未追蹤（untracked）的檔案會以 [`../redact/`](../redact/)
  進行掃描，一旦命中就會被略過（`archive.go`）；patch 內容則會針對 `+`／`-` 行內容做原地區段
  遮蔽（`patchredact.go`）。diff 標頭中的檔名，以及二進位 diff 標頭／內容行，則刻意不做改寫——
  這項可接受的殘留風險面記錄於 [ADR-042](../../docs/adr/0042-patch-redaction-residual-surface.md)。
- **原子性**（`atomicwrite.go`）—— artifact 先暫存於暫存目錄中、執行 fsync，再以原子方式重新
  命名（rename）到最終位置；一個 checkpoint 目錄要嘛是完整的，要嘛就不存在。`orphanscan.go`
  會在啟動時清除因行程被強制終止而遺留下來的暫存目錄。
- **`Verify`**（`verify.go`）—— 重新從磁碟讀取 `manifest.json` 與每一個 artifact 檔案，並重新
  計算大小與 SHA-256 摘要值；資料庫資料列本身的欄位絕不會被直接採信。
- **還原（Restore）** —— `restoredryrun.go` 是僅產生報告的預演（dry run）（ADD §19.6 檢查項目、
  `git apply --check`）；真正會造成異動的還原（`restore.go`、`RestoreApply`）隨 issue #6 ／
  ADR-048 一併推出：透過 `git apply` 重播 staged 與 unstaged 的 patch、以嚴格的不覆寫
  （no-clobber）規則解壓縮 untracked 檔案，且絕不觸碰任何 ref（不執行 checkout、reset、commit
  或 stash）。
- **`Service`**（`service.go`）—— 實作凍結的 `app.RepositoryCheckpointService` port
  （`Create` ／ `Verify` ／ `Restore`）；資料列透過 `store.go` 持久化
  （`migrations/0030_repository_checkpoints.sql`）。
