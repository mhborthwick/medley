set dotenv-path := 'authserver/.env'

start_auth_server:
  @go run authserver/cmd/main.go
