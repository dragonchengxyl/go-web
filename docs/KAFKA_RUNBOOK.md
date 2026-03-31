# Kafka Runbook

> 面向当前仓库的 Redis + Kafka 共存架构。
> 目标：让业务事件总线切到 Kafka，同时保留 Redis 用于限流、会话、排行榜和 WebSocket 广播。

## 1. 当前 Kafka 覆盖范围

Kafka 当前承接的是业务事件总线，不是全站所有实时能力。

已接入 Kafka 的事件流：

- `post.created`
- `post.moderated`
- `post.liked`
- `user.followed`
- `comment.created`
- `tip.sent`
- `audio.job.created`

已接入 Kafka 的消费者：

- `notification-svc`
- `moderation-svc`
- `audio-worker`
- `outbox-relay`

Redis 仍负责：

- token store / blacklist
- rate limit
- leaderboard
- WebSocket Pub/Sub 广播

## 2. Topic 设计

当前逻辑 topic：

- `furry-events.content`
  - `post.created`
  - `post.moderated`
- `furry-events.social`
  - `post.liked`
  - `user.followed`
  - `comment.created`
  - `tip.sent`
- `furry-events.audio`
  - `audio.job.created`
- `furry-events.dlq`
  - 死信事件

如果你修改 `STUDIO_KAFKA_TOPIC`，上面的前缀会一起变化。

## 3. 本地验证

### 3.1 启动基础设施

本地先起 PostgreSQL / Redis / MailHog / Kafka：

```bash
docker compose --profile kafka up -d postgres redis mailhog kafka
```

### 3.2 创建 topic

Kafka 首次启动后，手动创建 4 个 topic：

```bash
docker exec studio_kafka /opt/bitnami/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists --topic furry-events.content --partitions 3 --replication-factor 1

docker exec studio_kafka /opt/bitnami/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists --topic furry-events.social --partitions 3 --replication-factor 1

docker exec studio_kafka /opt/bitnami/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists --topic furry-events.audio --partitions 3 --replication-factor 1

docker exec studio_kafka /opt/bitnami/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists --topic furry-events.dlq --partitions 3 --replication-factor 1
```

检查：

```bash
docker exec studio_kafka /opt/bitnami/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --list
```

### 3.3 打开 Kafka 开关

临时环境变量：

```bash
export STUDIO_KAFKA_ENABLED=true
export STUDIO_KAFKA_BROKERS=localhost:9092
export STUDIO_KAFKA_TOPIC=furry-events
```

### 3.4 启动主服务与异步服务

主服务：

```bash
go run ./cmd/server -config configs/config.local.yaml
```

Outbox relay：

```bash
go run ./cmd/outbox-relay -config configs/config.local.yaml
```

通知服务：

```bash
go run ./cmd/notification-svc -config configs/config.local.yaml
```

音频 worker：

```bash
go run ./cmd/audio-worker -config configs/config.local.yaml
```

审核服务：

```bash
go run ./cmd/moderation-svc -config configs/config.local.yaml
```

说明：

- 如果本地没有有效的阿里云审核凭据，`moderation-svc` 会因为缺少配置而退出
- 这种情况下，可以先验证：
  - `social` topic
  - `audio` topic
  - outbox relay
  - notification / audio-worker 消费链路

### 3.5 检查 health / ready / metrics

```bash
curl http://127.0.0.1:18052/health
curl http://127.0.0.1:18052/ready
curl http://127.0.0.1:18052/metrics

curl http://127.0.0.1:18053/health
curl http://127.0.0.1:18053/ready
curl http://127.0.0.1:18053/metrics

curl http://127.0.0.1:18054/health
curl http://127.0.0.1:18054/ready
curl http://127.0.0.1:18054/metrics

curl http://127.0.0.1:18055/health
curl http://127.0.0.1:18055/ready
curl http://127.0.0.1:18055/metrics
```

### 3.6 做一次真实事件验证

启动主站后，做这些动作：

- 登录
- 发帖
- 点赞
- 评论
- 关注
- 创建音频任务

然后观察：

- `outbox_events` 是否从 `pending` 变成 `published`
- Kafka topic 是否有消息
- `notification-svc` 是否生成通知
- `audio-worker` 是否消费 `audio.job.created`

检查 outbox：

