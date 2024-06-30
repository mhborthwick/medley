FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/main.go ./cmd/main.go
RUN CGO_ENABLED=0 go build -o /bin/authserver ./cmd/main.go

FROM alpine:latest

WORKDIR /home/authserver
RUN addgroup -S spotify && \
  adduser -S authserver -G spotify

USER authserver
COPY --from=build /bin/authserver .local/bin/authserver
ENTRYPOINT [".local/bin/authserver"]
