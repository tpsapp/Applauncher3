# AppLauncher3 — Specification & Architecture Plan

## 1. Executive Scope & Non-Goals

### 1.1 Project Summary
**AppLauncher3** is a lightweight, cross-platform (Windows, Linux, macOS) desktop utility engineered to automate the sequential, staggered launching of applications, scripts, and system commands. It reimagines the legacy C# / .NET 3.5 Windows Forms utility (*AppLauncher2*) by providing a modern reactive UI, cross-platform OS detachment, robust multi-profile management, rich system tray ergonomics, and CLI automation.

### 1.2 Scope Boundaries (v1 MVP)
* **Sequential Staggered Queue:** Async execution loop running configured items in user-defined order with configurable per-item delays (milliseconds/seconds).
* **Process Independence & Detachment:** Child processes spawn completely detached (`SysProcAttr` process group separation) so terminating or restarting the launcher leaves child applications unharmed.
* **Per-Item Configuration:** Display name, executable/command target path, arguments, working directory, delay duration, and initial window state (Normal, Minimized, Hidden).
* **Multi-Profile Management:** Create, rename, clone, switch, and delete named execution profiles (e.g., "Workstation", "Development", "Gaming").
* **Queue Reordering & Manipulation:** Drag-and-drop reordering, move up/down, add, edit, insert at index, and delete items.
* **Rich System Tray:** Minimize-to-tray, quick profile switching menu, one-click execution trigger, and dynamic tooltip/menu execution status tracking.
* **CLI Automation & Diagnostic Logging:** `--autostart` (`-a`), `--config <path>` (`-c`), `--debug` (`-d`), `--dry-run`, and `--minimized` (`-m`).
* **Resilient Storage:** Human-readable `config.json` with JSON Schema validation, atomic disk writes (`temp` file + atomic rename), OS configuration folder defaults, and portable directory fallback.
* **Fault Policy (Continue with Warning):** If a process fails to launch (e.g. missing executable or permission error), the engine logs the diagnostic error, marks the item status as `Failed`, emits a desktop notification, and proceeds to the subsequent queue item after the specified delay.

### 1.3 Explicit Non-Goals (v1 Out-of-Scope)
* **TCP / HTTP Health Probes:** Socket readiness checks or HTTP 200 ping polling before advancing (slated for v1.1).
* **Process Exit Blocking:** Pausing the queue until a previous application terminates (slated for v1.1).
* **Remote Sync / Cloud Backends:** Zero telemetry, cloud sync, or remote accounts. Configuration remains strictly local.
* **Embedded Cron/Scheduler Service:** AppLauncher3 delegates recurring background triggers to standard OS schedulers (systemd timers, Windows Task Scheduler, launchd) via CLI flags.

---

## 2. Technical Architecture Specification

```mermaid
graph TD
    subgraph UI Layer [Frontend: Svelte 5 + TypeScript + Tailwind CSS]
        View[Main Window / Queue View]
        EditModal[Item Configuration & Native File Picker]
        ProfileBar[Profile Selector & Management]
        ProgressCard[Real-Time Progress & Telemetry]
    end

    subgraph IPC Bridge [Wails v2 Binding Layer]
        Bridge[Wails JS/Go IPC Bridge]
        Events[Go Event Emitter: QueueProgress, StatusEvents]
    end

    subgraph Core Engine [Go 1.22+ Backend]
        Engine[Execution Engine & Scheduler]
        ConfigMgr[Profile & Config Manager]
        CLIMgr[CLI Argument Parser]
        Logger[Diagnostic Trace Logger]
        TrayMgr[Native System Tray & Notification Controller]
    end

    subgraph OS Abstraction [Platform Process Spawner]
        WinSpawn[Windows: CREATE_NEW_PROCESS_GROUP / DETACHED]
        UnixSpawn[Linux/macOS: Setpgid / Daemonized Exec]
    end

    View <--> Bridge
    EditModal <--> Bridge
    ProfileBar <--> Bridge
    ProgressCard <--> Bridge
    Bridge <--> Engine
    Bridge <--> ConfigMgr
    Engine --> Events
    Events --> Bridge
    Engine --> OS Abstraction
    CLIMgr --> Engine
    TrayMgr <--> Engine
```

### 2.1 Technology Stack & Rationale
* **Backend Runtime:** **Go 1.22+**
  * *Rationale:* Produces zero-dependency native binaries across Windows, Linux, and macOS. High-performance asynchronous execution with native Goroutines and standard `os/exec` cross-platform hooks.
