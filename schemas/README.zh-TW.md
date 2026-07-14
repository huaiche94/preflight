# schemas/ — 凍結 wire shape 的 JSON Schema

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

針對以 JSON 檔案序列化的 `auspex.*.v1` wire shape（checkpoint manifest、
progress-tree 匯出檔）而制定的、機器可驗證的 JSON Schema 文件。每份 schema
自身的 `description` 欄位皆說明其鏡像對象；在此彙總如下：

| Schema | 釘選版本 | 鏡像對象 |
|---|---|---|
| `progress-tree.schema.json` | `auspex.progress-tree.v1` | `internal/progress` 的 Node/Edge Go 型別，以及 `progress_nodes`／`progress_edges` 資料表（migration 0020–0021）。形狀取自 `Auspex_ADD.md` 附錄 A、§18。 |
| `state-checkpoint.schema.json` | `auspex.state-checkpoint.v1` | `internal/statecheckpoint.Manifest` 與 `state_checkpoints` 資料表（migration 0023）。形狀取自 `Auspex_ADD.md` §18.8、附錄 B。在信任已儲存的 manifest 之前，必須獨立重新計算 `integrity_sha256`。 |
| `repository-checkpoint.schema.json` | `auspex.repository-checkpoint.v1` | `internal/repocheckpoint.Manifest`（某次 checkpoint 的 `manifest.json`）。形狀取自 `Auspex_ADD.md` §19、附錄 D。 |

schema 版本字串定義為 [`../pkg/protocol/v1/`](../pkg/protocol/v1/README.md)
中的 Go 常數，並由
[`CONTRACT_FREEZE.md`](../docs/implementation/vertical-slice/CONTRACT_FREEZE.md)
凍結。schema `description` 字串內對 `Auspex_ADD.md` 的引用，依 ADR-049
§Decision 3 保持逐字不變；該文件本身位於
[`../docs/design/Auspex_ADD.md`](../docs/design/Auspex_ADD.md)。

範例實例位於 [`../testdata/`](../testdata/README.md)
（`progress-trees/sample-task.json`、兩份 `sample-manifest.json`
fixture），並在產生時已完成 schema 驗證（記錄於
[`../docs/implementation/vertical-slice/checkpoint.md`](../docs/implementation/vertical-slice/checkpoint.md)）。
CI 目前尚無 JSON Schema 驗證工作——依
[`../.github/workflows/ci.yml`](../.github/workflows/ci.yml) 開頭的註解，
此事是刻意延後處理的。
