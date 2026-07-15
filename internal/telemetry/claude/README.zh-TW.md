# internal/telemetry/claude/ — 從 Claude Code payload 轉換成凍結 v1.Event envelope 的唯一路徑，並負責冪等持久化

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`claude` 套件會把 [`../../providers/claude`](../../providers/) 與
[`../../hooks/claude`](../../hooks/) 產生的中繼 struct（StatusLineSnapshot、
UserPromptSubmitEvent、StopEvent、StopFailureEvent）正規化成凍結的
`pkg/protocol/v1.Event` envelope（Auspex_ADD.md §11.1、CONTRACT_FREEZE.md；該 ADD
位於 [docs/design/Auspex_ADD.md](../../../docs/design/Auspex_ADD.md)）。套件契約就是
`normalizer.go` 檔案開頭的套件註解（沒有另外的 `doc.go`）：這是 repository
中唯一會從 Claude payload 建構 `v1.Event` 的套件，而且它只會送出 `pkg/protocol/v1`
封閉分類中已定義的 `EventType` 值。

分為兩部分：

- **`normalizer.go` / `managedrun.go` — 正規化。** `Normalizer`（`NewNormalizer(clock,
  ids)`）提供 `NormalizeStatusLine`、`NormalizeUserPromptSubmit`、`NormalizeStop`、
  `NormalizeStopFailure`，以及 `NormalizeManagedRun`。這個套件擁有確切的 idempotency-key
  摘要演算法（CONTRACT_FREEZE.md 凍結了這個欄位；由擁有該欄位的 provider
  角色定義摘要方式）。`managedrun.go` 負責處理 `auspex run` 的 terminal outcome
  （[`../../managed`](../../managed/) 會把 stream-json 的每一行解析成
  `ManagedRunOutcome` 再交給這裡處理）：產生一個 terminal turn 事件，此外若 provider
  的結果行帶有 usage 資訊，還會再產生一個以該 turn 為範圍的
  `provider.usage.observed` 事件 — 與累計性質的 statusline snapshot 不同，這是精確的單一 turn 歸因。
- **`store.go` — 持久化。** `EventStore`（`NewEventStore(db)`）會以
  `Event.IdempotencyKey` 為鍵，將事件持久且冪等地寫入 SQLite，一律透過
  [`../../storage/sqlite`](../../storage/sqlite/README.md) 的 `WithTx` /
  `app.TxRunner` 邊界（`Persist`、`PersistAll`）。

隱私：原始 prompt 文字絕不會傳到這個套件 — 上游的 parser 只會保留 hash 與衍生訊號，
`Event.Payload` 是依照 provider 角色自身的契約，在完成 redaction 之後才填入的。