* **Application Framework:** **Wails v2**
  * *Rationale:* Binds the Go engine directly to native webview runtimes (WebView2 on Windows, WebKit2GTK on Linux, WKWebView on macOS), yielding executable sizes (~10–15 MB) and minimal idle RAM usage (~25–35 MB) without the heavyweight baggage of Electron.
* **Frontend UI Framework:** **Svelte 5 + TypeScript + Tailwind CSS**
  * *Rationale:* Svelte 5's fine-grained reactivity (Runes) provides instantaneous DOM updates for queue timers and status badges without virtual DOM diffing. Tailwind CSS ensures dark/light desktop styling.

### 2.2 Cross-Platform Process Engine & Detached Spawning
AppLauncher3 spawns processes independently of the launcher's own PID tree:

| Target Platform | Command Resolution | Detachment Mechanism (`syscall.SysProcAttr`) | Window State Controls |
| :--- | :--- | :--- | :--- |
| **Windows** | Direct `.exe`, `.bat`, `.cmd`, `.lnk`, or PowerShell scripts via `cmd.exe /c` | `CreationFlags = syscall.CREATE_NEW_PROCESS_GROUP \| 0x00000008 (DETACHED_PROCESS)` | `HideWindow = true`, `ShowWindow = SW_SHOWMINIMIZED` |
| **Linux** | Standard ELF binaries, `.desktop` files, `.sh` shell scripts | `Setpgid: true`, `Setsid: true` (creates fresh session leader, immune to parent SIGHUP) | Emits standard freedesktop hints where supported |
| **macOS** | Mach-O binaries, `.command`, `.app` bundles (via `/usr/bin/open -a`) | `Setpgid: true` | Managed via `open --hide` or `open --background` flags |

### 2.3 Data Contract & State Management
Configuration is validated against an embedded JSON schema.

#### JSON Storage Location Resolution Hierarchy:
1. `--config <path>` flag passed via CLI.
2. Portable Mode: `./config.json` adjacent to the executable if present.
3. OS User Directory:
   * **Linux:** `${XDG_CONFIG_HOME:-~/.config}/applauncher3/config.json`
   * **Windows:** `%APPDATA%\AppLauncher3\config.json`
   * **macOS:** `~/Library/Application Support/AppLauncher3/config.json`

#### Configuration Schema (`config.json`):
```json
{
  "$schema": "https://raw.githubusercontent.com/tpsapp/applauncher3/main/schema.json",
  "version": 1,
  "activeProfile": "Default",
  "settings": {
    "startMinimized": false,
    "closeToTray": true,
    "confirmOnCancel": true,
    "theme": "system",
    "debugLogging": false
  },
  "profiles": [
    {
      "name": "Default",
      "items": [
        {
          "id": "c1f7a28e-8b92-49f8-b30a-3a1e2f893a01",
          "name": "Web Browser",
          "path": "/usr/bin/google-chrome-stable",
          "arguments": ["--new-window", "https://github.com"],
          "workingDirectory": "",
          "delaySeconds": 2.5,
          "launchState": "Normal",
          "enabled": true
        }
      ]
    }
  ]
}
```

#### Safe Persistence Strategy:
* Atomic Write: State changes are serialized, written to `config.json.tmp.<pid>`, fsynced, and renamed over `config.json` to prevent corruption during sudden power-off or OS crashes.

---

## 3. Work Breakdown Structure (WBS)

### Phase 1: Core Engine & Data Contracts (Go Backend)
* **1.1 Data Models & Validation:** Define `LauncherConfig`, `Profile`, and `LaunchItem` structs with custom JSON marshalling, default initializers, and UUID generation.
* **1.2 Config Repository:** Implement config resolution (Portable vs OS directory), atomic file read/write, schema validation, and config migrations.
* **1.3 Process Spawner & OS Abstraction:** Build platform-specific spawners (`spawner_windows.go`, `spawner_linux.go`, `spawner_darwin.go`) ensuring independent detachment and working directory inference.
* **1.4 Execution Loop & State Machine:** Implement the asynchronous queue worker supporting states (`Idle`, `Running`, `Paused`, `Cancelling`) using Go contexts and channels.
* **1.5 Unit & Engine Testing:** Validate sequence timing, delay pauses, missing binary graceful warnings, and config serialization under Go's testing harness.

