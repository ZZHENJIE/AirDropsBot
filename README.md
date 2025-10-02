# Airdrops Bot

一个自动监控并提醒 Binance 空投活动的工具。

## 功能特点

- 自动获取币安平台上的空投活动信息
- 在空投结束前发送邮件提醒（10分钟、5分钟、1分钟）
- 支持数据库存储和查询空投状态
- 可配置的监控间隔

## 系统要求

- Rust 编译环境
- SQLite 数据库支持

## 安装方法

```bash
git clone <repository-url>
cd airdropsbot
cargo build --release
```

## 配置说明

创建一个配置文件 `airdropsbot.json`：

```json
{
    "log": "日志输出路径",
    "interval": 1,
    "database": "数据库路径",
    "port": 8080,
    "email": {
        "smtp_code": "your_smtp_code",
        "smtp_code_type": "qq",
        "smtp_email": "your_email@qq.com",
        "cola_key": "获取网站 https://luckycola.com.cn/public/docs/shares/api/mail.html",
        "tomail": [
            "recipient@example.com"
        ]
    }
}
```

## 使用方法

```bash
./airdropsbot airdropsbot.json
```

## 项目结构

- `src/main.rs`: 主程序入口
- `src/app.rs`: 应用程序配置和邮件功能
- `src/binance.rs`: 币安空投数据获取和处理逻辑
- `src/lib.rs`: 库模块导出

## 许可证

本项目采用 GNU 通用公共许可证 v3.0 (GPL-3.0) 授权。

这意味着您可以自由地使用、修改和分发本软件。但是，​​任何衍生作品也必须是开源的，并采用相同的 GPL v3 许可证。​

更多详细信息，请参阅 LICENSE文件。

## 联系方式
[发送邮件](mailto:zhongzhenjie0729@outlook.com)
[Bilibili](https://space.bilibili.com/1362205077)
