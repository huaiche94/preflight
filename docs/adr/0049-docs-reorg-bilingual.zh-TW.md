# ADR-049 — 文件重整：design 文件移至 `docs/design/`、各資料夾新增 README、新增繁體中文翻譯

> 🌐 [English](0049-docs-reorg-bilingual.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

狀態：已接受（Accepted）
日期：2026-07-14
負責人：lead
核准人：repository owner，2026-07-14（文件重整請求）

## 背景

Repository 根目錄原本放著三份大型 design 文件（`Auspex_ADD.md`、`Auspex_Predictor_Design_Supplement.md`、`Auspex_Parallel_Execution_Plan.md`），以及 GitHub 與 agent 工具預期會存在於根目錄的九個 community／process 檔案（`README.md`、`AGENTS.md`、`CHANGELOG.md`、`CODE_OF_CONDUCT.md`、`CONSTITUTION.md`、`CONTRIBUTING.md`、`GOVERNANCE.md`、`SECURITY.md`、`SUPPORT.md`）。一位初次造訪 repo 的使用者，完全無法從這十二個根目錄 markdown 檔案中判斷哪一個才是入口點，而大多數資料夾也完全沒有任何介紹文件。Repository owner 提出了以下要求（2026-07-14）：為初次造訪者重寫一份 README、將根目錄的 markdown 整理進 `docs/`、為每個資料夾新增一份 `README.md` 介紹，以及為每一份 markdown 文件新增繁體中文版本。

變更 `Auspex_ADD.md` 的存放位置，需要編輯 Constitution 中的路徑參照（§1、§2、§8），而依 Constitution §3，這需要一份 ADR。

## 決策

1. **`docs/design/` 是這三份 design 文件的新家。** 檔名維持不變，因此以文件名稱引用章節（例如 `Auspex_ADD.md §31`）依然清楚明確、可被 grep 到。
2. **仍在維護中的文件會引用新路徑。** `CONSTITUTION.md`、`CONTRIBUTING.md`、`GOVERNANCE.md`、`SECURITY.md`、`SUPPORT.md`、`AGENTS.md`、`README.md` 以及 `agents/*.md`，現在都會參照 `docs/design/Auspex_ADD.md`（及其他同批文件）。
3. **歷史紀錄不會被改寫**——已核准的 ADR（依 Constitution §3.3 為不可變）、`docs/archive/**`、`docs/implementation/**` 的進度紀錄、Go 原始碼註解、JSON schema 說明字串，以及已 checksum 的測試 fixture（`testdata/**`），都維持原本的引用方式。文件名稱依然可以透過 grep 解析找到；只有仍在維護中的文件裡的超連結，需要保持有效。
4. **每個資料夾都會新增一份 `README.md` 介紹**，但內容會被測試逐一列舉或 checksum 的 fixture 目錄除外（`testdata/*` 的葉節點目錄、`internal/cli/testdata/`、`internal/managed/testdata/`）——這些目錄改由最近的上層 README 說明，如此一來新增檔案就不會破壞測試。
5. **雙語文件政策。** 每一份文件性質的 markdown 檔案，都會有一份名為 `<name>.zh-TW.md` 的繁體中文姊妹翻譯檔，並在兩份檔案的開頭互相交叉連結。
   - **文件原始撰寫語言即為規範性文本。** 對於每一份以英文撰寫的文件——也就是除了下方列出的兩份文件之外的所有文件——英文文件才具有規範性，其 `.zh-TW.md` 姊妹檔則是非規範性的閱讀輔助：若兩者有出入，以英文文件為準，翻譯本身即為一項待修的錯誤（bug）。
   - **有兩份文件是以繁體中文撰寫，並且其原文本身即具規範性：** `docs/design/Auspex_ADD.md`（架構權威文件——其內文主體為繁體中文，並搭配英文的章節標籤與程式碼）與 `docs/DECISION_LOG.md`。這兩份文件不會有 `.zh-TW.md` 姊妹檔（那會產生重複、且容易彼此漂移的版本）。若未來為其中任一份文件新增英文翻譯，該英文翻譯才是非規範性的一側。
   - 程式碼區塊、JSON payload、指令名稱、檔案路徑、schema 版本字串與識別字，一律不翻譯。
   - 作為測試輸入的 markdown 檔案（例如 `testdata/checkpoints/state/add-section-18-*.md`）屬於 fixture，而非文件，不進行翻譯。

## 影響

- 根目錄的 markdown 數量從 12 個降為 9 個；留在根目錄的檔案，全部都是 GitHub community 慣例檔案、agent 工具慣例檔案（`AGENTS.md`），或是流程權威文件（`CONSTITUTION.md`）。
- Constitution §1 的 source-of-truth 對照表現在指向 `docs/design/Auspex_ADD.md`；權威**內容**本身並未改變——本 ADR 變更的是位置、並新增翻譯政策，並未變更任何架構或流程規則。
- 在本 ADR 之後，任何人新增文件檔案時，都應在同一次變更中一併新增其 `.zh-TW.md` 姊妹檔；此慣例記載於 `docs/repository_inventory.md`。
- 翻譯的即時性採盡力而為（best-effort）原則：修改英文文件的 PR 應同步更新其翻譯，但過時的翻譯只是一個文件性質的 bug，絕不構成權威性的問題。
