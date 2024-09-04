package main

import (
	"log"
	"tt_fraudsters_suspender/cmd"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}
	cmd.Execute()
}
