# PUSH_QQ 项目说明文档

## 项目简介

PUSH_QQ 是一个基于 Go 语言开发的 QQ 消息推送 API 服务，允许通过 HTTP API 接口向 QQ 好友或群聊发送消息。该项目使用 LagrangeGo 库实现 QQ 协议，提供简单易用的 RESTful API 接口。

## 功能特性

- ✅ 支持发送 QQ 私聊消息
- ✅ 支持发送 QQ 群聊消息
- ✅ 支持二维码扫码登录
- ✅ 支持跨平台部署（Linux/macOS/Windows）

## 技术栈

- **后端框架**: Go + Fiber v2
- **QQ 协议**: LagrangeGo

## 安装部署

### 环境要求

- Linux/Mac/Windows/Andriod
- 可个人电脑部署，也可以放在服务器或挂机宝等运行
- [主程序，参见项目 release](https://github.com/Baiyuetribe/PushQQ/release)

### 首次登录

程序启动后，如果没有有效的登录状态，会自动生成二维码：

1. 程序会在当前目录生成 `qrcode.png` 文件
2. 使用手机 QQ 扫描二维码登录
3. 登录成功后，登录状态会保存到 `sig.bin` 文件
4. 下次启动时会自动使用保存的登录状态

## API 接口说明

### 1. 私聊消息发送

```bash
curl -X POST http://localhost:3206/api/msg \
  -H "Content-Type: application/json" \
  -d '{
    "method": "qq",
    "uid": "填好友QQ号",
    "msg": "Hello, 这是一条私聊消息!"
  }'
```

### 2. 发送群聊消息

```bash
curl -X POST http://localhost:3206/api/msg \
  -H "Content-Type: application/json" \
  -d '{
    "method": "group",
    "uid": "填群号",
    "msg": "Hello, 这是一条群消息!"
  }'
```

### 3. 健康检查

QQ 好友发送消息"ping"，如果返回“pong”代表服务正常

### QQ 客户端配置

- **协议版本**: Linux 3.2.15-30366
- **签名服务器**: https://sign.lagrangecore.org/api/sign/30366

## 常见问题

### 1. 掉线问题

- 经过测试，qsign 签名服务不稳定，容易导致掉线，该 qsign 使用 java 开发，个人不懂 java，如有懂 java 的看看，并描述基础实现逻辑。

### 2. 部署相关

**Q: 如何在服务器上后台运行？**
A:

```bash
# 使用 nohup
nohup ./push_qq > output.log 2>&1 &

# 使用 systemd（推荐）
# 创建服务文件后使用 systemctl 管理
```

**Q: 如何修改端口？**
A: 修改 `main.go` 中的 `app.Listen(":3206")` 部分

### 3. 是否考虑更多功能

否，Qsign 签名逻辑未搞清楚，需要本地实现后，才会考虑。
