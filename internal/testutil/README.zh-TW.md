# internal/testutil/ — 共用測試輔助程式碼

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

跨角色共用的測試輔助套件的總稱目錄。此目錄本身沒有任何 Go 檔案，也沒有
doc.go；目前唯一的套件是 [fakes/](fakes/)——針對
[../app/ports.go](../app/ports.go) 中凍結的跨元件服務介面，手工撰寫、
可設定的測試替身（test double）。

沒有任何 production 程式碼會 import 此目錄下的任何內容；使用者是
internal/ 底下各套件的測試（例如 internal/app/wiring、internal/pause、
internal/managed），以及 [../integrationtest/](../integrationtest/) 中的
整合測試套件。

fake 的模式與其保證，請見 [fakes/README.md](fakes/README.md)——以及作為
套件契約說明的 fakes/doc.go。
