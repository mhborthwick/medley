set dotenv-path := 'authserver/.env'

start_auth_server:
  @go run authserver/cmd/main.go

start_auth_server_docker:
  @docker compose -f authserver/compose.yml up -d

stop_auth_server_docker:
  @docker compose -f authserver/compose.yml stop