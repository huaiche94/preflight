# A08 — 跨元件 QA、安全性、可靠性與 CI

> 🌐 [English](08-qa-security-ci.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。本文件屬歷史封存，非現行實作指引。

## 模型

fixtures／CI 工作使用較便宜的模型；最終的對抗式（adversarial）審查使用 Fable。

## ADD 負責範圍

§§27–30、§32 的 DoD（完成定義）、§29 中跨元件的部分、Appendix G／H／I 的驗證。

## 專屬路徑

```text
.github/**
internal/integrationtest/**
testdata/e2e/**
testdata/security/**
docs/security/**
docs/implementation/day1/A08.md
SECURITY.md
CONTRIBUTING.md
CODE_OF_CONDUCT.md
GOVERNANCE.md
```

在第一輪不要變更功能性正式程式碼。針對負責人提出缺陷報告；只有 A00 能授權跨負責人修正。

## 任務

提供客觀證據，證明這個垂直切片（vertical slice）是安全的、可重啟的、冪等的，且與供應商相容。

## 產出物

1. 跨平台的基本 CI：format、vet、test、build；在支援的情況下加上 race。
2. 一個端對端的高風險 Claude fixture 流程：
   - 狀態列（status-line）擷取；
   - 提示詞的 preflight 阻擋；
   - state／repo 檢查點；
   - 一次性允許；
   - Stop 結果；
   - 暫停請求／喚醒復原。
3. 使用同一個 SQLite 資料庫的重啟測試。
4. 重複／順序錯亂事件測試。
5. 針對資料庫匯出／記錄／檢查點清單的原始提示詞與機密外洩掃描器。
6. 路徑穿越／符號連結，以及惡意 fixture 測試。
7. 排程器雙 worker／租約競爭測試。
8. 若 A07 有提供，測試 support-bundle／doctor 的隱私基準線。
9. `go test ./...` 的證據，以及未解決風險報告。

## 安全性斷言

- 若存在 HTTP，需具備 loopback／API 身分驗證；
- 預設不含提示詞文字；
- bearer token／API 金鑰皆遮蔽；
- hook payload 大小限制；
- 在支援的情況下，SQLite 與產出物權限採限制性設定；
- 外部指令使用 argv，不進行 shell 內插（interpolation）；
- 儲存庫產出物解壓縮不能逃逸出目的目錄；
- 自動接續（auto-resume）需要明確設定的同意。

## 最終報告

建立依嚴重程度排序的報告，內容包含：

```text
P0 blocks merge
P1 must fix before demo
P2 documented follow-up
```

每項發現都必須包含確切的檔案、重現步驟、預期不變性（invariant），以及負責的代理人。
