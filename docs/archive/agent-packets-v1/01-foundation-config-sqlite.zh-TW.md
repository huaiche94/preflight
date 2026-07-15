# A01 — 基礎建設、設定、路徑與核心 SQLite

> 🌐 [English](01-foundation-config-sqlite.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。本文件屬歷史封存，非現行實作指引。

## 模型

使用較便宜的程式碼模型即可；僅在遷移／復原（recovery）審查時使用 Fable。

## ADD 負責範圍

§10、§12 的核心部分、§26、§30 的啟動（bootstrap）部分、M0／M1。

## 任務

建立可建置的 Go 應用程式基礎架構，以及供其他所有功能套件使用的 SQLite 執行期。

## 專屬路徑

```text
go.mod
go.sum
cmd/preflight/main.go
internal/buildinfo/**
internal/config/**
internal/paths/**
internal/storage/sqlite/db.go
internal/storage/sqlite/migrate.go
internal/storage/sqlite/migrations/0000-0009_*.sql
internal/clock/**
internal/idgen/**
internal/lock/**
Makefile
Taskfile.yml
.golangci.yml
LICENSE
NOTICE
```

不要編輯 A00 的合約或功能遷移範圍。

## 產出物

1. Go module 與最小可行的 `preflight version`。
2. 依作業系統正確設置的 config／data／cache／runtime 路徑，並可注入環境變數／home 目錄。
3. YAML 設定載入，以及 day-one 流程所需欄位的優先順序文件。
4. SQLite 開啟／遷移／交易輔助函式。
5. 依 ADD 規定的 WAL、busy timeout、外鍵（foreign keys）與遷移版本檢查。
6. 外鍵所需的核心資料表：repositories、worktrees、sessions、turns、tasks、供應商安裝／設定中繼資料。
7. 供功能代理人使用的遷移測試框架（test harness）。
8. 透過窄範圍的應用程式建構子（app constructor），支援基本的儲存庫初始化指令；使用者介面指令由 A07 負責。

## 必要測試

- 從空白資料庫進行遷移；
- 重新開啟與冪等遷移；
- 安全地拒絕較新的 schema；
- 鎖定／忙碌（busy）行為；
- 無效權限與損毀資料庫的錯誤分類；
- Windows／macOS／Linux 路徑表測試；
- 設定優先順序與未知欄位行為；
- version 指令。

## 交接事項

在 `docs/implementation/day1/A01.md` 中記錄資料庫建構子、交易 API、遷移命名慣例，以及相依需求。

## 範圍外

- 超出核心 session／turn 參照範圍的遙測功能資料表；
- HTTP daemon；
- 完整的發行封裝（release packaging）；
- 任何預測器／供應商／檢查點的商業邏輯。
