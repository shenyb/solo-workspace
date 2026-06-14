#!/bin/bash
# ─── start.sh ───
# 用法: ./bin/start.sh [dev|prod]
# 默认 dev 环境

APP_NAME="{{projectName}}"
PROFILE="${1:-dev}"
JAR_DIR="$(cd "$(dirname "$0")/.." && pwd)"
JAR_FILE="${JAR_DIR}/{{projectName}}-web/target/${APP_NAME}-web-1.0.0.jar"
LOG_DIR="${JAR_DIR}/logs"

mkdir -p "${LOG_DIR}"

JAVA_OPTS="-Xms256m -Xmx512m"
GC_OPTS="-XX:+UseG1GC -XX:+PrintGCDetails -XX:+PrintGCDateStamps -Xloggc:${LOG_DIR}/gc.log"

echo "Starting ${APP_NAME} (profile=${PROFILE})..."
nohup java ${JAVA_OPTS} ${GC_OPTS} \
    -jar "${JAR_FILE}" \
    --spring.profiles.active="${PROFILE}" \
    > "${LOG_DIR}/console.log" 2>&1 &

PID=$!
echo $PID > "${LOG_DIR}/app.pid"
echo "PID: ${PID}"
