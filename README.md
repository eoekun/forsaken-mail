Forsaken-Mail
==============
A self-hosted disposable mail service.

[Online Demo](http://disposable.dhc-app.com)
[中文文档](./README.zh-CN.md)

### Installation

#### Setting up your DNS correctly

In order to receive emails, your smtp server address should be made available somewhere. Two records should be added to your DNS records. Let us pretend that we want to receive emails at ```*@subdomain.domain.com```:
* First an MX record: ```subdomain.domain.com MX 10 mxsubdomain.domain.com```. This means that the mail server for addresses like ```*@subdomain.domain.com``` will be ```mxsubdomain.domain.com```.
* Then an A record: ```mxsubdomain.domain.com A the.ip.address.of.your.mailin.server```. This tells at which ip address the mail server can be found.

You can use an [smtp server tester](http://mxtoolbox.com/diagnostic.aspx) to verify that everything is correct.

#### Let's Go
general way:
```
npm install && npm start
```
if you want to run this inside a docker container
```
docker build -t denghongcai/forsaken-mail .
docker run --name forsaken-mail -d -p 25:25 -p 3000:3000 denghongcai/forsaken-mail
```

#### Run with docker compose + basic auth

This repository now includes an `nginx` reverse proxy with HTTP Basic Authentication.

1. Set auth credentials with environment variables:
```
export BASIC_AUTH_USERNAME=admin
export BASIC_AUTH_PASSWORD=your-strong-password
```

Or create a local env file:
```
cp .env.example .env
# then edit .env
```

2. Start with docker compose:
```
docker compose up -d --build
```

3. Open in browser:
```
http://localhost
```

The browser will ask for username/password before accessing the web UI.

Notes:
* SMTP is exposed on port `25` from the `app` service.
* Web UI is exposed on port `80` through `nginx`.
* If env vars are not set, compose defaults to `admin / change-me` (only for quick local testing).
Open your browser and type in
```
http://localhost:3000
```
This `:3000` URL is for the direct `docker run` mode above.  
For `docker compose + nginx auth`, use:
```
http://localhost
```

Enjoy!
