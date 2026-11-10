#!/bin/bash
# SmartCDN 宝塔 Go 项目一键安装脚本
# 用法：bash install.sh [版本号]
# 要求：已安装宝塔面板 + Go 环境（建议 Go 1.23+）

set -euo pipefail

SMARTCDN_VERSION="${1:-latest}"
INSTALL_DIR="/www/server/smartcdn"
DATA_DIR="/www/wwwroot/smartcdn"
SERVICE_NAME="smartcdn"
BINARY="smartcdn"

GOPATH_BIN="$(go env GOPATH)/bin"
export PATH="$PATH:$GOPATH_BIN"

echo "=== SmartCDN 宝塔安装脚本 (版本: $SMARTCDN_VERSION) ==="

# 1. 创建目录
mkdir -p "$INSTALL_DIR" "$DATA_DIR/migrations" "$DATA_DIR/certs"

# 2. 编译主程序
echo "[1/5] 编译 SmartCDN..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

cd "$PROJECT_ROOT"
CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INSTALL_DIR/$BINARY" ./cmd/smartcdn
CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INSTALL_DIR/node-agent" ./cmd/node-agent
CGO_ENABLED=0 go build -ldflags="-s -w" -o "$INSTALL_DIR/probe-agent" ./cmd/probe-agent
cp -r migrations "$DATA_DIR/migrations/"
echo "    ✓ 编译完成"

# 3. 创建配置文件（若不存在）
ENV_FILE="/etc/smartcdn/smartcdn.env"
if [ ! -f "$ENV_FILE" ]; then
    mkdir -p /etc/smartcdn
    RANDOM_KEY="$(openssl rand -hex 32)"
    cat > "$ENV_FILE" <<EOF
# SmartCDN 环境变量
PORT=8080
SMARTCDN_ENCRYPT_KEY=$RANDOM_KEY
DB_DSN=smartcdn:CHANGE_ME@tcp(127.0.0.1:3306)/smartcdn?charset=utf8mb4&parseTime=True&loc=Local
REDIS_URL=redis://127.0.0.1:6379/0
LOG_LEVEL=info
EOF
    chmod 600 "$ENV_FILE"
    echo "    ✓ 已生成 $ENV_FILE（请修改 DB_DSN 中的数据库密码）"
fi

# 4. 安装 systemd 服务
echo "[3/5] 安装 systemd 服务..."
cp "$SCRIPT_DIR/../systemd/smartcdn.service" /etc/systemd/system/smartcdn.service
# 修正 WorkingDirectory 为实际安装目录
sed -i "s|WorkingDirectory=.*|WorkingDirectory=$INSTALL_DIR|" /etc/systemd/system/smartcdn.service
sed -i "s|ExecStart=.*|ExecStart=$INSTALL_DIR/$BINARY|" /etc/systemd/system/smartcdn.service

systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
echo "    ✓ 服务已注册"

# 5. 执行数据库迁移
echo "[4/5] 执行数据库迁移..."
if command -v goose &>/dev/null; then
    DB_DSN_VAL="$(grep DB_DSN $ENV_FILE | cut -d= -f2-)"
    goose -dir "$DATA_DIR/migrations" mysql "$DB_DSN_VAL" up && echo "    ✓ 迁移完成" || echo "    ⚠ 迁移失败，请手动执行"
else
    echo "    ⚠ goose 未安装，跳过自动迁移。请手动执行："
    echo "       go install github.com/pressly/goose/v3/cmd/goose@latest"
    echo "       goose -dir $DATA_DIR/migrations mysql \"\$DB_DSN\" up"
fi

# 6. 启动服务
echo "[5/5] 启动 SmartCDN..."
systemctl start "$SERVICE_NAME"
sleep 2
if systemctl is-active --quiet "$SERVICE_NAME"; then
    echo ""
    echo "=== 安装完成 ==="
    echo "    SmartCDN 正在运行，访问 http://YOUR_IP:8080"
    echo "    查看日志：journalctl -u smartcdn -f"
    echo "    配置文件：$ENV_FILE"
else
    echo ""
    echo "⚠ 服务启动失败，查看日志："
    journalctl -u "$SERVICE_NAME" --no-pager -n 20
fi
