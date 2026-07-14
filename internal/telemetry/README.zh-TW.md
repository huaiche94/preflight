# internal/telemetry/ — provider telemetry 正規化與持久性事件儲存

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

本目錄為各個 per-provider 的 telemetry pipeline 命名空間，這些 pipeline 會把原始的 provider
payload 轉換成 Auspex 凍結的傳輸事件封裝 `pkg/protocol/v1.Event`（Auspex_ADD.md
§11.1；該 ADD 位於 [docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)），並以冪等（idempotent）方式持久化。這一層沒有任何 Go 程式碼。

它唯一的子套件是 [`claude/`](claude/README.md)，依其自身的套件契約，它是從 Claude Code
provider payload 轉換成 `v1.Event` 的**唯一**路徑 — repository 中沒有其他套件會從 Claude
payload 建構出 `v1.Event`。

與相鄰套件的分工：

- [`../providers/claude`](../providers/) 與 [`../hooks/claude`](../hooks/)
  負責把原始的 provider JSON 解析成中繼、隱私安全的 Go struct（僅限解析這一步）；
- `telemetry/claude` 會把這些 struct 正規化成 `v1.Event` 值（包含建構 idempotency-key），並透過
  [`../storage/sqlite`](../storage/sqlite/README.md) 的 `WithTx` 邊界寫入 SQLite；
- 消費者讀取的是最終產生的 `events` 資料列（migration `0010_events.sql`）— 它們絕不會重新解析
  provider payload。
