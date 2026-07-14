# docs/implementation/ — 已執行實作工作的建置紀錄

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

每個已執行（或執行中）的實作工作各自對應一個子目錄。每個子目錄都保存
該次工作實際建置過程的完整證據軌跡：凍結後的任務 DAG、契約凍結
（contract freeze）、各角色逐節點驗證紀錄的進度產出物、回顧報告，以及
建置期間的分析文件。這些屬於**歷史紀錄** —— 對追溯某項功能「如何」及
「為何」被建置很有幫助，但絕不能取代目前的設計文件。

| 目錄 | 內容說明 |
|---|---|
| [`vertical-slice/`](vertical-slice/README.md) | Auspex 第一次建置的完整紀錄：85 個 DAG 節點、橫跨七個角色，由多個平行代理人歷經 13 個整合波次執行，從 Bootstrap 一路到 Stage-5 Final 關卡。目前是此處唯一記錄的工作。 |

## 相關文件

- 產品「本身是什麼」：[`../design/Auspex_ADD.md`](../design/Auspex_ADD.md)
  及其在 [`../design/`](../design/README.md) 中的同層文件。
- 建置過程中變更契約的決策：[`../adr/`](../adr/README.md)（ADR-041 起的
  多項 ADR 修訂了 vertical slice 凍結後的 DAG 與埠）。
- 這些紀錄所體現、且已廣義化以供重複使用的流程：
  [`../methodology/`](../methodology/README.md)。
- Constitution §6.7 說明了這些紀錄存在的理由：產品任務狀態的證據原則
  （「completed 代表有證據佐證」，§6.2）依此類推適用於此處每個角色的進
  度產出物 —— 僅存在於對話中的進度並不算數。
