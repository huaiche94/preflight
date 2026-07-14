# internal/pricing/ — 本機的逐模型價格表：token 推估 → 估計的美元成本範圍

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

ADR-043 increment 1（「cost forecast first」）：一個小型、本機、手動維護的價格表，把
token 推估值轉換成估計的成本範圍（RANGE）——絕不是單一數值，也絕不是實際量測到的成本。
價格是撰寫當下取自 Anthropic 公開牌價的預留（placeholder）預設值，絕不會在執行期間即時
抓取（local-first），且預期會隨時間漂移；它們對 caching、batch 折扣或訂閱方案一無所知。

主要進入點（`pricing.go`）：

- `Table` / `DefaultTable()` / `NewTable(overrides)`：不可變的 family → price 對照表
  （fable/mythos、opus、sonnet、haiku），以 model ID 的小寫子字串、依確定性的排序順序
  做匹配。
- `Table.Price(modelID)` 會解析出一個 `ModelPrice`（每 MTok 的美元價格，區分
  input/output），以及它所匹配到的 family；未知的模型會退回到 `DefaultFamily`
  （sonnet 等級——也就是 Claude Code 的預設等級；用 opus 會系統性高估 5 倍，用 haiku
  則會低估）。
- `Table.EstimateTurnCost(modelID, tokensLow, tokensHigh)` 產生一個 `CostRange`：
  由於 token 推估值是未區分 input/output 的總量，下界會把每個 token 都以 input
  價格、依低分位數計價，上界則以 output 價格、依高分位數計價——這個區間反映的是真實的
  不確定性，因此即使 P50 == P90 也會產生真正的範圍。`ModelFamily` 與 `Source` 會揭露
  是哪一個價格假設產生了這個數字。

消費端：[`internal/evaluation`](../evaluation/README.md) 的 forecast card（會把該範圍
標示為「uncalibrated estimate」，Constitution principle #2），以及
[`internal/policy`](../policy/README.md) 的 cost-budget 規則（`costbudget.go`），兩者
都由 [`internal/predictor/token`](../predictor/token/README.md) 的第 2 階段推估提供
資料。

以 YAML config 覆寫是一項已記載在案的後續工作，目前尚未建置——今天 `cmd/auspex` 尚未
接上任何正式環境的 config loader；`NewTable` 的 overrides 參數是目前既有的程式化
seam。本套件沒有 `doc.go`；套件合約就是 `pricing.go` 開頭的套件註解。
