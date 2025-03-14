# FullCycle Desafio Técnico - Rate Limiter

## Solução:

### projeto fullcycle-desafio-tecnico-rate-limiter

* docker-compose.yml com configuração para redis e serverrl
* o projeto fez amplo uso de `log/slog` que pode ser observado no log do container serverrl

#### subprojeto rate_limiter

* `cmd/main.go` com http server (serverrl) de exemplo de como usar o middleware
  * conectar ao redis
  * configurar o DB
  * inicializar o ratelimiter com o DB
  * envolver os handlers com o middelware passando o rate limiter como parâmetro
  * iniciar o server na porta 8080
* `internal/infra/database/db_interface.go` com a interface do banco de dados
* `internal/infra/database/db_redis.go` com a implementação do banco de dados
* `internal/limiter/limiter.go` com a implementação do rate limiter
* `internal/model/rate_limiter.go` com a struct do rate limiter e seus métodos
* `internal/web/middleware/rate_limiter.go` com a implementação do middleware
* `Dockerfile` com o build do serverrl
* `.env` com configurações de ambiente
  * IP_QT_REQS_SECOND=5 - nível de acesso mais baixo (por ip) - 5 reqs por segundo
  * IP_BLOCK_DURATION=2s - tempo de bloqueio se exceder o limite é 2s
  * TOKEN_1_QT_REQS_SECOND=10 - nível de acesso mais alto (por token) - 10 reqs por segundo
  * TOKEN_1_BLOCK_DURATION=2s - tempo de bloqueio se exceder o limite é 2s
  * TOKEN_2_QT_REQS_SECOND=20 - nível de acesso mais alto (por token) - 20 reqs por segundo
  * TOKEN_2_BLOCK_DURATION=1s - tempo de bloqueio se exceder o limite é 1s
  * TOKEN_3_QT_REQS_SECOND=50 - nível de acesso mais alto (por token) - 50 reqs por segundo
  * TOKEN_3_BLOCK_DURATION=500ms - tempo de bloqueio se exceder o limite é 500ms
  * TOKEN_4_QT_REQS_SECOND=100 - nível de acesso mais alto (por token) - 100 reqs por segundo
  * TOKEN_4_BLOCK_DURATION=100ms - tempo de bloqueio se exceder o limite é 100ms
  * USE_ONLY_IP_LIMITER=false - se for true, o rate limiter utiliza apenas o IP para limitar as requisições
  * USE_ONLY_TOKEN_LIMITER=false - se for true, o rate limiter utiliza apenas o token para limitar as requisições
* `test/main_test.go` com testes automatizados para cada nível de acesso. Aguardar 10 segundos entre as execuções dos testes para dar tempo do redis limpar as requisições anteriores. Meu computador não é muito forte e não suporta altas cargas. Conseguir lidar bem com 500 requisições simultâneas.
* `api/requests.http` com exemplos de requisições para o serverrl no padrão 'REST Client'

#### subprojeto tester

* `main.go` implementação cliente que permitiu fazer testes mais rápidos ao longo do desenvolvimento

## Objetivo:

Desenvolver um rate limiter em Go que possa ser configurado para limitar o número máximo de requisições por segundo com base em um endereço IP específico ou em um token de acesso.

## Descrição:

O objetivo deste desafio é criar um rate limiter em Go que possa ser utilizado para controlar o tráfego de requisições para um serviço web. O rate limiter deve ser capaz de limitar o número de requisições com base em dois critérios:

1. Endereço IP: O rate limiter deve restringir o número de requisições recebidas de um único endereço IP dentro de um intervalo de tempo definido.
1. Token de Acesso: O rate limiter deve também poderá limitar as requisições baseadas em um token de acesso único, permitindo diferentes limites de tempo de expiração para diferentes tokens. O Token deve ser informado no header no seguinte formato:
    1. `API_KEY: <TOKEN>`
1. As configurações de limite do token de acesso devem se sobrepor as do IP. Ex: Se o limite por IP é de 10 req/s e a de um determinado token é de 100 req/s, o rate limiter deve utilizar as informações do token.

## Requisitos:

* O rate limiter deve poder trabalhar como um middleware que é injetado ao servidor web
* O rate limiter deve permitir a configuração do número máximo de requisições permitidas por segundo.
* O rate limiter deve ter ter a opção de escolher o tempo de bloqueio do IP ou do Token caso a quantidade de requisições tenha sido excedida.
* As configurações de limite devem ser realizadas via variáveis de ambiente ou em um arquivo “.env” na pasta raiz.
* Deve ser possível configurar o rate limiter tanto para limitação por IP quanto por token de acesso.
* O sistema deve responder adequadamente quando o limite é excedido:
  * Código HTTP: 429
  * Mensagem: you have reached the maximum number of requests or actions allowed within a certain time frame
* Todas as informações de "limiter” devem ser armazenadas e consultadas de um banco de dados Redis. Você pode utilizar docker-compose para subir o Redis.
* Crie uma “strategy” que permita trocar facilmente o Redis por outro mecanismo de persistência.
* A lógica do limiter deve estar separada do middleware.

## Exemplos:

1. Limitação por IP: Suponha que o rate limiter esteja configurado para permitir no máximo 5 requisições por segundo por IP. Se o IP 192.168.1.1 enviar 6 requisições em um segundo, a sexta requisição deve ser bloqueada.
1. Limitação por Token: Se um token abc123 tiver um limite configurado de 10 requisições por segundo e enviar 11 requisições nesse intervalo, a décima primeira deve ser bloqueada.
1. Nos dois casos acima, as próximas requisições poderão ser realizadas somente quando o tempo total de expiração ocorrer. Ex: Se o tempo de expiração é de 5 minutos, determinado IP poderá realizar novas requisições somente após os 5 minutos.

## Dicas:

* Teste seu rate limiter sob diferentes condições de carga para garantir que ele funcione conforme esperado em situações de alto tráfego.

## Entrega:

* O código-fonte completo da implementação.
* Documentação explicando como o rate limiter funciona e como ele pode ser configurado.
* Testes automatizados demonstrando a eficácia e a robustez do rate limiter.
* Utilize docker/docker-compose para que possamos realizar os testes de sua aplicação.
* O servidor web deve responder na porta 8080.
