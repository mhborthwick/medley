set dotenv-load

start_auth_server_docker:
  @docker compose -f authserver/compose.yml up --build -d

stop_auth_server_docker:
  @docker compose -f authserver/compose.yml stop

run_cli:
  @go run cli/cmd/main.go create cli/config/example.pkl

run_test:
  @go test github.com/mhborthwick/medley/... -cover
