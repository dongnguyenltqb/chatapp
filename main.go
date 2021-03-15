package main

import (
	"chatapp/handler"
	"chatapp/infra"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
}

func main() {
	// Setup infra
	infra.Setup()
	// Run handler
	handler.Run()
}
