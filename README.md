# AirDropsBot（空投监控助手）

一个用 Go 语言编写的币安空投自动监控和通知系统，能够实时追踪新空投项目并通过邮件发送提醒。

## 🌟 主要特性

- 🔄 实时监控币安空投
- ⏰ 智能定时提醒（10分钟、5分钟、1分钟、开始时）
- 📧 邮件通知系统
- 🎯 高性能内存管理
- 🔒 线程安全设计
- ⚡ 支持配置热更新

## 📋 系统要求

- Go 1.16 或更高版本
- 支持 SMTP 的邮箱账号（推荐使用 QQ 邮箱）

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/yourusername/AirDropsBot-Backend.git
cd AirDropsBot-Backend
```

### 2. 配置文件

创建 `config.json` 文件：

```json
{
  "password": "your-password",
  "interval_seconds": 30,
  "email": {
    "smtpCode": "你的邮箱授权码",
    "smtpCodeType": "qq",
    "smtpEmail": "你的邮箱@qq.com",
    "ColaKey": "你的ColaKey",
    "tomail": [
      "接收通知的邮箱1@qq.com",
      "接收通知的邮箱2@qq.com"
    ]
  }
}
```

### 3. 构建项目

```bash
go build ./cmd/airdropsbot
```

### 4. 运行服务

```bash
./airdropsbot -addr :8080 -config ./config.json
```

## 🔧 API 接口

所有接口都需要在请求体中包含 password 字段进行认证。

### POST /start
启动监控服务

请求示例：
```json
{
  "password": "your-password"
}
```

### POST /stop
停止监控服务

### POST /status
获取服务状态，返回定时器运行状态、CPU和内存使用情况

### POST /config/get
获取当前配置

### POST /config/update
更新配置（需要先停止服务）

## 📨 邮件通知

系统会在以下时间点发送通知：
- 空投开始前 10 分钟
- 空投开始前 5 分钟
- 空投开始前 1 分钟
- 空投开始时

每封通知邮件包含：
- 项目名称和代币符号
- 项目简介
- 空投数量
- 链信息和合约地址
- 代币经济学数据
- 开始和结束时间

## 🗃️ 项目结构

```
airdropsbot/
├── cmd/
│   └── airdropsbot/        # 主程序入口
├── internal/
│   ├── airdrop/           # 空投监控核心逻辑
│   ├── config/            # 配置管理
│   ├── email/             # 邮件服务
│   ├── scheduler/         # 定时任务调度
│   ├── server/            # HTTP服务器
│   └── task/             # 任务执行逻辑
├── config.json            # 配置文件
├── go.mod
└── README.md
```

## ⚙️ 配置说明

### 主配置
- `password`: API接口认证密码
- `interval_seconds`: 检查间隔（建议 30-60 秒）

### 邮件配置
- `smtpCode`: 邮箱授权码
- `smtpCodeType`: 邮箱类型（目前支持 qq）
- `smtpEmail`: 发件人邮箱
- `ColaKey`: Cola 服务密钥
- `tomail`: 接收通知的邮箱列表

## 🔐 安全特性

- API 接口密码保护
- 线程安全的缓存管理
- 配置更新保护机制
- 请求超时控制

## 💡 性能优化

- 使用缓存减少 API 调用
- 智能的内存管理
- 自动清理过期数据
- 并发安全的数据访问

## 🤝 贡献指南

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 提交 Pull Request

## 📝 许可证

根据需要选择合适的开源许可证

## 📮 联系方式

如果您有任何问题或建议，请通过以下方式联系：
- Email: zhongzhenjie0729@outlook.com
- GitHub Issues: [创建新 issue](https://github.com/ZZHENJIE/AirDropsBot/issues)
