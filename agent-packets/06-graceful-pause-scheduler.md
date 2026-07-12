# A06 — Graceful Pause, Safe Points, and Durable Scheduler

## Model

Use Fable.

## ADD ownership

§20, pause parts of §§15/17/28/29, Appendix C, ADR-031 through ADR-040.

## Exclusive paths

```text
internal/pause/**
internal/scheduler/**
schemas/pause.schema.json
testdata/pause-scenarios/**
internal/storage/sqlite/migrations/0050-0059_*.sql
docs/implementation/day1/A06.md
```

## Mission

Implement the provider-neutral pause/resume state machine and durable wake scheduling. Depend only on frozen ports for predictor, progress/state checkpoint, repository checkpoint, provider interrupt/resume, quota read, clock, and leases.

## Required state path

```text
observing
→ pause_requested
→ quiescing
→ safe_point_reached
→ persisting
→ interrupting
→ sleeping
→ wake_due
→ validating
→ resuming
→ resumed
```

Include terminal/conflict/cancel/failure states from ADD.

## P0 deliverables

1. State transition validator.
2. `Observe` handling with debounce/hysteresis state.
3. `RequestPause` idempotency.
4. Safe-point coordinator interface and implementation for turn/section boundary observations.
5. Persist phase orchestration:
   - Progress Tree snapshot;
   - State Checkpoint;
   - Repository Checkpoint;
   - Pause Record;
   - Wake Job.
6. Durable scheduler lease with claim/renew/complete/fail/retry.
7. Restart recovery of overdue/leased jobs.
8. Resume validation:
   - quota safe;
   - repository fingerprint compatible;
   - session/provider capability valid;
   - authorization/consent valid.
9. Duplicate wake exactly-once behavior.
10. Cancel prevents future resume.
11. Provider interrupter/resumer fake contract tests.

## Day-one realism

A calibrated auto-pause may be unavailable due insufficient data. Support both:

- calibrated trigger: `P_hit_10m >= threshold` for consecutive observations;
- explicit uncalibrated emergency policy with a different reason code.

Implement durable wake and fake resumer first. Actual managed Claude resume is stretch and must not weaken state-machine tests.

## Required tests

- two qualifying observations trigger request;
- one spike does not;
- safe point persists checkpoints before interrupt;
- crash after every phase resumes/reconciles correctly;
- restart recovers wake job;
- unsafe quota reschedules;
- repo overlap blocks;
- unrelated repo change follows configured policy;
- duplicate workers yield one resume;
- expired lease reclaimed;
- cancel wins race with wake;
- provider interrupt failure leaves recoverable state.

## Boundary

Do not parse Claude events or implement checkpoint internals. Use ports/fakes.
