# internal/predictor/quota/ — 第 3 階段：確定性的 quota 與 context-window 推估

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`RuleQuotaForecaster`（`forecaster.go`）實作了凍結的 `app.QuotaForecaster` port（ADR-041、
predictor-05c）：它推估下一個 turn 結束後的 provider-quota 位置（ADD §15.3 差值模型）與
context-window 位置（ADD §15.9），並把兩者一併產出於單一的 `domain.QuotaForecast` 中，
因為它們共用相同的差值技術，且都提供給 [`risk/`](../risk/README.md) 的
quota_risk/context_risk 項。

它是無狀態的——`app.ForecastQuotaRequest` 已經帶有目前的
`QuotaObservation`／`ContextObservation`，以及來自 [`token/`](../token/README.md) 上游
第 2 階段的 `TokenForecast`，因此這裡不需要 `FeatureSource` 這層抽象。

運作機制：

- Quota：每個 limit window 都會得到目前使用率百分比＋預設的 P90 差值（`coldstart.go`：
  P50/P90 分別為 2／6 個百分點——這是本套件自行記載的 bootstrap 常數，ADD §15.3 並未
  指定）。推估結果最差的 window 決定唯一的純量輸出。若某個 window 的 `ResetsAt` 落在
  10 分鐘的 turn horizon 之內，就不會把重設後的用量一併累加（§15.8）。
- Context：預設淨成長量以 token 數表示（P50/P90 分別為 6k／20k，依 decision D-14——
  先前是以 window 比例表示，這在 1M window 上會高估一個數量級），再透過觀測到的 window
  大小換算成百分點；D-14 之前的比例備援方案，僅在 window 大小未知時才會套用。精確的
  `UsedTokens/WindowTokens` 優先於 provider 四捨五入後的 `UsedPercent`。
- 預設差值會依 token 推估值相對於 6000-token 的名目 turn 做縮放，並限制在 [0.5, 3.0]
  範圍內，避免單一極端推估把預設值抹除或炸開。
- 未知維持未知：缺少觀測值時會產生 nil 推估值，並附上
  `QUOTA_UNKNOWN`／`CONTEXT_UNKNOWN`，絕不會捏造出一個 0。

本波每一個結果都是 `Calibrated=false`，Confidence 為 low，並附上
`PREDICTION_COLD_START`——目前尚不存在任何經驗性的差值分布，這正是 CONTRACT_FREEZE.md
為第一版實作所允許的狀態。上述引用的 ADD 章節見
[Auspex_ADD.md](../../../docs/design/Auspex_ADD.md)。套件合約詳見 `doc.go`。
