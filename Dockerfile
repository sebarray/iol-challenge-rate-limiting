FROM golang:1.24-alpine AS build

WORKDIR /src

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/ratelimitd ./cmd/ratelimitd

FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app

WORKDIR /app

COPY --from=build /out/ratelimitd /app/ratelimitd

USER app

EXPOSE 8085

ENV HTTP_ADDR=:8085

ENTRYPOINT ["/app/ratelimitd"]
