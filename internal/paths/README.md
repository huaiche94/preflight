# internal/paths/ — per-OS resolution of Auspex's global config/data/cache/runtime directories

> 🌐 English | [繁體中文](README.zh-TW.md)

Package `paths` resolves the OS-correct, non-repository-local directories Auspex
uses for global user configuration, persistent data, cache, and runtime
(socket/pid/lock) files. The package contract is the package comment at the top of
`paths.go` (no separate `doc.go`). Repository-local paths (`.auspex/config.yaml`
etc., Auspex_ADD.md §26.3 — the ADD lives at
[docs/design/Auspex_ADD.md](../../docs/design/Auspex_ADD.md)) are explicitly not
this package's concern.

Key entry points:

- `Resolve(goos, env) (Dirs, error)` — per-OS branch: XDG for Linux and other
  POSIX, `resolveDarwin` for macOS (Application Support et al.), `resolveWindows`
  for Windows; tests pass a fixed `goos` to exercise all three families from any
  host. `ResolveHost(env)` wraps `Resolve(runtime.GOOS, env)`.
- `Dirs{Config, Data, Cache, Runtime}` — the resolved set. `Runtime` holds
  ephemeral single-boot files (daemon socket, pidfile, lockfile); on POSIX it
  prefers a permissions-restricted per-user runtime dir (e.g. `XDG_RUNTIME_DIR`)
  and falls back to `Cache` when none exists (notably macOS).
- `Env` — the injectable source of environment variables and the home directory
  (`OSEnv`/`NewOSEnv` in production, a fake in tests), per agents/foundation.md's
  injectable-environment requirement. `ErrNoHomeDir` is returned (wrapped) when a
  home directory is required but unavailable.

Consumers: [`../config`](../config/README.md) locates global user config here; the
SQLite store, logs, and [`../daemon`](../daemon/)'s socket/lock/pid files use the
data/runtime directories.
