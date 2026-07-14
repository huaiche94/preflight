# internal/policy/ — 管線的最終階段：risk ＋ runway → 一個凍結的 policy action

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`Decider.Decide`（`decide.go`）把來自 [`internal/predictor/risk`](../predictor/risk/README.md)
的綜合風險結果，以及來自 [`internal/predictor/runway`](../predictor/runway/README.md) 的
獨立 runway 推估——這是兩個訊號唯一合理匯集之處（ADR-041）——轉換成八個凍結的
`app.PolicyAction` 值（`internal/app/ports.go`，ADD §17.2）之一：

`RUN`、`WARN`、`REQUIRE_CONFIRMATION`、`CHECKPOINT_AND_RUN`、`SPLIT`、`PAUSE`、
`PAUSE_AND_AUTO_RESUME`、`BLOCK`。（目前這個 Decider 只會輸出其中六個；`SPLIT` 與
`PAUSE_AND_AUTO_RESUME` 是凍結的 enum 值，這裡目前還沒有任何程式路徑會產出它們，儘管
`context.go` 的 action-strength 階梯已經把它們納入排序。）

各個 gate 依照固定的 ADD §17.3 優先順序執行，第一個命中者勝出：explicit deny →
integrity failure（皆為呼叫端提供的布林值；兩者都會 fail closed 到 `BLOCK`）→
runway pause → 強制性的 checkpoint 邊界 → 接著才是 ADD §16.5 的風險區間（<0.45 為
RUN，0.45–0.65 為 WARN，0.65–0.85 在 blast radius 也偏高時為 REQUIRE_CONFIRMATION 或
CHECKPOINT_AND_RUN，>=0.85 為 CHECKPOINT_AND_RUN）。有兩條疊加規則只能加強、絕不能
削弱最終選定的 action：D-08 的 context-utilization 門檻（`context.go`；推估 P90
context >85% → WARN，>95% → CHECKPOINT_AND_RUN；預設啟用但受 confidence 把關，因此
現行的 cold-start forecaster 從未觸發它們），以及選擇性加入（opt-in）的單次 turn
成本預算（`costbudget.go`，ADR-043 increment 3，價格由
[`internal/pricing`](../pricing/README.md) 提供）。

Runway 的 PAUSE 有兩條路徑：一是觀測到經校準的 hit-probability >= 0.80，並搭配 §17.6
的雙樣本 debounce（`PriorRunwayHitConfirmed`，狀態由呼叫端持有）；二是未校準的緊急
狀況（limit 已達到、used >= 98%，或 time-to-limit P50 <= 60s），此路徑會跳過
debounce，且一律附上原因 `emergency_threshold`——絕非機率宣稱。

Cold-start policy（具承重意義的不變式，Constitution §7 rule 7）：只要任何一個上游輸入
是 `Calibrated == false`，無論選出哪個 action，`Decision.Probability` 一律無條件為
nil。恰好只有一條程式路徑會設定非 nil 的機率值，且它會直接檢查
`RunwayForecast.Calibrated == true`；風險分數絕不會被複製進 Probability。
`coldstart.go` 的 `ColdStartExample` 以一個可測試的字面值，固定住這個合約的形狀。
`Decide` 絕不回傳 error——每一個缺口都會降級成最保守的適用決策。套件合約詳見
`doc.go`；上述引用的 ADD 章節見 [Auspex_ADD.md](../../docs/design/Auspex_ADD.md)。
