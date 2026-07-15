# internal/clock/ — domain.Clock 的真實 wall-clock 實作

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`clock` 套件提供 [`../domain`](../domain/) 的 `Clock` 介面之真實 wall-clock 實作。套件契約就是
`clock.go` 檔案開頭的套件註解（沒有另外的 `doc.go`）。

- `System` — 以 `time.Now()` 為基礎的無狀態實作；可安全並行使用。
- `New() domain.Clock` — production wiring 所使用的建構函式。

Production 程式碼依賴的是 `domain.Clock`，絕不會直接依賴這個套件本身，這樣測試才能替換成 fake 並維持決定性
— 例如 [`../retention`](../retention/README.md) 與
[`../telemetry/claude`](../telemetry/claude/README.md) 都是接收 `domain.Clock`，而不是直接呼叫
`time.Now()`。
