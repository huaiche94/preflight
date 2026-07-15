# internal/hooks/claude/ — Claude Code hook payload 解析與回應編碼

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

解析 Claude Code 原生的 lifecycle-hook stdin payload，並編碼與 provider
相容的 stdout 回應。此套件沒有 doc.go；userpromptsubmit.go 檔案開頭的
package comment 即說明其契約。

Auspex 處理四種 Claude Code hook 事件，以 `auspex hook claude ...` 子指令
的形式對外提供（internal/cli/hook.go）：

- `user-prompt-submit` — `ParseUserPromptSubmit` → `UserPromptSubmitEvent`。
  從設計上即具備隱私安全性：解析呼叫內即把原始 prompt 化約為 SHA-256
  雜湊值、大小訊號，以及衍生出的 `features.PromptFeatures`；原始文字絕不會
  存活超出該堆疊框（stack frame）（Constitution §7 第 2 條）。
  `EncodeUserPromptSubmitResponse` 會產生 allow/block 回應（ADD §22.3）；
  `FallbackAllowResponse` 則是 Auspex 本身發生失敗時所使用的 fail-open
  回應本文。
- `stop` — `ParseStop` → `StopEvent`（一次乾淨的 turn/session 結束）。
- `stop-failure` — `ParseStopFailure` → `StopFailureEvent`，將 provider
  錯誤分類至凍結的 `domain.FailureClass` enum（此對應關係是 Auspex 的
  判斷法則，並非凍結契約）。
- `statusline` — 由相鄰套件
  [../../providers/claude/](../../providers/claude/) 解析，而非本套件。

Session bootstrap（issue #17）：在事件被持久化或評估執行之前，hook 處理器
會依 payload 回報的目錄，冪等（idempotently）地註冊該 session 的
repositories/worktrees/provider_sessions 資料列
（internal/orchestrator/sessionbootstrap.go——以既有的凍結唯一鍵約束執行
upsert；此步驟為 fail-open，因此非 git 目錄或 SQL 錯誤都只會靜默地無動作，
絕不會造成 hook 失敗）。

本套件僅負責解析與編碼；持久化、評估、以及 forecast 卡片屬於
internal/orchestrator/hooks.go 這一層的職責。（ADD 引用皆指
`Auspex_ADD.md`，現位於
[docs/design/Auspex_ADD.md](../../../docs/design/Auspex_ADD.md)。）
