> **ARCHIVED — 已被取代。** 這個編號九角色（`A00`–`A08`）結構，已被 `agents/` 下語意化命名的七角色結構取代（進度／狀態檢查點與儲存庫檢查點合併為 `checkpoint.md`；暫停／排程器與 CLI／API 編排合併為 `runtime.md`）。僅保留作為歷史參考；請勿再將這些檔案交給新的代理人——請改用 `agents/`。

# 代理人封包（Agent Packets）

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。本文件屬歷史封存，非現行實作指引。

這個目錄存放**每個 Day-1 代理人（`A00`–`A08`）各一份的交接封包**。每個檔案都是設計成單獨交給某個隔離的代理人／worktree，讓對方不需要在情境（context）中放入完整的架構文件或完整的執行計畫。

依照 `Preflight_Day1_Parallel_Execution_Plan.md` §3，一個 worker 應該只收到：

1. `Preflight_Day1_Parallel_Execution_Plan.md`（共同計畫）；
2. `docs/implementation/day1/CONTRACT_FREEZE.md`（在 A00 產出之後）；
3. 其被指派的 `Preflight_ADD.md` 章節；
4. 這個目錄中屬於它的封包檔案。

## 檔案清單

| 檔案 | 代理人 |
|---|---|
| `00-contract-integrator.md` | A00 — 合約凍結與整合 |
| `01-foundation-config-sqlite.md` | A01 — 基礎建設、設定、路徑、核心 SQLite |
| `02-claude-telemetry-hooks.md` | A02 — Claude 遙測、Hooks、供應商正規化 |
| `03-progress-state-checkpoint.md` | A03 — 進度樹與狀態檢查點 |
| `04-repository-checkpoint.md` | A04 — Git 觀察與儲存庫檢查點 |
| `05-predictor-policy.md` | A05 — 範圍估算器、預測器、風險、政策、授權 |
| `06-graceful-pause-scheduler.md` | A06 — 優雅暫停、安全點、持久排程器 |
| `07-runtime-cli-api.md` | A07 — 應用程式編排、CLI、本機 API |
| `08-qa-security-ci.md` | A08 — 跨元件 QA、安全性、可靠性、CI |
| `CONTRACT_FREEZE_TEMPLATE.md` | A00 的 `docs/implementation/day1/CONTRACT_FREEZE.md` 產出物範本 |

## 真實來源

這些封包檔案是每個代理人任務、所有權路徑、產出物與必要測試的**權威、可編輯版本**。Day-1 計畫中的「Agent Packets」章節只是一份摘要索引，會連結回這裡——如果你要變更某個封包的範圍，請在這裡編輯，而不是在計畫文件中編輯。

若要了解整體情境（願景、範圍邊界、拓撲、合併順序、刪減清單），請參閱位於儲存庫根目錄的 `Preflight_Day1_Parallel_Execution_Plan.md`。若要了解架構，請參閱 `Preflight_ADD.md`。
