#!/bin/bash

# Docker 镜像加速器配置脚本
# 解决国内无法访问 Docker Hub 的问题

set -e

echo "🐳 配置 Docker 镜像加速器..."

# 检查是否有 root 权限
if [ "$EUID" -ne 0 ]; then
  echo "❌ 请使用 sudo 运行此脚本"
  exit 1
fi

# 备份原配置
if [ -f /etc/docker/daemon.json ]; then
  echo "📦 备份原配置文件..."
  cp /etc/docker/daemon.json /etc/docker/daemon.json.backup.$(date +%Y%m%d_%H%M%S)
fi

# 创建配置目录
mkdir -p /etc/docker

# 写入配置
echo "✍️  写入镜像加速器配置..."
cat > /etc/docker/daemon.json <<'EOF'
{
  "registry-mirrors": [
    "https://docker.mirrors.ustc.edu.cn",
    "https://hub-mirror.c.163.com",
    "https://mirror.baidubce.com",
    "https://docker.m.daocloud.io"
  ],
  "log-driver": "json-file",
  "log-opts": {
    "max-size": "10m",
    "max-file": "3"
  }
}
EOF

echo "🔄 重启 Docker 服务..."
systemctl daemon-reload
systemctl restart docker

echo "✅ Docker 镜像加速器配置完成！"
echo ""
echo "📊 验证配置："
docker info | grep -A 10 "Registry Mirrors" || echo "配置已生效"

echo ""
echo "🧪 测试拉取镜像："
docker pull hello-world

echo ""
echo "🎉 配置成功！现在可以正常拉取镜像了。"
