# docs/archive/ — 已被取代的歷史文件

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。本目錄屬歷史封存，非現行實作指引。

這裡收錄的是已被完整取代、僅供**歷史參考——絕非實作依據**的文件。每個
檔案開頭都有一個 `ARCHIVED` 橫幅，說明取代它的內容為何。內容刻意保留舊
有的產品名稱（「AgentGuard」、「Preflight」）與舊有的檔案參照：ADR-045
已決議 `docs/archive/` 與 git 歷史在更名時不予重寫，而 ADR-049 在本次文
件重組中延續了此規則 —— 舊有的引用仍可透過 grep 找到對應內容。

| 檔案／目錄 | 內容說明 |
|---|---|
| [`AgentGuard_Architecture.md`](AgentGuard_Architecture.md) | 最早期的架構草稿，使用前身名稱「AgentGuard」 —— 已被 ADD 完整取代（其橫幅仍引用它為 `Preflight_ADD.md`，現為 [`../design/Auspex_ADD.md`](../design/Auspex_ADD.md)）。 |
| [`execution_prompt.md`](execution_prompt.md) | 早期的四人團隊、兩波次啟動提示草稿 —— 與已核准的九代理人（A00–A08）拓撲架構相牴觸，且從未實際執行。 |
| [`agent-packets-v1/`](agent-packets-v1/README.md) | 以編號區分的九角色（A00–A08）交接文件（九個交接檔案，外加一份契約凍結範本），已由位於 [`../../agents/`](../../agents/) 的語意命名七角色檔案取代。 |

## 相關文件

- 取代這些內容的來源：[`../design/`](../design/README.md)（權威性的
  ADD、predictor 補充文件、執行計畫）以及位於 [`../../agents/`](../../agents/)
  的現行角色交接文件。
- 凍結這些內容的更名決策：
  [`../adr/0045-rename-to-auspex.md`](../adr/0045-rename-to-auspex.md)。
- 實際的建置紀錄（未封存，同樣是歷史紀錄，但內容仍屬正確）：
  [`../implementation/vertical-slice/`](../implementation/vertical-slice/README.md)。

若你發現自己正依據此目錄中的任何內容進行實作，請立即停下 —— ADD 與已
通過的 ADR 才具有優先效力（Constitution §1–§3）。
