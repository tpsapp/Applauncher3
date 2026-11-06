# Contributing to AppLauncher3

Thank you for your interest in contributing to **AppLauncher3**! AppLauncher3 is a lightweight, modern, cross-platform desktop utility that automates the sequential, staggered launching of applications, scripts, and system commands.

---

## 🧭 Core Engineering Principles

Before opening a pull request or proposing changes, keep these core principles in mind:

1. **Lightweight & High Performance:** AppLauncher3 produces a self-contained ~7.0 MB binary with minimal memory footprint (~25 MB idle). Avoid introducing heavy external frameworks or unnecessary bloat.
2. **Robust Fault Isolation:** The spawner and queue execution engine strictly adhere to the *Continue-with-Warning* policy. Application launch errors must never crash the launcher or halt subsequent tasks unexpectedly.
3. **Cross-Platform Consistency:** Changes should maintain compatibility across Linux (X11 & Wayland / KDE / GNOME), Windows 10/11, and macOS. Platform-specific code should be isolated in build-guarded files (`*_linux.go`, `*_windows.go`, `*_darwin.go`, `*_other.go`).
4. **Data Race Free:** Concurrency in Go backend engines must be strictly verifiable with ThreadSanitizer (`go test -race`).

---

## 📋 Code of Conduct

We are committed to providing a welcoming, inclusive, and harassment-free experience for everyone. Please be respectful, constructive, and courteous in all interactions, issue reports, and pull requests.

---

## 🛠️ Local Development Setup

### Prerequisites
- **Go 1.22+**
- **Node.js 18+** & npm
- **GCC / Build Essentials**
- **WebKitGTK** (on Linux: `webkit2gtk-4.1` or `webkit2gtk-4.0`)
- **Wails CLI v2** (optional, standard `make` builds without global CLI)

### Quick Setup Steps

1. **Clone the repository:**
   ```bash
   git clone https://github.com/tpsapp/Applauncher3.git
   cd Applauncher3
   ```

2. **Install frontend dependencies:**
   ```bash
   cd frontend
   npm install
   cd ..
   ```

3. **Compile application binary:**
   ```bash
   make
   ```
   The compiled binary will be placed at `bin/applauncher3`.

4. **Run tests:**
   ```bash
   make test-race
   ```

---

## 📐 Coding Standards & Guidelines

### 1. Go Backend Standards
- **Hermetic Unit Tests:** Ensure all new logic has corresponding automated tests in `pkg/*` or `tests/`.
- **Concurrency Safety:** Always run tests with race detection enabled: `go test -race ./...`.
- **Platform Separation:** Keep POSIX and Windows system-level process handling segregated behind clean interface contracts (`pkg/runner`).

### 2. Frontend Svelte 5 Standards
- **Svelte 5 Runes:** Use modern Svelte 5 runes (`$state`, `$derived`, `$props`) rather than legacy Svelte 3/4 stores or reactive assignments (`let:`, `$: `).
- **Tailwind CSS Styling:** Follow existing Tailwind design tokens and ensure all UI elements support both Dark and Light mode cleanly.

---

## 🔄 Pull Request Workflow

1. **Fork & Branch:** Create a feature branch with a descriptive name (e.g., `feat/custom-hotkeys` or `fix/tray-menu-icon`).
2. **Verify Tests:** Ensure all unit tests and builds pass:
   ```bash
   make test-race
   ```
3. **Commit Cleanly:** Write concise, conventional commit messages (e.g., `feat: ...`, `fix: ...`, `docs: ...`).
4. **Submit PR:** Open a Pull Request on GitHub with a clear description of the problem solved and test verification steps.
