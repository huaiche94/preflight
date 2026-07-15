# internal/paths/ — 依作業系統解析 Auspex 全域的 config/data/cache/runtime 目錄

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

`paths` 套件負責解析 Auspex 用於全域使用者設定、持久性資料、快取，以及 runtime（socket／pid／lock）檔案時，所使用的、依作業系統而定且非
repository-local 的目錄。套件契約就是 `paths.go` 檔案開頭的套件註解（沒有另外的 `doc.go`）。
Repository-local 的路徑（例如 `.auspex/config.yaml`，Auspex_ADD.md §26.3 — 該 ADD
位於 [docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)）明確不在這個套件的職責範圍內。

主要進入點：

- `Resolve(goos, env) (Dirs, error)` — 依作業系統分支：Linux 與其他 POSIX 系統使用
  XDG、macOS 使用 `resolveDarwin`（Application Support 等等）、Windows 使用
  `resolveWindows`；測試會傳入固定的 `goos`，讓任何主機都能演練這三種系統家族。
  `ResolveHost(env)` 則是包裝了 `Resolve(runtime.GOOS, env)`。
- `Dirs{Config, Data, Cache, Runtime}` — 解析後的結果集合。`Runtime`
  存放僅在單次開機期間存在的暫時性檔案（daemon socket、pidfile、lockfile）；在 POSIX
  系統上，會優先使用具權限限制的 per-user runtime 目錄（例如 `XDG_RUNTIME_DIR`），若不存在則退回使用
  `Cache`（macOS 上尤其如此）。
- `Env` — 可注入的環境變數與家目錄來源（production 環境使用
  `OSEnv`/`NewOSEnv`，測試中使用 fake），依 agents/foundation.md 的可注入環境要求而設計。當需要家目錄但無法取得時，會回傳（包裝過的）
  `ErrNoHomeDir`。

使用端：[`../config`](../config/README.md) 會在這裡定位全域使用者設定；SQLite
store、記錄檔，以及 [`../daemon`](../daemon/) 的 socket／lock／pid 檔案，都使用這裡的
data／runtime 目錄。
