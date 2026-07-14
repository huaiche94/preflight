# internal/storage/ — storage adapters (currently SQLite only)

> 🌐 English | [繁體中文](README.zh-TW.md)

This directory is the storage-adapter branch of the dependency direction fixed by
Auspex_ADD.md §7.6 (the ADD lives at
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)): adapters sit at the
bottom of the stack, beneath domain services and interfaces. There is no Go code at
this level — it only namespaces concrete storage backends.

Its single subpackage is [`sqlite/`](sqlite/README.md), the frozen "SQLite runtime"
import path from
[CONTRACT_FREEZE.md](../../docs/implementation/vertical-slice/CONTRACT_FREEZE.md)'s
import-path table: connection setup and durability pragmas, the `app.TxRunner`
transaction boundary every storage-touching operation goes through, and the
forward-only migration engine with its embedded
[`sqlite/migrations/`](sqlite/migrations/README.md) files.

Higher layers do not import `database/sql` details directly; they either depend on
the frozen ports in [`../app`](../app/) (e.g. `TxRunner.WithTx`) or, for role-owned
stores (e.g. [`../progress`](../progress/), [`../telemetry/claude`](../telemetry/claude/)),
on `sqlite.DB`'s transaction/querier seam.
