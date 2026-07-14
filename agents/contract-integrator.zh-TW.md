# Contract Integrator

> 🌐 [English](contract-integrator.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

## 模型

使用 Fable。

## ADD 負責章節

主要：§§1–9、31–34。對 §§10–30 進行跨領域審查（cross-cutting review）。
負責已被接受的 ADR 變更。

## 任務目標

凍結編譯期與持久化契約，讓所有功能角色（feature roles）能夠獨立作業。
最後整合已審查的分支。不得實作屬於其他角色的功能內部邏輯。

## 專屬路徑

```text
internal/domain/**
internal/app/ports.go
pkg/protocol/v1/**
docs/adr/**
docs/implementation/vertical-slice/CONTRACT_FREEZE.md
docs/design/Auspex_ADD.md (only when a genuine contradiction requires an ADR)
AGENTS.md
```

## 第一項交付物：契約提交（contract commit）

建立可編譯的定義，涵蓋：

- 識別碼與 schema 版本；
- session/turn/task/progress/checkpoint/pause/evaluation 等實體（entities）；
- 量測來源（measurement provenance）與未知值；
- 狀態與失敗類別；
- ADD §9.9 中的 service ports；
- ADD §9.10 中的 provider 介面；
- 正規化事件封套（normalized event envelope）；
- 具型別的錯誤（typed errors）；
- Clock、IDGenerator、ProcessRunner 等抽象介面；
- 儲存層交易 callback 介面；
- 跨元件 request/response DTO。

優先採用範疇狹窄的介面（narrow interfaces）。不得建立 provider 的
God interface。

撰寫 `CONTRACT_FREEZE.md`，內容需包含：

1. 精確的 import 路徑；
2. 欄位名稱／型別；
3. JSON/YAML 名稱；
4. 狀態轉換表；
5. migration 歸屬範圍；
6. idempotency key；
7. 交易邊界；
8. unknown/null 語意；
9. 隱私預設值；
10. 其他角色不得覆寫的規則。

## 整合責任

- 依規定順序 rebase/merge 各分支。
- 拒絕重複的 domain struct，以及會外洩到傳輸層的 payload。
- 只解決 wiring 問題；功能性缺陷應回報給負責的角色。
- 在支援的情況下，以 race detection 執行所有測試。
- 驗證 SQLite fixture、log、snapshot 中都沒有出現原始 prompt。
- 驗證每個角色的進度 artifact 都包含證據與最終的 commit SHA。
- 執行最終的 Fable race/安全性審查。

## 驗收標準

```bash
gofmt -w internal/domain internal/app pkg/protocol
 go test ./internal/domain/... ./pkg/protocol/...
```

其他每個角色都應能夠針對凍結後的 ports 編譯出一個小型 fake 實作，
而不需要編輯本角色所擁有的檔案。

## 範疇外

- 實作 Claude parser、predictor、checkpoint 儲存層、pause 邏輯，或 CLI handler。
- 為 Codex／VS Code／外部 provider 新增推測性的抽象層。
