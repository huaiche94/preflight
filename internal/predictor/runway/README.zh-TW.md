# internal/predictor/runway/ — 十分鐘 quota 耗盡 runway 分數，獨立於主管線之外

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`Scorer.Score`（`runway.go`）依照 ADD §15.4–15.5 計算出 `domain.RunwayForecast`：「是否有
任何啟用中的 quota window 即將在接下來的 H 秒內耗盡」（`DefaultHorizon` = 600s）。這依
設計刻意不屬於 Scope → Token → Quota → Risk 這條鏈（ADR-041）：它由
`internal/app.GracefulPauseService.Observe`（由 runtime 角色擁有）消費，而
[`internal/policy`](../../policy/README.md) 只把它當作第二個、獨立的輸入來使用。
[`internal/evaluation`](../../evaluation/README.md) 的 `EvaluateTurn` 絕不會重新執行
它——只會讀回最近一次已計算好的推估結果。

運作機制：

- Burn rate 是同一個 limit window 最近兩次觀測值之間的瞬時 Δused% / Δminutes
  （`ScoreRequest.Current`／`Previous`）。Scorer 本身無狀態；觀測歷史由呼叫端自行持有。
- ADD §15.4 的離群值規則：間隔 < 2 秒不計入；負的差值視為 reset／校正；超過 50 pp/min
  這個合理性上限的速率會被視為異常而捨棄；超過 5 分鐘的樣本會降低 confidence。只有單一
  區間時，P50 = P90（不捏造任何分散度）。
- 分數來自 ADD §15.7 未校準的備援門檻：目前使用率 >= 95% → 1.0；horizon 內推估 P90
  >= 100% → 0.85；推估 >= 95% → 0.65；其餘則是依 headroom 平滑縮放的數值。若 reset
  落在 horizon 之內，則覆寫為低分（表示尚有 headroom 可用，§15.8）。provider 回報的
  `Reached` 會立即給出 1.0。

Cold-start 合約（ADD §15.6–15.7）：在沒有持久、經校準的 burn-rate 歷史紀錄之前（需
>= 20 筆有效樣本、held-out evaluation、ECE <= 0.08——本波從未達成），`HitProbability`
維持 nil，`Calibrated` 維持 false；`RiskScore` 是確定性的 0–1 分數，絕不會被當成機率
呈現。這裡的 ReasonCodes 是純字串，與凍結的（ADR-041 之前的）`RunwayForecast` 形狀一致。

上述引用的 ADD 章節見 [Auspex_ADD.md](../../../docs/design/Auspex_ADD.md)。套件合約詳見
`doc.go`。