```bash
psql "$STUDIO_DATABASE_DSN" -c "select status, event_type, topic, count(*) from outbox_events group by 1,2,3 order by 1,2;"
```

消费查看：

```bash
docker exec studio_kafka /opt/bitnami/kafka/bin/kafka-console-consumer.sh \
  --bootstrap-server localhost:9092 \
  --topic furry-events.social \
  --from-beginning
```

## 4. 服务器启用 Kafka

### 4.1 修改 `.env.prod`

至少配置：

```env
STUDIO_KAFKA_ENABLED=true
STUDIO_KAFKA_BROKERS=kafka:9092
STUDIO_KAFKA_TOPIC=furry-events
STUDIO_KAFKA_CLIENT_ID=studio-platform-prod
STUDIO_OBSERVABILITY_NOTIFICATION_HTTP_PORT=18052
STUDIO_OBSERVABILITY_MODERATION_HTTP_PORT=18053
STUDIO_OBSERVABILITY_AUDIO_WORKER_HTTP_PORT=18054
STUDIO_OBSERVABILITY_OUTBOX_HTTP_PORT=18055
```

如果你使用外部 Kafka 集群，把 `STUDIO_KAFKA_BROKERS` 换成真实 broker 列表。

### 4.2 启动 Kafka 与异步服务

如果使用 compose 内置 Kafka：

```bash
cd /opt/furry-app
sudo docker compose --env-file .env.prod -f docker-compose.prod.yml --profile kafka up -d kafka
```

也可以直接使用仓库内脚本：

```bash
cd /opt/furry-app
bash scripts/enable-kafka-prod.sh
```

创建 topic：

```bash
sudo docker exec furry-app-kafka-1 /opt/bitnami/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists --topic furry-events.content --partitions 3 --replication-factor 1

sudo docker exec furry-app-kafka-1 /opt/bitnami/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists --topic furry-events.social --partitions 3 --replication-factor 1

sudo docker exec furry-app-kafka-1 /opt/bitnami/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists --topic furry-events.audio --partitions 3 --replication-factor 1

sudo docker exec furry-app-kafka-1 /opt/bitnami/kafka/bin/kafka-topics.sh \
  --bootstrap-server localhost:9092 \
  --create --if-not-exists --topic furry-events.dlq --partitions 3 --replication-factor 1
```

启动应用与异步服务：

```bash
sudo docker compose --env-file .env.prod -f docker-compose.prod.yml up -d backend frontend
sudo docker compose --env-file .env.prod -f docker-compose.prod.yml --profile async up -d outbox-relay notification-svc audio-worker
```

如果审核凭据已配置，再补：

```bash
sudo docker compose --env-file .env.prod -f docker-compose.prod.yml --profile async up -d moderation-svc
```

### 4.3 检查容器健康

```bash
curl http://127.0.0.1:18052/health
curl http://127.0.0.1:18053/health
curl http://127.0.0.1:18054/health
curl http://127.0.0.1:18055/health
```

### 4.4 检查 metrics

```bash
curl http://127.0.0.1:18052/metrics | rg 'eventbus|outbox'
curl http://127.0.0.1:18053/metrics | rg 'eventbus|outbox'
curl http://127.0.0.1:18054/metrics | rg 'eventbus|outbox'
curl http://127.0.0.1:18055/metrics | rg 'eventbus|outbox'
```

如果服务器没有 `rg`，用：

```bash
curl http://127.0.0.1:18055/metrics | grep eventbus
```

## 5. 启用后的判断标准

满足以下条件，才算 Kafka 真正启用成功：

1. `outbox_events` 出现 `published` 记录
2. `furry-events.content/social/audio` 三类 topic 能看到消息
3. `notification-svc`、`audio-worker` 至少能消费到对应消息
4. `/ready` 返回 `ready`
5. `/metrics` 里能看到：
   - `studio_eventbus_publish_total`
   - `studio_eventbus_consume_total`
   - `studio_outbox_dispatch_total`

## 6. 当前边界

当前 Kafka 改造已经做到：

- 业务事件总线可切 Kafka
- topic 分层
- outbox relay
- DLQ topic
- metrics 埋点
- async worker 观测端口

但这仍然不是最终治理形态，后续仍建议继续补：

- topic ACL
- Kafka broker 持久化与副本策略
- 消费 lag 监控
- 更细的事件 schema/version 管理
- 自动 topic provisioning
