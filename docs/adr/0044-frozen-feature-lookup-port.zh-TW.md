# ADR-044 — 凍結 repository/session feature-lookup port（REC-01）

> 🌐 [English](0044-frozen-feature-lookup-port.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

狀態：已接受（Accepted）
日期：2026-07-13
負責人：contract-integrator（由 lead 執行）
核准人：repository owner，2026-07-13（issue #4 決策會議）

## 背景

Bootstrap 刻意延後了 repository/session/progress feature-lookup port（「What Bootstrap did NOT freeze」，`CONTRACT_FREEZE.md`）。因此有三個套件各自為相同能力長出了自己的本地 seam：

- `internal/predictor/scope.FeatureSource`（predictor-05）
- `internal/predictor/token.FeatureSource`（predictor-05b）
- `internal/evaluation.DataSource`（predictor-09；為前兩者的超集，共 11 個方法，由 `SQLDataSource` 實作，並在 Final integration gate 時接進真正的 binary）

`wave2-analysis/ADR_Recommendations.md` 的 REC-01 將此標記為排名第一、可被關閉的缺口：未來任何 predictor tier 或其他角色若需要相同的資料，要嘛得重新發明一個介面，要嘛得 import 另一個套件的內部實作——這正是 narrow-ports 這項紀律要防止的耦合。`Feature_Gap_Report.md` §1.1 也獨立地將其所促成的 repository-features 接線，列為第一優先的關鍵缺口。

## 決策

1. **將 `evaluation.DataSource` 的形狀原封不動地升格進入凍結合約**，成為 `internal/app/ports.go` 中的 `app.FeatureDataSource`（＋`app.ResolvedSession`）。採取原封不動升格而非重新設計，是因為這個形狀已經在真實使用中存活下來：它由正式環境的 SQLite adapter 實作、由完整的 evaluation 管線消費，並由 E2E 測試套件驗證過。
2. **`internal/evaluation` 為凍結型別建立別名**（`type DataSource = app.FeatureDataSource`），讓每一個 consumer、實作與測試都能維持原樣、不需修改即可編譯通過。
3. **predictor 端的兩個 `FeatureSource` 介面維持不變**，作為 consumer 端的窄視圖（interface segregation）：各自只宣告自己會用到的子集，而正式環境的 adapter 則以同一份 `app.FeatureDataSource` 實作同時支撐兩者。此次凍結固定的是 canonical 形狀，而非消費模式。
4. `internal/app` 新增對 `internal/features` 的 import（純 DTO 套件，僅 import `domain`——相依方向維持乾淨：`app → features → domain`）。

## 影響

- 依 Constitution §3，feature-lookup 的形狀現在是一項相容性承諾；變更需要 ADR，而不是套件內部的修改。
- 未來的 consumer（Statistical/ML predictor tier、issue-#14 的 forecast 介面、ADR-043 的 cost forecaster）將相依於 `app.FeatureDataSource`，而不是伸手進入 `internal/evaluation`。
- `CONTRACT_FREEZE.md` 新增一個 Amendments 小節記錄此變更；Bootstrap 當初的延後註記以引用方式關閉，而非被改寫。

## 已考慮的替代方案

- **在凍結前重新設計 port**——已否決：目前沒有任何 consumer 提出現有形狀無法滿足的需求；在沒有明確需求驅動下重新設計，屬於臆測性抽象化（README/ADD 貢獻規則所禁止）。
- **完全整併掉 predictor 端的介面**——已否決：它們的狹窄性對測試用的 fake 與誠實的相依關係而言是承重的；Go 的結構化型別讓這些視圖本身沒有額外成本。
