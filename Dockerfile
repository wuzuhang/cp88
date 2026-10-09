# syntax=docker/dockerfile:1

FROM docker.m.daocloud.io/library/node:22-alpine AS frontend-builder
WORKDIR /src/cfront
ARG NPM_REGISTRY=https://registry.npmjs.org
COPY cfront/package.json cfront/package-lock.json ./
RUN npm ci --registry=${NPM_REGISTRY} --no-audit --no-fund
COPY cfront/ ./
RUN npm run build

FROM golang:1.25-alpine AS backend-builder
WORKDIR /src
ARG GOPROXY=https://goproxy.cn,direct
ENV GOPROXY=${GOPROXY}
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
COPY backend/ ./backend/
COPY config/ ./config/
COPY cron/ ./cron/
COPY middleware/ ./middleware/
COPY pkg/ ./pkg/
COPY router/ ./router/
COPY utils/ ./utils/
COPY cmd/ ./cmd/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ssq-server . \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/import-history ./cmd/import-history

FROM alpine:3.22
RUN apk add --no-cache ca-certificates nginx supervisor tzdata \
    && mkdir -p /run/nginx
ENV TZ=Asia/Shanghai SSQ_SERVER_MODE=release SSQ_SERVER_PORT=8080
WORKDIR /app
COPY --from=backend-builder /out/ssq-server /out/import-history ./
COPY deploy/config.docker.yaml ./config.yaml
COPY --from=frontend-builder /src/cfront/dist/ /usr/share/nginx/html/
COPY deploy/nginx.conf /etc/nginx/http.d/default.conf
COPY deploy/supervisord.conf /etc/supervisord.conf
RUN chmod -R a+rX /usr/share/nginx/html
EXPOSE 80
HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
    CMD wget -q -O /dev/null http://127.0.0.1/ && wget -q -O /dev/null http://127.0.0.1/health || exit 1
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisord.conf"]
