#!/bin/bash
set -e

# 配置
BACKEND_DIR="$(cd "$(dirname "$0")" && pwd)"
FRONTEND_DIR="$(cd "$BACKEND_DIR/../desk-app" && pwd)"
BUILD_DIR="$FRONTEND_DIR/resources/server"

echo "========================================"
echo "  Village Helper Windows 打包脚本"
echo "========================================"
echo ""

# 检查前端项目是否存在
if [ ! -d "$FRONTEND_DIR" ]; then
    echo "❌ 错误: 找不到前端项目目录: $FRONTEND_DIR"
    echo "   请确保 desk-app 和 village-helper-server-go 在同一目录下"
    exit 1
fi

# 1. 编译 Go 后端
echo "🔨 步骤 1/4: 编译 Go 后端..."
cd "$BACKEND_DIR"
go mod tidy
go mod download
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o backend.exe .
echo "   ✅ backend.exe 编译完成"

# 2. 复制到前端项目
echo ""
echo "📦 步骤 2/4: 复制资源到前端..."
mkdir -p "$BUILD_DIR/data"
cp "$BACKEND_DIR/backend.exe" "$BUILD_DIR/"
if [ -f "$BACKEND_DIR/data/app.db" ]; then
    cp "$BACKEND_DIR/data/app.db" "$BUILD_DIR/data/"
    echo "   ✅ 后端 exe + 数据库已复制"
else
    echo "   ⚠️ 警告: 未找到 data/app.db，将使用空数据库"
fi

# 3. 打包前端
echo ""
echo "🖥️  步骤 3/4: 构建前端..."
cd "$FRONTEND_DIR"
if [ ! -d "node_modules" ]; then
    echo "   📥 安装依赖中..."
    npm install
fi
npm run build
echo "   ✅ 前端构建完成"

# 4. 打包 Windows 安装包
echo ""
echo "📦 步骤 4/4: 打包 Windows 安装包..."
npx electron-builder --win --x64
echo "   ✅ 打包完成"

# 输出结果
echo ""
echo "========================================"
echo "  🎉 打包成功！"
echo "========================================"
echo ""

INSTALLER=$(ls -1 "$FRONTEND_DIR/dist/"*.exe 2>/dev/null | head -1)
if [ -f "$INSTALLER" ]; then
    echo "📁 安装包位置:"
    echo "   $INSTALLER"
    echo ""
    echo "📊 文件大小:"
    ls -lh "$INSTALLER" | awk '{print "   " $5 "  " $9}'
else
    echo "📁 输出目录:"
    echo "   $FRONTEND_DIR/dist/"
fi

echo ""
echo "💡 提示:"
echo "   如果第一次运行，electron-builder 会自动下载 Windows 打包工具"
echo "   （约 150MB），请耐心等待。"
echo ""
