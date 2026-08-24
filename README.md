# Go Redis Rate Limiter

Rate limiter HTTP em Go, com persistência Redis e separação entre regra de negócio, middleware e estratégia de armazenamento.

## Executar

```bash
docker compose up --build
```

O serviço fica disponível em `http://localhost:8080`. A rota protegida de demonstração é `GET /`; `GET /health` serve como health check.

```bash
curl http://localhost:8080/
curl -H 'API_KEY: premium-token' http://localhost:8080/
```

Para executar os testes usando apenas Docker Compose:

```bash
docker compose run --rm test
```

## Configuração

| Variável | Padrão | Descrição |
| --- | --- | --- |
| `RATE_LIMIT_IP_REQUESTS_PER_SECOND` | `10` | Limite aplicado quando não há `API_KEY`. |
| `RATE_LIMIT_TOKEN_REQUESTS_PER_SECOND` | `100` | Limite para qualquer token informado. |
| `RATE_LIMIT_TOKEN_OVERRIDES` | `premium-token:200` | Limites específicos no formato `token:limite`, separados por vírgula. |
| `RATE_LIMIT_WINDOW` | `1s` | Janela de contagem. |
| `RATE_LIMIT_BLOCK_DURATION` | `5m` | Duração do bloqueio após exceder o limite. |
| `REDIS_ADDR` | `redis:6379` | Endereço do Redis. |
| `REDIS_PASSWORD` / `REDIS_DB` | vazio / `0` | Credenciais e database Redis. |

Ao receber `API_KEY`, o limiter identifica a requisição pelo token e nunca pelo IP: portanto Token > IP. Um token com override em `RATE_LIMIT_TOKEN_OVERRIDES` substitui também o limite padrão de token.

Após a requisição que ultrapassa o limite, a chave recebe bloqueio Redis por `RATE_LIMIT_BLOCK_DURATION`. Durante esse intervalo a resposta é `429` e o corpo é exatamente:

```text
you have reached the maximum number of requests or actions allowed within a certain time frame
```

## Arquitetura

- `internal/ratelimiter/limiter.go`: regra de escolha IP/token e seus limites.
- `internal/ratelimiter/middleware.go`: adaptação HTTP.
- `internal/ratelimiter/redis_store.go`: estratégia Redis, com script Lua atômico para contagem e bloqueio.

A porta de persistência é a interface `ratelimiter.Store`. Para trocar Redis por memória, banco SQL ou outro serviço, implemente `Allow(ctx, key, limit, window, blockDuration)` e injete a nova estratégia em `ratelimiter.New`; middleware e regra de negócio não precisam mudar.

## Testes

Os testes cobrem o bloqueio persistido no Redis, resposta HTTP 429 com corpo exato e a precedência Token > IP, incluindo override por token.
