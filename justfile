set dotenv-load

start_auth_server_docker:
  @docker compose -f authserver/compose.yml up --build -d

stop_auth_server_docker:
  @docker compose -f authserver/compose.yml stop

run_cli_create:
  @go run cli/cmd/main.go create cli/config/example.pkl

run_cli_sync:
  @go run cli/cmd/main.go sync cli/config/example.pkl

run_test:
  @go test github.com/mhborthwick/medley/... -cover
