package main

import (
	"github.com/outdoorsafetylab/gpxtoolkit/cmd"

	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cmd.Execute()
}
