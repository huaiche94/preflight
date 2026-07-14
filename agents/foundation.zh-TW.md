# Foundation

> 🌐 [English](foundation.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

## 模型

較便宜的 coding model 即已足夠；僅在 migration/recovery 審查時使用 Fable。

## ADD 負責章節

§10、§12 的核心部分、§26、§30 中 bootstrap 相關部分、M0/M1。

## 任務目標

建立可建置（buildable）的 Go 應用程式基礎，以及供其他每個角色的 package
使用的 SQLite runtime。

## 專屬路徑

```text
go.mod
go.sum
cmd/auspex/main.go
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

不得編輯 contract-integrator 的檔案，也不得編輯其他角色的 migration 範圍。

## 交付項目

1. Go module 與最小可行的 `auspex version`。
2. 依作業系統正確判斷的 config/data/cache/runtime 路徑，且環境變數／home 可注入替換。
3. YAML config 載入，並記錄 day-one 流程所需欄位的優先順序（precedence）。
4. SQLite 的 open/migrate/transaction 輔助函式。
5. 依 ADD 規範實作 WAL、busy timeout、外鍵（foreign keys），以及 migration 版本檢查。
6. 外鍵所需的核心資料表：repositories、worktrees、sessions、turns、tasks、
   provider 安裝／設定中繼資料。
7. 供其他每個角色使用的 migration 測試工具（test harness）。
8. 透過一個範疇狹窄的 app constructor，提供基本的 repository 初始化指令支援；
   使用者面向的指令由 runtime 角色擁有。

## 必要測試

- 從空白資料庫進行 migration；
- 重新開啟並確認 migration 的冪等性；
- 較新的 schema 應被安全地拒絕；
- locked/busy 行為；
- 無效權限與資料庫損毀時的錯誤分類；
- Windows/macOS/Linux 的路徑對照表測試；
- config 優先順序與未知欄位的行為；
- version 指令。

## 交接事項

在 `docs/implementation/vertical-slice/foundation.md` 中記錄 DB constructor、
交易 API、migration 命名慣例，以及依賴需求（dependency requests）。

## 範疇外

- 核心 session/turn 參照以外的 telemetry 功能資料表；
- HTTP daemon；
- 完整的 release 打包流程；
- 任何 predictor/provider/checkpoint 的商業邏輯。
