# Preflight Day-1 合約凍結

> 🌐 [English](CONTRACT_FREEZE_TEMPLATE.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。本文件屬歷史封存，非現行實作指引。

狀態：DRAFT（草稿）——A00 必須在功能分支 rebase 之前，替換掉所有預留位置（placeholder）。
合約提交（commit）：`<sha>`
Go module：`<module>`
Schema 基準版本：`<version>`

## Import 路徑

| 關注點 | 套件 |
|---|---|
| 領域實體（Domain entities） | `<path>` |
| 跨元件埠（ports） | `<path>` |
| 事件協定 | `<path>` |
| SQLite 執行期 | `<path>` |

## Schema 版本字串

```text
preflight.event.v1
preflight.progress-tree.v1
preflight.state-checkpoint.v1
preflight.repository-checkpoint.v1
preflight.pause.v1
preflight.api.v1
```

## ID 與冪等性規則

記錄每個實體 ID、操作／事件的冪等性金鑰（idempotency key），以及重播（replay）行為。

## 未知／null 語意

記錄未知用量、配額、上下文（context）、機率、供應商能力，以及重設時間戳記的處理方式。

## 交易邊界

記錄 `CompleteNode`、檢查點建立、授權消費（consumption）、暫停持久化，以及喚醒租約（wake lease）等交易。

## 錯誤合約

記錄穩定的錯誤代碼，以及 fail-open／fail-closed 分類。

## 隱私合約

原始提示詞、逐字稿（transcripts）、機密資料，以及儲存庫產出物的政策。

## 遷移範圍

- 0000–0009 A01
- 0010–0019 A02
- 0020–0029 A03
- 0030–0039 A04
- 0040–0049 A05
- 0050–0059 A06

## 凍結的狀態轉換

插入回合（turn）、進度節點、檢查點、暫停、喚醒工作（wake job），以及授權狀態資料表。
