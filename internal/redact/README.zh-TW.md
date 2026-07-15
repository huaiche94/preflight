# internal/redact/ — 針對 checkpoint artifact 的機密內容偵測

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

在工作樹內容進入具持久性的 checkpoint artifact 之前，先套用以樣式（pattern）為基礎的機密
內容偵測（Auspex_ADD.md §19.5「secret scan」與 §27.8 偵測器清單——ADD 現位於
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)）。套件合約請見 `doc.go`，其中
明確說明了這些偵測器能抓到什麼、抓不到什麼。

兩類偵測器：

- **檔名樣式**（`filename.go`、`MatchesSecretFilename`）—— §27.8 的名單（`.env`、`.env.*`、
  `*.pem`、`*.key`、`*.pfx`、`*.p12`、`id_rsa`、`id_ed25519`、`credentials.json`、`auth.json`、
  `secrets.*`），在讀取任何內容之前，先比對檔案的基本檔名（base filename）。
- **內容偵測器**（`patterns.go`、`Detectors`）—— 針對 bearer token、PEM 私鑰標頭、
  GitHub／OpenAI／Anthropic API key 的格式、Azure 儲存體帳戶金鑰、JWT 格式的權杖，以及
  密碼／連線字串樣式的固定正規表示式。`ScanPath`／`ScanContent`（`scan.go`）會回傳
  `Finding`；每個檔案最多掃描 `MaxContentScanBytes`（1 MiB）的內容，且疑似為二進位的內容會
  被略過（`IsLikelyBinary`）。

使用者，兩者皆位於 [`../repocheckpoint/`](../repocheckpoint/)：untracked 歸檔路徑會略過命中
的檔案並記錄一筆略過帳本項目（`archive.go`）；patch 路徑則會對 staged／unstaged patch 中
`+`／`-` 行內容裡偵測到的區段做原地遮蔽（`patchredact.go`）。

可接受的殘留風險：[ADR-042](../../docs/adr/0042-patch-redaction-residual-surface.md) 刻意將
patch 檔案路徑標頭行與二進位 diff 標頭／內容行排除在遮蔽範圍之外——改寫這些內容會破壞
`git apply`，並摧毀 checkpoint 的證據價值；此邊界由
`../repocheckpoint/patchredact_internal_test.go` 中的測試釘死。

這是一組有文件記載、依格式特徵（shape-based）的偵測器，而非一個窮舉式（exhaustive）
掃描器——它是縱深防禦（defense-in-depth）的其中一層，與 QA 團隊獨立的洩漏掃描互相搭配
（`doc.go`）。
