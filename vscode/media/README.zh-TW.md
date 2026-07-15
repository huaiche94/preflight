# vscode/media/ — VS Code 擴充套件的靜態資源

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

[`../package.json`](../package.json) 的 `contributes` 區段所參照的靜
態檔案。目前只有一個檔案：

- `auspex.svg` —— Auspex 活動列（activity-bar）容器及其 Status 檢視
  的圖示（`viewsContainers.activitybar[].icon` 與
  `views.auspex[].icon` 皆指向 `media/auspex.svg`）。繪製時只使用線
  條（stroke-only）與 `currentColor`，因此會跟隨編輯器佈景主題變化；
  圖案主題是鳥占師（bird-augur）望向地平線的觀測弧線（詳見 SVG 檔案
  自身的註解）。

`package.json` 中的路徑是相對於擴充套件的根目錄，因此在此重新命名或
搬移檔案時，必須在同一次變更中一併更新該 manifest 檔案。擴充套件的行
為與開發流程記載於 [`../README.md`](../README.md)。
