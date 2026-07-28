# ADR-0005: Local AI via Ollama for Autonomous RCA

**Date**: 2026-07-28
**Status**: Accepted
**Deciders**: @shahriarhaqueabir

## Context

UniversalOps aimed to provide not just raw telemetry but actionable intelligence — automatically analyzing system state, detecting anomalies, and suggesting remediation steps. Cloud-based AI solutions (OpenAI, Anthropic, etc.) were incompatible with the "100% local, zero telemetry" principle. The AI integration needed to run entirely on the user's machine, support structured data analysis (metrics, logs, alerts), and work offline.

## Decision

Integrate with **Ollama** as the local AI runtime, using a custom model (`universalops` based on Qwen 2.5) fine-tuned for system operations analysis. The AIOps layer (`internal/aiops/`) provides autonomous root cause analysis (RCA), anomaly detection, state querying, and remediation workflow generation — all via Ollama's REST API at `http://localhost:11434`.

## Alternatives Considered

### Alternative 1: Cloud AI APIs (OpenAI, Anthropic, Gemini)
- **Pros**: Most capable models, no local hardware requirements, always up-to-date
- **Cons**: Sends system telemetry to third-party servers, requires internet, ongoing API costs, violates privacy-first principle
- **Why not**: Fundamentally incompatible with the project's core value proposition of 100% local operation

### Alternative 2: llama.cpp embedded (pure Go bindings)
- **Pros**: No external process needed, fully self-contained, maximum privacy
- **Cons**: Significant binary size increase (~4-8GB for a useful model), complex cross-compilation, GPU acceleration requires CGo/CUDA dependencies
- **Why not**: The binary size and build complexity were unacceptable. Ollama handles model management, GPU acceleration, and API serving — we'd be recreating that infrastructure

### Alternative 3: ONNX Runtime + small model
- **Pros**: Cross-platform, smaller footprint than full LLM
- **Cons**: Limited to smaller models with reduced reasoning capability, less suitable for complex RCA tasks
- **Why not**: System operations RCA requires multi-step reasoning (correlating CPU spikes with network drops with recent process launches) — small models struggled with this complexity

## Consequences

- **Easier**: Ollama handles model download, quantization, GPU acceleration, and API serving. Users can swap models without changing UniversalOps. The REST API is simple and well-documented. Falls back gracefully if Ollama is not installed.
- **Harder**: Users must install Ollama separately (~500MB). Model download requires internet on first use. Performance depends on local hardware (CPU/GPU). The custom `universalops` modelfile must be maintained alongside the application.
- **Risks**: Ollama API changes could break integration. Large models (30B+ parameters) may be too slow for real-time analysis on consumer hardware. Mitigated by model fallback chain and configurable model selection.