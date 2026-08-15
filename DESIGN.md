# Diseno del Rate Limiter

## Objetivo

Implementar un prototipo funcional de rate limiting en Go, con un nucleo simple, testeable y facil de extender.

La solucion prioriza:

- logica deterministica;
- separacion clara entre HTTP, dominio y almacenamiento;
- manejo explicito de errores;
- tests sobre el algoritmo y el comportamiento HTTP.

## Arquitectura

La estructura sigue un monolito modular:

- `cmd/ratelimitd`: entrypoint HTTP.
- `cmd/config`: carga de configuracion desde variables de entorno.
- `internal/ratelimit/entity`: tipos de dominio.
- `internal/ratelimit/service`: caso de uso y algoritmo token bucket.
- `internal/ratelimit/repository`: contrato de almacenamiento.
- `internal/ratelimit/repository/memory`: almacenamiento in-memory.
- `internal/ratelimit/handler/http`: adaptador HTTP.
- `internal/platform`: utilidades compartidas.

`main` carga configuracion, crea el logger, inicializa el modulo de rate limiting y expone `rateLimitModule.Routes()` como handler HTTP. Asi el entrypoint no conoce los detalles internos de rutas, middleware o almacenamiento.

## Responsabilidades

`entity` contiene los conceptos del dominio:

- `Bucket`
- `Decision`
- `ErrEmptyKey`

`service` contiene la logica principal y expone el contrato:

```go
type Limiter interface {
    Allow(ctx context.Context, key string) (entity.Decision, error)
}
```

`Allow` valida la clave, respeta la cancelacion del contexto, recarga tokens, consume un token cuando corresponde y devuelve una `Decision`.

`repository` encapsula el estado de los buckets. La implementacion in-memory usa un `map` protegido por mutex y expone una actualizacion atomica para que el service no manipule el mapa directamente.

`handler/http` traduce entre HTTP y el caso de uso:

- obtiene la key del request;
- llama a `Allow`;
- escribe headers `RateLimit-*`;
- responde `429 Too Many Requests` cuando el limite fue alcanzado;
- registra requests HTTP con logging middleware.

El logger se mantiene en la capa HTTP, donde existen datos utiles como metodo, path, status y duracion. El service no depende de logging.

## Configuracion

La configuracion vive en `cmd/config` y se carga desde variables de entorno:

- `HTTP_ADDR`: direccion HTTP, por defecto `:8080`.
- `LOG_LEVEL`: nivel de logs, por defecto `info`.
- `RATE_LIMIT_CAPACITY`: capacidad maxima del bucket, por defecto `10`.
- `RATE_LIMIT_REFILL_RATE`: tokens recargados por segundo, por defecto `5`.
- `RATE_LIMIT_KEY_HEADER`: header confiable usado como key, por defecto vacio.

Separar la configuracion del dominio permite testear defaults, overrides y valores invalidos sin levantar el servidor.

## Algoritmo

La implementacion usa token bucket.

Cada key tiene un bucket con tokens disponibles y la fecha del ultimo refill. En cada request:

1. Se busca o crea el bucket.
2. Se recargan tokens segun el tiempo transcurrido.
3. Si hay al menos un token, se consume y el request se permite.
4. Si no hay tokens, se rechaza y se calcula `RetryAfter`.

Token bucket permite rafagas acotadas sin dejar de respetar una tasa promedio. Por eso encaja mejor que una ventana fija, que puede permitir rafagas en los bordes de ventana, y es mas simple que un sliding log.

## Concurrencia

El repository in-memory protege el mapa con un mutex. Esto mantiene correcta la actualizacion de buckets ante requests concurrentes.

Una version de mayor throughput podria usar shards o locks por bucket, pero no se agrega esa complejidad sin evidencia de contencion.

## Trade-offs

- **In-memory vs Redis**: se usa memoria local por simplicidad y baja latencia. La contra es que el limite no se comparte entre multiples instancias. Para un despliegue distribuido, Redis con operaciones atomicas seria la evolucion natural.
- **Token bucket vs ventanas**: token bucket permite rafagas controladas y mantiene una tasa promedio. Una ventana fija es mas simple, pero puede permitir picos en los bordes; sliding log es mas preciso, pero consume mas memoria.
- **Mutex unico vs locks por bucket**: el mutex unico simplifica la correccion concurrente. Si la contencion creciera, se podria evolucionar a sharding o locks por key.
- **Logger en HTTP vs service**: el logging vive en la capa HTTP porque ahi estan metodo, path, status y duracion. El service queda enfocado en la decision de rate limiting.
- **Middleware vs API Gateway completo**: el prototipo implementa el rate limiting como middleware reutilizable. Ruteo, reverse proxy y service discovery quedan fuera de alcance.

## HTTP

El rate limiter se implementa como middleware para envolver cualquier `http.Handler`.

La key por defecto se obtiene desde `RemoteAddr`. Si la aplicacion corre detras de infraestructura confiable, `RATE_LIMIT_KEY_HEADER` permite usar un header como identidad del rate limit.

Cuando el request es aceptado, se agregan:

- `RateLimit-Limit`
- `RateLimit-Remaining`
- `RateLimit-Reset`

Cuando es rechazado, se devuelve `429` y se agrega `Retry-After`.

## Limitaciones

Esta version es single-node e in-memory. Si corren multiples instancias, cada una mantiene sus propios buckets.

Para rate limiting distribuido, el estado podria moverse a Redis con operaciones atomicas o scripts Lua. Eso queda fuera del alcance del prototipo para mantener el foco en la correccion del algoritmo local.

El prototipo tampoco implementa funciones completas de API Gateway, como reverse proxy o ruteo a servicios upstream. El rate limiting queda encapsulado como middleware HTTP para que pueda integrarse en un gateway mas adelante.

## Uso de IA

Se uso IA como apoyo para estructurar la implementacion, los tests y este documento. Las decisiones finales del diseno quedan reflejadas en el codigo y en las pruebas.
