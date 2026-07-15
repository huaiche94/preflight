# internal/artifacts/ — 節點完成背後的 artifact 證據驗證器

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

把一份「宣稱的證據」轉變成「已驗證的證據」的具體檢查。「完成即需有證據」（Constitution
§6.2）就是在這裡被強制執行的：每個驗證器都會實際檢查檔案系統或檔案內容本身，絕不僅憑呼叫端
的宣稱。本套件是 [`../progress/`](../progress/) `CompleteNode` 協定所呼叫的純驗證介接點
（seam）；本身不負責協調交易或持久化（那部分留在 `internal/progress`）。

關鍵進入點（`validator.go`）：

- **`Validator`** —— 一個狹窄的介面：`Kind() string`，外加 `Validate(ctx, Candidate) (Result, error)`。
- **`Candidate`** —— 待測試的證據（路徑、預期的 SHA-256、預期的標題……）。
- **`Result`** —— `Passed`，外加人類可讀的 `Reasons`；驗證失敗的結果一定會附帶至少一個原因。

內建驗證器，以 `Kind()` 字串為鍵，儲存為 `artifacts.validator_id`，並供驗收標準
（acceptance criteria）使用（Auspex_ADD.md §18.5——ADD 現位於
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)）：

- `file_exists`（`file_exists.go`）—— 該路徑是一個實際存在的一般檔案。
- `checksum_matches`（`checksum.go`）—— 檔案實際的 SHA-256 與記錄的證據摘要值相符；正是這項
  檢查，讓「agent 自稱完成」不足以單獨成立。
- `heading_exists`（`heading.go`）—— 某個 Markdown 檔案中，宣稱的標題確實以真正的 ATX 標題行
  存在，且忽略出現在圍籬式（fenced）程式碼區塊內的比對結果。
- `fence_balance`（`fence_balance.go`）—— 每一個開啟的 Markdown 程式碼圍籬都有對應的結尾
  （依 CommonMark 規則）。

**`Registry`**（`registry.go`）依 kind 進行分派，預先內建上述四種驗證器，並且可以在不變更
schema 的情況下註冊自訂的 `Validator`。各驗證器彼此獨立可呼叫——每一個都會自行執行
stat／open，而不會假設其他驗證器已經先執行過。
