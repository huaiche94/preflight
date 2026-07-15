# A02 — Claude 遙測、Hooks 與供應商正規化

> 🌐 [English](02-claude-telemetry-hooks.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。本文件屬歷史封存，非現行實作指引。

## 模型

使用 Fable 進行 hook 語意審查；一般的解析器／fixture 工作可使用較便宜的模型。

## ADD 負責範圍

§11 中供應商正規化的部分、整個 §22、Appendix E.2／E.3、§29.7 中的 Claude 案例。

## 專屬路徑

```text
internal/providers/claude/**
internal/hooks/claude/**
internal/telemetry/claude/**
integrations/claude/**
testdata/provider-events/claude/**
internal/storage/sqlite/migrations/0010-0019_*.sql
docs/providers/claude/**
docs/implementation/day1/A02.md
```

## 任務

在不擷取（scrape）TUI 畫面的情況下，實作以 fixture 為基礎的 Claude Code 整合。將狀態列（status-line）與生命週期 hook 的 payload 正規化為凍結的 Preflight 事件，以及與供應商相容的 hook 回應。

## P0 產出物

1. 對以下項目具備未知欄位容忍度的解析器：
   - 狀態列快照；
   - `UserPromptSubmit`；
   - `Stop`；
   - `StopFailure`；
   - 若有可用 fixture，可選擇性支援 `TaskCreated`、`TaskCompleted`、`PreCompact`。
2. 正規化：
   - session／prompt 識別碼；
   - 若存在則正規化五小時用量百分比／重設時間戳記；
   - 上下文用量／視窗（window）；
   - 若存在則正規化輸入／輸出／快取 tokens；
   - 累計成本／耗時／LOC；
   - 失敗類別，包含速率限制（rate limit）；
   - 回合（turn）邊界與供應商能力觀測值。
3. 以供應商事件識別碼或確定性摘要（deterministic digest）為鍵的冪等遙測持久化。
4. 針對 `UserPromptSubmit` 的供應商相容 allow／block 回應編碼器。
5. 呼叫 `preflight hook claude ...` 的 Claude plugin／hooks 範例。
6. 涵蓋正常、缺漏／null、已壓縮（compacted）、高用量、重複、未知欄位、Stop，以及速率限制 StopFailure payload 的 fixtures。

## 隱私

- 預設絕不持久化原始提示詞。
- 只產生提示詞的 SHA-256、位元組長度，以及概略的 token 數量。
- 遮蔽（redact）fixture 中的機密資料。
- 逐字稿（transcript）路徑只是中繼資料，不代表有權讀取逐字稿內容。

## 測試

- 表格驅動（table-driven）的解析器測試；
- 重複事件的冪等性；
- null 配額／上下文行為；
- 接受未知欄位；
- 格式錯誤的 payload 會產生型別化錯誤，並提供有效的 hook 後備方案（fallback）；
- block／allow 回應的 golden files；
- 在持久化資料列／記錄輸出中，斷言原始提示詞不存在。

## 介面行為

A02 不會呼叫具體的預測器。它會發出正規化的 `EvaluateTurnRequest`，或呼叫 A00 的 evaluation port。在套件測試中使用假物件（fake）。

## 加分項目（Stretch）

受管理的 stream-json runner、訊號中斷（signal interruption），以及 session 接續（resume）轉接器。不要為了完成這些項目而犧牲 P0 的 hook 路徑。
