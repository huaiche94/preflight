# ADR-046 — 分層式 telemetry 保留機制：hot raw window → rollup → gzip archive → delete

> 🌐 [English](0046-tiered-telemetry-retention.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

狀態：已接受（Accepted）
日期：2026-07-13
負責人：lead（storage/CLI 相關介面），新的 migration 範圍分配則由 contract-integrator 負責
核准人：repository owner，2026-07-13（issue #19 決策會議）

## 背景

Auspex 的 SQLite 資料庫會無止盡地成長。每一次 hook 呼叫都會新增 `events` 資料列（每次 statusline snapshot 最多四筆）；每一個經過 evaluate 的 prompt 都會新增 `feature_vectors` + `predictions` + `policy_decisions`（若跑到 runway 路徑，還會加上 `runway_forecasts`）；每一次核發的 gate 都會新增 `authorizations`；每一次節點完成都會新增 `state_checkpoints`/`node_completions` 資料列，且對 repository checkpoint 而言，還會在 `<data-dir>/checkpoints/` 下新增一個磁碟上的產物目錄。這些資料從來沒有任何東西會刪除。一個 dogfooding 安裝環境（issue #17/#12）在每一個 prompt 都會持續累積這些資料，所以無止盡成長是一個「現在」就存在的問題，而不是「將來某天」的問題。

與此同時，M13 校準（issue #11）**需要**長期存活的 prediction-vs-actual 配對資料，而 Progress Tree/checkpoint 的不變式（Constitution §6）禁止刪除一個可續行任務仍然依賴的證據。因此，保留機制不能是一句粗暴的 `DELETE WHERE old`。

## 決策

採用**三層式保留機制**，由新增的 `internal/retention` 引擎執行，並以 `auspex gc` 對外呈現（schema 版本化輸出 `auspex.gc.v1`）：

1. **Hot raw window**——比保留視窗新的原始資料列不會被觸碰。預設視窗：**90 天**（`retention.Policy`、`DefaultRetentionDays`），可透過 `auspex gc --retention-days` 覆寫。所有類別共用同一個視窗；刻意不提供逐類別覆寫，直到出現真正的需求為止（不做臆測性抽象化）。
2. **Rollup 彙總表**（migration `0060_retention.sql`）——在原始資料列離開 hot tier 之前，值得永久保留的彙總值會在同一個 delete transaction 中被萃取進精簡的資料表：
   - `usage_rollups_daily`：依 (UTC 日期、provider、session、事件類型) 計數，並且只包含目前已持久化的 payload 誠實能提供的彙總值（詳細的逐欄位推導方式與其缺口，請見該 migration 檔頭的說明）。
   - `calibration_samples`：每一個過期的 prediction 對應一列，將其預測分位數與實際 turn 結果配對，**只在可推導時**才配對，否則就是 `actual_known = 0` 加上 NULL 的實際值——這是 ADD principle 1（「unknown 不等於 zero」）誠實 cold start 的展現。正是這個機制，讓 M13 校準所需的原始 prediction-vs-actual 配對資料，能夠跨越封存（archival）而保留下來。
3. **Gzip JSONL archive，然後才刪除——failure closed（失敗時關閉/保守）。** 過期的原始資料列會被寫入 `<data-dir>/archive/<table>/<YYYY-MM>/<table>-<UTC timestamp>-<runID>.jsonl.gz`，每一列對應一個 JSON 物件、保留完整欄位精確度，寫入方式與 `internal/repocheckpoint/atomicwrite.go` 相同，遵循 temp-file → fsync → rename 的紀律。接著這份 archive 會被**重新開啟並重新讀取**，其資料列數與 SHA-256 內容摘要會對照原先選取的內容進行驗證。只有在每一個類別的 archive 都通過驗證之後，delete 才會執行——所有類別在同一個 transaction 中一併執行，並將受影響的資料列數與選取集合核對。在該 transaction 之前發生的任何失敗，都會讓所有原始資料列維持不變，並記錄一筆失敗的 `retention_runs` 資料列。不存在「部分刪除」的狀態。

### 涵蓋的類別與其逐類別規則

| 類別 | 過期規則 | Rollup |
|---|---|---|
| `events` | `occurred_at` 早於視窗 | `usage_rollups_daily` |
| `feature_vectors` | `created_at` 早於視窗 | — |
| `predictions` + `policy_decisions` | prediction 的 `created_at` 早於視窗；綁定到已過期 prediction 的 decision 會隨之一併處理（反正本來就會 cascade——明確地將它們一併封存，可讓 archive 與 delete 集合保持一致）；孤兒 decision（`prediction_id IS NULL`）則依其自身的 `decided_at` 判定 | `calibration_samples` |
| `runway_forecasts` | `created_at` 早於視窗，**除非**該列仍被某筆存活的 `policy_decisions` 列參照（刪除這類列會讓一筆我們要保留的資料列被 `ON DELETE SET NULL` 變異） | — |
| `authorizations` | 僅限**同時**已被消費（`consumed_at IS NOT NULL`）**且**已過期（`expires_at`）超過視窗的資料列——未消費或尚未過期的 authorization 絕不會被 GC | — |
| checkpoints（`state_checkpoints`、`node_completions`、`repository_checkpoints` ＋磁碟上的產物目錄） | 僅限 `tasks.status` 為終態（`completed`/`failed`）、且 `completed_at` 早於視窗的任務 | — |

Checkpoint 的完整防護措施如下：

- **每個任務最新的 state checkpoint 與最新的 repository checkpoint 永遠會被保留**，無論年齡多大——這是可續行的安全錨點。被保留的 state checkpoint 所參照的 repository checkpoint（`repository_checkpoint_id`）也會一併保留，讓錨點永遠不會懸空。
- 一個 `completed_at IS NULL` 的終態任務會被**完全跳過**，並在執行摘要中被具名列出——此時完成時間的年齡無法被乾淨地推導出來，而 Constitution 的證據規則認為，用猜測代替保留是更糟的做法。
- 非終態任務無論多舊都不會被觸碰。
- 磁碟上的產物目錄（`repository_checkpoints.artifact_root`）只有在 delete transaction commit **之後**才會被移除（一個孤兒目錄是安全的；但一個存活資料列背後的目錄被刪除則不安全），且僅限於路徑解析後確實位於 Auspex 資料目錄之內——若根目錄在此範圍之外，則會原地保留並記錄在摘要中，而不是憑一廂情願地執行 `RemoveAll`。

### 空間回收

`internal/storage/sqlite/db.go` 的 pragma 初始化並未設定 `auto_vacuum`，因此每一個 Auspex 資料庫都以 SQLite 預設的 `auto_vacuum = NONE` 執行：delete 只會把分頁移到 freelist，檔案本身不會自動縮小。此引擎在執行期會實際讀取 `PRAGMA auto_vacuum`，而不是憑假設行事：若資料庫曾處於 `incremental` 模式，則在每次刪除動作之後會自動執行 `PRAGMA incremental_vacuum`；而對於目前實際上都是 `NONE` 模式的資料庫，則改為提供 `auspex gc --vacuum`，執行一次完整的 `VACUUM`（在獨佔鎖下重寫整個檔案——結果正確，但會短暫阻塞，因此設計為選擇性啟用）。`auspex.gc.v1` 的輸出會誠實地回報由 freelist 推算出的 `reclaimable_bytes_estimate`，而不是在位元組其實並未歸還給作業系統時，卻宣稱已經歸還。

### Dry run

`auspex gc --dry-run` **真正做到沒有任何副作用**：它會選取並回報每個資料表的計數與「原本會產生」的 rollup 結果，但不寫入任何 archive 檔案、不寫入任何 rollup 資料列，也不寫入任何 `retention_runs` 資料列。（另一個替代方案——記錄一筆 `dry_run=1` 的稽核資料列——曾被考慮過但遭到否決：一個名為「dry run」的模式若仍會寫入資料庫，正好會招致這項功能存在的目的所要防止的那種意外。）

### Migration 範圍 0060–0069 分配給 retention/gc

保留機制是跨切面（cross-cutting）的（它觸及每一個角色的資料表），且不屬於任何 vertical-slice 角色所擁有，因此它獲得自己專屬的範圍，而不是佔用別人的範圍。`CONTRACT_FREEZE.md` 的 migration 範圍表新增一列 `0060–0069 retention/gc`，並引用本 ADR。`0060_retention.sql` 是其第一個 migration。

### 校準的誠實性：目前的 payload 能與不能填入什麼

已對照 `internal/telemetry/claude/normalizer.go`（唯一產生持久化事件 payload 的來源）與 `internal/orchestrator/hooks.go`（唯一標記 `Event.TurnID` 的位置）驗證過：

- **可推導：** `actual_outcome`（來自帶有該 prediction `turn_id` 的 `provider.turn.*` 事件，值為 `completed`/`failed`/`interrupted`）、`actual_failure_class`（`provider.turn.failed` 上的 `failure_class` payload 欄位）、`actual_outcome_at`，以及 `session_id`（來自該 turn 自身的事件）。
- **目前無法推導，因此不設為欄位：** 每個 turn 實際的 token 使用量。沒有任何持久化的 payload 帶有這項資訊——`provider.turn.completed` 只帶有 `stop_hook_active`；statusline 的 `provider.usage.observed` 帶有 session 累計的 cost/duration/lines，而 `provider.context.observed` 帶有 context window 的填充量，這兩者都無法歸屬到單一 turn。新增一個捏造或歸屬錯誤的數字將違反 ADD principle 1；當某個 payload 真正取得逐 turn 的 token 資訊時，會在 0060–0069 範圍內新增一個 migration 來加入該欄位。
- **Cold-start 的現實：** 目前只有 `provider.turn.started` 事件會被標記 `turn_id`（Stop/StopFailure 的 hook payload 不帶有 turn 身分），因此真實的 Claude session 目前會產生 `actual_known = 0` 的樣本。這就是誠實的答案，並如實記錄下來——join 的方式是依 `turn_id`，一旦結果事件獲得 turn correlation（issue #1 的工作項目），樣本就會自動升級。

## 已考慮的替代方案

- **僅用 TTL（`DELETE WHERE old`）**——已否決：會摧毀 M13 校準（#11）所需要的原始 prediction-vs-actual 配對資料，也會摧毀 ADR-043 的 budget forecast 所需要的使用歷史，且沒有任何復原路徑。
- **僅用 Rollup（先彙總再刪除、不做 archive）**——已否決：rollup schema 本質上是一種賭注，賭的是哪些彙總值才重要；gzip JSONL archive 則是讓「賭錯」也能夠復原（可從 archive 重新推導）而非造成永久資料遺失的避險機制。
- **容量上限式環狀緩衝區（ring buffer）**——已否決：依容量而非時間來淘汰資料，會讓一個話多的 session 把另一個 session 仍在 hot 狀態的資料列擠掉；而且它與「僅用 TTL」一樣，具有摧毀校準資料的問題。

## 影響

- 資料庫不再無止盡地成長；長期訊號（每日使用量、校準配對資料）存活於精簡的資料表中，其餘的一切則以 `.jsonl.gz` 的形式存活於資料目錄下，使用者可自行決定是否刪除——Auspex 本身絕不會刪除 archive。
- `retention_runs` 為每一次執行都提供一筆持久化的稽核資料列（無論成功或失敗），與 Constitution §6 的證據紀律一致。
- 完整的 `VACUUM` 仍為選擇性啟用；若要將資料庫切換為 `auto_vacuum = incremental`，需要從空白重建，或是一項明確的 migration 決策，這已超出本文件範圍。
- 從 Auspex 的角度來看，archive 目錄是只增不減（append-only）的；archive 的裁剪／壓縮若日後有需要，屬於未來的決策。
