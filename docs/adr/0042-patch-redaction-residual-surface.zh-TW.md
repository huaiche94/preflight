# ADR-042 — Patch redaction 不含檔名與 binary-diff 標頭：已接受的殘留曝露面

> 🌐 [English](0042-patch-redaction-residual-surface.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

狀態：已接受（Accepted）
日期：2026-07-13
負責人：checkpoint（Part B／repocheckpoint），決策由 lead 記錄
核准人：repository owner，2026-07-13（issue #5 決策會議）

## 背景

qa-09 的最終嚴重度報告（P2 #2，`docs/implementation/vertical-slice/qa.md`）記載，`internal/repocheckpoint/patchredact.go` 僅對 staged/unstaged patch 中 `+`/`-` 的**行內容（line bodies）**進行 secret-shaped 內容的 redaction。它從未改寫下列內容：

- `diff --git a/... b/...`、`--- a/...`、`+++ b/...` 標頭行中的檔案路徑（路徑本身也可能是 secret-shaped）；
- binary-diff 的標頭／標記行（`Binary files a/X and b/Y differ`、`GIT binary patch` 及其 base85 payload）；
- context 行（開頭為空白）——這部分已另行被固定（pinned），因為 context 必須與目標檔案逐位元組相符，`git apply` 才能運作。

qa-09 將此標記為理論上的殘留曝露面，而非已確認的洩漏：沒有任何測試建構過 secret-shaped 的檔名，也未曾觀察到真實案例。

## 決策

**接受此殘留曝露面。不將 redaction 擴及 patch 標頭、檔名或 binary-diff payload。**

理由如下，依權重遞減排列：

1. **Patch 的有效性是承重的（load-bearing）。** Restore dry-run（checkpoint-b08）與任何未來的真實 restore（issue #6）都依賴對這些 patch 執行 `git apply --check`。改寫路徑或 binary payload 會產生一個無法再套用回其擷取來源 repository 的 patch——為了移除一個祕密而摧毀整份 checkpoint 的證據價值，而就檔名的情況而言，這個祕密*早已公開存在於使用者自己的 working tree 與 `git status` 輸出中*。
2. **威脅模型不同。** Redaction 這道程序存在的目的，是阻止祕密**值**（token、金鑰）被複製進入持久化的 checkpoint 產物中。一個 secret-shaped 的**檔名**並非靜置於產物中的憑證；它是使用者自己選擇的 repository 中繼資料。Untracked-archive 路徑（checkpoint-b06）——會複製整份檔案**內容**——仍保留其完整的逐檔祕密掃描與跳過紀錄（skip ledger）。
3. **發生率。** Secret-shaped 的檔名在實務上極為罕見；每一個曾觀察到的真實洩漏類別（qa-05 的 tracked-diff P1，已於 `f981bde` 修復）都涉及行內容，而這部分已受涵蓋。

## 影響

- 這條邊界已由測試固定，任何無聲的行為變化都會被抓到：`internal/repocheckpoint/patchredact_internal_test.go` 中的 `TestRedactPatchSecrets_ContextLine_NeverModified`、`TestRedactPatchSecrets_FileHeaderLines_NeverTreatedAsContent`，以及（隨本 ADR 新增的）`TestRedactPatchSecrets_ADR042_SecretShapedFilenameAndBinaryHeaders_AcceptedBoundary`。
- `patchredact.go` 的 doc comment 已記載了行範圍規則；現在它有了一份可引用的正式決策紀錄。

## 重新檢視觸發條件

若發生下列任一情況，應重新開啟此決策：

1. 觀察到或收到通報，發生透過檔名或 binary-diff 標頭的真實（非理論性）洩漏。
2. `internal/redact` 具備結構化的 patch 改寫能力（例如 re-hunking），能在**不**破壞 `git apply --check` 的前提下對標頭進行 redaction。
3. Checkpoint 產物新增了預設會離開本機的匯出／分享路徑（目前僅存於使用者資料目錄下，僅限本機），這將改變曝露模型。
