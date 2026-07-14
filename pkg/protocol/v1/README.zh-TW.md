# pkg/protocol/v1/ — 凍結的公開 wire protocol（`auspex.*.v1`）

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

套件 `v1` 是凍結的公開 wire protocol（其自身的 package comment 亦如此
說明）。它是一項相容性承諾，而不僅僅是實作細節：其基準線記錄於
[`CONTRACT_FREEZE.md`](../../../docs/implementation/vertical-slice/CONTRACT_FREEZE.md)，
此處任何內容的變更都需要一份 ADR（Constitution §3）。

## 檔案

- `event.go` —
  - 六個 `SchemaVersion*` 常數（`auspex.event.v1`、
    `auspex.progress-tree.v1`、`auspex.state-checkpoint.v1`、
    `auspex.repository-checkpoint.v1`、`auspex.pause.v1`、
    `auspex.api.v1`）；
  - `EventType`，一個封閉、具版本號的分類體系（taxonomy）
    （[`Auspex_ADD.md`](../../../docs/design/Auspex_ADD.md) §11.3）——
    新的事件型別一律經由 contract-integrator 加入，絕不從功能程式碼中
    臨時（ad hoc）冒出字串；
  - `Event`，每個 provider payload 在抵達 domain/storage 程式碼之前，
    都會被轉換成這個正規化 envelope（ADD §11.1）。
- `event_test.go` — 逐位元組（byte-for-byte）釘住 schema 版本字串。

## 契約重點

- provider 的 wire payload 絕不能在未經正規化的情況下滲入這些型別中，
  `Event.Payload` 只有在完成過濾（redaction）之後才會被填入
  （Constitution §7）。這裡絕不會有原始 prompt 文字的欄位。
- `Event.IdempotencyKey` 對每個 provider 事件身分而言是確定性
  （deterministic）的（若存在穩定的 provider ID 就使用它，否則使用內容
  摘要）。
- `nil`／欄位缺席代表未知，絕不是以替代零值表示。

checkpoint／progress wire shape 的 JSON Schema 鏡像版本位於
[`../../../schemas/`](../../../schemas/README.md)。
