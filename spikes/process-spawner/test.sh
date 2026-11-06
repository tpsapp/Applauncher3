#!/usr/bin/env bash
set -euo pipefail

PID_FILE="/tmp/applauncher3_spike1_child.pid"
rm -f "$PID_FILE"

echo "[Spike 1] Compiling spawner harness..."
go build -o /tmp/spawner spikes/process-spawner/main.go

echo "[Spike 1] Running spawner to launch 'sleep 15'..."
SPAWN_OUT=$(/tmp/spawner "$PID_FILE" sleep 15)
echo "$SPAWN_OUT"

CHILD_PID=$(cat "$PID_FILE")
echo "[Spike 1] Verifying child PID $CHILD_PID is alive after parent exit..."

if kill -0 "$CHILD_PID" 2>/dev/null; then
    echo "PASS: Child process $CHILD_PID is actively running after parent terminated."
    CHILD_PPID=$(ps -o ppid= -p "$CHILD_PID" | tr -d ' ')
    echo "INFO: Child PPID is $CHILD_PPID (reparented to subreaper/init)."
    kill "$CHILD_PID" 2>/dev/null || true
    echo "PASS: Cleaned up child process."
    rm -f "$PID_FILE" /tmp/spawner
    exit 0
else
    echo "FAIL: Child process $CHILD_PID terminated when parent exited."
    rm -f "$PID_FILE" /tmp/spawner
    exit 1
fi
