#!/bin/bash
# ─── deploy.sh ───
# 打包 → scp 上传 → SSH 重启
#
# 用法:
#   1. 先配置以下变量
#   2. ./bin/deploy.sh

# ─── 服务器配置（请修改） ───
SERVER_HOST="{{SERVER_HOST}}"
SERVER_USER="{{SERVER_USER}}"
DEPLOY_PATH="{{DEPLOY_PATH}}"       # e.g. /home/{{SERVER_USER}}/my-generate-demo

# ─── 开始部署 ───
set -e

echo "=== 1. 打包 ==="
cd "$(dirname "$0")/.."
mvn clean package -DskipTests -q

JAR_FILE="my-generate-demo-web/target/my-generate-demo-web-1.0.0.jar"

echo "=== 2. 上传 ==="
ssh "${SERVER_USER}@${SERVER_HOST}" "mkdir -p ${DEPLOY_PATH}"
scp "${JAR_FILE}" "${SERVER_USER}@${SERVER_HOST}:${DEPLOY_PATH}/"

echo "=== 3. 重启 ==="
ssh "${SERVER_USER}@${SERVER_HOST}" "
    cd ${DEPLOY_PATH}
    ./bin/stop.sh 2>/dev/null || true
    ./bin/start.sh prod
"

echo "=== 4. 查看日志 ==="
ssh "${SERVER_USER}@${SERVER_HOST}" "tail -f ${DEPLOY_PATH}/logs/console.log"
