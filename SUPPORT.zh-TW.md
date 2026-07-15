# 支援（Support）

> 🌐 [English](SUPPORT.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

Auspex 是一個早期階段的開源專案。支援是社群基礎、盡力而為（best-effort）性質的。

## 該去哪裡詢問

- **錯誤（bug）與功能請求（feature request）**——開一個
  [GitHub issue](https://github.com/huaiche94/auspex/issues)。請附上
  `auspex version` 的輸出、你的作業系統，以及（針對執行期問題）
  `auspex doctor --json` 的輸出，並移除其中任何敏感資訊。
- **關於架構或貢獻的問題**——請先讀
  [`README.md`](README.md)、[`CONSTITUTION.md`](CONSTITUTION.md)，以及
  [`docs/design/Auspex_ADD.md`](docs/design/Auspex_ADD.md)；如果答案不在其中，再開一個 issue。
- **安全性漏洞**——請**不要**開公開的 issue。請依照
  [`SECURITY.md`](SECURITY.md) 的流程處理。

## 可以預期什麼

- Issue 會依維護者的時間許可進行分流（triage）；沒有 SLA 保證。
- Auspex 將所有狀態都儲存在本機（位於你作業系統使用者資料目錄下的
  SQLite）。依設計，support bundle 與診斷資訊絕不會包含原始 prompt——在
  issue 中貼上日誌時，請維持這個原則。
