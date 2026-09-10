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
Também verificam que a consulta por ID retorna os dados do lobby criado
e preserva o erro `ErrLobbyNotFound` quando ele não existe, permitindo
identificá-lo com `errors.Is`.
A listagem é verificada tanto sem lobbies quanto com múltiplos lobbies
criados, comparando os dados por ID sem depender da ordem de retorno.

Falhas de armazenamento são simuladas por uma implementação de teste
da interface do repository. Os testes verificam que criação, consulta
e listagem preservam o erro original e retornam dados nulos.

Os testes HTTP usam `httptest` para verificar health e criação de
lobbies, incluindo status HTTP, Content-Type, conteúdo JSON e dados
armazenados. Também verificam a rejeição de entradas inválidas com
`400`, sem armazenamento, e falhas do repository com `500`, sem
expor detalhes internos na resposta.
A consulta por ID é testada com `ServeMux`, cobrindo sucesso, IDs
inválidos, lobby inexistente e falha de armazenamento, com verificação
dos dados retornados e das mensagens de erro.
A listagem HTTP é testada vazia, com múltiplos lobbies e com falha de
armazenamento, verificando o array JSON, os dados retornados e a
resposta `500` com mensagem pública.

Os endpoints foram verificados manualmente com `curl`.

## PostgreSQL local

Requer Docker e Docker Compose.

Para iniciar o banco:

```bash
docker compose up -d --wait
```

Para consultar seu estado:

```bash
docker compose ps
```

O banco está disponível em `localhost:5432`, com banco e usuário
`lobby` e senha de desenvolvimento `lobby_dev`.

Para parar e remover o container, preservando os dados no volume:

```bash
docker compose down
```

A API ainda utiliza o repository em memória. A integração com
PostgreSQL será adicionada na próxima etapa.

### Criar a tabela

Com o banco iniciado, aplique a primeira migração na raiz do projeto:

```bash
docker compose exec -T postgres psql -U lobby -d lobby \
  -v ON_ERROR_STOP=1 --single-transaction \
  < migrations/001_create_lobbies.up.sql
```

Execute uma vez por banco novo. As migrações ainda são aplicadas
manualmente, sem controle automático de versões.

Para conferir a estrutura:

```bash
docker compose exec postgres psql -U lobby -d lobby -c "\d lobbies"
```

## Próximos passos

- Ampliar os testes do repository, do service e dos handlers.
- Adicionar persistência com PostgreSQL.
