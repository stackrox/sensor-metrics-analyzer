FROM golang:1.25-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG BUILD_TIME=""

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags "-X main.buildVersion=${VERSION} -X main.buildTime=${BUILD_TIME}" \
    -o /out/web-server ./web/server

FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates nginx
RUN adduser -D -H -u 10001 appuser

COPY --from=builder /out/web-server /app/web-server
COPY web/static /app/web/static
COPY automated-rules /app/automated-rules
COPY templates/markdown.tmpl /app/templates/markdown.tmpl
COPY deploy/nginx.container.conf /etc/nginx/nginx.conf
COPY deploy/container-entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh \
    && chown -R 10001:10001 /app /etc/nginx

EXPOSE 8080

USER 10001

ENTRYPOINT ["/app/entrypoint.sh"]
