# A06 — 優雅暫停、安全點與持久排程器

> 🌐 [English](06-graceful-pause-scheduler.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。本文件屬歷史封存，非現行實作指引。

## 模型

使用 Fable。

## ADD 負責範圍

§20、§§15／17／28／29 中暫停相關的部分、Appendix C、ADR-031 至 ADR-040。

## 專屬路徑

```text
internal/pause/**
internal/scheduler/**
schemas/pause.schema.json
testdata/pause-scenarios/**
internal/storage/sqlite/migrations/0050-0059_*.sql
docs/implementation/day1/A06.md
```

## 任務

實作與供應商無關的暫停／接續狀態機，以及持久化的喚醒排程。僅依賴凍結的 ports 來存取預測器、progress／state checkpoint、repository checkpoint、供應商中斷／接續、配額讀取、時鐘與租約（leases）。

## 必要狀態路徑

```text
observing
→ pause_requested
→ quiescing
→ safe_point_reached
→ persisting
→ interrupting
→ sleeping
→ wake_due
→ validating
→ resuming
→ resumed
```

納入 ADD 中定義的終止／衝突／取消／失敗狀態。

## P0 產出物

1. 狀態轉換驗證器。
2. 具去抖動（debounce）／遲滯（hysteresis）狀態的 `Observe` 處理。
3. `RequestPause` 的冪等性。
4. 針對回合／區段邊界觀測值的安全點（safe-point）協調器介面與實作。
5. 持久化階段的編排：
   - Progress Tree 快照；
   - State Checkpoint；
   - Repository Checkpoint；
   - Pause Record（暫停紀錄）；
   - Wake Job（喚醒工作）。
6. 具備 claim／renew／complete／fail／retry 的持久排程器租約。
7. 重啟後復原逾期／已租用的工作。
8. 接續驗證：
   - 配額安全；
   - 儲存庫指紋相容；
   - session／供應商能力有效；
   - 授權／同意有效。
9. 重複喚醒的恰好一次（exactly-once）行為。
10. 取消可阻止未來的接續。
11. 供應商中斷器／接續器的假合約測試。

## Day-one 現實考量

由於資料不足，已校準的自動暫停可能無法使用。需同時支援：

- 已校準的觸發條件：連續觀測值 `P_hit_10m >= threshold`；
- 具備不同原因代碼的明確未校準緊急政策。

先實作持久化喚醒與假接續器（fake resumer）。實際受管理的 Claude 接續屬於加分項目，且不得削弱狀態機測試的嚴謹度。

## 必要測試

- 兩次符合條件的觀測值會觸發請求；
- 單一尖峰不會觸發；
- 安全點會在中斷前持久化檢查點；
- 每個階段後發生當機都能正確接續／協調；
- 重啟能復原喚醒工作；
- 配額不安全時重新排程；
- 儲存庫重疊時阻擋；
- 不相關的儲存庫變更依設定政策處理；
- 重複的 worker 只產生一次接續；
- 過期租約被回收；
- 取消在與喚醒的競爭中勝出；
- 供應商中斷失敗會留下可復原的狀態。

## 邊界

不要解析 Claude 事件，也不要實作檢查點內部邏輯。請使用 ports／fakes。
