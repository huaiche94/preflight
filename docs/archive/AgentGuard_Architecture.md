> **ARCHIVED — obsolete.** This is an early, differently-named precursor
> ("AgentGuard") to the current project. It has been superseded in full by
> `Preflight_ADD.md`, which is the sole authoritative architecture document.
> Kept for historical reference only; do not implement against this file.

# AgentGuard --- Runtime Guard for AI Coding Agents

## Vision

AgentGuard is **not** another checkpoint manager or agent memory system.

Its purpose is to become the **runtime operating system for AI coding
agents**, sitting between the user and providers such as Codex CLI,
Claude Code, Gemini CLI, Cursor CLI, etc.

    User
      │
      ▼
    AgentGuard
      │
      ├── Telemetry
      ├── Forecast
      ├── Policy
      ├── Checkpoint
      └── Execution
      │
      ▼
    Codex / Claude / Gemini / ...

Core value:

-   Predict execution risk before a turn starts.
-   Recommend checkpointing only when justified.
-   Learn from historical execution telemetry.
-   Eventually optimize routing, task splitting, and runtime policy.

------------------------------------------------------------------------

# Product Positioning

## What we are NOT

-   Agent Memory
-   Another Resume Tool
-   Another Checkpoint Tool
-   Another Prompt Manager

Projects already occupy much of this space.

Our differentiator is **prediction**, not persistence.

------------------------------------------------------------------------

# Core Principle

Current tools answer:

> "How do we continue?"

AgentGuard answers:

> "Should we even start this turn?"

------------------------------------------------------------------------

# Long-term Goal

Become the runtime middleware shared by multiple AI coding agents.

                     VS Code
                        │
                        ▼
               AgentGuard Runtime
          ┌─────────────┼─────────────┐
          ▼             ▼             ▼
       Codex         Claude         Gemini

------------------------------------------------------------------------

# Architecture

    cmd/
        agentguard/

    internal/
        telemetry/
        predictor/
        policy/
        checkpoint/
        providers/
            codex/
            claude/
            gemini/
        git/
        storage/
        runtime/
        scheduler/
        execution/

    pkg/
        sdk/

    research/
        notebooks/
        evaluation/

------------------------------------------------------------------------

# Language Choice

Production runtime: **Go**

Research and experimentation: **Python**

Reasoning:

-   Go is ideal for developer tooling.
-   Single static binary.
-   Excellent CLI ecosystem.
-   Easy Homebrew/Winget distribution.
-   Strong concurrency.
-   Small memory footprint.

Python is retained only for:

-   feature engineering
-   notebooks
-   offline evaluation
-   ML experiments
-   exporting model weights

------------------------------------------------------------------------

# Runtime Pipeline

    User Prompt
          │
          ▼
    Telemetry Collector
          │
          ▼
    Scope Estimator
          │
          ▼
    Risk Predictor
          │
          ▼
    Policy Engine
          │
          ▼
    Checkpoint Decision
          │
          ▼
    Provider

------------------------------------------------------------------------

# Major Components

## 1. Telemetry Engine

Collect:

-   prompt length
-   input/output tokens
-   tool calls
-   files read
-   files modified
-   lines added/deleted
-   git diff
-   execution duration
-   build/test results
-   retries
-   session metadata

Persist into SQLite.

------------------------------------------------------------------------

## 2. Scope Estimator

Estimate before execution:

-   files likely to read
-   files likely to change
-   changed LOC
-   integration tests required
-   dependency fan-out
-   uncertainty score

Output:

``` json
{
  "estimatedFilesReadP50": 8,
  "estimatedFilesReadP90": 19,
  "estimatedFilesChangedP50": 4,
  "estimatedFilesChangedP90": 9,
  "uncertainty": 0.42
}
```

------------------------------------------------------------------------

## 3. Risk Predictor

Initially rule-based.

Later:

-   quantile regression
-   gradient boosted trees
-   logistic regression

Interface:

``` go
type Predictor interface {
    Estimate(input PredictionInput) PredictionResult
}
```

Implementations:

-   RulePredictor
-   StatisticalPredictor
-   MLPredictor
-   RemotePredictor

------------------------------------------------------------------------

## 4. Policy Engine

Consumes prediction.

Returns:

-   RUN
-   WARN
-   CHECKPOINT
-   SPLIT
-   ABORT

Example:

    IF
    rolling_usage > 70%
    AND
    estimated_files > 8
    AND
    integration_tests == true

    => CHECKPOINT

------------------------------------------------------------------------

## 5. Checkpoint Engine

Stores durable state.

    .agentguard/

    current.yaml

    checkpoints/

    telemetry/

    logs/

Checkpoint contains:

-   objective
-   acceptance criteria
-   completed
-   in progress
-   pending
-   git HEAD
-   changed files
-   verification
-   next action

------------------------------------------------------------------------

## 6. Provider Layer

Abstract all coding agents.

``` go
type Provider interface {

    Execute()

    Resume()

    CollectTelemetry()

    Checkpoint()

}
```

Providers:

-   Codex
-   Claude
-   Gemini
-   Cursor
-   OpenCode

------------------------------------------------------------------------

# Data Storage

SQLite

Tables:

-   sessions
-   turns
-   telemetry
-   predictions
-   outcomes
-   checkpoints
-   providers

------------------------------------------------------------------------

# Runtime Prediction

Three independent risks.

1.  Usage quota risk
2.  Context pressure
3.  Task completion risk

Overall:

    overall = max(
        quotaRisk,
        contextRisk,
        completionRisk,
    )

------------------------------------------------------------------------

# UX

Low

Run immediately.

Medium

Show warning.

High

Recommend checkpoint.

Critical

Checkpoint by default.

------------------------------------------------------------------------

# Future

Eventually AgentGuard can answer:

-   Should I checkpoint?
-   Should I split this task?
-   Should I switch models?
-   Should I create another agent?
-   Should I delay execution?
-   Which provider is cheapest?
-   Which provider has the highest completion probability?

------------------------------------------------------------------------

# Why Go

This is a developer infrastructure tool, closer to:

-   git
-   lazygit
-   fzf
-   gh
-   kubectl
-   helm

than to an ML framework.

------------------------------------------------------------------------

# Research Pipeline

Python only.

    Telemetry

    ↓

    Notebook

    ↓

    Training

    ↓

    Evaluation

    ↓

    Export

    ↓

    Go Runtime

Model artifacts:

-   JSON coefficients
-   ONNX
-   tree model

Runtime never depends on Python.

------------------------------------------------------------------------

# Roadmap

## Phase 1

-   Telemetry
-   Rule predictor
-   SQLite
-   Codex adapter
-   Claude adapter
-   Checkpoint
-   CLI

## Phase 2

-   VS Code extension
-   Better prediction
-   Dashboard
-   Calibration

## Phase 3

-   ML predictor
-   Multi-provider routing
-   Task splitting
-   Team policy
-   Cloud sync

------------------------------------------------------------------------

# Success Criteria

A successful AgentGuard project should become:

> The runtime operating system for AI coding agents.

rather than merely another checkpoint/resume utility.
