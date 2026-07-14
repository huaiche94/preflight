# ADR-043 — 將 quota runway 廣義化為多資源 forecast（context window、cost budget、rate limits）

> 🌐 [English](0043-multi-resource-runway.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

狀態：已接受（Accepted）（方向性決策）；實作分階段隨 issue #14 進行
日期：2026-07-13
負責人：lead（predictor + policy 相關介面），任何凍結 port 的變更則由 contract-integrator 負責
核准人：repository owner，2026-07-13（issue #13 決策會議）

## 背景

Auspex 的 evaluation 管線將 **provider quota** 視為主要的可耗盡資源：`QuotaForecaster`（ADR-041 第 3 階段，`internal/predictor/quota`）推估滾動視窗（rolling-window）的 quota 與 context 百分比，而 Graceful Pause 的主要觸發條件是「quota 上限經校準後，判定為很可能即將命中」。

Provider 生態正朝著這個假設之外的方向演變：Codex 已經移除其 5 小時滾動限制，Claude 也可能跟進。當硬性的 provider 限制被弱化或消失時，使用者最關切的問題會從「*我會不會被切斷？*」轉變為「*這次執行會不會無限制地燒掉金錢、context 或時間？*」——而在一個沒有限制的 provider 之下，沒有任何機制能阻止一個過夜的 agent 執行花掉數百美元。

這個架構其實早已預見此一轉向：prediction 與 policy 是分離的（ADD §6.6）、provider 能力是明確且可降級的（§6.7，§8.6 中 `RollingQuotaUsage: false` 是合法狀態），而 `domain.QuotaForecast` 早已同時承載 quota 推估**與** context 推估——quota 一直以來就只是多個資源中的其中一個。

## 決策

1. **資源集合。** Forecast 層涵蓋四類可耗盡的資源，各自產生相同形狀的輸出（推估的 P50/P90 使用率、是否能在剩餘視窗內完成的判定、reason codes、`DataQuality`）：
   - **Provider quota／rate limits**——維持現行行為，只要 provider 仍公開限制就照舊（週上限與 API 429 類別不會隨 5 小時視窗消失）。
   - **Context window**——已由 `projectContext` 推估；由次要欄位升格為一級資源（first-class resource）。
   - **Cost budget**——新增：token forecast × 定價／消耗模型，與 config 中*使用者宣告*的預算（per-turn、per-day）比較。不需要任何 provider 訊號。
   - **Wall-clock time budget**——新增、可選：推估的 turn 耗時 vs. 使用者宣告的時間預算。最後才會出貨；需要目前尚不存在的耗時 telemetry（issue #15／#11）。
2. **Policy 輸入從 provider 授予的限制轉為使用者宣告的預算。** Policy 引擎的八個凍結 action 不變；預算被突破時會對應到相同的 action（`WARN`、`CHECKPOINT_AND_RUN`、`PAUSE`、`BLOCK`……）。預算存放在 YAML config 中，套用既有的優先序規則；未設定預算則代表該資源在 policy 上單純不啟用（明確降級，絕不用猜測代替）。
3. **Pause/wake 機制原封不動地重複使用。** Wake job 的觸發條件新增一個資源類別區別子（resource-class discriminator）（quota 重置時間 → budget 重置時間／provider 恢復），但持久化的 scheduler、lease 與 resume 驗證語意完全維持原樣。
4. **對合約的影響是純增量式的。** `domain.QuotaForecast` 維持不變（凍結合約）；此次廣義化在 evaluation 結果上新增一個同層級的 `domain.ResourceForecast` 清單，依資源類別填入。既有項目沒有任何重新命名或移除；REC-05 的 multi-window 問題得到的答案是「是——每個 window/資源各有一筆 forecast 項目，形狀相同」，而不是靠擴寬單一 struct 來解決。

## 影響

- Auspex 的價值主張不再依賴 provider 維持硬性限制；逐 prompt 的 cost/scope/risk 估算（issue #14）成為主要介面，而本 ADR 提供其資源／成本模型。
- 定價表成為一份需要維護的產物（依 provider/model 區分，可由 config 覆寫，預設絕不在執行期抓取——local-first）。
- 未經校準的 forecast 仍標示為分數／範圍，絕不宣稱為機率（Constitution principle #2），且與資源類別無關。

## 排序

依 issue #14 這條線逐步實作：cost forecast 最先（使用者價值最大，不需要新的 telemetry）、context-window 升格居次、time budget 最後。Issue #13 追蹤本 ADR；#14 追蹤介面；#11/#12 提供校準資料。

依 D-08，Increment 2（context-window 升格）已於 2026-07-13 出貨：`internal/policy/context.go` 中預設啟用的門檻（推估的 P90 context >85% 觸發 WARN／>95% 觸發 CHECKPOINT_AND_RUN，並以 confidence 為關卡使 cold-start 推估保持靜默，可透過 `internal/policy.Config` 調整／停用），推估值已持久化（migration 0045）並顯示於每一個 forecast-card 介面上。

Increment 3（cost budget）已於 2026-07-14 出貨（issue #13）：`policy.Config.TurnCostBudgetUSD`——依本 ADR「未設定預算即代表 policy 未啟用」的原則，在零值時為不啟用——`internal/policy/costbudget.go` 中有一套兩層規則，作用於管線現在會在做決策前、依 session 已標記的 model（#20 Phase 0）計算出的誠實成本範圍：最壞情況估計超出預算 → WARN，即使是樂觀估計超出預算 → CHECKPOINT_AND_RUN；沿用與 increment 2 相同的「絕不降級」階梯與 reason-code 揭露方式，沒有 cold-start confidence 關卡（宣告預算屬使用者自願加入，且決策的 Calibrated/Confidence 欄位仍會揭露估計品質）。Config 介面：程式化的 `Service.Policy` seam——YAML config 串接仍是已記錄在案的 composition-root 缺口。Time budget（increment 4）仍在等待耗時 telemetry（#15/#11）；`domain.ResourceForecast` 的同層清單仍與 contract-integrator 一同保持開放。
