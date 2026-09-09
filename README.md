# Go Game Lobby Service

API REST para gerenciamento de lobbies de jogos, desenvolvida como
projeto de aprendizado de Go.

O projeto evolui incrementalmente, com separação entre transporte HTTP,
regras de negócio e persistência.

## Estado atual

- Criação de lobbies com validação de nome e número de jogadores.
- Listagem de lobbies.
- Consulta de lobby por ID.
- Endpoint de health.
- Armazenamento em memória protegido por `sync.RWMutex`.

Os dados são perdidos quando o servidor é encerrado.
A listagem não possui ordenação garantida.

## Requisitos

- Go 1.25.1 ou superior.

## Executar

Na raiz do projeto:

```bash
go run ./cmd/api
```

O servidor escuta na porta `8080`.

## Endpoints

| Método | Caminho | Descrição |
|---|---|---|
| POST | `/lobbies` | Cria um lobby |
| GET | `/lobbies` | Lista os lobbies |
| GET | `/lobbies/{id}` | Consulta um lobby |
| GET | `/health` | Confirma que o servidor está respondendo |

### Criar um lobby

```bash
curl -i -X POST http://localhost:8080/lobbies \
  -H 'Content-Type: application/json' \
  -d '{"name":"Sala de estudo","max_players":4}'
```

Retorna `201 Created` com o lobby criado.

O nome não pode ficar vazio após remover espaços das extremidades,
e `max_players` deve ser pelo menos `2`.
Novos lobbies recebem o status `waiting`.

### Listar lobbies

```bash
curl -i http://localhost:8080/lobbies
```

Retorna `200 OK` com um array JSON, ou `[]` quando não existem lobbies.

### Consultar por ID

Use o ID retornado na criação:

```bash
curl -i http://localhost:8080/lobbies/1
```

Retorna `200 OK` quando encontrado, `404 Not Found` quando não existe
e `400 Bad Request` quando o ID é inválido.

### Health

```bash
curl -i http://localhost:8080/health
```

Retorna `200 OK` com:

```json
{"status":"ok"}
```

Esse endpoint confirma apenas que o servidor HTTP está respondendo.

## Organização do código

- `cmd/api`: inicialização, conexão das dependências e registro das rotas.
- `internal/domain`: estrutura e estados de um lobby.
- `internal/service`: regras de negócio e coordenação das operações.
- `internal/repository`: contrato de persistência e erro compartilhado.
- `internal/repository/memory`: implementação em memória.
- `internal/http`: handlers, leitura das requisições e respostas HTTP.

## Verificações

```bash
go test ./...
go vet ./...
go test -race ./...
```

Os testes automatizados atuais verificam o isolamento por cópias na
criação, consulta por ID e listagem, além da busca por ID inexistente
e da listagem vazia ou com múltiplos lobbies no repository em memória.
Também verificam que contextos previamente cancelados impedem a
criação, a consulta e a listagem, retornando `context.Canceled`.

Há também um teste de 100 criações concorrentes, verificando a
quantidade armazenada e IDs positivos e únicos. A suíte foi executada
com o detector de corridas de dados (`-race`).

Os testes do service verificam a rejeição de nomes vazios e quantidades
de jogadores abaixo de 2, sem armazenar lobbies inválidos. Também
verificam a criação válida com 2 jogadores, a remoção de espaços nas
extremidades do nome, o status inicial, o preenchimento da data de
criação e o armazenamento do lobby.
Também verificam que a consulta por ID preserva o erro
`ErrLobbyNotFound`, permitindo identificá-lo com `errors.Is`.

Os endpoints foram verificados manualmente com `curl`.

## Próximos passos

- Ampliar os testes do repository e do service e adicionar testes dos handlers.
- Adicionar persistência com PostgreSQL.
