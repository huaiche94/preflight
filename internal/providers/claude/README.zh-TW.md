# internal/providers/claude/ — Claude Code status-line payload 解析器

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

解析 Claude Code 原生的 status-line JSON 快照（ADD §22.5；`Auspex_ADD.md`
現位於 [docs/design/Auspex_ADD.md](../../../docs/design/Auspex_ADD.md)）並轉為
中介 Go struct。此套件沒有 doc.go；statusline.go 檔案開頭的 package comment
即說明其契約：本套件僅止於解析步驟，正規化為凍結的
`pkg/protocol/v1.Event` envelope 是
[../../telemetry/claude/](../../telemetry/claude/) 的職責。

進入點：`ParseStatusLine(raw []byte) (StatusLineSnapshot, error)`。此快照
透過 `ContextObservation`（context window 使用量）、`QuotaObservations`
（每個 rate-limit window 對應一個 `domain.QuotaObservation`，數量依實際
回傳而定——issue #21），以及 `WeeklyLimitUsedPercent`（status line 的週用量
區段）投影至凍結的 domain 結構。

解析原則：

- 每個可選欄位皆為 pointer；`nil` 代表未知，絕不以替代零值表示
  （ADD §22.10；CONTRACT_FREEZE.md 的「Unknown/null semantics」）；
- 任何巢狀層級中的未知欄位皆會被容忍；
- 已知欄位若出現無法辨識的編碼方式，該欄位會降級為未知，而不會讓整個快照
  解析失敗（`flexTimestamp`；issue #27 事件）；
- 語法無效的 payload 會回傳附帶 `ErrCodeValidation` 的 `domain.Error`，
  讓 hook wrapper 得以退回安全回應，而不是直接崩潰。

由 internal/orchestrator/hooks.go 在 `auspex hook claude statusline` 之下
呼叫使用。相鄰的 hook payload 解析器（UserPromptSubmit/Stop/StopFailure）
位於 [../../hooks/claude/](../../hooks/claude/)。
