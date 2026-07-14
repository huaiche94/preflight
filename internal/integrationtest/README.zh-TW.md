# internal/integrationtest/ — 跨角色整合測試與端對端測試

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

純測試用套件（沒有非測試用的 Go 檔案，也沒有 doc.go）：驗證各自獨立建置、
也各自完成單元測試的元件，在端對端組合時仍能正確運作。這些測試套件刻意
驅動「真正的」production 程式碼——真實的落地 SQLite 檔案（絕不使用
`:memory:`）、真實的 migration、真實的暫存 git repository、真實的子行程
（subprocess）——而非套件內部的 fake。這是 qa 的專屬路徑之一
（agents/qa.md）；套件層級的單元測試則與各套件放在一起。

目前的測試套件：

- e2e_highrisk_test.go — 垂直切片（vertical-slice）示範（qa-02）：驅動一個
  高風險 turn 端到端流程，從 status-line 擷取、經評估封鎖、checkpoint、
  一次性允許（one-time allow）、stop，到 pause/wake 復原。
- restart_sameDB_test.go — 對同一個 SQLite 檔案，跨多個角色的儲存層進行
  重啟測試（qa-03）。
- duplicate_outoforder_test.go — 冪等事件持久化與重複／失序
  （out-of-order）進度處理的組合測試（qa-04）。
- leakage_scanner_test.go — 針對落地 DB bytes（含 WAL）與 checkpoint
  artifact 的原始 prompt／機密掃描（qa-05）。
- malicious_fixture_test.go — 針對凍結 checkpoint 契約的路徑穿越
  （path-traversal）／符號連結（symlink）／惡意 fixture 攻擊測試（qa-06）。
- scheduler_doubleworker_test.go — 整合層級的 scheduler 雙 worker／lease
  競態測試（qa-07）。
- hookbootstrap_test.go — issue #17 驗收：hook 路徑在沒有任何測試預先塞入
  資料（seeding）的情況下，能自行 bootstrap repository/worktree/session
  資料列。
- evaluate_privacy_test.go — issue #14：`auspex evaluate` 絕不持久化原始
  prompt 文字（以雜湊存在性作為負向對照的 canary 掃描）。
- forecast_prompt_conditioned_test.go — issue #42：token forecast 會依
  prompt 內容而變化。
- managedrun_test.go — issue #8：`auspex run` 針對編譯出的 fake-provider
  子行程進行 managed one-shot 測試；這是此處唯一同時使用
  [../testutil/fakes/](../testutil/fakes/) 中 double 的測試套件。
