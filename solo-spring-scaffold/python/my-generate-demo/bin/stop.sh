#!/bin/bash
# ─── stop.sh ───

APP_NAME="my-generate-demo"
JAR_DIR="$(cd "$(dirname "$0")/.." && pwd)"
PID_FILE="${JAR_DIR}/logs/app.pid"

if [ ! -f "${PID_FILE}" ]; then
    PIDS=$(pgrep -f "${APP_NAME}-web" 2>/dev/null)
    if [ -z "${PIDS}" ]; then
        echo "${APP_NAME} is not running."
        exit 0
    fi
    echo "Found PIDs: ${PIDS}. Killing..."
    kill ${PIDS} 2>/dev/null
    echo "Stopped."
    exit 0
fi

PID=$(cat "${PID_FILE}")
echo "Stopping ${APP_NAME} (PID=${PID})..."
kill "${PID}" 2>/dev/null

for i in $(seq 1 10); do
    if ! kill -0 "${PID}" 2>/dev/null; then
        echo "Stopped."
        rm -f "${PID_FILE}"
        exit 0
    fi
    sleep 1
done

echo "Force killing..."
kill -9 "${PID}" 2>/dev/null
rm -f "${PID_FILE}"
echo "Force stopped."
