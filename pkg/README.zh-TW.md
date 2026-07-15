# pkg/ — 可對外公開匯入的 Go 套件

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

在 `github.com/huaiche94/auspex` module 中，唯一允許外部 Go 程式碼 import
的部分。其餘皆位於 [`../internal/`](../internal/) 之下，並由編譯器強制
設為私有。

內容：

- [`protocol/`](protocol/README.md) — 具版本號的公開 wire protocol；目前
  只有 [`protocol/v1/`](protocol/v1/README.md)，也就是凍結的
  `auspex.*.v1` 契約。

刻意讓這棵樹保持精簡：放在這裡的型別即代表一項公開的相容性承諾（見
[`CONTRACT_FREEZE.md`](../docs/implementation/vertical-slice/CONTRACT_FREEZE.md)），
因此各種形狀（shape）會持續留在 `internal/` 中，直到真的必須對外共用為止。
