# Predictor

> 🌐 [English](predictor.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

Scope Estimator、Predictor、Risk、Policy 與 Authorization。

## 模型

使用 Fable。

## ADD 負責章節

§§13–17、§29.9–29.11 中與預測相關的部分、ADR-018/019/020/026/032/033。

## 專屬路徑

```text
internal/features/**
internal/predictor/**
internal/policy/**
internal/evaluation/**
testdata/models/**
internal/storage/sqlite/migrations/0040-0049_*.sql
schemas/model.schema.json
docs/implementation/vertical-slice/predictor.md
```

若凍結後的目錄結構中沒有 `internal/evaluation`，請使用 contract-integrator
指定的確切路徑；不得另外建立一個互相競爭的 package。

## 任務目標

實作一個具決定性（deterministic）、可解釋、且對 cold-start 安全的
predictor/policy 迴圈。day-one 的輸出是風險分數與分位數（quantile）估計，
而非經校準（calibrated）的機率值。

## 交付項目

1. 不儲存原始 prompt 的 prompt 特徵擷取器（feature extractor）。
2. Repository/session/progress 特徵 DTO。
3. 具備明確 `unknown` 值的簡易任務分類器。
4. 具單調性（monotonic）保證的經驗 P50/P80/P90 工具。
5. 針對讀取／變更檔案數與 LOC 的 scope 估計。
6. 依近期 turn 推算的 quota-delta 與 context 成長估計。
7. 十分鐘 runway **分數**與預測紀錄。
8. 風險組成要素：quota、context、completion、blast radius。
9. 原因代碼（reason codes）與信心水準（confidence）。
10. 政策動作：ALLOW、WARN、CHECKPOINT、SPLIT、PAUSE、ABORT。
11. Evaluation 持久化。
12. 具備 prompt/session/evaluation 綁定、到期時間，以及 replay 拒絕機制的
    一次性 authorization 發放／消耗。

## Cold-start 契約

當樣本數／校準門檻未達標時：

```json
{
  "calibrated": false,
  "confidence": "low",
  "risk_score": 0.84,
  "probability": null,
  "reason_codes": ["insufficient_history", "quota_headroom_low"]
}
```

絕不可依此數值輸出「84% probability」這類敘述。

## 初始政策建議

- quota 壓力低且無完整性（integrity）問題：ALLOW；
- 中度壓力：WARN；
- 預測的 P90 超過可用餘裕（headroom），或 blast radius 偏高：CHECKPOINT；
- 經校準的十分鐘命中機率連續兩次 >= 設定門檻：PAUSE；
- 未經校準的緊急狀況：以 `emergency_threshold` 作為原因回傳 PAUSE，
  而非宣稱機率值。

## 必要測試

- 分位數單調性（monotonicity）的 property test；
- 缺失值／unknown 行為；
- 不得出現 divide-by-zero/NaN/Inf；
- 相同輸入產生決定性（deterministic）輸出；
- reason-code 的 golden test；
- 政策優先順序與 fail-open/fail-closed；
- authorization 恰好只能被消耗一次；
- 過期／錯誤 prompt／錯誤 session 的 authorization 應被拒絕；
- 以時間為界的到期測試；
- 快速路徑（fast path）的 benchmark。

## 邊界

不得進行 provider JSON 解析、Git 指令操作、checkpoint 建立，或流程中斷。
決策一律透過凍結後的 ports 回傳。
