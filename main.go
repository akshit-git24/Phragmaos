package main

import (
	"fmt"
	"log"

	p "phragmaos"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	fmt.Println("Hello from Phragmaos!")
	p.Run()
}
