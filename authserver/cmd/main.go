package main

import (
	"fmt"
	"os"

	"github.com/google/uuid"
)

func GetRandomString() string {
	return uuid.NewString()
}

func main() {
	clientID := os.Getenv("CLIENT_ID")
	randID := GetRandomString()
	fmt.Println(clientID)
	fmt.Println(randID)
}
