# AWS 部署手册 — aatist 后端(单 EC2 自建全栈)

目标架构:一台 EC2 用 docker-compose 跑全部容器(gateway + 4 业务服务 + postgres/redis/rabbitmq/minio),
宿主机 nginx 终结 TLS 并反代到网关。前端在 Vercel,API 域名 `https://api.aatist.fi`。

```
Vercel 前端 ──HTTPS──┐
                     ▼
        nginx(443, Let's Encrypt)
          │                    │
          ├─ /avatars/ /files/ → MinIO(127.0.0.1:9000)
          └─ 其余             → gateway(127.0.0.1:8080)
                                   └─(docker 内网)→ backend / file-service / chat-service
```

---

## 1. 开 EC2

- 镜像:Ubuntu 22.04 LTS;机型:**t3.medium**(2C/4G,跑全栈含数据库更稳;最低 t3.small 2G)
- 存储:30 GB gp3
- 弹性 IP:分配一个并绑定(重启不变 IP)

**安全组(入站):**

| 端口 | 来源 | 用途 |
|------|------|------|
| 22 | 你的 IP | SSH |
| 80 | 0.0.0.0/0 | HTTP(certbot 校验 + 跳转) |
| 443 | 0.0.0.0/0 | HTTPS |

> ⚠️ 不要开放 8080/8081/8086/8088/5432/6379/5672/9000/9001。这些只在本机/docker 内网用。
> 结合后端的 `X-User-ID` 头信任模型,内部端口一旦公网可达 = 任意用户可被冒充。

## 2. DNS

在域名解析里把 `api.aatist.fi` 的 A 记录指向 EC2 弹性 IP。

## 3. 装 Docker

```bash
sudo apt update && sudo apt install -y ca-certificates curl git
curl -fsSL https://get.docker.com | sudo sh
sudo usermod -aG docker $USER && newgrp docker   # 之后免 sudo 用 docker
```

## 4. 拉代码 + 配置

```bash
git clone <你的仓库地址> aatist && cd aatist/backend

# 1) 配置文件(可选但推荐:auto_verified_domains / jwt ttl 只能从 yaml 读)
cp configs/config.example.yaml configs/config.yaml

# 2) 生产环境变量
cp .env.production.example .env
# 生成两个强密钥填进 JWT_SECRET 和 INTERNAL_API_TOKEN:
openssl rand -hex 32
openssl rand -hex 32
nano .env   # 把所有 <CHANGE_ME> 改成真实值,确认域名/邮箱/OAuth
```

`.env` 必填项检查:`JWT_SECRET`、`INTERNAL_API_TOKEN`、`POSTGRES_PASSWORD`、`MQ_PASS`、
`MINIO_ROOT_USER/PASSWORD`、`CORS_ORIGINS`、`FRONTEND_URL`。SendGrid/Google 没配可留空(对应功能停用)。

## 5. 启动

```bash
docker compose -f docker-compose.prod.yml --env-file .env up -d --build
docker compose -f docker-compose.prod.yml ps          # 都应 healthy/running
docker compose -f docker-compose.prod.yml logs -f gateway backend   # 看启动日志
```

`migrate` 容器会自动建表并退出(exit 0)。本机自测:

```bash
curl http://127.0.0.1:8080/gateway/health     # {"data":{"status":"ok",...}}
```

## 6. nginx + TLS

```bash
sudo apt install -y nginx certbot python3-certbot-nginx
sudo mkdir -p /var/www/certbot

sudo cp deploy/nginx/api.aatist.fi.conf /etc/nginx/sites-available/api.aatist.fi.conf
sudo ln -s /etc/nginx/sites-available/api.aatist.fi.conf /etc/nginx/sites-enabled/
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t && sudo systemctl reload nginx

# 申请证书(确认 DNS 已生效再跑)
sudo certbot --nginx -d api.aatist.fi
```

certbot 会装好证书并自动续期。验证:

```bash
curl https://api.aatist.fi/gateway/health
```

## 7. Google OAuth(如启用)

Google Cloud Console → 凭据 → OAuth 客户端,**Authorized redirect URIs** 加:

```
https://api.aatist.fi/auth/callback/google
```

与 `.env` 里的 `GOOGLE_OAUTH_REDIRECT_URI` 必须一字不差。

## 8. 前端(Vercel)

在 Vercel 项目环境变量里设(变量名见 [client.js](../frontend/src/shared/api/client.js)):

```
VITE_API_URL=https://api.aatist.fi/api/v1
```

然后重新部署。WebSocket 由同一变量派生,会自动走 `wss://api.aatist.fi/api/v1/ws`。

确认前端真实域名已加进后端 `.env` 的 `CORS_ORIGINS`(含 `www`、必要时加 Vercel 预览域名)。

---

## 冒烟测试

```bash
# 注册 → 登录 → 带 token 访问受保护接口
curl -X POST https://api.aatist.fi/api/v1/auth/register -H 'Content-Type: application/json' \
  -d '{"email":"test@aalto.fi","password":"test123","name":"Test","role":"student","school":"Aalto","faculty":"SCI","major":"CS"}'
# 浏览器实际走前端注册/登录,再点开 Talents/Opportunities/Messages 验证
```

## 日常运维

```bash
# 更新代码后重建
git pull && docker compose -f docker-compose.prod.yml --env-file .env up -d --build

# 看某服务日志 / 重启
docker compose -f docker-compose.prod.yml logs -f backend
docker compose -f docker-compose.prod.yml restart gateway

# 数据库备份(强烈建议挂 cron 每天跑)
docker compose -f docker-compose.prod.yml exec -T postgres \
  pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" | gzip > backup_$(date +%F).sql.gz
```

## 上线前再确认

- [ ] `8081/8086/8088/5432/6379/5672/9000` 没有公网入站规则
- [ ] `JWT_SECRET` / `INTERNAL_API_TOKEN` 是随机强值,且三个服务的 `JWT_SECRET` 一致
- [ ] postgres / minio / rabbitmq 都改了强密码(没有 guest / minioadmin / aatn:aatn)
- [ ] `CORS_ORIGINS` 是精确的前端域名,不是 `*`
- [ ] `DISABLE_EMAIL_VERIFICATION` 按上线策略设(测试 true / 正式 false)
- [ ] MinIO 有 `minio_data` 卷(已在 prod compose 中,重启不丢文件)
- [ ] 配了每日 pg_dump 备份
```
