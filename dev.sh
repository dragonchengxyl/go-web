#!/usr/bin/env bash
# dev.sh — 一键本地开发环境启动脚本
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="$ROOT/.dev-logs"
mkdir -p "$LOG_DIR"

# ── 自动检测 docker compose 命令版本 ──────────────────
if docker compose version &>/dev/null 2>&1; then
  DC="docker compose"
elif command -v docker-compose &>/dev/null; then
  DC="docker-compose"
else
  DC="docker compose"  # 留给后续依赖检查报错
fi

# ── 颜色 ──────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'
BLUE='\033[0;34m'; CYAN='\033[0;36m'; NC='\033[0m'

info()    { echo -e "${BLUE}[INFO]${NC}  $*"; }
success() { echo -e "${GREEN}[OK]${NC}    $*"; }
warn()    { echo -e "${YELLOW}[WARN]${NC}  $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*" >&2; }
step()    { echo -e "\n${CYAN}▶ $*${NC}"; }

# ── 清理函数（Ctrl+C 时自动停止子进程）────────────────
BACKEND_PID=""
FRONTEND_PID=""
cleanup() {
  echo ""
  step "正在停止所有服务..."
  [[ -n "$BACKEND_PID" ]]  && kill "$BACKEND_PID"  2>/dev/null && info "后端已停止"
  [[ -n "$FRONTEND_PID" ]] && kill "$FRONTEND_PID" 2>/dev/null && info "前端已停止"
  info "日志保存在 $LOG_DIR"
  exit 0
}
trap cleanup SIGINT SIGTERM

# ── 帮助 ─────────────────────────────────────────────
usage() {
  echo ""
  echo "用法: $0 [选项]"
  echo ""
  echo "  --no-docker     跳过 Docker 基础设施启动（已有 PG/Redis 时使用）"
  echo "  --no-migrate    跳过数据库迁移"
  echo "  --backend-only  只启动后端"
  echo "  --frontend-only 只启动前端（需后端已运行）"
  echo "  --stop          停止 Docker 基础设施"
  echo "  --logs          查看实时日志"
  echo "  -h, --help      显示此帮助"
  echo ""
}

# ── 解析参数 ──────────────────────────────────────────
NO_DOCKER=false
NO_MIGRATE=false
BACKEND_ONLY=false
FRONTEND_ONLY=false

for arg in "$@"; do
  case $arg in
    --no-docker)     NO_DOCKER=true ;;
    --no-migrate)    NO_MIGRATE=true ;;
    --backend-only)  BACKEND_ONLY=true ;;
    --frontend-only) FRONTEND_ONLY=true ;;
    --stop)
      info "停止 Docker 基础设施..."
      cd "$ROOT" && $DC down
      success "已停止"
      exit 0
      ;;
    --logs)
      echo "== 后端日志 =="
      tail -f "$LOG_DIR/backend.log" &
      echo "== 前端日志 =="
      tail -f "$LOG_DIR/frontend.log"
      exit 0
      ;;
    -h|--help) usage; exit 0 ;;
    *) error "未知参数: $arg"; usage; exit 1 ;;
  esac
done

# ═══════════════════════════════════════════════════════
echo ""
echo -e "${CYAN}╔═══════════════════════════════════════╗${NC}"
echo -e "${CYAN}║     Furry 社区平台 — 本地开发环境     ║${NC}"
echo -e "${CYAN}╚═══════════════════════════════════════╝${NC}"

# ── 1. 检查依赖 ───────────────────────────────────────
step "检查依赖工具"
MISSING=()
command -v docker   &>/dev/null || MISSING+=("docker")
command -v go       &>/dev/null || MISSING+=("go")
command -v pnpm     &>/dev/null || MISSING+=("pnpm")
command -v migrate  &>/dev/null || MISSING+=("golang-migrate (go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest)")

