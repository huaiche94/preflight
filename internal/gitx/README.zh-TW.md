# internal/gitx/ — 為 Repository Checkpoint（儲存庫檢查點）層提供、僅使用 argv 的 Git 底層操作（plumbing）

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

每一次 Git 呼叫都是以 argv 陣列形式，透過 `domain.ProcessRunner` 執行——本套件絕不會組成或
執行 shell 指令字串（Constitution §7 規則 5）。`ExecRunner`（`runner.go`）是以 `exec.Command`
為基礎的實作；非零的結束碼（exit code）被視為資料（`ProcessResult.ExitCode`），而不是錯誤。

`Client`（`client.go`）將 runner 包裝成 checkpoint 層所需的一組固定 Git 操作，所有旗標
（flag）都是釘死（pinned）的，因此使用者的 Git 設定永遠不會影響擷取到的內容
（Auspex_ADD.md §19.4——ADD 現位於 [docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)）：

- **`Status`** —— `git status --porcelain=v2 -z --branch --untracked-files=all --find-renames`，由 `porcelain.go` 解析成型別化的項目，外加 `BranchInfo`。
- **`DiffNumstat`**（`numstat.go`）—— `git diff --numstat -z --no-ext-diff --find-renames`，可用於 staged（`--cached`）或 unstaged。
- **`DiffPatch` ／ `ApplyCheck` ／ `Apply` ／ `ListUntracked`**（`patch.go`）—— 對二進位安全的 patch（`--binary --full-index --no-ext-diff`）、預演式的套用檢查（dry-run apply check），以及還原流程會用到的兩個會造成異動的 `git apply` 呼叫（皆無法移動 ref）。
- **`ResolveRepo`**（`resolver.go`）—— 將工作樹中的任意路徑對應到 `RepoInfo`（worktree 根目錄、git 目錄、共用目錄、連結 worktree 偵測）；需要 git >= 2.31。
- **`Fingerprint`**（`fingerprint.go`）—— repository 狀態的確定性 SHA-256 摘要值（schema 為 `auspex.gitx.fingerprint.v1`），涵蓋 HEAD、status 項目、numstat 計數，以及 untracked 列舉政策，用於 checkpoint 身分識別與變更偵測。

主要使用者：[`../repocheckpoint/`](../repocheckpoint/)，其擷取路徑僅限於唯讀子集
（status、diff、ls-files）。
