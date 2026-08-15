# Rate Limiter

Prototipo de rate limiter en Go usando token bucket.

Para detalles de diseno y decisiones tecnicas, ver [DESIGN.md](./DESIGN.md).

## Ejecutar

```bash
go run ./cmd/ratelimitd
```

La API escucha en `:8080` por defecto.

## Endpoints

- `GET /healthz`: health check sin rate limiting.
- `GET /`: endpoint protegido por rate limiting.

## Configuracion

| Variable | Default | Descripcion |
| --- | ---: | --- |
| `HTTP_ADDR` | `:8080` | Direccion HTTP. |
| `LOG_LEVEL` | `info` | Nivel de logs: `debug`, `info`, `warn` o `error`. |
| `RATE_LIMIT_CAPACITY` | `10` | Capacidad maxima del bucket por cliente. |
| `RATE_LIMIT_REFILL_RATE` | `5` | Tokens recargados por segundo. |
| `RATE_LIMIT_KEY_HEADER` | vacio | Header confiable usado como key del rate limit. |

Ejemplo en PowerShell:

```powershell
$env:RATE_LIMIT_CAPACITY = "10"
$env:RATE_LIMIT_REFILL_RATE = "5"
$env:LOG_LEVEL = "info"
$env:HTTP_ADDR = ":8080"
go run ./cmd/ratelimitd
```

## Docker

```bash
docker compose up --build
```

Para levantarlo en background:

```bash
docker compose up -d --build
```

El compose publica la API en `127.0.0.1:8080`.

## Probar

```bash
curl -i http://localhost:8080/
```

Cuando el limite se alcanza, la API responde `429 Too Many Requests` con `Retry-After` y headers `RateLimit-*`.

## Tests

```bash
go test ./...
```
