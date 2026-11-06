# Security Policy

The AppLauncher3 project takes the security and integrity of user environments, system processes, and application configurations seriously. This document outlines our security policies, supported versions, vulnerability disclosure procedures, and local process isolation model.

---

## Supported Versions

Security patches and bug fixes are actively provided for the following releases:

| Version / Branch | Supported          | Go Target  | Wails Target |
|:-----------------|:------------------:|:-----------|:-------------|
| `main` (Latest)  | :white_check_mark: | Go 1.22+   | Wails v2     |
| `< 1.0.0`        | :x:                | N/A        | N/A          |

We recommend always running the latest release or building from `main` to ensure up-to-date runtime protections.

---

## Reporting a Vulnerability

If you discover a security vulnerability in AppLauncher3, please **do not report it via a public GitHub issue, pull request, or public forum**. Instead, report it privately so we can evaluate and remediate the issue promptly.

### How to Report Privately
- **Contact:** Thomas Sapp (Sappsworld Programming)
- **Email:** tom.sapp@sappsworld.com
- **Subject line format:** `[SECURITY] AppLauncher3 - Vulnerability Report: <Brief Description>`

### What to Include in Your Report
To help us triage and resolve the issue quickly, please include:
1. **Description:** A detailed explanation of the vulnerability and its potential impact.
2. **Steps to Reproduce:** Exact reproduction steps, configuration file snippet, or proof-of-concept (PoC) script.
3. **Affected Subsystems:** Specific components involved (e.g. `pkg/runner`, `pkg/storage`, `pkg/tray`, IPC bridge).
4. **Environment Details:** Operating System (Linux distribution/desktop environment, Windows version, or macOS version), Go runtime version, WebKitGTK version.
5. **Mitigation Suggestions:** Any recommended code changes or configuration mitigations, if known.

### Response Timeline & SLAs
- **Initial Acknowledgment:** Within **48 hours** of report receipt.
- **Triage & Severity Assessment:** Within **5 business days**.
- **Fix & Release Coordination:** We aim to release a patched version within **14 days** of triage confirmation, coordinating public disclosure timelines with the reporter.

---

## Security Architecture & Threat Model

AppLauncher3 is designed to run locally on desktop workstations to orchestrate process execution. Its core design emphasizes defense-in-depth:

### 1. Detached Process Execution Isolation
- Child processes are launched in independent process groups (`Setpgid: true` on POSIX systems; `CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS` on Windows).
- Executables are spawned via explicit system calls without passing through arbitrary shell interpreters (`/bin/sh -c` or `cmd.exe /c`), mitigating shell injection vulnerabilities when launching applications with arguments.

### 2. Atomic Configuration Persistence
- Configuration files are written using an atomic write pipeline: payload is written to a temporary file (`config.json.tmp.*`), committed via `fsync`, and atomically renamed to the destination.
- Corrupted configuration files are safely quarantined (`config.json.corrupt.*`) rather than blindly overwritten or crashing the application.
- Configuration directory paths are restricted to standard OS user configuration boundaries (`~/.config/applauncher3`, `%APPDATA%\applauncher3`) or local portable directory.

### 3. IPC & Frontend Boundary
- AppLauncher3 uses Wails v2 to bind Go backend logic to the Svelte 5 frontend over a local WebKit IPC bridge.
- The IPC interface exposes strongly typed RPC methods with input validation on model fields before any backend action is triggered.
- WebKit assets are embedded directly into the Go binary at compile time (`//go:embed`), eliminating external network attack surfaces or unvetted asset injection.
