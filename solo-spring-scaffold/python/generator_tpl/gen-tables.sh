#!/bin/bash
# ─── gen-tables.sh ───
# 从数据库表生成 MyBatis-Plus 代码（Entity / Mapper / Service / Controller）
#
# 前提：已配置 MySQL 连接信息，maven 依赖已下载
# 用法: ./bin/gen-tables.sh

set -e

cd "$(dirname "$0")/.."

echo "=== 编译 Generator ==="
mvn compile -pl {{projectName}}-dao -am -q

echo "=== 运行 Generator ==="
mvn exec:java -pl {{projectName}}-dao \
    -Dexec.mainClass="{{basePackage}}.generator.Generator" \
    -Dexec.classpathScope=compile \
    -q

echo "=== 完成 ==="
echo "已生成: entity / mapper / service / controller"
