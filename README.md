# rss2email


## 问题

> rsshub 是海外的，国内无法访问

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

