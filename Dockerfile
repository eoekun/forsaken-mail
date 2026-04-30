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
COPY . .
COPY --from=frontend /app/dist ./embed
RUN CGO_ENABLED=1 go build -o /app/forsaken-mail ./cmd/server

FROM alpine:3.21
RUN apk add --no-cache ca-certificates sqlite-libs
COPY --from=backend /app/forsaken-mail /usr/local/bin/forsaken-mail
VOLUME /data
EXPOSE 25 3000
CMD ["forsaken-mail"]