### Phase 2: CLI Automation & Diagnostics
* **2.1 Flag Parser:** Integrate CLI flag handling for `--autostart`, `--config`, `--debug`, `--dry-run`, and `--minimized`.
* **2.2 Trace Logging Engine:** Build thread-safe `launcher-debug.log` writer activated by `-d`/`--debug` capturing process stdout/stderr error streams and execution milestones.
* **2.3 Dry-Run Simulator:** Ensure `--dry-run` executes exact delay queues and logs synthetic process calls without triggering OS process spawns.

### Phase 3: Modern Desktop UI (Svelte 5 + TypeScript + Tailwind)
* **3.1 Layout & Theme Engine:** Modern shell featuring header, profile dropdown switcher, queue progress bar, and system dark/light theme detection.
* **3.2 Reactive Queue Grid:** Queue view showing status badges (`Pending`, `Launching`, `Running`, `Failed`, `Skipped`), drag-and-drop handles, and action buttons (`Move Up`, `Move Down`, `Delete`).
* **3.3 Add/Edit Item Modal:** Modal dialog with executable path input, argument tags, delay slider/input, window state selector, and native OS file browser picker integration (`wails/runtime.OpenFileDialog`).
* **3.4 Control Bar:** Ergonomic sticky control footer with `Start`, `Pause`, `Resume`, and `Cancel` buttons, live execution step indicator (`Step 3 of 8`), and timer countdown.

### Phase 4: System Tray & OS Ergonomics
* **4.1 System Tray Controller:** Native tray icon with dynamic context menu (Current Profile, Start Queue, Show/Hide Window, Exit).
* **4.2 Minimize-to-Tray Behavior:** Intercept close events to minimize window to tray when enabled in preferences.
* **4.3 Native Desktop Notifications:** Trigger platform-native toast notifications on queue completion or process launch failures.

### Phase 5: Verification, Packaging & Cross-Compilation
* **5.1 End-to-End Test Suite:** Automated testing of full queue runs, pause/resume state transitions, and config round-tripping.
* **5.2 Cross-Platform Build Pipelines:** Configure Wails packaging targets for Windows (`.exe` with application manifest), Linux (`.deb`, `.tar.gz`, AppImage), and macOS (`.app` bundle).

---

## 4. Risk Matrix

| Risk / Failure Mode | Likelihood | Impact | Architectural Mitigation Strategy |
| :--- | :--- | :--- | :--- |
| **Child Process Leaks / SIGHUP on Exit** | High | High | Explicit platform abstraction using `CREATE_NEW_PROCESS_GROUP` on Windows and `Setpgid: true` on Linux/macOS to ensure clean process detachment. |
| **WebKit2GTK Missing on Minimal Linux Runtimes** | Medium | High | Clearly document distribution prerequisites; provide comprehensive instructions and package dependencies (`libwebkit2gtk-4.1`). |
| **Configuration File Corruption on Sudden Crash** | Low | High | Atomic write pipeline (`tempfile` + `Sync` + atomic file rename) paired with automatic `.bak` snapshot generation before overwrite. |
| **Invalid Executable Paths in Queue** | High | Medium | Non-blocking execution policy: engine flags item as `Failed`, records stderr/reason to diagnostic log, sends toast alert, and advances to next item. |
| **GUI Freeze during Process Spawning** | Low | High | Queue worker runs strictly in a background Goroutine communicating over non-blocking Wails event emitters to the Svelte UI. |

---

## 5. Phase 1 Execution Checklist

- [ ] Initialize Wails v2 project structure (`wails init -n applauncher3 -t svelte-ts`).
- [ ] Configure Tailwind CSS and Lucide icons in the frontend workspace.
- [ ] Implement `pkg/models/config.go` containing models: `LauncherConfig`, `Profile`, and `LaunchItem`.
- [ ] Implement `pkg/storage/config_manager.go` with cross-platform directory resolution and atomic JSON persistence.
- [ ] Implement platform-segregated process spawning engine in `pkg/runner/`:
  - `runner.go` (common interface, async loop, state machine)
  - `spawner_windows.go` (Windows detachment flags)
  - `spawner_unix.go` (Linux & macOS session decoupling)
- [ ] Author unit tests in `pkg/runner/runner_test.go` and `pkg/storage/config_manager_test.go`.

---

**Next Logical Skill in Progression:** `architecture-spikes` (To implement and validate the cross-platform detached process spawner proof-of-concept harness before frontend wiring).

