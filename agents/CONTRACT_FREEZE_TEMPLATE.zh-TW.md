# Auspex Vertical-Slice Contract Freeze

> 🌐 [English](CONTRACT_FREEZE_TEMPLATE.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

狀態：DRAFT——在其他角色的分支 rebase 之前，contract-integrator 必須先替換掉每一個 placeholder。
Contract commit：`<sha>`
Go module：`<module>`
Schema baseline：`<version>`

## 匯入路徑

| 關注點 | Package |
|---|---|
| Domain entities | `<path>` |
| Cross-component ports | `<path>` |
| Event protocol | `<path>` |
| SQLite runtime | `<path>` |

## Schema 版本字串

```text
auspex.event.v1
auspex.progress-tree.v1
auspex.state-checkpoint.v1
auspex.repository-checkpoint.v1
auspex.pause.v1
auspex.api.v1
```

## ID 與 idempotency 規則

記錄每個實體的 ID、operation/event 的 idempotency key，以及 replay 行為。

## Unknown/null 語意

記錄 unknown usage、quota、context、probability、provider capability，
以及 reset 時間戳的處理方式。

## 交易邊界

記錄 `CompleteNode`、checkpoint 建立、authorization 消耗、pause persist，
以及 wake lease 等交易。

## 錯誤契約

記錄穩定的錯誤代碼，以及 fail-open/fail-closed 的分類方式。

## 隱私契約

原始 prompt、transcript、機密內容，以及 repository artifact 的政策。

## Migration 編號範圍

- 0000–0009 foundation
- 0010–0019 claude-provider
- 0020–0029 checkpoint（Part A — progress/state）
- 0030–0039 checkpoint（Part B — repository）
- 0040–0049 predictor
- 0050–0059 runtime（Part A — pause/scheduler）

## 凍結狀態轉換

插入 turn、progress node、checkpoint、pause、wake job，以及 authorization
狀態表。
