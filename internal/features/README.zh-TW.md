# internal/features/ — 有 privacy 邊界的特徵擷取與 task classifier

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

本套件從 prompt、repository、session 與 Progress Tree（ADD §14.2）中萃取出作為預測輸入
的訊號，並把每一個 turn 分類到固定的 ADD §14.3 task 分類法（taxonomy）上。
[`internal/predictor`](../predictor/README.md) 即建構於此層之上。

Privacy 邊界（Constitution §7 rule 2，凍結的 privacy 合約）：原始 prompt 文字只透過
唯一一個函式 `ExtractPromptFeatures`（`prompt.go`）進入本套件，且絕不會離開。沒有任何
匯出型別會攜帶原始 prompt 文字、其子字串，或任何可還原的編碼——只有衍生出的計數、旗標、
分數，以及一個 SHA-256 摘要。新增一個能夠承載原始 prompt 文字的欄位，是違反合約，而不是
一項功能。

主要組成：

- `PromptFeatures`：大小訊號、ADD §14.7 不需 tokenizer 的 `ApproxTokens` 估計值
  （一律為 ConfidenceLow）、結構計數（路徑、清單項目、驗收標準）、動詞布林值
  （fix/implement/refactor/investigate/migrate——詞彙依 issue #42 擴充），以及領域
  指標（tests、schema/API、security、performance、docs、open-ended、cross-layer、
  repository-wide）。
- `ClassifyTask`（`classifier.go`）：把衍生特徵以確定性、固定優先順序的方式，對應到
  16 個 `TaskClass` 值（`taskclass.go`）。動詞衝突時由 action verb 勝出（#49）；
  訊號不足時回傳 `TaskClassUnknown` 並附上 ConfidenceUnavailable——絕不用猜的。
- `RepositoryFeatures` / `SessionFeatures` / `ProgressFeatures`（`dto.go`）：只包含
  衍生資料的 DTO，quantile 欄位採用 pointer 型別，nil 代表未知，絕不會用零來替代。

涉及本套件、寫作當下仍為開啟狀態的 issue：#50——持久化的 prompt-feature payload
schema，在 hook 端的寫入者（`internal/hooks/claude`、`internal/telemetry/claude`）
與讀回端的解碼器（`internal/evaluation/datasource_sql.go`）之間，重複了同一組欄位
字面值（key literals），且沒有 extraction-version 標記；#51——`ExtractPromptFeatures`
會對 prompt 做好幾次完整遍歷（行分割、轉小寫複本、word set、欄位掃描、rune 迴圈），
並在會阻塞的 hook 路徑上，每次呼叫都有記憶體配置。

套件合約詳見 `doc.go`；上述引用的 ADD 章節見
[Auspex_ADD.md](../../docs/design/Auspex_ADD.md)。
