# internal/testutil/fakes/ — 針對凍結 app port 的測試替身

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

針對 [../../app/ports.go](../../app/ports.go) 中凍結的跨元件服務介面，
手工撰寫、可設定的測試替身（test double），供各角色共用，避免各自重複
造輪子。套件契約請見 doc.go。

每個 fake 都遵循同一套模式（doc.go 有詳細說明）：

- 每個凍結介面對應一個匯出的 struct，命名為 `Fake<InterfaceName>`，並針對
  每個方法提供一個可選的 `<Method>Func` 欄位——測試只需設定自己需要的方法
  即可；
- 一個編譯期斷言 `var _ app.X = (*FakeX)(nil)`，確保凍結契約一旦變動，
  會在建置期讓本套件失敗，而不是等到下游測試才發現；
- 呼叫某個 `Func` 欄位為 `nil` 的方法時，會以凍結的 `domain.Error` 型態
  （`ErrCodeUnavailable`；unconfigured.go）大聲失敗，而不是靜默回傳零值；
- 沒有內建的呼叫紀錄或同步機制——若測試需要這些功能，需自行在其 `Func`
  closure 中實作。

提供的 fake：`FakeEvaluationService`、`FakeGracefulPauseService`、
`FakeProgressTreeService`、`FakeRepositoryCheckpointService`、
`FakeStateCheckpointService`、`FakeTurnInterrupter`、
`FakeSessionResumer`。

providercontract.go 另外匯出一組可重複使用的契約測試套組
（`ProviderInterrupterContract`、`ProviderSessionResumerContract`），
目前用於針對兩個 provider fake 執行測試（providercontract_test.go），
未來也可用於任何真正的 interrupter/resumer adapter。
