# Auspex 貢獻者指引（Contributor instructions）

> 🌐 [English](AGENTS.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

Auspex 是為 AI 編碼代理（AI coding agents）打造的本地優先（local-first）預測式執行期守門系統（predictive runtime guard）。

## 事實來源（Source of truth）

先讀 `CONSTITUTION.md`——它是本 repository 至高的流程權威（supreme process authority）（單一事實來源層級、ADR 規則、路徑所有權、Progress Tree 不變量）。在進行任何架構相關工作之前，請先讀
`docs/design/Auspex_ADD.md`。`docs/adr/` 下已被接受的 ADR 可以修訂它。如果你被指派了某個角色（role），也請讀 `agents/` 下對應的檔案。

## 範疇紀律（Scope discipline）

一次只實作一個路線圖里程碑（roadmap milestone）。在對應里程碑到來之前，不要引入雲端服務、不透明的 ML 相依套件，或未來 provider 的抽象層。

## 必要原則（Required principles）

- Go 是正式環境（production）的執行期語言。
- TypeScript 僅侷限於 VS Code 延伸模組。
- Python 僅用於離線研究。
- 原始 prompt 與工具輸出預設不會被持久化。
- Provider 的能力落差（capability gaps）必須明確標示。
- 不得在穩定路徑（stable paths）中解析未經文件化的逐字稿（transcripts）。
- 不得使用 shell 指令字串來執行 Git／provider 相關操作。
- Repository checkpoint 必須是原子性的（atomic）。
- Progress Tree 是具規範性的任務狀態（canonical task state）。
- 一個節點要標記完成，需要持久性的產物／證據（durable artifact/evidence）以及 State Checkpointing。
- 長篇文件要一段一段持久化寫入。
- 未經校準的風險分數（uncalibrated risk scores）不是機率（probabilities）。
- Graceful Pause（優雅暫停）僅在受管模式（managed mode）下才有完整保證。
- Auto-resume（自動恢復）是選擇性加入（opt-in）的，且需要 repository／配額／session 驗證。

## 動手修改前

1. 確認目前所在的里程碑。
2. 檢視既有的程式碼／測試。
3. 說明與 ADD 之間的衝突之處。
4. 建立一份持久性的實作進度產物（implementation progress artifact）。
5. 產出一份聚焦的計畫。

## 工作進行中

每完成一個邏輯上的工作單元後：

1. 把原始碼／文件／測試寫入實體檔案；
2. 執行本機驗證；
3. 更新進度狀態與下一步動作；
4. 不要把已完成的工作只留在對話上下文（conversation context）中。

## 收尾前

執行相關指令：

- `gofmt`
- `go vet ./...`
- `go test ./...`
- `go build ./cmd/auspex`
- 該里程碑專屬的驗收檢查

回報哪些測試沒有執行。
