# internal/lock/ — 單機、PID 檔案風格的 advisory file lock

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`lock` 套件提供單機、單一 daemon 適用的 advisory file lock。套件契約就是 `lock.go`
檔案開頭的套件註解（沒有另外的 `doc.go`）。

Auspex 是一個 local-first 的 modular monolith（Auspex_ADD.md §1.4；該 ADD 位於
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)）：設計上，每台機器（或每個
runtime directory）在同一時間應該只有一個 daemon 擁有特定的 SQLite 資料庫與 runtime
directory。這個套件讓該 daemon — 以及短生命週期的 CLI 呼叫 — 擁有一種能在當機情況下仍安全偵測並防止並行擁有的方式。

- `Acquire(path) (*FileLock, error)` — 取得鎖，若另一個*仍在存活*的行程已持有該鎖，則回傳
  `ErrLocked`（來自已死亡 PID 的過期鎖會被偵測並回收）。
- `FileLock.Release()` / `FileLock.Path()`。

它刻意**不是**分散式鎖、不是網路鎖，也不是用來取代 SQLite 自身的 WAL／busy-timeout
並行控制機制（那是 [`../storage/sqlite`](../storage/sqlite/README.md) 的職責）。它主要的使用端是
[`../daemon`](../daemon/)：其位於 runtime directory（由 [`../paths`](../paths/README.md)
解析）中的 `daemon.lock`，用來落實「每個 runtime directory 只能有一個 daemon」。
