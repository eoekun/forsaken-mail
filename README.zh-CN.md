Forsaken-Mail
==============
一个可自托管的临时邮箱服务。

[在线演示](http://disposable.dhc-app.com)
[English README](./README.md)

### 安装

#### 正确配置 DNS

为了能接收邮件，你需要先让 SMTP 服务可被外部投递。假设你要接收 `*@subdomain.domain.com` 的邮件，需要添加两条 DNS 记录：

* MX 记录：`subdomain.domain.com MX 10 mxsubdomain.domain.com`  
  表示 `*@subdomain.domain.com` 的邮件服务器是 `mxsubdomain.domain.com`。
* A 记录：`mxsubdomain.domain.com A 你的邮件服务器IP`  
  表示该邮件服务器对应的 IP 地址。

你可以使用 [smtp tester](http://mxtoolbox.com/diagnostic.aspx) 验证配置是否正确。

#### 开始运行
通用方式：
```bash
npm install && npm start
```

如果你想用 Docker 运行：
```bash
docker build -t denghongcai/forsaken-mail .
docker run --name forsaken-mail -d -p 25:25 -p 3000:3000 denghongcai/forsaken-mail
```

#### 使用 docker compose + basic auth 运行

本仓库已集成 `nginx` 反向代理和 HTTP Basic Authentication。

1. 用环境变量设置鉴权账号密码：
```bash
export MAIL_HOST=mail.example.com
export MAILIN_HOST=0.0.0.0
export MAILIN_PORT=25
export KEYWORD_BLACKLIST=admin,postmaster,system,webmaster,administrator,hostmaster,service,server,root
export DINGTALK_WEBHOOK_TOKEN=your-dingtalk-access-token
export BASIC_AUTH_USERNAME=admin
export BASIC_AUTH_PASSWORD=your-strong-password
export NGINX_PORT=80
```

也可以使用本地环境文件：
```bash
cp .env.example .env
# 然后编辑 .env
```

现在配置统一走环境变量，`config-default.json` 已不再使用。

2. 使用 docker compose 启动：
```bash
docker compose up -d --build
```

3. 浏览器访问：
```text
http://localhost
```

访问网页前会弹出用户名/密码验证。

说明：
* SMTP 端口 `25` 由 `app` 服务暴露。
* Web 端口 `80` 由 `nginx` 服务暴露。
* 若未设置环境变量，compose 默认使用 `admin / change-me`（仅用于本地快速测试）。
* `MAIL_HOST` 用于页面展示的邮箱域名（例如 `abc@mail.example.com`），即使你临时通过服务器 IP 访问页面也会按该域名展示。
* 若设置 `DINGTALK_WEBHOOK_TOKEN`，每封新邮件会触发一次固定内容的钉钉机器人文本通知；为空则不发送请求。

浏览器也可以访问：
```text
http://localhost:3000
```

`3000` 端口仅对应上面的直接 `docker run` 方式。  
如果是 `docker compose + nginx auth`，请使用：
```text
http://localhost
```

Enjoy!
