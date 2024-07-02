set dotenv-load

start_auth_server_docker:
  @docker compose -f authserver/compose.yml up --build -d

stop_auth_server_docker:
  @docker compose -f authserver/compose.yml stop

run_cli:
  @go run cli/cmd/main.go