# internal/predictor/ — 確定性、可解釋的預測基本元件，以及四階段推估管線

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

此目錄樹存放 Predictor 管線，負責把一則 prompt 轉換成 turn 開始前的推估（pre-turn forecast）。
目前這裡的每一個實作都屬於 Version 1（基於規則、確定性），依循
[Auspex_Predictor_Design_Supplement.md](../../docs/design/Auspex_Predictor_Design_Supplement.md)
（本管線的設計文件）中的演進路線圖（Evolution Roadmap）；管線本身的凍結形狀則定義於
[ADR-041](../../docs/adr/0041-predictor-forecast-layer.md)。程式碼註解中引用的「ADD §…」，
對應章節見 [Auspex_ADD.md](../../docs/design/Auspex_ADD.md)。

管線（每個階段都是凍結的 `internal/app` port，由
[`internal/evaluation`](../evaluation/README.md) 的 `EvaluateTurn` 全程串接）：

1. Prompt features 與 task classifier — [`internal/features`](../features/README.md)
   （`ExtractPromptFeatures`、`ClassifyTask`）；本套件即位於此層之上。
2. Scope estimator — [`scope/`](scope/README.md)（`app.ScopeEstimator`）：預期的檔案數／行數。
3. Token forecaster — [`token/`](token/README.md)（`app.TokenForecaster`）：該 turn 的 token 成本。
4. Quota forecaster — [`quota/`](quota/README.md)（`app.QuotaForecaster`）：turn 結束後推估的
   quota 與 context-window 位置。
5. Risk combiner — [`risk/`](risk/README.md)（`app.RiskCombiner`）：四個風險組成項＋整體風險。
6. Policy — [`internal/policy`](../policy/README.md)：最終階段，將風險與 runway 對應到一個 action。

[`runway/`](runway/README.md) 依設計刻意不在這條鏈之內（ADR-041）：它回答的是「quota window
是否即將在 horizon 內耗盡」，並提供給 `GracefulPauseService.Observe`，而非 `RiskCombiner` 的輸入。

根套件本身包含一個共用的基本元件：`Quantiles` / `EmpiricalQuantiles`
（`quantile.go`），也就是凍結的 P50/P80/P90 三元組，具備無條件的 P50 <= P80 <= P90 保證，
且不會輸出 NaN/Inf。

Cold-start 合約（套件合約詳見 `doc.go`）：第一天的輸出是風險分數與 quantile 估計值，絕不是
經校準的機率——對應 Constitution §7 rule 7：「未經校準的風險分數絕不會被標示為機率。」
