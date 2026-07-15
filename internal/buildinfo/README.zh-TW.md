# internal/buildinfo/ — `auspex version` 背後的 build/version 中繼資料

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`buildinfo` 套件保存支撐 `auspex version` 指令所需、最精簡的 build/version 中繼資料。套件契約就是
`buildinfo.go` 檔案開頭的套件註解（沒有另外的 `doc.go`）。

- `Version` — 目前是固定的開發用佔位值 `"0.0.0-dev"`；依照 agents/foundation.md，串接真正由
  ldflags 注入的值（release tag、commit SHA、build date）明確不在 foundation-01 的範圍內，仍待正式的發行封裝完成。
- `String()` — 該指令所輸出、人類可讀的版本字串。

使用端：[`../cli`](../cli/) 的 `auspex version` 指令會輸出 `String()`，而
`cmd/auspex/wire.go` 會把 `Version` 傳入 [`../daemon`](../daemon/) 的設定與
[`../httpapi`](../httpapi/) 的 version endpoint。
