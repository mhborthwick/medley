package main

import (
	"fmt"
	"os"
)

func main() {
	clientID := os.Getenv("CLIENT_ID")
	fmt.Println(clientID)
	fmt.Println("hello, world")
}
