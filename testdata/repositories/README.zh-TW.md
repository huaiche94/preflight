# testdata/repositories 測試固定資料（fixtures）

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

此目錄存放 `internal/repocheckpoint` 與 `internal/gitx` 測試所需的
fixture，這些測試需要範例儲存庫（repository）的*內容*，而非範例
checkpoint 的 *manifest*（後者請見 `testdata/checkpoints/repository/`）。

checkpoint-b04 並不會在此提交真正的 `.git` 目錄：把一個真實的 Git
儲存庫以純目錄樹的形式提交進另一個 Git 儲存庫，會讓工具（以及在
這個儲存庫本身執行的 `git status`）行為異常，因此 Auspex 自身的測試
套件改為在執行時按需建立真正的暫存儲存庫（詳見 `internal/gitx` 的
`repoBuilder`，以及 `internal/repocheckpoint` 在 `helpers_test.go`
中對應的測試輔助函式）——這才是此角色測試中「真實儲存庫」覆蓋率的
真正來源，每次執行測試都會在暫存目錄中重新產生，而非凍結在此處成為
可能與受測程式碼預期不同步的靜態 fixture 資料。

此角色範疇中較後續的節點（checkpoint-b05 的二進位差異 patch
邊界情境、checkpoint-b06 的機密資訊／路徑過濾、checkpoint-b09 的
路徑穿越（path-traversal）／符號連結逃逸（symlink-escape）安全性
關卡）未來可能會在此加入具體的 fixture 檔案，前提是某個邊界情境
確實需要一個凍結、可供檢視的範例，而非動態產生的暫存儲存庫——例如
帶有特定位元組樣式的二進位檔案，或是路徑中含有難以在 Go 測試程式碼
中以可攜方式建構的字元。此檔案的存在，是為了在那些節點真正加入內容
之前，先把這項慣例記錄下來。
