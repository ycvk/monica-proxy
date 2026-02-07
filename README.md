# 十分抱歉，由于某些原因，将停止维护本项目

---

# Monica Proxy

<div align="center">

![Go](https://img.shields.io/badge/go-1.24-00ADD8)
![License](https://img.shields.io/badge/license-MIT-green)
![Docker](https://img.shields.io/badge/docker-ready-2496ED)

**Monica AI 代理服务**

将 Monica AI 转换为 ChatGPT 兼容的 API，支持完整的 OpenAI 接口兼容性

[快速开始](#-快速开始) • [功能特性](#-功能特性) • [部署指南](#-部署指南) • [配置参考](#-配置参考)

</div>

---

## 🚀 **快速开始**

### 一键启动

```bash
docker run -d \
  --name monica-proxy \
  -p 8080:8080 \
  -e MONICA_COOKIE="your_monica_cookie" \
  -e BEARER_TOKEN="your_bearer_token" \
  neccen/monica-proxy:latest
```

### 测试API

```bash
curl -H "Authorization: Bearer your_bearer_token" \
     http://localhost:8080/v1/models
```

## ✨ **功能特性**

### 🔗 **API兼容性**

- ✅ **完整的System Prompt支持** - 通过Custom Bot Mode实现真正的系统提示词
- ✅ **ChatGPT API完全兼容** - 无缝替换OpenAI接口，支持所有标准参数
- ✅ **流式响应** - 完整的SSE流式对话体验，支持实时输出
- ✅ **Monica模型支持** - GPT-4o、Claude-4、Gemini等主流模型完整映射

## 🏗️ **部署指南**

### 🐳 **Docker Compose部署（推荐）**

#### 部署配置

```yaml
# docker-compose.yml
services:
  monica-proxy:
    build: .
    container_name: monica-proxy
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - MONICA_COOKIE=${MONICA_COOKIE}
      - BEARER_TOKEN=${BEARER_TOKEN}
      - RATE_LIMIT_RPS=100          # 启用限流：每秒100请求
      # Custom Bot模式配置（可选）
      # - ENABLE_CUSTOM_BOT_MODE=true
      # - BOT_UID=${BOT_UID}
```

### 🔧 **源码编译**

```bash
# 克隆项目
git clone https://github.com/ycvk/monica-proxy.git
cd monica-proxy

# 编译
go build -o monica-proxy main.go

# 运行
export MONICA_COOKIE="your_cookie"
export BEARER_TOKEN="your_token"
# export BOT_UID="your_bot_uid"  # 可选，用于Custom Bot模式
./monica-proxy
```

## ⚙️ **配置参考**

### 🌍 **环境变量配置**

| 变量名                      | 必需 | 默认值       | 说明                                               |
|--------------------------|----|-----------|--------------------------------------------------|
| `MONICA_COOKIE`          | ✅  | -         | Monica登录Cookie                                   |
| `BEARER_TOKEN`           | ✅  | -         | API访问令牌                                          |
| `ENABLE_CUSTOM_BOT_MODE` | ❌  | `false`   | 启用Custom Bot模式，支持系统提示词                           |
| `BOT_UID`                | ❌* | -         | Custom Bot的UID（*当ENABLE_CUSTOM_BOT_MODE=true时必需） |
| `RATE_LIMIT_RPS`         | ❌  | `0`       | 限流配置：0=禁用，>0=每秒请求数限制                             |
| `TLS_SKIP_VERIFY`        | ❌  | `true`    | 是否跳过TLS证书验证                                      |
| `LOG_LEVEL`              | ❌  | `info`    | 日志级别：debug/info/warn/error                       |
| `SERVER_PORT`            | ❌  | `8080`    | HTTP服务监听端口                                       |
| `SERVER_HOST`            | ❌  | `0.0.0.0` | HTTP服务监听地址                                       |

### 📄 **配置文件示例**

```yaml
# config.yaml
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout: "30s"
  write_timeout: "30s"

monica:
  cookie: "your_monica_cookie"
  enable_custom_bot_mode: false   # 启用后支持系统提示词
  bot_uid: "your_custom_bot_uid"  # Custom Bot模式必需

security:
  bearer_token: "your_bearer_token"
  rate_limit_enabled: true
  rate_limit_rps: 100
  tls_skip_verify: false

http_client:
  timeout: "3m"
  max_idle_conns: 100
  max_idle_conns_per_host: 20
  retry_count: 3

logging:
  level: "info"
  format: "json"
  mask_sensitive: true
```

## 🔌 **API使用**

### 支持的端点

- `POST /v1/chat/completions` - 聊天对话（兼容ChatGPT）
- `GET /v1/models` - 获取模型列表
- `POST /v1/images/generations` - 图片生成（兼容DALL-E）

### 认证方式

```http
Authorization: Bearer YOUR_BEARER_TOKEN
```

### 聊天API示例

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {"role": "system", "content": "你是一个有帮助的助手"},
      {"role": "user", "content": "你好"}
    ],
    "stream": true
  }'
```

### 支持的模型

| 模型系列         | 模型名称                                                                                             | 说明                 |
|--------------|--------------------------------------------------------------------------------------------------|--------------------|
| **GPT系列**    | `gpt-5`, `gpt-4o`, `gpt-4o-mini`, `gpt-4.1`, `gpt-4.1-mini`, `gpt-4.1-nano`, `gpt-4-5`           | OpenAI GPT模型       |
| **Claude系列** | `claude-4-sonnet`, `claude-4-opus`, `claude-3-7-sonnet`, `claude-3-5-sonnet`, `claude-3-5-haiku` | Anthropic Claude模型 |  
| **Gemini系列** | `gemini-2.5-pro`, `gemini-2.5-flash`, `gemini-2.0-flash`, `gemini-1`                             | Google Gemini模型    |
| **O系列**      | `o1-preview`, `o3`, `o3-mini`, `o4-mini`                                                         | OpenAI O系列模型       |
| **其他**       | `deepseek-reasoner`, `deepseek-chat`, `grok-3-beta`, `grok-4`, `sonar`, `sonar-reasoning-pro`    | 专业模型               |

## 🛠️ **高级功能**

### Custom Bot Mode（系统提示词支持）

通过启用 Custom Bot Mode，可以让所有的聊天请求都支持系统提示词（system prompt）功能：

```bash
# 启用 Custom Bot Mode
export ENABLE_CUSTOM_BOT_MODE=true
export BOT_UID="your-bot-uid"  # 必需

⬇️ 启动项目后 ⬇️

# 现在所有 /v1/chat/completions 请求都支持 system prompt
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer your_token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [
      {
        "role": "system",
        "content": "你是一个海盗船长，用海盗的口吻说话"
      },
      {
        "role": "user",
        "content": "介绍一下你自己"
      }
    ]
  }'
```

**优势：**

- 无需修改客户端代码，保持完全兼容
- 所有请求都可以动态设置不同的 prompt
- 支持流式和非流式响应

### 限流配置

```bash
# 启用限流（每秒50请求）
export RATE_LIMIT_RPS=50
docker-compose restart monica-proxy

# 测试限流效果
for i in {1..100}; do curl -H "Authorization: Bearer token" http://localhost:8080/v1/models & done
```

## 📈 **监控和运维**

### 日志查看

```bash
# 查看实时日志
docker-compose logs -f monica-proxy

# 查看错误日志  
docker-compose logs monica-proxy | grep -i error

# 查看JSON格式结构化日志
docker-compose logs monica-proxy | jq .
```

### 服务状态检查

```bash
# 测试API可用性
curl -H "Authorization: Bearer your_token" \
     http://localhost:8080/v1/models

# 测试限流状态（查看HTTP响应头）
curl -I -H "Authorization: Bearer your_token" \
     http://localhost:8080/v1/models
```

### 基础监控

```bash
# 查看容器资源使用情况
docker stats monica-proxy

# 简单的API压力测试
for i in {1..10}; do
  curl -s -H "Authorization: Bearer your_token" \
       http://localhost:8080/v1/models > /dev/null && echo "OK" || echo "FAIL"
done
```

## 🔧 **故障排查**

### 常见问题

1. **认证失败**
   ```bash
   # 检查Token配置
   docker-compose exec monica-proxy env | grep BEARER_TOKEN
   ```

2. **限流过于严格**
   ```bash
   # 调整限流参数
   export RATE_LIMIT_RPS=200
   docker-compose restart monica-proxy
   ```

## 🤝 **贡献指南**

欢迎提交Issue和Pull Request！

1. Fork本项目
2. 创建特性分支：`git checkout -b feature/amazing-feature`
3. 提交更改：`git commit -m 'Add amazing feature'`
4. 推送分支：`git push origin feature/amazing-feature`
5. 提交Pull Request

## 📄 **许可证**

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情

---

<div align="center">

**如果这个项目对你有帮助，请给个 ⭐️ Star！**

</div>
