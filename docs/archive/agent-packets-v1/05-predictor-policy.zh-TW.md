# A05 — 範圍估算器、預測器、風險、政策與授權

> 🌐 [English](05-predictor-policy.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。本文件屬歷史封存，非現行實作指引。

## 模型

使用 Fable。

## ADD 負責範圍

§§13–17、§29.9–29.11 中預測相關的部分、ADR-018／019／020／026／032／033。

## 專屬路徑

```text
internal/features/**
internal/predictor/**
internal/policy/**
internal/evaluation/**
testdata/models/**
internal/storage/sqlite/migrations/0040-0049_*.sql
schemas/model.schema.json
docs/implementation/day1/A05.md
```

若凍結的目錄配置中沒有 `internal/evaluation`，請使用 A00 指派的確切路徑；不要建立互相競爭的套件。

## 任務

實作具確定性、可解釋、對冷啟動安全的預測器／政策迴圈。day-one 的輸出是風險分數與分位數估計值，而非已校準的機率。

## 產出物

1. 不儲存原始提示詞的提示詞特徵擷取器。
2. 儲存庫／session／progress 特徵 DTO。
3. 具明確 `unknown` 值的簡易任務分類器。
4. 具單調性保證的經驗 P50／P80／P90 工具函式。
5. 讀取／變更檔案數與 LOC 的範圍估計。
6. 依近期回合估算配額增減量與上下文成長。
7. 十分鐘可運作時間（runway）**分數**與預測紀錄。
8. 風險組成項目：配額、上下文、完成度、影響範圍（blast radius）。
9. 原因代碼與信心水準。
10. 政策動作：ALLOW、WARN、CHECKPOINT、SPLIT、PAUSE、ABORT。
11. 評估結果持久化。
12. 一次性授權的發放／消費，並與 prompt／session／evaluation 綁定，具備到期時間與重播（replay）拒絕機制。

## 冷啟動合約

當樣本／校準門檻未達成時：

```json
{
  "calibrated": false,
  "confidence": "low",
  "risk_score": 0.84,
  "probability": null,
  "reason_codes": ["insufficient_history", "quota_headroom_low"]
}
```

絕不能將此數值輸出成「84% 的機率」。

## 初始政策建議

- 配額壓力低且無完整性問題：ALLOW；
- 中度壓力：WARN；
- 預測的 P90 超出可用餘裕（headroom）或影響範圍過大：CHECKPOINT；
- 已校準的十分鐘命中機率連續兩次 >= 設定門檻：PAUSE；
- 未校準的緊急情況：以 `emergency_threshold` 為原因 PAUSE，而非宣稱機率。

## 必要測試

- 分位數單調性的性質測試；
- 缺失值／unknown 行為；
- 不出現除以零／NaN／Inf；
- 相同輸入產生確定性輸出；
- 原因代碼的 golden 測試；
- 政策優先順序與 fail-open／fail-closed；
- 授權恰好消費一次；
- 過期／錯誤 prompt／錯誤 session 的授權遭拒；
- 以時鐘為基準的到期測試；
- 快速路徑的效能基準測試（benchmark）。

## 邊界

不進行供應商 JSON 解析、Git 指令、檢查點建立，或處理程序中斷。透過凍結的 ports 回傳決策。
