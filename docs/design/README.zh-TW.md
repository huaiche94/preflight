# docs/design/ — 權威設計文件

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

Auspex 的三份治理性設計文件。它們原本位於儲存庫根目錄，直到
ADR-049 將它們搬移至此處（2026-07-14）；檔名維持不變，因此散見於
程式碼註解與歷史文件中、像是 `Auspex_ADD.md §31` 這樣的章節引用，
仍然能明確指向這些檔案。

| 文件 | 角色 |
|---|---|
| [`Auspex_ADD.md`](Auspex_ADD.md) | **唯一具權威性的架構與實作規格書**——涵蓋產品架構、領域模型、功能需求、路線圖。當程式碼、issue、PR 或評論與其牴觸時，架構面以本文件為準。僅能透過 [`../adr/`](../adr/) 下已核准的 ADR 修訂。**以繁體中文撰寫**（本文內容為 zh-TW，章節標籤與程式碼為英文）；中文文本具規範性，且沒有另外的 `.zh-TW.md` 版本（ADR-049）。 |
| [`Auspex_Predictor_Design_Supplement.md`](Auspex_Predictor_Design_Supplement.md) | ADD（§14–§17）的補充文件：詳述 predictor 管線——範圍估計、token／quota 預測、風險合併。由 ADR-041 正式化。 |
| [`Auspex_Parallel_Execution_Plan.md`](Auspex_Parallel_Execution_Plan.md) | 第一個垂直切片建置案的次要執行計畫：七角色拓撲、擁有權邊界，以及合併順序。其實際執行紀錄位於 [`../implementation/vertical-slice/`](../implementation/vertical-slice/README.md)。 |

## 擁有權

這些檔案是共用的跨領域產物（cross-cutting artifact），專屬由
`contract-integrator` 角色擁有（Constitution §4.3）。`Auspex_ADD.md`
只有在確實存在矛盾、非改不可時才能編輯，且對應的 ADR 必須隨同一次
變更一併送出（Constitution §3.5）。
</content>
