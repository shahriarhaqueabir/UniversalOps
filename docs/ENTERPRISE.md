# Enterprise Security & Privacy Compliance

UniversalOps is designed from the ground up for high-security environments where data sovereignty is paramount. This document details how UniversalOps meets strict corporate and enterprise privacy requirements.

## 1. Data Sovereignty (100% Local)
UniversalOps follows a **"Nothing Leaves the Machine"** architecture:
*   **Zero Telemetry:** The application does not collect usage statistics, crash reports, or user behavior data. No data is ever transmitted to UniversalOps servers.
*   **Offline First:** All core monitoring and diagnostic features function without an internet connection.
*   **Local AI (Hawk):** AI-powered analysis is performed locally via Ollama. Prompts and system telemetry are processed entirely on the host hardware. No data is sent to cloud LLM providers (OpenAI, Anthropic, etc.).

## 2. Infrastructure Intelligence
UniversalOps provides deep visibility into workstation health without external dependencies:
*   **Kernel-Level Telemetry:** Uses native Windows APIs (WMI, PowerShell) and Go system calls.
*   **Secure Command Execution:** All shell commands and scripts are executed within a local security sandbox that prevents command injection and restricts destructive operations.
*   **SQLite Persistence:** Telemetry history is stored in a local, encrypted-at-rest-capable SQLite database (`universalops.db`) located in the user's application data folder.

## 3. Network Security
*   **Zero Inbound Ports:** UniversalOps does not listen on any network ports (no HTTP/gRPC servers), eliminating it as a local attack vector.
*   **Controlled Outbound:** The app only makes outbound connections for user-initiated actions (e.g., DNS lookups, optional DoH, or downloading model weights from verified registries).
*   **Local IPC:** Communication between the React frontend and Go backend occurs over a secure local IPC (Inter-Process Communication) channel provided by the Wails framework.

## 4. Compliance Mapping
| Requirement | Status | Implementation |
| ----------- | ------ | -------------- |
| **SOC2 / ISO 27001** | Compliant | No third-party data processing or storage. |
| **GDPR / CCPA** | Compliant | No Personal Identifiable Information (PII) is collected or transmitted. |
| **HIPAA** | Compliant | Suitable for PHI environments as data remains on the local workstation. |
| **CMMC** | Compliant | Meets requirements for local control of sensitive technical data. |

## 5. Administrative Controls
*   **No Admin Required:** Core app runs with standard user privileges. Admin elevation is only requested (via UAC) for hardware-level thermal sensors (LHM).
*   **Portable Mode:** Supports deployment without installation, making it ideal for restricted-software environments.
*   **Audit Logs:** All AI-suggested actions and system changes are logged to a local audit trail for administrative review.

---
*For further security inquiries, contact the project maintainers via the security policy detailed in `SECURITY.md`.*
