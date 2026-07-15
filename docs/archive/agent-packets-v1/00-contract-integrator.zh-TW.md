# A00 — 合約凍結與整合代理人

> 🌐 [English](00-contract-integrator.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。本文件屬歷史封存，非現行實作指引。

## 模型

使用 Fable。

## ADD 負責範圍

主要負責：§§1–9、31–34。跨領域審查 §§10–30。負責已核准 ADR 的變更。

## 任務

凍結編譯期與持久化合約，讓所有功能代理人能夠獨立作業。在最後階段整合已審查的分支。不要實作屬於其他代理人的功能內部邏輯。

## 專屬路徑

```text
internal/domain/**
internal/app/ports.go
pkg/protocol/v1/**
docs/adr/**
docs/implementation/day1/CONTRACT_FREEZE.md
Preflight_ADD.md (only when a genuine contradiction requires an ADR)
AGENTS.md
```

## 第一項產出物：合約提交（commit）

為以下項目建立可編譯的定義：

- 識別碼與 schema 版本；
- session／turn／task／progress／checkpoint／pause／evaluation 實體；
- 量測來源（provenance）與未知值；
- 狀態與失敗類別；
- ADD §9.9 中的服務埠（ports）；
- ADD §9.10 中的供應商介面；
- 正規化的事件封套（envelope）；
- 型別化錯誤；
- Clock、IDGenerator、ProcessRunner 抽象介面；
- 儲存層交易回呼（callback）介面；
- 跨元件的請求／回應 DTO。

優先採用窄介面（narrow interfaces）。不要建立供應商的萬能介面（God interfaces）。

撰寫 `CONTRACT_FREEZE.md`，內容包含：

1. 精確的 import 路徑；
2. 欄位名稱／型別；
3. JSON／YAML 名稱；
4. 狀態轉換表；
5. 遷移所有權範圍；
6. 冪等性金鑰；
7. 交易邊界；
8. 未知／null 語意；
9. 隱私預設值；
10. 代理人不得覆寫的規則。

## 整合責任

- 依規定順序 rebase／合併分支。
- 拒絕重複的領域（domain）結構與 wire payload 洩漏。
- 只處理接線（wiring）問題；將功能缺陷退回給負責的代理人。
- 在支援的情況下，以 race detection 執行所有測試。
- 驗證 SQLite fixtures、記錄或快照中沒有出現原始提示詞。
- 驗證每個代理人的進度產出物都包含證據與最終的 commit SHA。
- 進行最終的 Fable race／安全性審查。

## 驗收

```bash
gofmt -w internal/domain internal/app pkg/protocol
 go test ./internal/domain/... ./pkg/protocol/...
```

所有其他代理人都能針對凍結的埠（ports）編譯出一個小型假實作（fake implementation），而不需編輯 A00 所擁有的檔案。

## 範圍外

- 實作 Claude 解析器、預測器、檢查點儲存、暫停邏輯或 CLI 處理常式。
- 為 Codex／VS Code／外部供應商新增臆測性的抽象介面。
