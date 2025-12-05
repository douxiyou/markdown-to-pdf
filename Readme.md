# Markdown to PDF Converter

这是一个基于 Go 和 Chrome Headless 的 Markdown 到 PDF 转换服务。该服务可以通过 HTTP 接口接收 Markdown 文本并将其转换为 PDF 文档。

## 功能特点

- 将 Markdown 文本转换为格式化的 PDF 文档
- 使用 Chrome Headless 浏览器确保高质量的 PDF 输出
- 提供 RESTful API 接口
- 支持长时间运行的服务，具备自动健康检查和 Chrome 实例恢复功能
- 高效的 Chrome 实例管理，避免重复初始化

## 技术栈

- Go 语言
- Echo v4 Web 框架
- Chromedp - Chrome DevTools Protocol 控制库
- Blackfriday v2 - Markdown 解析器

## 安装与部署

### 环境要求

- Go 1.25+
- Chrome 或 Chromium 浏览器

### 构建项目

```bash
# 克隆项目
git clone <repository-url>
cd md-to-pdf

# 下载依赖
go mod tidy

# 构建
go build -o md-to-pdf main.go

# 运行
./md-to-pdf
```

### 使用 PM2 部署 (可选)

项目包含 [ecosystem.config.js](file:///Users/ningwei/go-repo/md-to-pdf/ecosystem.config.js) 配置文件，可用于 PM2 部署：

```bash
pm2 start ecosystem.config.js
```

## API 接口

### 转换 Markdown 到 PDF

```
POST /utils/convert-pdf
```

请求体应包含 Markdown 格式的文本内容。响应将是生成的 PDF 文件。

示例：

```bash
curl -X POST \
  http://localhost:3000/utils/convert-pdf \
  -H 'Content-Type: text/markdown' \
  -d '# Hello World
  
  This is a sample markdown document.' \
  --output output.pdf
```

### 健康检查

```
GET /health
```

返回服务和 Chrome 实例的状态信息。

## 项目结构

```
.
├── chrome/                 # Chrome 实例管理
│   └── instance.go
├── convert/                # PDF 转换逻辑
│   └── toPdf.go
├── main.go                 # 主程序入口
├── ecosystem.config.js     # PM2 配置文件
└── go.mod                  # Go 模块文件
```

## Chrome 实例管理

为了提高性能和服务稳定性，该项目实现了 Chrome 实例的持久化管理：

1. 服务启动时初始化 Chrome 实例
2. 自动健康检查，每30秒检查一次实例状态
3. 如果实例长时间未使用（超过1小时）会自动刷新
4. 当检测到实例不可用时会自动重新初始化

## 常见问题

### Chrome 实例未初始化错误

如果服务长时间运行后出现 "Chrome实例未初始化" 错误，服务会自动尝试重新初始化 Chrome 实例。

### 性能优化建议

1. 服务会在空闲时保持 Chrome 实例活动状态以提高响应速度
2. 对于高并发场景，考虑使用负载均衡部署多个实例

## 许可证

[MIT License](LICENSE)