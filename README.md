# canis-lupus-arctos

A small CR~~UD~~ REST API.

## Requirements

- Go: [download](https://go.dev/dl/)
- Air (optional live reload): [installation](https://github.com/air-verse/air/tree/master?tab=readme-ov-file#via-go-install-recommended)
- Templ (board web view): [installation](https://templ.guide/quick-start/installation#go-install)

## Using the Project

### Setup environment

#### Download modules

```sh
go mod download
```

#### Run the server

The entrypoint for the program is in [cmd/cmd.go](/cmd/cmd.go)

We can start the server with live reload for local development

```sh
air
```

or, run the command directly

```sh
go run cmd/cmd.go
```

or, compile the binary and run it

```sh
go build -o risk-server ./cmd
./risk-server
```

### Consume the API

The [OpenAPI spec file](/openapi.yml) details the API endpoints and their expected payloads and outputs.

To create a new risk:

```sh
curl --request POST \
  --url http://localhost:8080/v1/risk
```

response:

```json
{
  "id": "501bb792-787e-4318-a4af-923860b3355c",
  "state": "closed",
  "title": "A thing that happened",
  "description": "it was pretty bad"
}
```

fetch a risk by id:

```sh
curl --request GET \
  --url http://localhost:8080/v1/risk/501bb792-787e-4318-a4af-923860b3355c
```

response:

```json
{
  "id": "501bb792-787e-4318-a4af-923860b3355c",
  "state": "closed",
  "title": "A thing that happened",
  "description": "it was pretty bad"
}
```

fetch all risks:

```sh
curl --request GET \
  --url http://localhost:8080/v1/risk
```

response:

```json
[
  {
    "id": "501bb792-787e-4318-a4af-923860b3355c",
    "state": "closed",
    "title": "A thing that happened",
    "description": "it was pretty bad"
  },
  {
    "id": "41546f04-7628-40e3-b116-6f1203b68a48",
    "state": "open"
  },
  {
    "id": "da077100-a3c1-4904-b3a5-603a52534411",
    "state": "closed",
    "title": "Another thing that happened",
    "description": "it was pretty good"
  }
]
```

### Testing

Run at project root:

```sh
go test ./...
```
