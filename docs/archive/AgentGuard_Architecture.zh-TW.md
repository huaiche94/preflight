> **ARCHIVED — 已過時。** 這是目前專案早期、使用不同名稱（「AgentGuard」）的前身文件。已被 `Preflight_ADD.md` 完整取代，該文件才是唯一權威的架構文件。僅保留作為歷史參考；請勿依此檔案進行實作。

# AgentGuard --- AI 程式碼代理人的執行期防護工具

> 🌐 [English](AgentGuard_Architecture.md) | 繁體中文
>
> 本文件為非規範翻譯，內容以英文版為準（ADR-049）。本文件屬歷史封存，非現行實作指引。

## Vision

AgentGuard **不是**另一個檢查點管理工具，也不是代理人記憶系統。

它的目的是成為 **AI 程式碼代理人的執行期作業系統（runtime operating system）**，位於使用者與 Codex CLI、Claude Code、Gemini CLI、Cursor CLI 等供應商之間。

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

核心價值：

-   在回合（turn）開始前預測執行風險。
-   只在有正當理由時才建議建立檢查點。
-   從歷史執行遙測資料中學習。
-   最終優化路由、任務拆分與執行期政策。

------------------------------------------------------------------------

# Product Positioning

## 我們不是什麼

-   代理人記憶（Agent Memory）
-   另一個接續（resume）工具
-   另一個檢查點工具
-   另一個提示詞管理工具

已經有許多專案佔據了這個領域的大部分位置。

我們的差異化在於**預測（prediction）**，而非持久化（persistence）。

------------------------------------------------------------------------

# Core Principle

現有工具回答的問題是：

> 「我們該如何繼續？」

AgentGuard 回答的問題是：

> 「我們究竟該不該開始這個回合？」

------------------------------------------------------------------------

# Long-term Goal

成為多個 AI 程式碼代理人共用的執行期中介層（middleware）。

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

正式執行期（production runtime）：**Go**

研究與實驗：**Python**

理由：

-   Go 非常適合開發者工具。
-   單一靜態執行檔。
-   優秀的 CLI 生態系。
-   容易透過 Homebrew/Winget 散布。
-   強大的並行（concurrency）能力。
-   記憶體佔用小。

保留 Python 僅用於：

-   特徵工程
-   notebooks
-   離線評估
-   ML 實驗
-   匯出模型權重

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

收集項目：

-   提示詞長度
-   輸入／輸出 tokens
-   工具呼叫（tool calls）
-   讀取的檔案
-   修改的檔案
-   新增／刪除的行數
-   git diff
-   執行耗時
-   建置／測試結果
-   重試次數
-   工作階段（session）中繼資料

持久化至 SQLite。

------------------------------------------------------------------------

## 2. Scope Estimator

在執行前估算：

-   可能讀取的檔案
-   可能變更的檔案
-   變更的程式碼行數（LOC）
-   所需的整合測試
-   相依擴散範圍（dependency fan-out）
-   不確定性分數

輸出：

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

初期採用規則式（rule-based）。

之後：

-   分位數迴歸（quantile regression）
-   梯度提升樹（gradient boosted trees）
-   邏輯迴歸（logistic regression）

介面：

``` go
type Predictor interface {
    Estimate(input PredictionInput) PredictionResult
}
```

實作：

-   RulePredictor
-   StatisticalPredictor
-   MLPredictor
-   RemotePredictor

------------------------------------------------------------------------

## 4. Policy Engine

消費（consume）預測結果。

回傳：

-   RUN
-   WARN
-   CHECKPOINT
-   SPLIT
-   ABORT

範例：

    IF
    rolling_usage > 70%
    AND
    estimated_files > 8
    AND
    integration_tests == true

    => CHECKPOINT

------------------------------------------------------------------------

## 5. Checkpoint Engine

儲存持久狀態。

    .agentguard/

    current.yaml

    checkpoints/

    telemetry/

    logs/

檢查點內容包含：

-   目標（objective）
-   驗收標準
-   已完成項目
-   進行中項目
-   待處理項目
-   git HEAD
-   變更的檔案
-   驗證結果
-   下一步動作

------------------------------------------------------------------------

## 6. Provider Layer

抽象化所有程式碼代理人。

``` go
type Provider interface {

    Execute()

    Resume()

    CollectTelemetry()

    Checkpoint()

}
```

供應商：

-   Codex
-   Claude
-   Gemini
-   Cursor
-   OpenCode

------------------------------------------------------------------------

# Data Storage

SQLite

資料表：

-   sessions
-   turns
-   telemetry
-   predictions
-   outcomes
-   checkpoints
-   providers

------------------------------------------------------------------------

# Runtime Prediction

三種獨立的風險。

1.  用量配額風險（usage quota risk）
2.  上下文壓力（context pressure）
3.  任務完成風險（task completion risk）

整體：

    overall = max(
        quotaRisk,
        contextRisk,
        completionRisk,
    )

------------------------------------------------------------------------

# UX

Low

立即執行。

Medium

顯示警告。

High

建議建立檢查點。

Critical

預設建立檢查點。

------------------------------------------------------------------------

# Future

最終 AgentGuard 應能回答：

-   我該建立檢查點嗎？
-   我該拆分這個任務嗎？
-   我該切換模型嗎？
-   我該建立另一個代理人嗎？
-   我該延後執行嗎？
-   哪個供應商最便宜？
-   哪個供應商的完成機率最高？

------------------------------------------------------------------------

# Why Go

這是一個開發者基礎設施工具，性質更接近：

-   git
-   lazygit
-   fzf
-   gh
-   kubectl
-   helm

而非機器學習框架。

------------------------------------------------------------------------

# Research Pipeline

僅使用 Python。

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

模型產出物（model artifacts）：

-   JSON 係數
-   ONNX
-   樹模型（tree model）

執行期絕不依賴 Python。

------------------------------------------------------------------------

# Roadmap

## Phase 1

-   遙測
-   規則式預測器
-   SQLite
-   Codex 轉接器（adapter）
-   Claude 轉接器
-   檢查點
-   CLI

## Phase 2

-   VS Code 擴充套件
-   更好的預測
-   儀表板
-   校準（calibration）

## Phase 3

-   ML 預測器
-   多供應商路由
-   任務拆分
-   團隊政策
-   雲端同步

------------------------------------------------------------------------

# Success Criteria

一個成功的 AgentGuard 專案應該成為：

> AI 程式碼代理人的執行期作業系統。

而不僅僅是另一個檢查點／接續（resume）工具。
