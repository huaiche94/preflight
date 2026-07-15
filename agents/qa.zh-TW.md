# QA

> 🌐 [English](qa.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

跨元件 QA、安全性、可靠性與 CI。

## 模型

fixture/CI 工作使用較便宜的模型；最終的對抗式（adversarial）審查使用 Fable。

## ADD 負責章節

§§27–30、§32 的 DoD（Definition of Done）、§29 中的跨元件部分、
附錄 G/H/I 的驗證項目。

## 專屬路徑

```text
.github/**
internal/integrationtest/**
testdata/e2e/**
testdata/security/**
docs/security/**
docs/implementation/vertical-slice/qa.md
SECURITY.md
CONTRIBUTING.md
CODE_OF_CONDUCT.md
GOVERNANCE.md
```

初期不得變更功能性的 production code。針對負責的角色提出缺陷回報即可；
只有 contract-integrator 能授權跨角色的修正。

## 任務目標

提供客觀證據，證明這個 vertical slice 是安全、可重新啟動（restartable）、
冪等（idempotent），且與 provider 相容的。

## 交付項目

1. 跨平台的基本 CI：format、vet、test、build；在支援的平台上加上 race。
2. 一條端對端（end-to-end）的高風險 Claude fixture 流程：
   - status-line 擷取（ingestion）；
   - prompt 中的 auspex block；
   - state/repo checkpoint；
   - 一次性 allow；
   - Stop 結果；
   - pause 請求／wake 復原。
3. 使用同一個 SQLite DB 的重新啟動測試。
4. 重複／順序錯亂（out-of-order）事件測試。
5. 針對 DB export/logs/checkpoint manifest 的原始 prompt 與機密內容外洩掃描器。
6. Path traversal/symlink 與惡意 fixture 測試。
7. Scheduler 的 double-worker/lease 競態測試。
8. 若 runtime 角色有提供 support-bundle/doctor，需驗證其隱私基準線。
9. `go test ./...` 的證據，以及尚未解決風險的報告。

## 安全性斷言

- 若存在 HTTP，需有 loopback/API 驗證；
- 預設不含 prompt 文字；
- bearer token/API key 需被遮蔽；
- hook payload 大小限制；
- 在支援的情況下，SQLite 與 artifact 權限應盡量限縮；
- 外部指令一律使用 argv 形式，不得做 shell interpolation；
- repository artifact 解壓縮不得逸出目的地路徑；
- 自動 resume 需要明確設定過的同意（consent）。

## 最終報告

建立一份依嚴重程度分級的報告：

```text
P0 blocks merge
P1 must fix before demo
P2 documented follow-up
```

每項發現都須包含確切檔案、重現步驟、預期不變條件（invariant），
以及負責的角色。
