# RSS2Email

RSS2Email 是一个用 Go 编写的 RSS 订阅服务，可以定期获取 RSS 源并将更新通过电子邮件发送给订阅者。

## 功能特点

- 定时抓取 RSS 源内容
- 通过电子邮件发送更新
- 支持多个 RSS 源
- 使用 SQLite 数据库存储订阅信息
- 避免重复发送已处理的内容
- 支持 Docker 部署

## RSS 源示例

- 阮一峰周刊 (ruanyifeng)
- DecoHack (decohack)
- 少数派 (sspai)
- 知乎 (zhihu)
- Kitekagi 世界 (kitekagi)
- Kitekagi 人工智能 (kitekagi-ai)

## 快速开始

### 使用 Docker（推荐）

1. 构建镜像：
   ```bash
   make docker-image-build
   ```

2. 运行容器：
   ```bash
   make docker-run
   ```

### 本地构建运行

1. 克隆项目：
   ```bash
   git clone https://github.com/weirwei/rss2email.git
   cd rss2email
   ```

2. 构建项目：
   ```bash
   go build -o rss2email main.go
   ```

3. 运行服务：
   ```bash
   ./rss2email
   ```

## 配置

### 邮箱配置

在 `conf/yaml/email.yaml` 文件中配置邮箱信息：

```yaml
host: smtp.163.com
port: 465
user: your-email@163.com
pass: your-password
```

### RSS 源配置

RSS 源与调度从数据库读取，需先写入 `feed_sources` 表。

手动插入示例：

```sql
INSERT INTO feed_sources (subscription_id, name, feed_url, content_field, schedule_type, cron_spec, created_at, updated_at, deleted)
VALUES ("ruanyifeng", "阮一峰周刊", "https://www.ruanyifeng.com/blog/atom.xml", "description", "cron", "0 10/3 * * 5", datetime('now'), datetime('now'), 0);
```

### 翻译配置（英文转中文）

复制示例配置并按需填写：

```bash
cp conf/yaml/translation.example.yaml conf/yaml/translation.yaml
```

关键配置：

- `enabled`：是否启用翻译
- `provider`：`google` / `microsoft` / `baidu` / `llm`
- `source_lang` / `target_lang`：默认 `en -> zh-CN`
- `only_translate_english`：只翻译看起来是英文的内容
- `append_translated`：
  - `true`：保留原文并追加中文翻译
  - `false`：直接替换为中文

示例（使用大模型翻译）：

```yaml
enabled: true
provider: llm
source_lang: en
target_lang: zh-CN
only_translate_english: true
append_translated: true
llm:
  endpoint: https://api.openai.com/v1/chat/completions
  api_key: sk-xxxx
  model: gpt-4o-mini
  temperature: 0
```

## 使用方法

### 订阅 RSS 源

使用以下命令为指定邮箱订阅 RSS 源：

```bash
./rss2email register your-email@example.com ruanyifeng sspai zhihu
```

支持的订阅源以 `feed_sources` 表为准。

### 添加订阅源

```bash
./rss2email add-feed <subscription_id> <feed_url> <name> [content_field] [schedule_type] [cron_spec]
```

示例：

```bash
./rss2email add-feed ruanyifeng https://www.ruanyifeng.com/blog/atom.xml "阮一峰周刊" description cron "0 10/3 * * 5"
```

### 启动服务

```bash
./rss2email
```

服务启动后会根据配置的定时任务自动抓取 RSS 源并发送邮件。

### 启动 Web 配置页面

如果不想手动操作 `feed_sources` / `user_subscriptions`，可启动内置管理页面：

```bash
./rss2email web --addr :8080
```

打开 `http://127.0.0.1:8080`，即可在页面中新增/更新/删除订阅源与用户订阅。

## 定时任务

服务从数据库读取调度策略：

- `schedule_type=startup`：启动时执行一次
- `schedule_type=live`：每小时执行一次
- `schedule_type=cron`：按 `cron_spec` 执行

## 数据库

项目使用 SQLite 数据库存储订阅信息，数据库文件位于 `db/rss.db`。

表结构：
- `user_subscriptions`：存储用户订阅信息和处理进度
- `feed_sources`：存储 RSS 源配置

## 开发

### 项目结构

```
├── cmd/              # 命令行接口
├── conf/             # 配置文件
├── constants/        # 常量定义
├── db/               # 数据库文件
├── email/            # 邮件发送功能
├── helpers/          # 辅助函数
├── models/           # 数据库模型
├── rss/              # RSS 抓取功能
├── service/          # 业务逻辑
└── test/             # 测试初始化
```

### 添加新的 RSS 源

只需在 `feed_sources` 表中添加 RSS 源 URL 与调度信息，服务会自动加载并调度，不需要改代码。运行中会按分钟级轮询 `feed_sources` 并增量重载（新增/修改/删除都可生效）。

### 测试说明

- 默认测试：`go test ./...`，仅运行稳定单元测试。
- 集成测试：依赖本地 DB 表和外部 RSS，建议通过 build tag 单独运行。

## 问题记录

> rsshub 是海外的，国内无法访问

（这个方法长期会无法访问，推荐自己部署）
使用 cloudflare 代理，worker 代码：
```js
// 定义重定向的目标 URL
const RUANYIFENG_FEED_URL = 'http://feeds.feedburner.com/ruanyifeng';
const RSSHUB_FEED_URL = 'https://rsshub.app';

// 监听 'fetch' 事件，这是 Cloudflare Worker 的入口点
addEventListener('fetch', event => {
  event.respondWith(handleRequest(event.request));
});

/**
 * 处理传入的请求
 * @param {Request} request - 传入的 HTTP 请求对象
 * @returns {Promise<Response>} - 返回一个 Promise，解析为 HTTP 响应对象
 */
async function handleRequest(request) {
  // 解析请求的 URL
  const url = new URL(request.url);
  let targetURL;
  const pathname = url.pathname;
  if (pathname.startsWith('/rsshub')) {
    targetURL = new URL(RSSHUB_FEED_URL);
    const fixedPathname = pathname.substring('/rsshub'.length);
    targetURL.pathname = fixedPathname
    targetURL.search = url.search
  } else if (pathname === '/ruanyifeng') {
    targetURL = new URL(RUANYIFENG_FEED_URL);
    targetURL.pathname = url.pathname
    targetURL.search = url.search
  } else {
    return
  }
  let newRequest = new Request(targetURL, {
    method: request.method,
    headers: request.headers,
    body: request.body
  })

  let response = await fetch(newRequest)

  // 添加跨域支持
  let corsHeaders = {
    'Access-Control-Allow-Origin': '*',
    'Access-Control-Allow-Methods': 'GET,HEAD,POST,OPTIONS',
    'Access-Control-Allow-Headers': request.headers.get('Access-Control-Request-Headers'),
  }

  // 如果是预检请求，直接返回跨域头
  if (request.method === 'OPTIONS') {
    return new Response(null, { headers: corsHeaders })
  }

  // 复制响应以添加新的头
  let responseHeaders = new Headers(response.headers)
  for (let [key, value] of Object.entries(corsHeaders)) {
    responseHeaders.set(key, value)
  }

  return new Response(response.body, {
    status: response.status,
    statusText: response.statusText,
    headers: responseHeaders
  })
}
```

> rsshub 会报 http 429 too many request

> cloudflare 代理 rsshub 被网关拦住了 2025-07-25

最后使用本地部署的策略，发现部署后的开销意外的小。不过只是做了一个最基本的部署，外加一个自动更新，先用一段时间看看效果
