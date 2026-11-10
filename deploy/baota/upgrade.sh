#!/bin/bash
# SmartCDN 宝塔热升级脚本（零停机：编译新版后原子替换）
set -euo pipefail

INSTALL_DIR="/www/server/smartcdn"
SERVICE_NAME="smartcdn"
BINARY="smartcdn"
TMP_BIN="$INSTALL_DIR/${BINARY}.new"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=== SmartCDN 升级 ==="

cd "$PROJECT_ROOT"

echo "[1/3] 编译新版本..."
CGO_ENABLED=0 go build -ldflags="-s -w" -o "$TMP_BIN" ./cmd/smartcdn
echo "    ✓ 编译完成"

echo "[2/3] 执行数据库迁移..."
ENV_FILE="/etc/smartcdn/smartcdn.env"
if command -v goose &>/dev/null && [ -f "$ENV_FILE" ]; then
    DB_DSN_VAL="$(grep DB_DSN $ENV_FILE | cut -d= -f2-)"
    goose -dir "$INSTALL_DIR/migrations" mysql "$DB_DSN_VAL" up && echo "    ✓ 迁移完成"
else
    echo "    ⚠ 跳过迁移"
fi

echo "[3/3] 原子替换并重启服务..."
mv "$TMP_BIN" "$INSTALL_DIR/$BINARY"
systemctl restart "$SERVICE_NAME"
sleep 2
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "    ✓ 升级成功"
else
    echo "    ✗ 重启失败，请检查日志：journalctl -u $SERVICE_NAME -n 30"
    exit 1
fi
