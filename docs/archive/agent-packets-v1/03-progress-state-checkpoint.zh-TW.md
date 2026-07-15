# A03 — 進度樹與狀態檢查點

> 🌐 [English](03-progress-state-checkpoint.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。本文件屬歷史封存，非現行實作指引。

## 模型

使用 Fable，因為這是定義產品完整性的邊界。

## ADD 負責範圍

§18、Appendix A／B、§29.5 中的 State Checkpoint 情境，並以 ADR-027 至 ADR-030 及 ADR-039 作為限制條件。

## 專屬路徑

```text
internal/progress/**
internal/statecheckpoint/**
internal/artifacts/**
schemas/progress-tree.schema.json
schemas/state-checkpoint.schema.json
testdata/progress-trees/**
testdata/checkpoints/state/**
internal/storage/sqlite/migrations/0020-0029_*.sql
docs/implementation/day1/A03.md
```

## 任務

讓 Progress Tree 成為權威、持久的任務狀態，並強制規定：節點若沒有經過驗證的產出物證據，就不能變為完成狀態。

## 產出物

1. Task／node／edge／artifact 儲存區。
2. 具有明確合法轉換的節點狀態機。
3. 產出物驗證器：
   - 檔案存在；
   - 檢查碼（checksum）相符；
   - Markdown 標題存在；
   - Markdown 程式碼區塊（code fences）成對平衡；
   - 選用的自訂驗證器介面。
4. `CompleteNode` 原子協定：
   - 暫存／驗證產出物證據；
   - 更新節點；
   - 建立 State Checkpoint；
   - 在適用情況下於單一資料庫交易中提交；
   - 提交後發布正規化事件。
5. State Checkpoint 清單（manifest）序列化與檢查碼。
6. 針對暫存產出物與資料庫之間因當機（crash）造成落差的啟動時協調（reconciliation）。
7. 完成用的冪等性金鑰，以及重複供應商事件的處理。
8. Snapshot／load-latest／verify API。

## 必須拒絕

- 沒有產出物、僅憑「代理人宣稱完成」；
- 產出物缺失或已變更；
- 違反相依政策卻已完成的子節點；
- 證據互相矛盾的重複完成；
- 無效的狀態轉換；
- 檢查點清單參照到尚未提交的資料列。

## 必要測試

- 有效的 Markdown 區段完成並建立檢查點；
- 缺少標題或程式碼區塊不平衡時遭拒；
- 在完成流程各階段注入當機並進行協調；
- 100 個依序節點產生 100 個可驗證的檢查點；
- 相同的冪等性金鑰回傳相同結果；
- 互相矛盾的冪等性 payload 遭拒；
- 並行完成的競爭情況（race）。

## 邊界

儲存庫檔案／diff 擷取屬於 A04 的範疇。A03 透過凍結的 ports 儲存參照，並在測試中假造（fake）A04。
