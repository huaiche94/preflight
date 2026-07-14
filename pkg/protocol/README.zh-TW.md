# pkg/protocol/ — 具版本號的公開 wire protocol

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

每個協定版本都有自己的子目錄，讓使用端可以透過確切的 import path 釘選
（pin）特定契約。目前恰好只有一個：

- [`v1/`](v1/README.md) — 凍結的 `auspex.*.v1` 事件 envelope、事件型別
  分類（taxonomy），以及 schema 版本常數。

破壞性變更絕不會直接修改既有的 `v1/`——而是會在其旁邊新增一個 `v2/`
套件，並經由 ADR 核准（Constitution §3；「凍結」具體承諾了什麼，見
[`CONTRACT_FREEZE.md`](../../docs/implementation/vertical-slice/CONTRACT_FREEZE.md)）。
同一契約中、非 Go 語言 wire shape 的 JSON Schema 文件，則位於
[`../../schemas/`](../../schemas/README.md)。
