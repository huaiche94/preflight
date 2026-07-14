# Claude Provider

> 🌐 [English](claude-provider.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

## 模型

hook 語意審查使用 Fable；例行的 parser／fixture 工作可使用較便宜的模型。

## ADD 負責章節

§11 中 provider 正規化的部分、全部的 §22、附錄 E.2/E.3、§29.7 中的 Claude 案例。

## 專屬路徑

```text
internal/providers/claude/**
internal/hooks/claude/**
internal/telemetry/claude/**
integrations/claude/**
testdata/provider-events/claude/**
internal/storage/sqlite/migrations/0010-0019_*.sql
docs/providers/claude/**
docs/implementation/vertical-slice/claude-provider.md
```

## 任務目標

在不 scraping TUI 的前提下，實作以 fixture 為基礎的 Claude Code 整合。
將 status-line 與生命週期 hook 的 payload 正規化為凍結後的 Auspex 事件，
以及與 provider 相容的 hook 回應。

## P0 交付項目

1. 對未知欄位（unknown-field）具容忍度的 parser，涵蓋：
   - status-line snapshot；
   - `UserPromptSubmit`；
   - `Stop`；
   - `StopFailure`；
   - 若有可用 fixture，則含可選的 `TaskCreated`、`TaskCompleted`、`PreCompact`。
2. 正規化：
   - session/prompt 識別碼；
   - 五小時用量百分比／reset 時間戳（若存在）；
   - context 用量／window；
   - input/output/cache tokens（若存在）；
   - 累積 cost／duration／LOC；
   - 失敗類別，包含 rate limit；
   - turn 邊界與 provider 能力觀測值。
3. 以 provider 事件身分或確定性摘要（deterministic digest）為 key 的冪等 telemetry 持久化。
4. 針對 `UserPromptSubmit`、與 provider 相容的 allow/block 回應編碼器。
5. 呼叫 `auspex hook claude ...` 的 Claude plugin/hooks 範例。
6. 涵蓋正常、缺失/null、compacted、高用量、重複、未知欄位、Stop，
   以及 rate-limit StopFailure payload 的 fixture。

## 隱私

- 預設絕不持久化原始 prompt。
- 僅產生 prompt 的 SHA-256、位元組長度，以及概略 token 計數。
- 遮蔽 fixture 中的機密內容。
- Transcript 路徑僅是中繼資料，不代表擁有讀取 transcript 的權限。

## 測試

- table-driven 的 parser 測試；
- 重複事件的冪等性；
- null quota/context 的行為；
- 接受未知欄位；
- 格式錯誤的 payload 會產生具型別的錯誤，並提供合法的 hook fallback；
- block/allow 回應的 golden file；
- 針對持久化資料列／log 輸出，斷言原始 prompt 不存在。

## 介面行為

本角色不會呼叫具體的 predictor。它會發出正規化的 `EvaluateTurnRequest`，
或呼叫 contract-integrator 的 evaluation port。套件測試中請使用 fake 實作。

## 延伸目標

Managed stream-json runner、訊號中斷（signal interruption），以及 session
resume adapter。不得為了完成這些項目而犧牲 P0 的 hook 路徑。