if [[ ${#MISSING[@]} -gt 0 ]]; then
  error "缺少以下工具，请先安装："
  for m in "${MISSING[@]}"; do echo "  - $m"; done
  exit 1
fi
success "所有依赖工具已就绪"

# ── 2. 初始化 .env ────────────────────────────────────
step "检查环境变量文件"
if [[ ! -f "$ROOT/.env" ]]; then
  warn ".env 不存在，从 .env.example 复制"
  cp "$ROOT/.env.example" "$ROOT/.env"
  info "已生成 $ROOT/.env，如需自定义请编辑后重新运行"
fi

if [[ ! -f "$ROOT/apps/web/.env" ]]; then
  warn "apps/web/.env 不存在，从 .env.example 复制"
  cp "$ROOT/apps/web/.env.example" "$ROOT/apps/web/.env"
fi
success "环境变量文件已就绪"

# 加载后端 .env
set -o allexport
# shellcheck disable=SC1091
source "$ROOT/.env"
set +o allexport

# ── 3. Docker 基础设施（PG + Redis + MailHog）───────
if [[ "$NO_DOCKER" == false && "$FRONTEND_ONLY" == false ]]; then
  step "启动 Docker 基础设施（PostgreSQL + Redis + MailHog）"
  cd "$ROOT"

  # 检查端口是否已可连接（不管是容器还是本地服务）
  PG_OK=false; RD_OK=false; MH_OK=false
  pg_isready -h localhost -p 5432 -U studio &>/dev/null && PG_OK=true
  redis-cli -h localhost ping &>/dev/null                && RD_OK=true
  (echo > /dev/tcp/127.0.0.1/1025) &>/dev/null           && MH_OK=true

  if [[ "$PG_OK" == true && "$RD_OK" == true && "$MH_OK" == true ]]; then
    info "PostgreSQL、Redis 和 MailHog 已可连接，跳过 Docker 启动"
  else
    # 先清理已停止的同名容器
    $DC rm -f postgres redis mailhog 2>/dev/null || true
    $DC up -d postgres redis mailhog 2>&1 || { error "Docker 启动失败"; exit 1; }

    info "等待 PostgreSQL 就绪..."
    for i in $(seq 1 30); do
      pg_isready -h localhost -p 5432 -U studio &>/dev/null && { success "PostgreSQL 已就绪"; break; }
      [[ $i -eq 30 ]] && { error "PostgreSQL 启动超时"; exit 1; }
      sleep 1
    done

    info "等待 Redis 就绪..."
    for i in $(seq 1 15); do
      redis-cli -h localhost ping &>/dev/null && { success "Redis 已就绪"; break; }
      [[ $i -eq 15 ]] && { error "Redis 启动超时"; exit 1; }
      sleep 1
    done

    info "等待 MailHog 就绪..."
    for i in $(seq 1 15); do
      (echo > /dev/tcp/127.0.0.1/1025) &>/dev/null && { success "MailHog 已就绪"; break; }
      [[ $i -eq 15 ]] && { error "MailHog 启动超时"; exit 1; }
      sleep 1
    done
  fi
fi

# ── 4. 数据库迁移 ─────────────────────────────────────
if [[ "$NO_MIGRATE" == false && "$FRONTEND_ONLY" == false ]]; then
  step "运行数据库迁移"
  DSN="${STUDIO_DATABASE_DSN:-postgres://studio:password@localhost:5432/studio_db?sslmode=prefer}"
  # golang-migrate 的 pq 驱动不支持 sslmode=prefer，本地统一用 disable
  MIGRATE_DSN="${DSN//sslmode=prefer/sslmode=disable}"
  MIGRATE_DSN="${MIGRATE_DSN//sslmode=require/sslmode=disable}"
  if migrate -path "$ROOT/migrations" -database "$MIGRATE_DSN" up 2>&1; then
    success "迁移完成"
  else
    warn "迁移出现警告（可能是已存在的迁移），继续..."
  fi
fi

# ── 5. Go 依赖 ────────────────────────────────────────
if [[ "$FRONTEND_ONLY" == false ]]; then
  step "同步 Go 依赖"
  cd "$ROOT"
  go mod tidy -e 2>/dev/null || true
  success "Go 依赖已就绪"
fi

# ── 6. 前端依赖 ───────────────────────────────────────
if [[ "$BACKEND_ONLY" == false ]]; then
  step "安装前端依赖"
  cd "$ROOT"
  pnpm install --frozen-lockfile 2>/dev/null || pnpm install
  success "前端依赖已就绪"
fi

# ── 7. 启动后端 ───────────────────────────────────────
if [[ "$FRONTEND_ONLY" == false ]]; then
  step "启动后端 (Go + Gin) — :8080"
  cd "$ROOT"
  go run ./cmd/server/main.go > "$LOG_DIR/backend.log" 2>&1 &
  BACKEND_PID=$!

  info "等待后端就绪..."
  for i in $(seq 1 20); do
    if curl -sf http://localhost:8080/health &>/dev/null; then
      success "后端已启动 → http://localhost:8080"
      break
    fi
    # 检查进程是否崩溃
    if ! kill -0 "$BACKEND_PID" 2>/dev/null; then
      error "后端启动失败！日志："
      tail -30 "$LOG_DIR/backend.log"
      exit 1
    fi
    sleep 1
  done
  # 若 /health 不可达但进程存活，仍继续
  if ! curl -sf http://localhost:8080/health &>/dev/null; then
    warn "后端健康检查未响应，但进程仍在运行，继续..."
  fi
fi

# ── 8. 启动前端 ───────────────────────────────────────
if [[ "$BACKEND_ONLY" == false ]]; then
  step "启动前端 (Next.js) — :3000"
  cd "$ROOT/apps/web"
  pnpm dev > "$LOG_DIR/frontend.log" 2>&1 &
  FRONTEND_PID=$!

  info "等待前端就绪（首次编译约需 15s）..."
  for i in $(seq 1 40); do
    if curl -sf http://localhost:3000 &>/dev/null; then
      success "前端已启动 → http://localhost:3000"
      break
    fi
    if ! kill -0 "$FRONTEND_PID" 2>/dev/null; then
      error "前端启动失败！日志："
      tail -30 "$LOG_DIR/frontend.log"
      exit 1
    fi
    sleep 1
  done
fi

# ── 9. 完成 ───────────────────────────────────────────
echo ""
echo -e "${GREEN}╔═══════════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║            所有服务已启动！                   ║${NC}"
echo -e "${GREEN}╠═══════════════════════════════════════════════╣${NC}"
[[ "$FRONTEND_ONLY" == false ]] && \
echo -e "${GREEN}║  后端 API  →  http://localhost:8080           ║${NC}"
[[ "$BACKEND_ONLY" == false ]] && \
echo -e "${GREEN}║  前端页面  →  http://localhost:3000           ║${NC}"
[[ "$FRONTEND_ONLY" == false ]] && \
echo -e "${GREEN}║  MailHog   →  http://localhost:8025           ║${NC}"
echo -e "${GREEN}╠═══════════════════════════════════════════════╣${NC}"
echo -e "${GREEN}║  日志目录  →  .dev-logs/                      ║${NC}"
echo -e "${GREEN}║  按 Ctrl+C 停止所有服务                       ║${NC}"
echo -e "${GREEN}╚═══════════════════════════════════════════════╝${NC}"
echo ""

# ── 实时显示后端日志（便于调试）─────────────────────
if [[ "$FRONTEND_ONLY" == false && "$BACKEND_ONLY" == false ]]; then
  info "实时显示后端日志（前端日志见 .dev-logs/frontend.log）"
elif [[ "$BACKEND_ONLY" == true ]]; then
  info "实时显示后端日志"
else
  info "实时显示前端日志"
fi

if [[ "$FRONTEND_ONLY" == true ]]; then
  tail -f "$LOG_DIR/frontend.log"
else
  tail -f "$LOG_DIR/backend.log"
fi
