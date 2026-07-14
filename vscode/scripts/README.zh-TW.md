# vscode/scripts/ — 擴充套件的建置／測試輔助腳本

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

目前只有一支腳本：

- `run-tests.js` —— 已編譯單元測試的確定性啟動器；`npm test` 會執行
  它（見 [`../package.json`](../package.json) 中的 `scripts.test`，在
  `pretest` 先將 `src/` 建置成 `out/` 之後）。它會自行列舉
  `out/test/*.test.js`，**若找不到任何測試檔案，就明確失敗**（絕不能
  讓「零個測試被發現」看起來像是綠燈通過），並將明確的檔案清單交給
  `node --test`。

之所以需要這支腳本，是因為 `node --test` 的位置參數（positional-path）
語意在不同 Node 版本間並不一致（其檔頭註解記錄了此腳本所取代的 CI 迴
歸問題）：Node 20 會掃描目錄型引數，Node 22 會把它當成模組處理並以
`ERR_MODULE_NOT_FOUND` 中止，而在 Node ≥ 21 上，不相符的 glob 樣式會
執行零個測試並以結束碼 0 收尾。使用明確的檔案清單則能在各版本間表現
一致。

CI 在 [`../../.github/workflows/ci.yml`](../../.github/workflows/ci.yml)
的 `vscode` job 中執行同一個 `npm test` 進入點（Node 依本儲存庫的精確
版本鎖定政策，固定為 22.11.0）。測試本身位於
[`../src/test/`](../src/test/README.md)。
