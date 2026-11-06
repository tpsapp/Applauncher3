# Goal: Build a Modern Cross-Platform Application Launcher ("AppLauncher3")

Create a lightweight, modern, cross-platform desktop utility (Windows, Linux, macOS) that automates the sequential, staggered launching of applications, scripts, and system commands. 

This project is a modern reimagining of a legacy Windows Forms tool ("AppLauncher2"). It preserves the original tool's simplicity while adding cross-platform process management, modern UI ergonomics, multiple profiles, and resilient configuration handling.

---

### 1. Technology Stack Recommendations
Choose one of the following stacks (recommend the cleanest and most lightweight fit):
- **Option A (C# / .NET 8+):** Avalonia UI or Uno Platform (MVVM pattern), cross-platform `System.Diagnostics.Process`, compiled as a native Ahead-of-Time (AOT) or self-contained single-file binary.
- **Option B (Rust / Web):** Tauri + Lightweight Frontend (Svelte or Vue + TypeScript), using native OS process spawning via Rust's `std::process`.
- **Option C (Go):** Wails or Fyne, providing zero-dependency portable binaries across all three operating systems.

---

### 2. Core Features (Recreating AppLauncher2)

1. **Sequential Staggered Execution Queue**:
   - Queue of applications launched sequentially.
   - Per-application configurable delay/wait interval (in seconds or milliseconds) before triggering the next application in queue.
   - Real-time progress bar, item counter (`Step X of Y`), and currently active application display.
   - Interactive controls: `Start`, `Stop / Cancel`, and `Pause / Resume`.

2. **Per-Item Process Configuration**:
   - **Display Name**: Friendly label.
   - **Target Path / Command**: Full path to binary, script, or executable. Includes a native OS file picker that automatically deduces the item's name and working directory upon selection.
   - **Command-Line Arguments**: Space-delimited or array-based argument list with proper quote escaping.
   - **Working Directory**: Working directory path (auto-inferred or user-specified).
   - **Wait / Delay Time**: Delay interval after launch before proceeding to the next item.
   - **Launch State**: Start minimized, hidden, or standard window state where supported by the host OS.

3. **Queue Reordering & Management**:
   - Add, Edit, Insert at index, and Delete queue items.
   - Visual reordering (drag-and-drop or Move Up / Move Down buttons).

4. **CLI Flags & Headless Support**:
   - `--autostart` / `-a`: Automatically begin executing the queue immediately on application startup.
   - `--config <path>` / `-c <path>`: Load a custom profile/config file.
   - `--debug` / `-d`: Write granular diagnostic trace logs to a local file (`launcher-debug.log`).
   - `--dry-run`: Simulate the sequence and delays without actually spawning processes.
   - `--minimized` / `-m`: Launch the launcher itself minimized to the system tray.

---

### 3. Modern Enhancements & Cross-Platform Requirements

1. **Cross-Platform OS Abstraction**:
   - **Windows**: Support `.exe`, `.bat`, `.cmd`, `.lnk` shortcuts, and PowerShell scripts.
   - **macOS**: Support `.app` bundles (via `open -a` or direct binary execution), shell scripts (`.sh`, `.command`), and Mach-O binaries.
   - **Linux**: Support desktop entries (`.desktop`), standard ELF binaries, and shell scripts (`.sh`).
   - **Process Independence**: Launch processes detached so that closing the launcher does not terminate spawned applications (unless explicitly toggled by the user).

2. **Profile & Preset Management**:
   - Allow saving and switching between named launch profiles (e.g., "Workstation Setup", "Gaming", "Development Environment", "Media Kiosk") rather than restricting to a single configuration.

3. **Configuration Storage**:
   - Store settings in human-readable JSON or YAML in the OS standard user configuration folder:
     - Windows: `%APPDATA%/AppLauncher3/`
     - Linux: `~/.config/applauncher3/`
     - macOS: `~/Library/Application Support/AppLauncher3/`
   - Support a portable mode (reads `config.json` next to the executable if present).
   - Zero administrative/elevated privileges required.

4. **Advanced Launch Conditions (Optional per Item)**:
   - Delay by fixed time (default).
   - Wait until process exits before proceeding (blocking mode).
   - Wait for a local TCP port or health check endpoint to respond before continuing (useful for starting a backend before launching a frontend).

5. **UI / UX Standards**:
   - Clean, modern layout with system dark/light theme support.
   - System tray / menu bar integration with quick-launch profile menu and completion notifications.
   - Per-item status indicator icons: Pending, Running, Succeeded (with exit code), or Failed.
   - Built-in error logging panel showing detailed stderr/stdout when a process fails to launch.

---

### 4. Implementation Steps
1. Define the core data model (`LaunchItem`, `LaunchProfile`, `LauncherConfig`) and schema serialization.
2. Implement the cross-platform process runner engine with async delay loops and cancellation tokens.
3. Build the primary UI:
   - Queue list view with status indicators.
   - Add/Edit modal dialog with file picker.
   - Control panel with Start/Stop buttons and progress bar.
4. Implement CLI argument parsing and auto-start execution logic.
5. Add unit and integration tests for process spawning, delay sequencing, and config serialization across platforms.
