# Agents

> 🌐 [English](README.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。

本目錄包含角色定義（role definitions）。

每個檔案彼此獨立。

每個角色擁有一個界限明確的範疇（bounded context）。

切勿修改其他角色的檔案。

架構請參閱 `docs/design/Auspex_ADD.md`。

## Roles

| 檔案 | 角色 | 整併自（合併前依編號排列的歷史版本見 `docs/archive/agent-packets-v1/`） |
|---|---|---|
| `contract-integrator.md` | 凍結編譯期／持久化契約；整合已審查的分支 | （原為 `00-contract-integrator.md`） |
| `foundation.md` | Go module、設定、路徑、核心 SQLite runtime | （原為 `01-foundation-config-sqlite.md`） |
| `claude-provider.md` | Claude Code telemetry、hooks、provider 正規化 | （原為 `02-claude-telemetry-hooks.md`） |
| `checkpoint.md` | Progress Tree + State Checkpointing，**以及** Repository Checkpoint | （原為 `03-progress-state-checkpoint.md` + `04-repository-checkpoint.md`） |
| `predictor.md` | Scope estimator、predictor、risk、policy、authorization | （原為 `05-predictor-policy.md`） |
| `runtime.md` | Graceful Pause + durable scheduler，**以及** CLI/API/orchestration | （原為 `06-graceful-pause-scheduler.md` + `07-runtime-cli-api.md`） |
| `qa.md` | 跨元件 QA、安全性、可靠性、CI | （原為 `08-qa-security-ci.md`） |
| `CONTRACT_FREEZE_TEMPLATE.md` | contract-integrator 的 `docs/implementation/vertical-slice/CONTRACT_FREEZE.md` 交付物範本 | （不變） |

`checkpoint`、`runtime` 這兩個角色各自涵蓋了原本兩個獨立的 packet。
之所以維持為單一角色，是因為兩個半邊在實務上總是一起被使用；但每個檔案仍將兩
個半邊清楚分隔在 Part A / Part B 區段中——包括各自獨立的專屬路徑（exclusive
paths）與各自獨立的 migration 編號範圍——因此內部界線依然是真正的分界，
而不是合併成一團無法區分的整體。

## Spawning

一份 packet 檔案的設計目的，是可以單獨交給一個隔離的 agent/worktree
——不需要在 context 中提供完整的 `docs/design/Auspex_ADD.md` 或
`docs/design/Auspex_Parallel_Execution_Plan.md`。啟動一個角色時：
提供這份檔案的內容、其被指派的 `docs/design/Auspex_ADD.md` 章節，以及
contract-integrator 產出後的 `docs/implementation/vertical-slice/CONTRACT_FREEZE.md`。

若需整體脈絡（願景、範疇邊界、topology、合併順序、刪減清單），
請參閱 `docs/design/Auspex_Parallel_Execution_Plan.md`；架構請參閱
`docs/design/Auspex_ADD.md`；專案層級的治理與優先順序規則，請參閱
Auspex Repository Constitution。
