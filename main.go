package main

import (
	"github.com/fatindeed/proxy-validator/cmd"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	cmd.Execute()
}
