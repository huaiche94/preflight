# internal/predictor/risk/ — 第 4 階段：針對第 1–3 階段輸出的可解釋風險組合器

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`RuleRiskCombiner`（`combiner.go`）實作了凍結的 `app.RiskCombiner` port（ADR-041、
predictor-07）。無狀態：`app.CombineRiskRequest` 直接帶有三個上游輸出
（[`scope/`](../scope/README.md) 的 `ScopeEstimate`、[`token/`](../token/README.md) 的
`TokenForecast`、[`quota/`](../quota/README.md) 的 `QuotaForecast`）。

它逐字計算 ADD §16.2 的「Initial explainable formula」（係數位於 `coldstart.go`，並依公式
自身的變數命名）：

- quota_risk / context_risk = sigmoid((推估 P90 − 85) / 7)，來自兩個 `QuotaForecast`
  推估值。若推估值為 nil，則給出 sigmoid 中點 0.5 的分數，並附上
  `QUOTA_UNKNOWN`／`CONTEXT_UNKNOWN`——絕不捏造出一個 0（ADD §16.3）。
- completion_risk / blast_radius_risk：以 `ScopeEstimate` 欄位計算並夾限（clamp）的線性
  公式；沒有對應凍結欄位的項目（open-ended scope、retry/測試失敗率、progress 阻塞、
  public API 變更）則從 `ScopeEstimate.ReasonCodes` 讀成布林指標——這是一座有明文記載
  的橋接，因為凍結的 request 無法再擴增。（ADR-041 已固定 completion_risk 這個名稱；
  「execution_risk」是同一個概念，此處絕不使用。）
- overall ＝四者中的最大值；其 Confidence 取四者中最低者，ReasonCodes 則是四者去重後的
  聯集。分數一律落在 [0,1] 之內；NaN 會被夾限為 1.0（最保守的取值）。

不變式（Constitution §7 rule 7）：未校準分數絕不是機率。每一個 `domain.RiskComponent.Score`
都是 0–1 的風險分數，而非機率，除非每一個貢獻的輸入本身都是 `Calibrated=true`——本波不
可能出現這種情況，因為第 1–3 階段全部都僅有 cold-start。Calibrated/Confidence 一律誠實地
從上游傳遞，絕不憑空製造；在 cold-start 路徑上，下游揭露機率的介面
（[`internal/policy`](../../policy/README.md) 的 `Decision.Probability`、
[`internal/evaluation`](../../evaluation/README.md) 的 `ForecastCard.Probability`）都會
輸出 probability null。

上述引用的 ADD 章節見 [Auspex_ADD.md](../../../docs/design/Auspex_ADD.md)。套件合約詳見
`doc.go`。
