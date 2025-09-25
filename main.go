package main

import (
	"gpxtoolkit/cmd"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cmd.Execute()
}
