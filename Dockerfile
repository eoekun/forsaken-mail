FROM node:24-alpine AS frontend
WORKDIR /app
COPY web/package.json web/package-lock.json* ./
RUN npm install
COPY web/ .
RUN npm run build

FROM golang:1.24-alpine AS backend
RUN apk add --no-cache gcc musl-dev
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
COPY --from=frontend /embed ./embed
RUN CGO_ENABLED=1 go build -o /app/forsaken-mail ./cmd/server

FROM alpine:3.21
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates sqlite-libs
COPY --from=backend /app/forsaken-mail /usr/local/bin/forsaken-mail
COPY --from=frontend /embed /embed
VOLUME /data
EXPOSE 25 3000
CMD ["forsaken-mail"]
