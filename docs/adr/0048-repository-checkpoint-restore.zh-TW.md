# ADR-048 — 真正的 repository checkpoint restore（issue #6）

> 🌐 [English](0048-repository-checkpoint-restore.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

狀態：已接受（Accepted）
日期：2026-07-14
負責人：由 lead 執行
追蹤：issue #6；終結 checkpoint-b08「真正的 restore 屬於 stretch/deferred」這項 vertical-slice 延後事項

## 背景

過去 Repository Checkpoints 僅支援擷取：create、verify，以及完整的 ADD §19.6 dry-run 都已存在（checkpoint-b04/b08），但沒有任何機制能把 worktree 還原回已擷取的狀態——continuity 故事的最後一環（checkpoint → incident → restore），以及 Graceful Pause「repo 可以回到已知狀態」的 resume 假設，兩者都懸而未決。Constitution 不可協商項 #9（checkpoint 流程絕不可以在使用者不知情下 commit 目前所在的分支）使得 restore 的異動設計成為一項 ADR 等級的決策。

## 決策

### 執行模型

Restore 透過現存最狹窄的異動性基本操作（mutating primitives）來重播該 checkpoint：

1. **staged patch** → `git apply --binary --index`（index ＋ worktree），
2. **unstaged patch** → `git apply --binary`（僅 worktree），
3. **untracked.zip** → 依擷取時自身的路徑安全規則、加上嚴格的「不覆寫（no-clobber）」規則，逐項解壓縮。

`git apply` 無法移動 HEAD、切換分支或建立 commit——「不異動 ref」的保證是結構性的，而非只是一項慣例。Restore 絕不執行 checkout/reset/stash/commit。測試會斷言，一次 restore 前後 HEAD、分支與 commit 數量必須逐位元組完全相同。

### 關卡順序（維持不變，現在成為承重結構）

`Service.Restore` 會先執行既有的 §19.6 dry-run——checksum 驗證、repository 身分確認、dirty-target 政策、對兩份 patch 執行 `git apply --check`——除非該判定結果是乾淨的，否則無法進入 apply 步驟。Dry-run 仍是**預設值**：凍結的 request 新增一個純增量式的 `Apply bool` 欄位，其零值會完全維持 ADR-048 之前的行為（依循 ADR-044 的修訂紀律；已新增 CONTRACT_FREEZE.md 條目）。`RestoreResult` 新增 `SafetyCheckpointID` 與 `UntrackedSkipped`，皆為純增量式欄位。

### Dirty-target 規則（ADD §19.6「safety checkpoint/force」）

- Dirty target 且未帶 `AllowDirty` → 拒絕（維持不變）。
- Dirty target 且帶有 `AllowDirty` ＋ `Apply` → 會先無條件地、透過與正在被還原的來源相同的 `Create` 路徑，擷取一份 restore 前狀態的**安全 checkpoint（safety checkpoint）**——這是操作者的復原（undo）依據，會回傳於結果中，以及任何後續錯誤中。安全擷取若失敗，則整個流程中止且不會有任何異動（絕不會「沒有保險就硬幹」）。
- Clean target → 不需要安全 checkpoint：其狀態就是 HEAD，而 restore 無法移動 HEAD。

### 絕不刪除、絕不覆寫

ADD §19.6「除非帶 `--exact`，否則絕不刪除多餘檔案」的規則，是在結構上被強制執行的：restore 不會刪除任何東西，且 untracked 的解壓縮動作會跳過任何已存在的目的地（以 `O_EXCL` 為底層機制，並透過 Lstat 感知 symlink），並在結果中以 `exists_not_overwritten` 揭露每一次跳過。目前尚未建置 `--exact` 模式；若未來建置，將會修訂本 ADR。

### 惡意封存檔防禦（第二道防線）

在 service 路徑上，一個被竄改過的 `untracked.zip` 在解壓縮之前就已經無法通過 checksum 驗證。解壓縮動作仍然獨立強制執行以下規則：worktree 封閉性（不允許 `..`、不允許絕對路徑、絕不觸碰 `.git`）、僅處理一般檔案項目（symlink／特殊項目一律跳過）、拒絕存在著 symlink 的上層目錄，以及對實際解壓縮後的串流重新套用擷取時自身的大小上限（zip 標頭是攻擊者可控制的）。對抗性測試會使用精心構造的封存檔，直接驅動解壓縮動作。

### 部分套用的誠實揭露

ApplyCheck 在 apply 之前的一瞬間執行，但在這期間 tree 仍可能改變。若 unstaged 的重播在 staged 已經落地之後失敗（或是在兩份 patch 都套用完後解壓縮失敗），錯誤訊息會準確說明 restore 究竟進行到哪一步，並在存在安全 checkpoint 時附上其 ID——絕不會用一個籠統的失敗訊息，掩蓋一個只還原到一半的 tree。

### CLI

`auspex checkpoint restore --id <id> [--apply] [--allow-dirty]`——依 ADD §19.6，預設為 dry-run，輸出為 schema 版本化的 JSON（`auspex.checkpoint-restore.v1`），透過既有的 `CheckpointCreateDeps.RepositoryCheckpoint` service 接線。

## 已考慮的替代方案

- **以 `git stash`／`git checkout` 為基礎的 restore**——已否決：兩者都會在操作者不知情的情況下移動 ref 或建立 stash 狀態；`git apply` 是唯一一個影響範圍恰好等於 §19.2 所擷取範圍的基本操作。
- **每次 apply 都建立安全 checkpoint（連 clean target 也不例外）**——已否決：clean target 的狀態就是 HEAD，本身已經是持久化的；無條件擷取只會讓產物量翻倍，卻沒有帶來任何額外的復原價值。
- **完全拒絕 dirty target（不提供 AllowDirty apply）**——已否決：ADD §19.6 明確提供了 safety-checkpoint/force 這條路徑，且 pause/resume 流程需要它（一個被暫停的 session，其 tree 經常處於 dirty 狀態）。
- **預設刪除多餘檔案（「exact」語意）**——被 ADD §19.6 的 MUST 條款直接否決。

## 影響

- Continuity 的故事至此完整閉環：capture → verify → dry-run → 真正的 restore，全部都依循凍結的 port。
- Graceful Pause 的 resume 驗證現在可以假設 repository 狀態在機制上是可還原的。
- `RestoreResult.Applied` 不再必然為 false——先前把它當作裝飾性欄位處理的呼叫端，現在應該要讀取它。
