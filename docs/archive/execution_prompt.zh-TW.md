> **ARCHIVED — 已過時。** 這份早期草擬的啟動提示詞，指定的是四人、兩波次的團隊結構，與 `Preflight_Day1_Parallel_Execution_Plan.md` 及 `agent-packets/` 中核准的九代理人（A00–A08）拓撲相衝突。它也早於目前「Phase 0 期間不衍生隊友」的指示。僅保留作為歷史參考；請勿依原樣執行此提示詞。

> 🌐 [English](execution_prompt.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。本文件屬歷史封存，非現行實作指引。

建立一個恰好由四位隊友組成的代理人團隊。

未經我同意，不要衍生超過四位隊友。
使用 tmux 分割視窗模式。
若 Fable 可用，每位隊友都使用 Fable。

你是團隊負責人兼整合者。
不要親自實作隊友所擁有的正式程式碼。
在所有必要產出物與驗證結果都存在之前，不要宣告完成。

合約凍結提交（commit）為：

<INSERT_CONTRACT_COMMIT_SHA>

在建立任務之前，先閱讀以下檔案：

- Preflight_ADD.md
- Preflight_Day1_Parallel_Execution_Plan.md
- docs/implementation/day1/CONTRACT_FREEZE.md
- AGENTS.md

建立兩個執行波次。

第一波：
- foundation：A01 基礎建設、設定、路徑、SQLite
- claude-adapter：A02 Claude 遙測與 Hooks
- state-checkpoint：A03 進度樹與狀態檢查點
- repository-checkpoint：A04 Git 觀察器與儲存庫檢查點

第二波：
- foundation 負責 A05 預測器與政策
- claude-adapter 負責 A06 優雅暫停與排程器
- state-checkpoint 負責 A07 CLI 與應用程式編排
- repository-checkpoint 負責 A08 QA、安全性、可靠性與 CI

第二波任務必須維持阻塞狀態，直到其所需的第一波相依項目完成為止。

嚴格的所有權規則：

1. 每位隊友只能修改自己被指派的路徑。
2. 絕不修改其他隊友的檔案。
3. 只有 foundation 可以修改 go.mod 或 go.sum。
4. 只有負責人可以修改共用合約。
5. 隊友不得執行 git add 或 git commit。
6. 負責人是唯一的 Git 提交者。
7. 每位隊友完成每個邏輯節點後，都必須更新自己的進度產出物。
8. 除非所需檔案都存在且驗證指令都通過，否則任務不算完成。
9. 若共用合約不夠用，請傳訊息給負責人，而不是自行變更它。
10. 在整合或開始第二波之前，等待所有第一波隊友完成。

在實作之前，建立完整的共用任務圖並展示給我看：

- 任務 ID
- 相依關係
- 指派的隊友
- 擁有的路徑
- 驗證指令
- 預期產出物

在任何隊友撰寫程式碼之前，需要先核准計畫。
只有在計畫遵守凍結合約、遷移範圍、路徑所有權與測試要求時，才可以核准該計畫。
