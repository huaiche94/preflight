# internal/hooks/ — provider lifecycle-hook payload 處理

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

依 provider 分開的套件，負責解析原生 lifecycle-hook 的 stdin payload，並編碼
這些 hook 所預期、與 provider 相容的 stdout 回應。[claude/](claude/) 是目前
唯一的子套件；此目錄本身沒有任何 Go 檔案，也沒有 doc.go。

此目錄周邊的 hook 進入路徑：

1. provider 呼叫 `auspex hook claude <event>`（指令樹定義於
   internal/cli/hook.go），該指令會從 stdin 讀入完整的原始 payload，且絕不
   記錄或回顯（echo）其內容。
2. 該指令會呼叫對應的 `orchestrator.Handle*` 函式
   （internal/orchestrator/hooks.go），透過本目錄下的套件進行解析
   （status-line 解析則位於 [../providers/claude/](../providers/claude/)），
   經由 [../telemetry/claude/](../telemetry/claude/) 正規化為
   `pkg/protocol/v1.Event`，視情況將其持久化，並在 hook 語意需要做出決策時
   執行評估（evaluation）。
3. 該指令會將語法上有效、且與 provider 相容的 JSON 回應寫入 stdout，並在
   除了真正的指令使用錯誤（例如 stdin 無法讀取）以外的所有情況下，以結束碼
   0 結束。

Hook 採 fail open：Auspex 本身的失敗絕不會阻擋 provider 的 session。payload
格式錯誤或內部錯誤時，會產生安全的後備回應（例如
`claude.FallbackAllowResponse()`），絕不會造成崩潰或缺漏回應；語意上的
「封鎖（block）」決策內容只會出現在回應本文中，絕不會以非零的行程結束碼
表示——如此一來，provider 的 hook runner 就不會把一般的封鎖誤判為 Auspex
崩潰（處理器契約詳見 internal/orchestrator/hooks.go；ADD §17.5 的
telemetry-unavailable 規則——`Auspex_ADD.md`，現位於
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)）。
